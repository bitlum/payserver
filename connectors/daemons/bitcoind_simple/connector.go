package bitcoind_simple

import (
	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/bitlum/connector/metrics"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btclog"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// allAccounts denotes that request should aggregate response for all
	// accounts available.
	allAccounts = "*"

	// defaultAccount denotes default account of wallet.
	defaultAccount = ""

	// minimumFeeRate is the minimal satoshis which we should pay for one byte
	//  of information in blockchain.
	minimumFeeRate = decimal.NewFromFloat(1.0)
)

// Config is a bitcoind config.
type Config struct {
	// Net blockchain network this connector should operate with.
	Net string

	// MinConfirmations is a minimum number of confirmations which is needed
	// to treat transaction as confirmed.
	MinConfirmations int

	// RPCClient...
	RPCClient rpc.Client

	// Asset is an asset with which this connector is working.
	Asset connectors.Asset

	// FeePerByte fee sat/byte, which blockchain miners require for including
	// transactions in two coming blocks.
	//
	// NOTE: This is used only if internal system was unable to return fee rate.
	FeePerByte int

	Logger btclog.Logger

	// Metric is an metrics backend which is used for tracking the metrics of
	// connector.
	Metrics crypto.MetricsBackend

	// StateStorage is used to keep data which is needed for connector to
	// properly synchronise and track transactions.
	StateStore StateStorage

	// PaymentStorage is an external storage for payments, it is used by
	// connector to save payment as well as update its state.
	PaymentStore connectors.PaymentsStore
}

func (c *Config) validate() error {
	if c.Net == "" {
		return errors.New("net should be specified")
	}

	if c.MinConfirmations <= 0 {
		return errors.New("min confirmations shouldn't be less or equal " +
			" zero")
	}

	if c.RPCClient == nil {
		return errors.New("rpc client is not specified")
	}

	if c.Logger == nil {
		return errors.New("logger should be specified")
	}

	if c.Asset == "" {
		return errors.New("asset should be specified")
	}

	if c.FeePerByte == 0 {
		return errors.New("fee per unit should be specified")
	}

	if c.Metrics == nil {
		return errors.New("metrics backend should be specified")
	}

	if c.PaymentStore == nil {
		return errors.New("payment store should be specified")
	}

	if c.StateStore == nil {
		return errors.New("state store should be specified")
	}

	return nil
}

type Connector struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg    *Config
	client rpc.Client

	netParams *chaincfg.Params
	log       *common.NamedLogger
}

// A compile time check to ensure Connector implements the BlockchainConnector
// interface.
var _ connectors.BlockchainConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Connector{
		cfg:    cfg,
		quit:   make(chan struct{}),
		client: cfg.RPCClient,
		log: &common.NamedLogger{
			Name:   string(cfg.Asset),
			Logger: cfg.Logger,
		},
	}, nil
}

func (c *Connector) Start() (err error) {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		c.log.Warn("client already started")
		return nil
	}

	defer func() {
		// If start has failed than, we should oll back mark that
		// service has started, so that we could start server again if needed.
		if err != nil {
			atomic.SwapInt32(&c.started, 0)
		}
	}()

	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	resp, err := c.client.GetBlockChainInfo()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to get type of network: %v", err)
	}

	if !isProperNet(c.cfg.Net, resp.Chain) {
		return errors.Errorf("networks are different, desired: %v, "+
			"actual: %v", c.cfg.Net, resp.Chain)
	}

	c.log.Infof("Init connector working with '%v' net", c.cfg.Net)

	c.netParams, err = getParams(c.cfg.Asset, resp.Chain)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("failed to get net params: %v", err)
	}

	c.wg.Add(1)
	go func() {
		defer func() {
			c.log.Info("Quit sync tx goroutine")
			c.wg.Done()
		}()

		c.log.Info("Start sync tx goroutine")

		syncPaymentStateTicker := time.NewTicker(time.Second * time.Duration(10))
		defer syncPaymentStateTicker.Stop()

		for {
			select {
			case <-syncPaymentStateTicker.C:
				if err := c.syncPaymentState(); err != nil {
					m.AddError(metrics.MiddleSeverity)
					c.log.Errorf("unable to sync payment state: %v", err)
					continue
				}
			case <-c.quit:
				return
			}
		}
	}()

	c.wg.Add(1)
	go func() {
		defer func() {
			c.log.Info("Quit reporting  goroutine")
			c.wg.Done()
		}()

		c.log.Info("Starting reporting goroutine...")

		reportTicker := time.NewTicker(time.Second * 30)
		defer reportTicker.Stop()

		for {
			select {
			case <-reportTicker.C:
				if err := c.reportMetrics(); err != nil {
					c.log.Error(err)
					continue
				}
			case <-c.quit:
				return
			}
		}
	}()

	return err
}

