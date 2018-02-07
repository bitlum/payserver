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
	"github.com/bitlum/connector/common/db"
	"github.com/bitlum/btcd/btcec"
	"github.com/bitlum/btcutil"
	"github.com/btcsuite/btclog"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config...
type Config struct {
	// Port...
	Port int

	// Host...
	Host string

	// TlsCertPath...
	TlsCertPath string

	// Logger...
	Logger btclog.Logger
}

func (c *Config) validate() error {
	// Port...
	if c.Port == 0 {
		return errors.Errorf("port should be specified")
	}

	// Host...
	if c.Host == "" {
		return errors.Errorf("host should be specified")
	}

	// TlsCertPath...
	if c.TlsCertPath == "" {
		return errors.Errorf("tlc cert path should be specified")
	}

	if c.Logger == nil {
		return errors.Errorf("logger should be specified")
	}

	return nil
}

// Connector...
type Connector struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg    *Config
	client lnrpc.LightningClient
	db     *db.DB

	notifications chan *common.Payment
	log           *common.NamedLogger
	conn          *grpc.ClientConn
	nodeAddr      string
}

//...
var _ common.LightningConnector = (*Connector)(nil)

// NewConnector...
func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("config is invalid: %v", err)
	}

	return &Connector{
		cfg:           cfg,
		notifications: make(chan *common.Payment),
		quit:          make(chan struct{}),
		log: &common.NamedLogger{
			Logger: cfg.Logger,
			Name:   "LIGHTNING",
		},
	}, nil
}

// Start...
func (c *Connector) Start() error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		c.log.Warn("lightning client already started")
		return nil
	}

	creds, err := credentials.NewClientTLSFromFile(c.cfg.TlsCertPath, "")
	if err != nil {
		return errors.Errorf("unable to load credentials: %v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	target := net.JoinHostPort(c.cfg.Host, strconv.Itoa(c.cfg.Port))
	c.log.Infof("lightning client connection to lnd: %v", target)

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return errors.Errorf("unable to to dial grpc: %v", err)
	}
	c.client = lnrpc.NewLightningClient(conn)
	c.conn = conn

	reqInfo := &lnrpc.GetInfoRequest{}
	respInfo, err := c.client.GetInfo(context.Background(), reqInfo)
	if err != nil {
		return errors.Errorf("unable get lnd node info: %v", err)
	}

	c.nodeAddr = respInfo.IdentityPubkey

	c.log.Info("Subscribe on invoice updates")
	reqSubsc := &lnrpc.InvoiceSubscription{}
	invoiceSubscription, err := c.client.SubscribeInvoices(context.Background(), reqSubsc)
	if err != nil {
		return errors.Errorf("unable to subscribe on invoice updates: %v", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			invoiceUpdate, err := invoiceSubscription.Recv()
			if err != nil {
				c.log.Errorf("unable to read from invoice stream: %v", err)

				select {
				case <-c.quit:
					c.log.Info("Invoice receiver goroutine shutdown")
					return
				case <-time.After(time.Second * 5):
					// Trying to reconnect after receiving transport closing
					// error.
					invoiceSubscription, err = c.client.SubscribeInvoices(context.Background(), reqSubsc)
					if err != nil {
						c.log.Errorf("unable to re-subscribe on invoice"+
							" updates: %v", err)
						continue
					}

					c.log.Info("Re-subscribe on invoice updates")
					continue
				}
			}

			if !invoiceUpdate.Settled {
				c.log.Info("Received non-settled invoice update, " +
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
				case c.notifications <- payment:
					break repeat
				case <-time.After(time.Second):
					// TODO(andrew.shvv) add pending queue
					c.log.Errorf("unable to send notification for payment"+
						"(%v)", payment.Address)
				}
			}
		}
	}()

	c.log.Info("lightning client started")
	return nil
}

// Stop...
func (c *Connector) Stop(reason string) error {
	if !atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		c.log.Warn("lightning client already shutdown")
		return nil
	}

	close(c.quit)
	if err := c.conn.Close(); err != nil {
		return errors.Errorf("unable to close connection to lnd: %v", err)
	}

	c.wg.Wait()

	c.log.Infof("lightning client shutdown, reason(%v)", reason)
	return nil
}

// CreateInvoice...
func (c *Connector) CreateInvoice(account string, amount string) (string,
	error) {

	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		return "", err
	}

	invoice := &lnrpc.Invoice{
		Memo:  account,
		Value: satoshis,
	}

	invoiceResp, err := c.client.AddInvoice(context.Background(), invoice)
	if err != nil {
		return "", err
	}

	return invoiceResp.PaymentRequest, nil
}

// SendTo...
func (c *Connector) SendTo(invoice string) error {
	req := &lnrpc.SendRequest{
		PaymentRequest: invoice,
	}

	resp, err := c.client.SendPaymentSync(context.Background(), req)
	if err != nil {
		return errors.Errorf("unable to send payment: %v", err)
	}

	if resp.PaymentError != "" {
		return errors.Errorf("unable to send payment: %v", resp.PaymentError)
	}

	return nil
}

// ReceivedPayments returns channel with transactions which are passed
// the minimum threshold required by the client to treat the
// transactions as confirmed.
func (c *Connector) ReceivedPayments() <-chan *common.Payment {
	return c.notifications
}

func (c *Connector) Info() (*common.LightningInfo, error) {
	req := &lnrpc.GetInfoRequest{}
	info, err := c.client.GetInfo(context.Background(), req)
	return &common.LightningInfo{
		Host:            c.cfg.Host,
		Port:            strconv.Itoa(c.cfg.Port),
		MinAmount:       "0.00000001",
		MaxAmount:       "0.042",
		GetInfoResponse: info,
	}, err
}

func (c *Connector) QueryRoutes(pubKey, amount string) ([]*lnrpc.Route, error) {
	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		return nil, errors.Errorf("unable to convert amount: %v", err)
	}

	// First parse the hex-encdoed public key into a full public key object
	// to check that it is valid.
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, errors.Errorf(
			"unable decode identity key from string: %v", err)
	}

	if _, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256()); err != nil {
		return nil, errors.Errorf("unable decode identity key: %v", err)
	}

	req := &lnrpc.QueryRoutesRequest{
		PubKey: pubKey,
		Amt:    satoshis,
	}

	info, err := c.client.QueryRoutes(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return info.Routes, nil
}
