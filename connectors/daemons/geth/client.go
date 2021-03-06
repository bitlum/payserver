package geth

import (
	"fmt"
	"time"

	"sync"

	"sync/atomic"

	"math/big"

	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/rpc/ethereum"
	"github.com/bitlum/connector/metrics"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/btcsuite/btclog"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
)

var (
	// weiInEth is a number of wei in the one Ethereum.
	weiInEth = decimal.NewFromFloat(1e18)

	// defaultTxGas is the number of gas in ethereum which is needed to
	// propagate the transaction.
	defaultTxGas = int64(90000)
)

type internalAccount string

var (
	// defaultAccount is an account which is used to aggregate all money on
	// it. When service received money it redirect them on default address
	// so that later we could use only one transaction for all payments.
	defaultAccount internalAccount = "default"

	// zigzagAccount is an account where all users address belong to,
	// originally all money goes on this address, and later redirected on
	// default address.
	// TODO(andrew.shvv) rename, do not forget about db migration
	zigzagAccount internalAccount = "zigzag"
)

type DaemonConfig struct {
	Name       string
	ServerHost string
	ServerPort int
	Password   string
}

// Config is a connector config.
type Config struct {
	// Net blockchain network this connector should operate with.
	Net string

	// MinConfirmations is a minimum number of block on top of the
	// one where transaction appeared, before we consider transaction as
	// confirmed.
	MinConfirmations int

	// SyncTickDelay is for how long processing loop should sleep before
	// start syncing pending, confirmed and mempool transactions.
	SyncTickDelay int

	// LastSyncedBlockHash is the hash of block which were proceeded last.
	// In this field is specified, than hash will be initialized from it,
	// rather than from database.
	LastSyncedBlockHash string

	// DaemonCfg holds the information about how to connect to the
	// blockchain daemon.
	DaemonCfg *DaemonConfig

	// Asset denotes asset which is represented by this config.
	Asset connectors.Asset

	Logger btclog.Logger

	// Metrics is a metric backend which is used to collect metrics from
	// connector. In case of prometheus client they stored locally till
	// they will be collected by prometheus server.
	Metrics crypto.MetricsBackend

	// PaymentStorage is an external storage for payments, it is used by
	// connector to save payment as well as update its state.
	PaymentStorage connectors.PaymentsStore

	// AccountsStorage is used to keep track connections between addresses and
	// accounts, because of the reason of Ethereum client not having this mapping
	// internally.
	AccountStorage AccountsStorage

	// StateStorage is used to keep data which is needed for connector to
	// properly synchronise and track transactions.
	StateStorage connectors.StateStorage
}

func (c *Config) validate() error {
	if c.Net == "" {
		return errors.New("net should be specified")
	}

	if c.DaemonCfg == nil {
		return errors.New("daemon config should be specified")
	}

	if c.Logger == nil {
		return errors.New("logger should be specified")
	}

	if c.SyncTickDelay == 0 {
		c.SyncTickDelay = 5
	}

	if c.Asset == "" {
		return errors.New("asset should be specified")
	}

	if c.Metrics == nil {
		return errors.New("metrics backend should be specified")
	}

	if c.AccountStorage == nil {
		return errors.New("account store should be specified")
	}

	if c.PaymentStorage == nil {
		return errors.New("payment store should be specified")
	}

	if c.StateStorage == nil {
		return errors.New("state store should be specified")
	}

	return nil
}

// Connector is an implementation of BlockchainConnector which interacts with
// geth daemon.
type Connector struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg    *Config
	client *ExtendedEthRpc

	// defaultAddress is the address which is used as the aggregator address
	// for all incoming transaction. Every payment we receive will be redirected
	// on this address, so that later it could be used for sending transaction
	// from it.
	defaultAddress string

	// memPoolTxs contains transaction which are not yet in the blockchain
	// and still waiting to be included in the blocks.
	memPoolTxs pendingMap

	// unconfirmedTxs is the map of transaction which are already in the
	// blockchains, but not yet confirmed from our pov.
	unconfirmedTxs pendingMap
	pendingLock    sync.Mutex

	log *common.NamedLogger
}

