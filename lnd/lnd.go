package lnd

import (
	"context"

	"sync"

	"time"

	"sync/atomic"

	"net"
	"strconv"

	"encoding/hex"

	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/db"
	"github.com/bitlum/btcd/btcec"
	"github.com/bitlum/btcutil"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/bitlum/connector/metrics"
)

const (
	MethodCreateInvoice = "CreateInvoice"
	MethodSendTo        = "SendTo"
	MethodInfo          = "Info"
	MethodQueryRoutes   = "QueryRoutes"
	MethodStart         = "Start"
	MethodHandleInvoice = "HandlePayments"
)

// Config is a connector config.
type Config struct {
	// Net blockchain network this connector should operate with.
	Net string

	// Name of the daemon client.
	Name string

	// Port is gRPC port of lnd daemon.
	Port int

	// Host is gRPC host of lnd daemon.
	Host string

	// TlsCertPath is a path to certificate, which is needed to have a secure
	// gRPC connection with lnd daemon.
	TlsCertPath string

	// Metrics is a metric backend which is used to collect metrics from
	// connector. In case of prometheus client they stored locally till
	// they will be collected by prometheus server.
	Metrics crypto.MetricsBackend
}

func (c *Config) validate() error {
	if c.Net == "" {
		return errors.Errorf("net should be specified")
	}

	if c.Port == 0 {
		return errors.Errorf("port should be specified")
	}

	if c.Host == "" {
		return errors.Errorf("host should be specified")
	}

	if c.TlsCertPath == "" {
		return errors.Errorf("tlc cert path should be specified")
	}

	if c.Metrics == nil {
		return errors.Errorf("metricsBackend should be specified")
	}

	return nil
}

type Connector struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg    *Config
	client lnrpc.LightningClient
	db     *db.DB

	notifications chan *common.Payment

	conn     *grpc.ClientConn
	nodeAddr string
}

// Runtime check to ensure that Connector implements common.LightningConnector
// interface.
var _ common.LightningConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("config is invalid: %v", err)
	}

	return &Connector{
		cfg:           cfg,
		notifications: make(chan *common.Payment),
		quit:          make(chan struct{}),
	}, nil
}

// Start...
func (c *Connector) Start() error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		log.Warn("lightning client already started")
		return nil
	}

	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodStart, c.cfg.Metrics)
	defer m.Finish()

	creds, err := credentials.NewClientTLSFromFile(c.cfg.TlsCertPath, "")
	if err != nil {
		m.AddError(errToSeverity(ErrTLSRead))
		return errors.Errorf("unable to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	target := net.JoinHostPort(c.cfg.Host, strconv.Itoa(c.cfg.Port))
	log.Infof("lightning client connection to lnd: %v", target)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		m.AddError(errToSeverity(ErrGRPCConnect))
		return errors.Errorf("unable to to dial grpc: %v", err)
	}
	c.conn = conn
	c.client = lnrpc.NewLightningClient(c.conn)

	reqInfo := &lnrpc.GetInfoRequest{}
	respInfo, err := c.client.GetInfo(context.Background(), reqInfo)
	if err != nil {
		m.AddError(errToSeverity(ErrGetInfo))
		return errors.Errorf("unable get lnd node info: %v", err)
	}

	// TODO(andrew.shvv) not working for mainnet, as far response don't have
	// a mainnet param.
	lndNet := "simnet"
	if respInfo.Testnet {
		lndNet = "testnet"
	}

	if lndNet != c.cfg.Net {
		return errors.Errorf("hub net is '%v', but config net is '%v'",
			c.cfg.Net, lndNet)
	}

	log.Infof("Init connector working with '%v' net", lndNet)

	c.nodeAddr = respInfo.IdentityPubkey
	var invoiceSubscription lnrpc.Lightning_SubscribeInvoicesClient

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			balance, err := c.FundsAvailable()
			if err != nil {
				m.AddError(string(metrics.MiddleSeverity))
				log.Errorf("unable to get available funds: %v", err)
			}

			log.Infof("Asset(BTC), media(lightning), available funds(%v)",
				balance.Round(8).String())

			f, _ := balance.Float64()
			m.CurrentFunds(f)

			select {
			case <-time.After(time.Second * 10):
			case <-c.quit:
				return
			}
		}
	}()

	c.wg.Add(1)
	go func() {
		m := crypto.NewMetric(c.cfg.Name, "BTC", MethodHandleInvoice, c.cfg.Metrics)
		defer m.Finish()

		defer c.wg.Done()
		for {
			if invoiceSubscription == nil {
				log.Info("Subscribe on invoice updates...")

				// Trying to reconnect after receiving transport closing
				// error.
				reqSubsc := &lnrpc.InvoiceSubscription{}
				invoiceSubscription, err = c.client.SubscribeInvoices(context.Background(), reqSubsc)
				if err != nil {
					m.AddError(errToSeverity(ErrResubscribeInvoiceStream))
					log.Errorf("unable to re-subscribe on invoice"+
						" updates: %v", err)

					select {
					case <-c.quit:
						log.Info("Invoice receiver goroutine shutdown")
						return
					case <-time.After(time.Second * 5):
						continue
					}
				}
			}

			invoiceUpdate, err := invoiceSubscription.Recv()
			if err != nil {
				m.AddError(errToSeverity(ErrReadInvoiceStream))
				log.Errorf("unable to read from invoice stream: %v", err)
				invoiceSubscription = nil
				continue
			}

			if !invoiceUpdate.Settled {
				log.Info("Received non-settled invoice update, "+
					"invoice(%v)", invoiceUpdate.PaymentRequest)
				continue
			}

			amount := btcutil.Amount(invoiceUpdate.Value)
			payment := &common.Payment{
				ID:      invoiceUpdate.PaymentRequest,
				Amount:  decimal.NewFromFloat(amount.ToBTC()),
				Account: invoiceUpdate.Memo,
				Address: c.nodeAddr,
				Type:    common.Lightning,
			}

		repeat:
			for {
				select {
				case <-c.quit:
					log.Info("Invoice receiver goroutine shutdown")
					return
				case c.notifications <- payment:
					break repeat
				case <-time.After(time.Second):
					// TODO(andrew.shvv) add pending queue
					m.AddError(errToSeverity(ErrSendPaymentNotification))
					log.Errorf("unable to send notification for payment"+
						"(%v)", payment.Address)
				}
			}
		}
	}()

	log.Info("lightning client started")
	return nil
}

