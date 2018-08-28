package lnd

import (
	"context"
	"sync"

	"time"

	"sync/atomic"

	"encoding/hex"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/bitlum/connector/metrics"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/assets/bitcoin"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/lightningnetwork/lnd/lnwire"
)

const (
	MethodCreateInvoice    = "CreateInvoice"
	MethodSendTo           = "SendTo"
	MethodInfo             = "Info"
	MethodQueryRoutes      = "QueryRoutes"
	MethodStart            = "Start"
	MethodHandleInvoice    = "HandlePayments"
	MethodValidateInvoice  = "ValidateInvoice"
	MethodConfirmedBalance = "ConfirmedBalance"
	MethodPendingBalance   = "PendingBalance"
)

// Config is a connector config.
type Config struct {
	// PeerPort public port of the lnd via which other lightning network nodes
	// could connect.
	// TODO(andrew.shvv) Remove when lnd would return this info
	PeerPort string

	// PeerHost public host of the lnd via which other lightning network nodes
	// could connect.
	// TODO(andrew.shvv) Remove when lnd would return this info
	PeerHost string

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

	// MacaroonPath is path to macaroon which will be used to make authorizaed
	// RPC requests. Should be empty if lnd run with --no-macaroon option.
	MacaroonPath string

	// Metrics is a metric backend which is used to collect metrics from
	// connector. In case of prometheus client they stored locally till
	// they will be collected by prometheus server.
	Metrics crypto.MetricsBackend

	// PaymentStorage is an external storage for payments, it is used by
	// connector to save payment as well as update its state.
	PaymentStore connectors.PaymentsStore
}

func (c *Config) validate() error {
	if c.PeerHost == "" {
		return errors.Errorf("peer host should be specified")
	}

	if c.PeerHost == "" {
		return errors.Errorf("peer port should be specified")
	}

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

	if c.PaymentStore == nil {
		return errors.New("payment store should be specified")
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

	notifications chan *connectors.Payment

	conn     *grpc.ClientConn
	nodeAddr string
}

// Runtime check to ensure that Connector implements connectors.LightningConnector
// interface.
var _ connectors.LightningConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("config is invalid: %v", err)
	}

	return &Connector{
		cfg:           cfg,
		notifications: make(chan *connectors.Payment),
		quit:          make(chan struct{}),
	}, nil
}

func (c *Connector) Start() (err error) {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		log.Warn("lightning client already started")
		return nil
	}

	defer func() {
		// If start has failed than, we should oll back mark that
		// service has started.
		if err != nil {
			atomic.SwapInt32(&c.started, 0)
		}
	}()

	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodStart, c.cfg.Metrics)
	defer m.Finish()

	c.client, c.conn, err = c.getClient(c.cfg.MacaroonPath)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable get grpc client: %v", err)
	}

	reqInfo := &lnrpc.GetInfoRequest{}
	respInfo, err := c.client.GetInfo(context.Background(), reqInfo)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable get lnd node info: %v", err)
	}

	lndNet := "simnet"
	if respInfo.Testnet {
		lndNet = "testnet"
	}

	// TODO(andrew.shvv) not working for mainnet, as far response don't have
	// a mainnet param.
	if c.cfg.Net != "mainnet" {
		if lndNet != c.cfg.Net {
			return errors.Errorf("hub net is '%v', but config net is '%v'",
				c.cfg.Net, lndNet)
		}

		log.Infof("Init connector working with '%v' net", lndNet)
	} else {
		log.Info("Init connector working with 'mainnet' net")
	}

	c.nodeAddr = respInfo.IdentityPubkey
	var invoiceSubscription lnrpc.Lightning_SubscribeInvoicesClient

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			balance, err := c.ConfirmedBalance("")
			if err != nil {
				m.AddError(metrics.MiddleSeverity)
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
		defer c.wg.Done()

		for {
			if err := c.reportMetrics(); err != nil {
				log.Errorf("unable report metrics: %v", err)
				continue
			}

			select {
			case <-time.After(time.Second * 30):
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

		var err error

		for {
			if invoiceSubscription == nil {
				log.Info("Subscribe on invoice updates...")

				// Trying to reconnect after receiving transport closing
				// error.
				reqSubsc := &lnrpc.InvoiceSubscription{}
				invoiceSubscription, err = c.client.SubscribeInvoices(context.Background(), reqSubsc)
				if err != nil {
					m.AddError(metrics.MiddleSeverity)
					log.Errorf("unable to subscribe on invoice"+
						" updates: %v", err)

					select {
					case <-c.quit:
						log.Info("Invoice receiver goroutine shutdown")
						return
					case <-time.After(time.Second * 5):
						// Subscribe error usually happens because of the
						// dial connection being closed.
						client, conn, err := c.getClient(c.cfg.MacaroonPath)
						if err != nil {
							m.AddError(metrics.HighSeverity)
							log.Errorf("unable create gRPC client: %v", err)
							continue
						}

						c.client = client
						c.conn = conn
						continue
					}
				}
			}

			invoiceUpdate, err := invoiceSubscription.Recv()
			if err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("unable to read from invoice stream: %v", err)
				invoiceSubscription = nil
				continue
			}

			if !invoiceUpdate.Settled {
				log.Infof("Received invoice creation notification, "+
					"invoice(%v), amount(%v), receipt(%v), memo(%v)",
					invoiceUpdate.PaymentRequest,
					invoiceUpdate.Value, string(invoiceUpdate.Receipt),
					invoiceUpdate.Memo)
				continue
			}

			paymentHash := hex.EncodeToString(invoiceUpdate.RHash)
			invoice := invoiceUpdate.PaymentRequest
			amount := lnwire.MilliSatoshi(invoiceUpdate.AmtPaid).ToBTC()

			payment := &connectors.Payment{
				PaymentID: generatePaymentID(invoice, paymentHash),
				UpdatedAt: connectors.NowInMilliSeconds(),
				Status:    connectors.Completed,
				Direction: connectors.Incoming,
				Account:   string(invoiceUpdate.Receipt),
				Receipt:   invoice,
				Asset:     connectors.BTC,
				Media:     connectors.Lightning,
				MediaID:   paymentHash,
				Amount:    decimal.NewFromFloat(amount),
				MediaFee:  decimal.Zero,
			}

			if err := c.cfg.PaymentStore.SavePayment(payment); err != nil {
				log.Errorf("unable to add payment to storage: %v",
					payment.PaymentID)
			}
		}
	}()

	log.Info("lightning client started")
	return err
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
// NOTE: Part of the connectors.LightningConnector interface.
func (c *Connector) CreateInvoice(account, amount, description string) (string,
	error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodCreateInvoice, c.cfg.Metrics)
	defer m.Finish()

	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return "", err
	}

	invoice := &lnrpc.Invoice{
		Receipt: []byte(account),
		Value:   satoshis,
		Memo:    description,
	}

	invoiceResp, err := c.client.AddInvoice(context.Background(), invoice)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", err
	}

	return invoiceResp.PaymentRequest, nil
}