// A compile time check to ensure Connector implements the BlockchainConnector
// interface.
var _ connectors.BlockchainConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Connector{
		cfg:            cfg,
		quit:           make(chan struct{}),
		memPoolTxs:     make(pendingMap),
		unconfirmedTxs: make(pendingMap),
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
		// service has started.
		if err != nil {
			atomic.SwapInt32(&c.started, 0)
		}
	}()

	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	c.log.Info("Creating RPC client...")
	url := fmt.Sprintf("http://%v:%v", c.cfg.DaemonCfg.ServerHost,
		c.cfg.DaemonCfg.ServerPort)
	c.client = &ExtendedEthRpc{ethrpc.NewEthRPC(url)}

	version, err := c.client.NetVersion()
	if err != nil {
		return errors.Errorf("unable to get net version: %v", err)
	}

	if c.cfg.Net != convertVersion(version) {
		return errors.Errorf("networks are different, desired: %v, "+
			"actual: %v", c.cfg.Net, convertVersion(version))
	}

	c.log.Infof("Init connector working with '%v' net", convertVersion(version))

	c.log.Info("Getting last synced block hash...")
	var lastSyncedBlockHash string
	if c.cfg.LastSyncedBlockHash != "" {
		lastSyncedBlockHash = c.cfg.LastSyncedBlockHash

		c.log.Infof("Get synced block hash(%v) from config",
			lastSyncedBlockHash)
	} else {
		lastSyncedBlockHash, err = c.fetchLastSyncedBlockHash()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			return errors.Errorf("unable to fetch last block synced "+
				"hash: %v", err)
		}

		c.log.Infof("Last synced block hash(%v)", lastSyncedBlockHash)
	}

	defaultAddress, err := c.fetchDefaultAddress()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to fetch default address: %v", err)
	}
	c.log.Infof("Default address: %v", defaultAddress)
	c.defaultAddress = defaultAddress

	// Initialise default address nonce by asking ethereum about it.
	txCount, err := c.client.EthGetTransactionCount(c.defaultAddress, "pending")
	if err != nil {
		return errors.Errorf("unable to get default transactions count: %v",
			err)
	}

	dbNonce, err := c.cfg.AccountStorage.DefaultAddressNonce()
	if err != nil {
		return errors.Errorf("unable to put default account "+
			"nonce in db: %v", err)
	}

	// Only save transaction count if nonce db for some reason was erased
	// / lost.
	if txCount > dbNonce {
		if err := c.cfg.AccountStorage.PutDefaultAddressNonce(txCount); err != nil {
			return errors.Errorf("unable to put default account "+
				"nonce in db: %v", err)
		}
	}

	c.log.Infof("Default address nonce is: %v", txCount)

	c.wg.Add(1)
	go func() {
		syncBlockDelay := time.Duration(c.cfg.SyncTickDelay) * time.Second
		syncingBlockTicker := time.NewTicker(syncBlockDelay)
		reportTicker := time.NewTicker(time.Second * 30)

		defer func() {
			c.log.Info("Quit syncing transactions goroutine")

			syncingBlockTicker.Stop()
			reportTicker.Stop()

			c.wg.Done()
		}()

		c.log.Info("Starting syncing goroutine...")

		for {
			select {
			case <-syncingBlockTicker.C:
				prevLastSyncedBlockHash := lastSyncedBlockHash

				newLastSyncedBlock, err := c.syncBlock(prevLastSyncedBlockHash)
				if err != nil {
					c.log.Errorf("unable to sync: %v", err)
					continue
				}

				if newLastSyncedBlock.Hash != prevLastSyncedBlockHash {
					lastSyncedBlockHash = newLastSyncedBlock.Hash

					c.log.Infof("Last synced block hash (%v) number(%v)",
						newLastSyncedBlock.Hash, newLastSyncedBlock.Number)

					err := c.syncPendingTransactions(newLastSyncedBlock.Number)
					if err != nil {
						c.log.Errorf("unable to sync pending "+
							"transactions: %v", err)
						continue
					}
				}

			case <-reportTicker.C:
				if err := c.reportMetrics(); err != nil {
					c.log.Errorf("unable to report metric: %v", err)
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

func (c *Connector) WaitShutDown() <-chan struct{} {
	return c.quit
}

// CreateAddress is used to create deposit address.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) CreateAddress() (string, error) {
	return c.createAddress(zigzagAccount)
}

func (c *Connector) createAddress(account internalAccount) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	var address string
	var err error

	address, err = c.client.PersonalNewAddress(c.cfg.DaemonCfg.Password)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", errors.Errorf("unable to create address: %v", err)
	}

	err = c.cfg.AccountStorage.AddAddressToAccount(address, string(account))
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", err
	}

	return address, err
}