func (c *Connector) Stop(reason string) {
	if !atomic.CompareAndSwapInt32(&c.shutdown, 0, 1) {
		c.log.Warn("client already shutdown")
		return
	}

	c.log.Infof("client shutting down (reason: %v)...", reason)
	close(c.quit)

	c.wg.Wait()

	c.log.Info("client shutdown")
}

// CreateAddress is used to create deposit address.
func (c *Connector) CreateAddress() (string, error) {
	address, err := c.cfg.RPCClient.GetNewAddress(defaultAccount)
	if err != nil {
		return "", err
	}

	return address.String(), nil
}

// ConfirmedBalance return the amount of confirmed funds available for account.
func (c *Connector) ConfirmedBalance() (decimal.Decimal, error) {
	amount, err := c.cfg.RPCClient.GetBalanceByLabel(allAccounts, c.cfg.MinConfirmations)
	if err != nil {
		return decimal.Zero, err
	}

	return sat2DecAmount(amount), nil
}

// PendingBalance return the amount of funds waiting to be confirmed.
func (c *Connector) PendingBalance() (decimal.Decimal,
	error) {

	overallBalance, err := c.cfg.RPCClient.GetBalanceByLabel(allAccounts, 0)
	if err != nil {
		return decimal.Zero, err
	}

	confirmedBalance, err := c.cfg.RPCClient.GetBalanceByLabel(allAccounts, c.cfg.MinConfirmations)
	if err != nil {
		return decimal.Zero, err
	}

	return sat2DecAmount(overallBalance - confirmedBalance), nil
}

// SendPayment sends payment with given amount to the given address.
func (c *Connector) SendPayment(address, amount string) (*connectors.Payment, error) {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	decodedAddress, err := decodeAddress(c.cfg.Asset, address, c.netParams.Name)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("invalid address: %v", err)
	}

	amtInBtc, err := decimal.NewFromString(amount)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("unable to decode amount: %v", err)
	}

	txHash, err := c.cfg.RPCClient.SendToAddress(decodedAddress, decAmount2Sat(amtInBtc))
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable send transaction: %v", err)
	}

	tx, err := c.cfg.RPCClient.GetTransaction(txHash)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable get transaction by hash: %v", err)
	}

	payment := &connectors.Payment{
		UpdatedAt: connectors.ConvertTimeToMilliSeconds(time.Now()),
		Status:    connectors.Pending,
		Direction: connectors.Outgoing,
		System:    connectors.External,
		Receipt:   address,
		Asset:     c.cfg.Asset,
		Media:     connectors.Blockchain,
		Amount:    amtInBtc,
		MediaFee:  decimal.NewFromFloat(tx.Fee).Abs().Round(8),
		MediaID:   txHash.String(),
	}

	payment.PaymentID, err = payment.GenPaymentID()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable generate payment id: %v", err)
	}

	if err := c.cfg.PaymentStore.SavePayment(payment); err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable save payment id: %v", err)
	}

	return payment, nil
}

// DecodeAddress takes the blockchain address and ensure its validity.
func (c *Connector) ValidateAddress(address string) error {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	_, err := decodeAddress(c.cfg.Asset, address, c.netParams.Name)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return errors.Errorf("invalid address: %v", err)
	}

	return nil
}