// SendTo is used to send specific amount of money to address within this
// payment system.
//
// NOTE: Part of the connectors.LightningConnector interface.
func (c *Connector) SendTo(invoiceStr, amountStr string) (*connectors.Payment,
	error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodSendTo, c.cfg.Metrics)
	defer m.Finish()

	// Check that invoice is valid, and that amount which we are sending is
	// corresponding to what we expect.
	netParams, err := bitcoin.GetParams(c.cfg.Net)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	amount, err := btcToSatoshi(amountStr)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	invoice, err := zpay32.Decode(invoiceStr, netParams)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	if invoice.MilliSat.ToSatoshis() != btcutil.Amount(amount) {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("wrong amount")
	}

	// Send payment to the recipient and wait for it to be received.
	req := &lnrpc.SendRequest{
		PaymentRequest: invoiceStr,
	}

	// TODO(andrew.shvv) Use async version and return waiting payment after
	// 3-5 seconds.
	resp, err := c.client.SendPaymentSync(context.Background(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to send payment: %v", err)
	}

	if resp.PaymentError != "" {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to send payment: %v", resp.PaymentError)
	}

	paymentAmt, err := decimal.NewFromString(amountStr)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to parse amount(%v): %v",
			amount, err)
	}

	paymentHash := hex.EncodeToString(invoice.PaymentHash[:])
	payment := &connectors.Payment{
		PaymentID: generatePaymentID(invoiceStr, paymentHash),
		UpdatedAt: connectors.NowInMilliSeconds(),
		Status:    connectors.Completed,
		Direction: connectors.Outgoing,
		Receipt:   invoiceStr,
		Asset:     connectors.BTC,
		Media:     connectors.Lightning,
		Amount:    paymentAmt,
		MediaFee:  sat2DecAmount(btcutil.Amount(resp.PaymentRoute.TotalFees)),
		MediaID:   paymentHash,
	}

	if err := c.cfg.PaymentStore.SavePayment(payment); err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable add payment in store: %v", err)
	}

	return payment, nil
}

// ReceivedPayments returns channel with transactions which are passed
// the minimum threshold required by the client to treat as confirmed.
//
// NOTE: Part of the connectors.LightningConnector interface.
func (c *Connector) ReceivedPayments() <-chan *connectors.Payment {
	return c.notifications
}