// CreatePayment generates the payment, but not sends it,
// instead returns the payment id and waits for it to be approved.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) SendPayment(toAddress, amountStr string) (*connectors.Payment, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, errors.Errorf("unable parse amount: %v", err)
	}

	// If we send transaction too frequently ethereum transaction counter
	// is not working properly, for that reason we use internal nonce counter.
	nonce, err := c.cfg.AccountStorage.DefaultAddressNonce()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to get default nonce: %v", err)
	}

	details, fee, err := c.generateTransaction(c.defaultAddress, toAddress,
		amount, false, nonce)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	payment := &connectors.Payment{
		UpdatedAt: connectors.NowInMilliSeconds(),
		Status:    connectors.Waiting,
		Direction: connectors.Outgoing,
		System:    connectors.External,
		Receipt:   toAddress,
		Asset:     connectors.Asset(c.cfg.Asset),
		Media:     connectors.Blockchain,
		Amount:    amount.Round(8),
		MediaFee:  fee,
		MediaID:   details.TxID,
		Detail:    details,
	}

	payment.PaymentID, err = payment.GenPaymentID()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, err
	}

	if err := c.cfg.PaymentStorage.SavePayment(payment); err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable add payment(%v) in store: %v",
			payment.PaymentID, err)
	}

	c.log.Infof("Create payment %v", spew.Sdump(payment))

	return c.sendPayment(payment.PaymentID, true)
}

func (c *Connector) generateTransaction(fromAddress, toAddress string,
	amount decimal.Decimal, includeFee bool,
	nonce int) (*connectors.GeneratedTxDetails,
	decimal.Decimal, error) {

	// Fetch suggested by the daemon gas price.
	gp, err := c.client.EthGasPrice()
	if err != nil {
		return nil, decimal.Zero, err
	}

	gasPrice := big.NewInt(0)
	gasPrice, _ = gasPrice.SetString(gp, 0)

	weiAmount := big.NewInt(0)
	weiAmount.SetString(amount.Mul(weiInEth).String(), 0)
	txAmount := weiAmount

	gas := big.NewInt(defaultTxGas)
	txFee := new(big.Int).Mul(gas, gasPrice)

	// Ensure that we are not trying to send negative amount.
	if includeFee && txFee.Cmp(txAmount) > 0 {
		return nil, decimal.Zero, errors.New("fee is greater than amount")
	}

	// If transaction is redirected to the default account than we should use
	// "send all the available money" model.
	if includeFee {
		txAmount = new(big.Int).Sub(txAmount, txFee)
	}

	_, err = c.client.PersonalUnlockAddress(fromAddress, c.cfg.DaemonCfg.Password, 2)
	if err != nil {
		return nil, decimal.Zero, errors.Errorf("unable to unlock sender account: %v", err)
	}

	tx, rawTxStr, err := c.client.EthSignTransaction(ethrpc.T{
		From:     fromAddress,
		To:       toAddress,
		Gas:      int(gas.Int64()),
		GasPrice: gasPrice,
		Value:    txAmount,
		Data:     "",
		Nonce:    nonce,
	})
	if err != nil {
		return nil, decimal.Zero, errors.Errorf("unable to sign tx: %v", err)
	}

	c.log.Debugf("Generated transaction, from(%v), to(%v), amount(%v), "+
		"includeFee(%v), nonce(%v)", fromAddress, toAddress, amount,
		includeFee, nonce)

	requiredFee := decimal.NewFromBigInt(txFee, 0).Div(weiInEth).Round(8)
	return &connectors.GeneratedTxDetails{
		RawTx: []byte(rawTxStr),
		TxID:  tx.Hash,
	}, requiredFee, nil
}

// sendPayment sends created previously payment to the
// blockchain network.
func (c *Connector) sendPayment(paymentID string, isFromDefault bool) (*connectors.Payment, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	// We should be able to receive payment which we putted in storage
	// earlier on the stage of payment generation.
	payment, err := c.cfg.PaymentStorage.PaymentByID(paymentID)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable find payment(%v): %v", paymentID,
			err)
	}

	// Extract the detail about payment, which were putter on the stage
	// of creation of the payment, in order to use raw transaction
	// to send it in blockchain.
	details, ok := payment.Detail.(*connectors.GeneratedTxDetails)
	if !ok {
		return nil, errors.Errorf("unable get details for payment(%v)",
			paymentID)
	}

	_, err = c.client.EthSendRawTransaction(string(details.RawTx))
	if err != nil {
		payment.Status = connectors.Failed
		if err := c.cfg.PaymentStorage.SavePayment(payment); err != nil {
			m.AddError(metrics.HighSeverity)
			c.log.Errorf("unable update payment(%v) status: %v",
				paymentID, err)
		}

		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to execute send tx rpc call: %v", err)
	}

	payment.Status = connectors.Pending
	payment.UpdatedAt = connectors.NowInMilliSeconds()

	// If we are sending payment from default address, than we should increase
	// default nonce, because we are tracking nonce to avoid errors related to
	// nonce repeat.
	if isFromDefault {
		nonce, err := c.cfg.AccountStorage.DefaultAddressNonce()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			return nil, errors.Errorf("unable to get default nonce: %v", err)
		}

		err = c.cfg.AccountStorage.PutDefaultAddressNonce(nonce + 1)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			return nil, errors.Errorf("unable to save default nonce: %v", err)
		}

		c.log.Infof("Payment is sent and default address nonce is"+
			" increased to %v", nonce)
	}

	err = c.cfg.PaymentStorage.SavePayment(payment)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		c.log.Errorf("unable update payment(%v) status: %v", paymentID, err)
	}

	c.log.Infof("Sent payment %v", spew.Sdump(payment))

	return payment, nil
}