// EstimateFee estimate fee for the transaction with the given sending
// amount.
//
// NOTE: Fee depends on amount because of the number amount of inputs
// which has to be used to construct the transaction.
func (c *Connector) EstimateFee(amount string) (decimal.Decimal, error) {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	// Estimate fee for the median transaction size of 225 bytes.
	// TODO(andrew.shvv) Use amount to construct actual transaction and
	// calculate its size.
	size := decimal.New(225, 0)

	feeRateSatoshiPerByte := c.getFeeRate()
	feeInSatoshis := feeRateSatoshiPerByte.Mul(size)
	feeInBitcoin := feeInSatoshis.Div(satoshiPerBitcoin)

	return feeInBitcoin.Round(8), nil
}

// getFeeRate estimates the approximate rate in sat/byte needed for a
// transaction to begin confirmation within 2 blocks if possible.
//
// NOTE: Uses virtual transaction size as defined in BIP 141
// (witness data is discounted).
func (c *Connector) getFeeRate() decimal.Decimal {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	feeRate, err := c.client.EstimateFee()
	if err != nil {
		if c.cfg.Net == "mainnet" {
			// In case of mainnet such situation happens rarely for that
			// reason we should notify about that. But in testnet and simnet
			// usually not enough data to make fee proper fee estimation.
			c.log.Errorf("unable get fee rate: %v", err)
			m.AddError(metrics.HighSeverity)
		}

		// Take fee rate from config, which was initialised on the start of
		// payserver.
		feeRateSatoshiPerByte := decimal.New(int64(c.cfg.FeePerByte), 0).Round(8)
		if feeRateSatoshiPerByte.LessThan(minimumFeeRate) {
			feeRateSatoshiPerByte = minimumFeeRate
		}

		c.log.Debugf("Get fee rate(%v sat/byte) from config", feeRateSatoshiPerByte)
		return feeRateSatoshiPerByte
	}

	// Initially rate is returned as BTC/Kb, for convience we convert it
	// to sat/byte.
	feeRateBtcPerKiloByte := decimal.NewFromFloat(feeRate)
	bytesInKiloByte := decimal.NewFromFloat(1024)
	feeRateSatoshiPerKiloByte := feeRateBtcPerKiloByte.Mul(satoshiPerBitcoin)
	feeRateSatoshiPerByte := feeRateSatoshiPerKiloByte.Div(bytesInKiloByte).Round(8)

	if feeRateSatoshiPerByte.LessThan(minimumFeeRate) {
		feeRateSatoshiPerByte = minimumFeeRate
	}

	c.log.Debugf("Get fee rate(%v sat/byte) from daemon",
		feeRateSatoshiPerByte)

	return feeRateSatoshiPerByte
}