// Info returns the information about our lnd node.
//
// NOTE: Part of the connectors.LightningConnector interface.
func (c *Connector) Info() (*connectors.LightningInfo, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodInfo, c.cfg.Metrics)
	defer m.Finish()

	req := &lnrpc.GetInfoRequest{}
	info, err := c.client.GetInfo(context.Background(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	return &connectors.LightningInfo{
		Host:            c.cfg.PeerHost,
		Port:            c.cfg.PeerPort,
		MinAmount:       "0.00000001",
		MaxAmount:       "0.042",
		GetInfoResponse: info,
	}, nil
}

// QueryRoutes returns list of routes from to the given lnd node,
// and insures the the capacity of the channels is sufficient.
//
// NOTE: Part of the connectors.LightningConnector interface.
func (c *Connector) QueryRoutes(pubKey, amount string, limit int32) ([]*lnrpc.Route, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodQueryRoutes, c.cfg.Metrics)
	defer m.Finish()

	satoshis, err := btcToSatoshi(amount)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("unable to convert amount: %v", err)
	}

	// First parse the hex-encdoed public key into a full public key object
	// to check that it is valid.
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf(
			"unable decode identity key from string: %v", err)
	}

	if _, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256()); err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("unable decode identity key: %v", err)
	}

	req := &lnrpc.QueryRoutesRequest{
		PubKey:    pubKey,
		Amt:       satoshis,
		NumRoutes: limit,
	}

	info, err := c.client.QueryRoutes(context.Background(), req)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	return info.Routes, nil
}

// ValidateInvoice takes the encoded lightning network invoice and ensure
// its valid.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ValidateInvoice(invoiceStr, amountStr string) error {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodValidateInvoice, c.cfg.Metrics)
	defer m.Finish()

	netParams, err := bitcoin.GetParams(c.cfg.Net)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable load network params: %v", err)
	}

	amount, err := btcToSatoshi(amountStr)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return errors.Errorf("unable convert amount: %v", err)
	}

	invoice, err := zpay32.Decode(invoiceStr, netParams)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return errors.Errorf("unable decode invoice: %v", err)
	}

	if invoice.MilliSat.ToSatoshis() != btcutil.Amount(amount) {
		m.AddError(metrics.LowSeverity)
		return errors.Errorf("wrong amount")
	}

	return nil
}

// ConfirmedBalance return the amount of confirmed funds available for account.
// TODO(andrew.shvv) Show funds locked in the channels
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ConfirmedBalance(account string) (decimal.Decimal, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodConfirmedBalance, c.cfg.Metrics)
	defer m.Finish()

	req := &lnrpc.WalletBalanceRequest{}
	resp, err := c.client.WalletBalance(context.Background(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return decimal.Zero, err
	}

	balanceSatoshis := decimal.New(resp.ConfirmedBalance, 0)
	balanceBTC := balanceSatoshis.Div(satoshiPerBitcoin)
	return balanceBTC.Round(8), nil
}

// PendingBalance return the amount of funds waiting to be confirmed.
// TODO(andrew.shvv) Show funds locked in the channels
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) PendingBalance(account string) (decimal.Decimal, error) {
	m := crypto.NewMetric(c.cfg.Name, "BTC", MethodConfirmedBalance, c.cfg.Metrics)
	defer m.Finish()

	req := &lnrpc.WalletBalanceRequest{}
	resp, err := c.client.WalletBalance(context.Background(), req)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return decimal.Zero, err
	}

	balanceSatoshis := decimal.New(resp.UnconfirmedBalance, 0)
	balanceBTC := balanceSatoshis.Div(satoshiPerBitcoin)
	return balanceBTC.Round(8), nil
}

// reportMetrics is used to report necessary health metrics about internal
// state of the connector.
func (c *Connector) reportMetrics() error {
	asset := connectors.BTC
	m := crypto.NewMetric("lnd", string(asset),
		"ReportMetrics", c.cfg.Metrics)
	defer m.Finish()

	var overallSent decimal.Decimal
	var overallReceived decimal.Decimal
	var overallFee decimal.Decimal

	payments, err := c.cfg.PaymentStore.ListPayments(asset,
		connectors.Completed, "", connectors.Lightning)
	if err != nil {
		return errors.Errorf("unable to list payments: %v", err)
	}

	for _, payment := range payments {
		if payment.Direction == connectors.Incoming {
			overallReceived = overallReceived.Add(payment.Amount)
		}

		if payment.Direction == connectors.Outgoing {
			overallSent = overallSent.Add(payment.Amount)
			overallFee = overallFee.Add(payment.MediaFee)
		}

		if payment.Direction == connectors.Internal {
			overallFee = overallFee.Add(payment.MediaFee)
		}
	}

	overallReceivedF, _ := overallReceived.Float64()
	m.OverallReceived(overallReceivedF)

	overallSentF, _ := overallSent.Float64()
	m.OverallSent(overallSentF)

	overallFeeF, _ := overallFee.Float64()
	m.OverallFee(overallFeeF)

	log.Infof("Metrics reported, overall received(%v %v), "+
		"overall sent(%v %v), overall fee(%v %v)", overallReceivedF,
		asset, overallSentF, asset, overallFeeF, asset)

	return nil
}