// ConfirmedBalance returns number of funds available under control of
// connector.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ConfirmedBalance() (decimal.Decimal, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	balance := decimal.Zero

	addresses, err := c.cfg.AccountStorage.GetAddressesByAccount(string(defaultAccount))
	if err != nil {
		return decimal.Zero, err
	}

	for _, address := range addresses {
		weis, err := c.client.EthGetBalance(address, "latest")
		if err != nil {
			return decimal.Zero, err
		}

		amount := decimal.NewFromBigInt(&weis, 0).Div(weiInEth)
		balance = balance.Add(amount)
	}

	return balance, nil
}

// PendingBalance return the amount of funds waiting to be confirmed.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) PendingBalance() (decimal.Decimal,
	error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	var amount decimal.Decimal
	for _, tx := range c.memPoolTxs[string(defaultAccount)] {
		amount = amount.Add(tx.Amount)
	}

	for _, tx := range c.unconfirmedTxs[string(defaultAccount)] {
		amount = amount.Add(tx.Amount)
	}

	return amount.Round(8), nil
}

// syncUnconfirmed process blocks above the minimum confirmations threshold
// and creates the in-memory map of unconfirmed transactions.
func (c *Connector) syncUnconfirmed(bestBlockNumber,
lastSyncedBlockNumber int) (pendingMap, error) {

	unconfirmedTxs := make(pendingMap)
	for {
		select {
		case <-c.quit:
			return nil, errors.Errorf("sync unconfirmed quit")
		default:
		}

		if lastSyncedBlockNumber >= bestBlockNumber {
			return unconfirmedTxs, nil
		}

		nextBlockNumber := lastSyncedBlockNumber + 1
		confirmations := int64(bestBlockNumber - nextBlockNumber)
		block, err := c.client.EthGetBlockByNumber(nextBlockNumber, true)
		if err != nil {
			return nil, errors.Errorf("unable to get last sync block "+
				"from daemon: %v", err)
		}

		for _, tx := range block.Transactions {
			// If we could get account by the address that means that
			// transaction going to our service.
			account, err := c.cfg.AccountStorage.GetAccountByAddress(tx.To)
			if err != nil {
				return nil, err
			}

			if account == "" {
				continue
			}

			amount := decimal.NewFromBigInt(&tx.Value, 0).Div(weiInEth)
			gas := decimal.New(int64(tx.Gas), 0)
			gasPrice := decimal.NewFromBigInt(&tx.GasPrice, 0)
			fee := gas.Mul(gasPrice).Div(weiInEth)

			payment := &connectors.Payment{
				UpdatedAt: connectors.NowInMilliSeconds(),
				Status:    connectors.Pending,
				Account:   account,
				Receipt:   tx.To,
				Asset:     c.cfg.Asset,
				Media:     connectors.Blockchain,
				Amount:    amount,
				MediaFee:  fee,
				MediaID:   tx.Hash,
				Detail: &connectors.BlockchainPendingDetails{
					Confirmations:     confirmations,
					ConfirmationsLeft: int64(c.cfg.MinConfirmations) - confirmations,
				},
			}

			if account == string(defaultAccount) {
				payment.Direction = connectors.Incoming
				payment.System = connectors.Internal
				payment.MediaFee = decimal.Zero
				payment.PaymentID, err = payment.GenPaymentID()
				if err != nil {
					return nil, err
				}

			} else {
				payment.Direction = connectors.Incoming
				payment.System = connectors.External
				payment.MediaFee = decimal.Zero
				payment.PaymentID, err = payment.GenPaymentID()
				if err != nil {
					return nil, err
				}
			}

			if err := c.cfg.PaymentStorage.SavePayment(payment); err != nil {
				return nil, errors.Errorf("unable to save payment(%v): %v",
					payment.PaymentID, err)
			}

			unconfirmedTxs.add(payment)
		}

		lastSyncedBlockNumber = nextBlockNumber
	}
}