// syncPaymentState synchronise state of the payment and put in the db,
// in order to avoid fetching directly from bitcoind daemon.
func (c *Connector) syncPaymentState() error {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	allTXs, err := c.cfg.RPCClient.ListTransactionByLabel(allAccounts, math.MaxInt16, 0)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return err
	}

	txCounter, err := c.cfg.StateStore.LastTxCounter()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable get last tx counter: %v", err)
	}

	if len(allTXs) < txCounter {
		return errors.Errorf("unexpected behaviour: number of downloaded " +
			"transaction less than previously sync counter")
	}

	// TODO(andrew.shvv) it will stop syncing if blockchain daemon is cleaned
	//  and started again
	newTXS := allTXs[txCounter:]
	if len(newTXS) == 0 {
		return nil
	}

	c.log.Debugf("sync %v payments, last tx counter was %v",
		len(newTXS), txCounter)

	for _, tx := range newTXS {
		var status connectors.PaymentStatus
		if tx.Confirmations >= int64(c.cfg.MinConfirmations) {
			status = connectors.Completed
		} else {
			status = connectors.Pending
		}

		var direction connectors.PaymentDirection
		switch tx.Category {
		case "send":
			direction = connectors.Outgoing
		case "receive":
			direction = connectors.Incoming
		default:
			c.log.Errorf("unknown tx category: %v", tx.Category)
			m.AddError(metrics.HighSeverity)

			txCounter++
			err := c.cfg.StateStore.PutLastSyncedTxCounter(txCounter)
			if err != nil {
				return errors.Errorf("unable save last synced tx "+
					"counter: %v", err)
			}

			continue
		}

		fee := decimal.Zero
		if tx.Fee != nil {
			fee = decimal.NewFromFloat(*tx.Fee).Abs().Round(8)
		}

		p := &connectors.Payment{
			UpdatedAt: connectors.ConvertTimeToMilliSeconds(time.Now()),
			Status:    status,
			Direction: direction,
			System:    connectors.External,
			Receipt:   tx.Address,
			Asset:     c.cfg.Asset,
			Media:     connectors.Blockchain,
			Amount:    decimal.NewFromFloat(tx.Amount).Abs().Round(8),
			MediaFee:  fee,
			MediaID:   tx.TxID,
		}

		p.PaymentID, err = p.GenPaymentID()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			return errors.Errorf("unable to generate payment id, txid(%v): %v",
				tx.TxID, err)
		}

		if _, err := c.cfg.PaymentStore.PaymentByID(p.PaymentID); err != nil {
			c.log.Infof("New payment(%v) has been found: %v", p.PaymentID,
				spew.Sdump(p))
		}

		if err := c.cfg.PaymentStore.SavePayment(p); err != nil {
			m.AddError(metrics.HighSeverity)
			return errors.Errorf("unable to save payment(%v): %v",
				p.PaymentID, err)
		}

		// Increment tx synced counter only or confirmed transaction,
		// so that we updated pending transaction earlier.
		if tx.Confirmations >= int64(c.cfg.MinConfirmations) {
			c.log.Infof("Payment(%v) is completed: %v", p.PaymentID,
				spew.Sdump(p))

			txCounter++
			err := c.cfg.StateStore.PutLastSyncedTxCounter(txCounter)
			if err != nil {
				return errors.Errorf("unable save last synced tx "+
					"counter: %v", err)
			}
		}
	}

	return nil
}

// reportMetrics is used to report necessary health metrics about internal
// state of the connector.
func (c *Connector) reportMetrics() error {
	m := crypto.NewMetric(c.client.DaemonName(), string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	var overallSent decimal.Decimal
	var overallReceived decimal.Decimal
	var overallFee decimal.Decimal

	payments, err := c.cfg.PaymentStore.ListPayments(c.cfg.Asset,
		connectors.Completed, "", connectors.Blockchain, "")
	if err != nil {
		m.AddError(metrics.MiddleSeverity)
		return errors.Errorf("unable to list payments: %v", err)
	}

	for _, payment := range payments {
		if payment.Direction == connectors.Incoming &&
			payment.System == connectors.External {
			overallReceived = overallReceived.Add(payment.Amount)
		}

		if payment.Direction == connectors.Outgoing &&
			payment.System == connectors.External {
			overallSent = overallSent.Add(payment.Amount)
			overallFee = overallFee.Add(payment.MediaFee)
		}

		overallFee = overallFee.Add(payment.MediaFee)
	}

	overallReceivedF, _ := overallReceived.Float64()
	m.OverallReceived(overallReceivedF)

	overallSentF, _ := overallSent.Float64()
	m.OverallSent(overallSentF)

	overallFeeF, _ := overallFee.Float64()
	m.OverallFee(overallFeeF)

	balance, err := c.ConfirmedBalance()
	if err != nil {
		m.AddError(metrics.MiddleSeverity)
		return errors.Errorf("unable to get available funds: %v", err)
	}

	c.log.Infof("Asset(%v), media(blockchain), available funds(%v)",
		c.cfg.Asset, balance.Round(8).String())

	f, _ := balance.Float64()
	m.CurrentFunds(f)

	c.log.Infof("Metrics reported, overall received(%v %v), "+
		"overall sent(%v %v), overall fee(%v %v)", overallReceivedF,
		c.cfg.Asset, overallSentF, c.cfg.Asset, overallFeeF, c.cfg.Asset)

	return nil
}