// Stop gracefully stops the connection with lnd daemon.
func (c *Connector) Stop(reason string) error {
	if !atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		log.Warn("lightning client already shutdown")
		return nil
	}

	close(c.quit)
	if err := c.conn.Close(); err != nil {
		return errors.Errorf("unable to close connection to lnd: %v", err)
	}

	c.wg.Wait()

	log.Infof("lightning client shutdown, reason(%v)", reason)
	return nil
}

// CreateInvoice is used to create lightning network invoice.
//
// NOTE: Part of the common.LightningConnector interface.
func (c *Connector) CreateInvoice(account string, amount string) (string, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodCreateInvoice, c.cfg.Metrics)
	defer m.Finish()

	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		m.AddError(errToSeverity(ErrConvertAmount))
		return "", err
	}

	invoice := &lnrpc.Invoice{
		Memo:  account,
		Value: satoshis,
	}

	invoiceResp, err := c.client.AddInvoice(context.Background(), invoice)
	if err != nil {
		m.AddError(errToSeverity(ErrAddInvoice))
		return "", err
	}

	return invoiceResp.PaymentRequest, nil
}

// SendTo is used to send specific amount of money to address within this
// payment system.
//
// NOTE: Part of the common.LightningConnector interface.
func (c *Connector) SendTo(invoice string) error {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodSendTo, c.cfg.Metrics)
	defer m.Finish()

	req := &lnrpc.SendRequest{
		PaymentRequest: invoice,
	}

	resp, err := c.client.SendPaymentSync(context.Background(), req)
	if err != nil {
		m.AddError(errToSeverity(ErrSendPayment))
		return errors.Errorf("unable to send payment: %v", err)
	}

	if resp.PaymentError != "" {
		m.AddError(errToSeverity(ErrSendPayment))
		return errors.Errorf("unable to send payment: %v", resp.PaymentError)
	}

	return nil
}

// ReceivedPayments returns channel with transactions which are passed
// the minimum threshold required by the client to treat as confirmed.
//
// NOTE: Part of the common.LightningConnector interface.
func (c *Connector) ReceivedPayments() <-chan *common.Payment {
	return c.notifications
}

// Info returns the information about our lnd node.
//
// NOTE: Part of the common.LightningConnector interface.
func (c *Connector) Info() (*common.LightningInfo, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodInfo, c.cfg.Metrics)
	defer m.Finish()

	req := &lnrpc.GetInfoRequest{}
	info, err := c.client.GetInfo(context.Background(), req)
	if err != nil {
		m.AddError(errToSeverity(ErrGetInfo))
		return nil, err
	}

	return &common.LightningInfo{
		Host:            c.cfg.Host,
		Port:            strconv.Itoa(c.cfg.Port),
		MinAmount:       "0.00000001",
		MaxAmount:       "0.042",
		GetInfoResponse: info,
	}, nil
}

// QueryRoutes returns list of routes from to the given lnd node,
// and insures the the capacity of the channels is sufficient.
//
// NOTE: Part of the common.LightningConnector interface.
func (c *Connector) QueryRoutes(pubKey, amount string) ([]*lnrpc.Route, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodQueryRoutes, c.cfg.Metrics)
	defer m.Finish()

	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		m.AddError(errToSeverity(ErrConvertAmount))
		return nil, errors.Errorf("unable to convert amount: %v", err)
	}

	// First parse the hex-encdoed public key into a full public key object
	// to check that it is valid.
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		m.AddError(errToSeverity(ErrPubkey))
		return nil, errors.Errorf(
			"unable decode identity key from string: %v", err)
	}

	if _, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256()); err != nil {
		m.AddError(errToSeverity(ErrPubkey))
		return nil, errors.Errorf("unable decode identity key: %v", err)
	}

	req := &lnrpc.QueryRoutesRequest{
		PubKey: pubKey,
		Amt:    satoshis,
	}

	info, err := c.client.QueryRoutes(context.Background(), req)
	if err != nil {
		m.AddError(errToSeverity(ErrUnableQueryRoutes))
		return nil, err
	}

	return info.Routes, nil
}

// FundsAvailable returns number of funds available under control of
// connector.
//
// NOTE: Part of the common.Connector interface.
func (c *Connector) FundsAvailable() (decimal.Decimal, error) {
	req := &lnrpc.WalletBalanceRequest{}
	resp, err := c.client.WalletBalance(context.Background(), req)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.New(resp.ConfirmedBalance, 0), nil
}