// syncPending creates the in-memory map of transactions which
// are in the memory pool of the ethereum blockchain daemon.
func (c *Connector) syncPending() (pendingMap, error) {
	mempoolTxs := make(pendingMap)

	txs, err := c.client.EthGetPendingTxs()
	if err != nil {
		return nil, err
	}

	for _, tx := range txs {
		// If we could get account by the address that means that
		// transaction going to our service.
		account, err := c.cfg.AccountStorage.GetAccountByAddress(tx.To)
		if err != nil {
			return nil, err
		}

		if account == "" {
			continue
		}

		amount := decimal.NewFromBigInt(&tx.Value, 0).Div(weiInEth)
		gas := decimal.New(int64(tx.Gas), 0)
		gasPrice := decimal.NewFromBigInt(&tx.GasPrice, 0)
		fee := gas.Mul(gasPrice).Div(weiInEth)

		payment := &connectors.Payment{
			UpdatedAt: connectors.NowInMilliSeconds(),
			Status:    connectors.Pending,
			Account:   account,
			Receipt:   tx.To,
			Asset:     c.cfg.Asset,
			Media:     connectors.Blockchain,
			Amount:    amount,
			MediaFee:  fee,
			MediaID:   tx.Hash,
			Detail: &connectors.BlockchainPendingDetails{
				Confirmations:     0,
				ConfirmationsLeft: int64(c.cfg.MinConfirmations),
			},
		}

		if account == string(defaultAccount) {
			payment.Direction = connectors.Incoming
			payment.System = connectors.Internal
			payment.MediaFee = decimal.Zero
			payment.PaymentID, err = payment.GenPaymentID()
			if err != nil {
				return nil, err
			}
		} else {
			payment.Direction = connectors.Incoming
			payment.System = connectors.External
			payment.MediaFee = decimal.Zero
			payment.PaymentID, err = payment.GenPaymentID()
			if err != nil {
				return nil, err
			}
		}

		if err := c.cfg.PaymentStorage.SavePayment(payment); err != nil {
			return nil, errors.Errorf("unable to save payment(%v): %v",
				payment.PaymentID, err)
		}

		mempoolTxs.add(payment)
	}

	return mempoolTxs, nil
}

// syncConfirmed process new blocks and notify subscribed clients that
// transaction reached the minimum confirmation limit,
// and fail if notification listener haven't been initialized.
func (c *Connector) syncConfirmed(bestBlockNumber int,
	lastSyncedBlock *ethrpc.Block) (*ethrpc.Block, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)

	for {
		select {
		case <-c.quit:
			return nil, errors.Errorf("sync confirmed quit")
		default:
		}

		// Stop syncing if we reached the threshold of minimum confirmations.
		// Next block will contain transactions which are considered to be
		// unconfirmed.
		confirmations := int64(bestBlockNumber - lastSyncedBlock.Number)
		if confirmations < int64(c.cfg.MinConfirmations)+1 {
			return lastSyncedBlock, nil
		}

		nextBlockNumber := lastSyncedBlock.Number + 1
		block, err := c.client.EthGetBlockByNumber(nextBlockNumber, true)
		if err != nil {
			return nil, err
		}

		for _, confirmedTx := range block.Transactions {
			// By the given address identify is sender address belongs
			// to our system.
			senderAccount, err := c.cfg.AccountStorage.GetAccountByAddress(
				confirmedTx.From)
			if err != nil {
				return nil, err
			}

			// By the given address identify is receiver address belongs
			// to our system.
			receiverAccount, err := c.cfg.AccountStorage.GetAccountByAddress(
				confirmedTx.To)
			if err != nil {
				return nil, err
			}

			// Skip transactions which do not belongs to us.
			if senderAccount == "" && receiverAccount == "" {
				continue
			}

			makeDirection := func(sender, receiver string) string {
				switch sender {
				case "default":
					sender = "default"
				case "":
					sender = "unknown"
				default:
					sender = "accounts"
				}

				switch receiver {
				case "default":
					receiver = "default"
				case "":
					receiver = "unknown"
				default:
					receiver = "accounts"
				}

				return fmt.Sprintf("%v => %v", sender, receiver)
			}

			var (
				// Whether this payment should be considered as user
				// originated or not.
				isInternal bool

				needUpdateStatusOfIncoming bool
				needUpdateStatusOfOutgoing bool

				// Whether this payment has to be redirected on default address.
				needRedirect bool
			)

			d := makeDirection(senderAccount, receiverAccount)
			switch d {
			case "default => default":
				// Such payment doesn't make a lot of sense, but if it made
				// they should be considered as internal, and not be exposed outside.
				isInternal = true
				needUpdateStatusOfIncoming = true
				needUpdateStatusOfOutgoing = true

			case "default => accounts":
				// This payment from our aggregation address,
				// to one of addresses which belong to our wallet.
				// We should track both outgoing and incoming payments.
				isInternal = false

				// Track payment to account
				needUpdateStatusOfIncoming = true

				// Track payment from default to account
				needUpdateStatusOfOutgoing = true

				// After payment will be received on one of the accounts,
				// it should be redirected back to default address.
				needRedirect = true

			case "default => unknown":
				// Is the standard "send" payment from our aggregation
				// address to
				// some user in the network.
				isInternal = false
				needUpdateStatusOfOutgoing = true

			case "accounts => accounts":
				// This payment is done from one user account to another.
				isInternal = false
				needUpdateStatusOfIncoming = true
				needUpdateStatusOfOutgoing = true

			case "accounts => unknown":
				// We are not sending payment from accounts, this should be
				// unexpected behaviour. Payment are done only from default
				// address.
				return nil, errors.Errorf("unexpected behavior, received tx("+
					"%v) from one of the internal accounts", confirmedTx.Hash)

			case "unknown => unknown":
				// This should has been handled previously.
				// We shouldn't handle payment which do not touches our
				// system.
				continue

			case "accounts => default":
				// In this case we receive previously redirected payment
				// on our default address.
				isInternal = true
				needUpdateStatusOfIncoming = true
				needUpdateStatusOfOutgoing = true

			case "unknown => default":
				// Payment on default address from unknown address are
				// unusual. It might be deposit on connector by mistake,
				// because it should be done on account address.
				isInternal = false
				needUpdateStatusOfIncoming = true

			case "unknown => accounts":
				// It is standard "receive" payment from unknown user on the
				// network.
				isInternal = false
				needUpdateStatusOfIncoming = true
				needRedirect = true
			}

			c.log.Infof("Handling %v transaction(%v)", d, confirmedTx.Hash)

			// We need identify what gas was actually used by the network.
			receipt, err := c.client.EthGetTransactionReceipt(confirmedTx.Hash)
			if err != nil {
				return nil, errors.Errorf("unable to get "+
					"transaction receipt for tx(%v): %v", confirmedTx.Hash, err)
			}

			// Convert amount from wei representation to Ethereum.
			amount := decimal.NewFromBigInt(&confirmedTx.Value, 0).Div(weiInEth)
			gas := decimal.New(int64(receipt.GasUsed), 0)
			gasPrice := decimal.NewFromBigInt(&confirmedTx.GasPrice, 0)
			fee := gas.Mul(gasPrice).Div(weiInEth)

			payment := connectors.Payment{
				UpdatedAt: connectors.NowInMilliSeconds(),
				Status:    connectors.Completed,
				Account:   receiverAccount,
				Receipt:   confirmedTx.To,
				Asset:     c.cfg.Asset,
				Media:     connectors.Blockchain,
				Amount:    amount,
				MediaFee:  fee,
				MediaID:   confirmedTx.Hash,
			}

			if isInternal {
				payment.System = connectors.Internal
			} else {
				payment.System = connectors.External
			}

			if needUpdateStatusOfOutgoing {
				outgoingPayment := payment
				outgoingPayment.Direction = connectors.Outgoing
				outgoingPayment.PaymentID, err = outgoingPayment.GenPaymentID()
				if err != nil {
					return nil, err
				}

				if err := c.cfg.PaymentStorage.SavePayment(&outgoingPayment); err != nil {
					return nil, errors.Errorf("unable to add payment to storage: %v",
						outgoingPayment.PaymentID)
				}

				c.log.Infof("Confirm outgoing payment(%v)",
					spew.Sdump(outgoingPayment))
			}

			if needUpdateStatusOfIncoming {
				incomingPayment := payment
				incomingPayment.Direction = connectors.Incoming
				incomingPayment.MediaFee = decimal.Zero
				incomingPayment.PaymentID, err = incomingPayment.GenPaymentID()
				if err != nil {
					return nil, err
				}

				if err := c.cfg.PaymentStorage.SavePayment(&incomingPayment); err != nil {
					return nil, errors.Errorf("unable to add payment to storage: %v",
						incomingPayment.PaymentID)
				}

				c.log.Infof("Confirmed incoming payment(%v)",
					spew.Sdump(incomingPayment))

				if needRedirect {
					// In this case we received transaction on one of our
					// non-internal accounts we should make money
					// aggregation on default account.
					c.log.Infof("Make redirect of payment("+
						"%v)", incomingPayment.PaymentID)
					if err := c.makeRedirect(confirmedTx.To, amount); err != nil {
						c.log.Errorf("unable to make payment(%v) "+
							"redirection: %v", spew.Sdump(incomingPayment), err)
						m.AddError(metrics.HighSeverity)
					}
				}
			}
		}

		// Update database with last synced block hash.
		hash := []byte(block.Hash)
		if err := c.cfg.StateStorage.PutLastSyncedHash(hash); err != nil {
			return nil, errors.Errorf("unable to put block hash in db: %v",
				err)
		}

		lastSyncedBlock = block

		// After transaction has been consumed by other subsystem
		// overwrite cache.
		c.log.Infof("Process block hash(%v), number(%v)", block.Hash, block.Number)

		// Report last synchronised block number from daemon point of view.
		m.BlockNumber(int64(lastSyncedBlock.Number))
	}
}

// makeRedirect is used to make a redirect of previously received money on
// default address. Such aggregation is needed so that later we could use
// default address to send money with one transaction.
func (c *Connector) makeRedirect(initialAddress string, amount decimal.Decimal) error {
	// TODO(andrew.shvv) use persistence task queue on
	// case if fails.

	// Transaction count is used as a nonce to avoid transaction collision.
	txCount, err := c.client.EthGetTransactionCount(initialAddress, "pending")
	if err != nil {
		return errors.Errorf("unable to get transactions count: %v", err)
	}

	// Generate aggregate transaction which sends money from receive
	// address on default account.
	// TODO(andrew.shvv) What if fee is greater than sending amount?
	aggregateTx, fee, err := c.generateTransaction(initialAddress, c.defaultAddress,
		amount, true, txCount)
	if err != nil {
		return errors.Errorf("unable to generate transfer tx(%v): %v", err)
	}

	aggregatePayment := &connectors.Payment{
		UpdatedAt: connectors.NowInMilliSeconds(),
		Status:    connectors.Waiting,
		System:    connectors.Internal,
		Account:   string(defaultAccount),
		Receipt:   c.defaultAddress,
		Asset:     c.cfg.Asset,
		Media:     connectors.Blockchain,
		Amount:    amount.Sub(fee),
		MediaFee:  fee,
		MediaID:   aggregateTx.TxID,
		Detail:    aggregateTx,
	}

	// We have to track both outgoing and incoming for consistency with other
	// connectors.
	aggregatePayment.Direction = connectors.Incoming
	aggregatePayment.PaymentID, err = aggregatePayment.GenPaymentID()
	if err != nil {
		return err
	}

	if err := c.cfg.PaymentStorage.SavePayment(aggregatePayment); err != nil {
		return errors.Errorf("unable to add payment to storage: %v",
			aggregateTx.TxID)
	}

	aggregatePayment.Direction = connectors.Outgoing
	aggregatePayment.PaymentID, err = aggregatePayment.GenPaymentID()
	if err != nil {
		return err
	}

	if err := c.cfg.PaymentStorage.SavePayment(aggregatePayment); err != nil {
		return errors.Errorf("unable to add payment to storage: %v",
			aggregateTx.TxID)
	}

	c.log.Infof("Send redirect payment(%v)", spew.Sdump(aggregatePayment))

	if _, err = c.sendPayment(aggregatePayment.PaymentID, false); err != nil {
		return errors.Errorf("unable to send aggregate tx(%v): %v",
			aggregatePayment.PaymentID, err)
	}

	return nil
}

// fetchLastSyncedBlockHash returns hash of block which were handled in previous
// cycle of processing.
func (c *Connector) fetchLastSyncedBlockHash() (string, error) {
	c.log.Info("Restore hash from database...")
	lastHash, _ := c.cfg.StateStorage.LastSyncedHash()
	if lastHash != nil {
		return string(lastHash), nil
	}

	c.log.Info("Unable to find block in db, fetching best block...")
	bestBlockNumber, err := c.client.EthBlockNumber()
	if err != nil {
		return "", errors.Errorf("unable to request last best block "+
			"hash: %v", err)
	}

	block, err := c.client.EthGetBlockByNumber(bestBlockNumber, false)
	if err != nil {
		return "", errors.Errorf("unable to request last best block "+
			"hash: %v", err)
	}

	if err := c.cfg.StateStorage.PutLastSyncedHash([]byte(block.Hash)); err != nil {
		return "", errors.Errorf("unable to put best block in db: %v", err)
	}

	return block.Hash, nil
}

// fetchDefaultAddress fetch address which will be used for redirection of
// every incoming transaction. This address will be used as pool of liquadity.
func (c *Connector) fetchDefaultAddress() (string, error) {
	defaultAddress, err := c.cfg.AccountStorage.GetLastAccountAddress(
		string(defaultAccount))
	if err != nil && err != ErrAccountAddressNotFound {
		return "", errors.Errorf("unable to get default address: %v", err)
	}

	if defaultAddress == "" {
		c.log.Info("Unable to find default address in db, generating it...")
		defaultAddress, err = c.createAddress(defaultAccount)
		if err != nil {
			return "", errors.Errorf("unable to generate default address: %v", err)
		}
	}

	return defaultAddress, nil
}

// syncBlock synchronise latest blocks and update transactions states,
// returns the lat synced block.
func (c *Connector) syncBlock(lastSyncedBlockHash string) (*ethrpc.Block, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	bestBlockNumber, err := c.client.EthBlockNumber()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to fetch best block number: %v", err)
	}

	lastSyncedBlock, err := c.client.EthGetBlockByHash(lastSyncedBlockHash, false)
	if err != nil {
		// TODO(andrew.shvv) Check reoginizations
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to get last sync block from daemon"+
			": %v", err)
	}

	// Sync block below minimum confirmations threshold,
	// and update payment's states.
	lastSyncedBlock, err = c.syncConfirmed(bestBlockNumber, lastSyncedBlock)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to process blocks: %v", err)
	}

	return lastSyncedBlock, nil
}

// syncPendingTransactions updates states of pending transaction maps.
func (c *Connector) syncPendingTransactions(lastSyncedBlockNumber int) error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	bestBlockNumber, err := c.client.EthBlockNumber()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to fetch best block number: %v", err)
	}

	// Sync block above minimum confirmation threshold and
	// populate unconfirmed pending map with transactions.
	unconfirmedTxs, err := c.syncUnconfirmed(bestBlockNumber, lastSyncedBlockNumber)
	if err != nil {
		m.AddError(metrics.MiddleSeverity)
		return errors.Errorf("unable to sync unconfirmed txs: %v", err)
	}

	c.pendingLock.Lock()
	c.unconfirmedTxs.merge(unconfirmedTxs,
		func(payment *connectors.Payment) {
			details, ok := payment.Detail.(*connectors.BlockchainPendingDetails)
			if !ok {
				c.log.Warn("unable get payment(%v) pending tx details", payment.PaymentID)
				return
			}

			c.log.Infof("Unconfirmed tx(%v) were added, "+
				"account(%v), amount(%v), confirmations(%v), "+
				"left(%v)", payment.PaymentID, payment.Account, payment.Amount,
				details.Confirmations, details.ConfirmationsLeft)
		})
	c.pendingLock.Unlock()

	memPoolTxs, err := c.syncPending()
	if err != nil {
		m.AddError(metrics.MiddleSeverity)
		return errors.Errorf("unable to fetch mempool txs: %v", err)
	}

	c.pendingLock.Lock()
	c.memPoolTxs.merge(memPoolTxs, func(tx *connectors.Payment) {
		c.log.Infof("Mempool tx(%v) were added, "+
			"account(%v), amount(%v)", tx.PaymentID, tx.Account, tx.Amount)
	})
	c.pendingLock.Unlock()

	return nil
}

// ValidateAddress validates given blockchain address.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ValidateAddress(address string) error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	if err := ethereum.ValidateAddress(address); err != nil {
		m.AddError(metrics.LowSeverity)
		return err
	}

	return nil
}

// EstimateFee estimate fee for the transaction with the given sending
// amount.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) EstimateFee(amount string) (decimal.Decimal, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		common.GetFunctionName(), c.cfg.Metrics)
	defer m.Finish()

	// Fetch suggested by the daemon gas price.
	gp, err := c.client.EthGasPrice()
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return decimal.Zero, err
	}

	gasPrice := big.NewInt(0)
	gasPrice, _ = gasPrice.SetString(gp, 0)

	gas := big.NewInt(defaultTxGas)
	txFee := new(big.Int).Mul(gas, gasPrice)

	return decimal.NewFromBigInt(txFee, 0).Div(weiInEth), nil
}

// reportMetrics is used to report necessary health metrics about internal
// state of the connector.
func (c *Connector) reportMetrics() error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		"ReportMetrics", c.cfg.Metrics)
	defer m.Finish()

	var overallSent decimal.Decimal
	var overallReceived decimal.Decimal
	var overallFee decimal.Decimal

	payments, err := c.cfg.PaymentStorage.ListPayments(c.cfg.Asset,
		connectors.Completed, "", connectors.Blockchain, "")
	if err != nil {
		m.AddError(metrics.LowSeverity)
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

		if payment.System == connectors.Internal {
			overallFee = overallFee.Add(payment.MediaFee)
		}
	}

	overallReceivedF, _ := overallReceived.Float64()
	m.OverallReceived(overallReceivedF)

	overallSentF, _ := overallSent.Float64()
	m.OverallSent(overallSentF)

	overallFeeF, _ := overallFee.Float64()
	m.OverallFee(overallFeeF)

	c.log.Infof("Metrics reported, overall received(%v %v), "+
		"overall sent(%v %v), overall fee(%v %v)", overallReceivedF,
		c.cfg.Asset, overallSentF, c.cfg.Asset, overallFeeF, c.cfg.Asset)

	// Check number of funds available and track this metric in metric
	// backend for farther analysis.
	balance, err := c.ConfirmedBalance()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to "+
			"get available funds: %v", err)
	}

	c.log.Infof("Asset(%v), media(blockchain), available funds(%v)",
		c.cfg.Asset, balance.Round(8).String())

	f, _ := balance.Float64()
	m.CurrentFunds(f)

	return nil
}
