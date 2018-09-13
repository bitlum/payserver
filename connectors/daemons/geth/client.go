package geth

import (
	"fmt"
	"time"

	"sync"

	"sync/atomic"

	"math/big"

	"github.com/bitlum/connector/metrics"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/btcsuite/btclog"
	"github.com/go-errors/errors"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/assets/ethereum"
	"github.com/davecgh/go-spew/spew"
)

var (
	// weiInEth is a number of wei in the one Ethereum.
	weiInEth = decimal.NewFromFloat(1e18)

	// defaultAccount is an account which is used to aggregate all money on
	// it. When service received money it redirect them on default address
	// so that later we could use only one transaction for all payments.
	defaultAccount = "default"

	allAccounts = "all"

	// defaultTxGas is the number of gas in ethereum which is needed to
	// propagate the transaction.
	defaultTxGas = int64(90000)
)

const (
	MethodStart               = "Start"
	MethodAccountAddress      = "AccountAddress"
	MethodCreateAddress       = "CreateAddress"
	MethodPendingTransactions = "PendingTransactions"
	MethodCreatePayment       = "MethodCreatePayment"
	MethodSendPayment         = "SendPayment"
	MethodConfirmedBalance    = "ConfirmedBalance"
	MethodPendingBalance      = "PendingBalance"
	MethodSync                = "Sync"
	MethodEstimateFee         = "MethodEstimateFee"
	MethodValidateAddress     = "MethodValidateAddress"
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

	log *connectors.NamedLogger
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
		log: &connectors.NamedLogger{
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
		MethodStart, c.cfg.Metrics)
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

	if err := c.cfg.AccountStorage.PutDefaultAddressNonce(txCount); err != nil {
		return errors.Errorf("unable to put default account "+
			"nonce in db: %v", err)
	}

	c.log.Infof("Default address nonce is: %v", defaultAddress)

	c.wg.Add(1)
	go func() {
		delay := time.Duration(c.cfg.SyncTickDelay) * time.Second
		syncingTicker := time.NewTicker(delay)
		reportTicker := time.NewTicker(time.Second * 30)

		defer func() {
			c.log.Info("Quit syncing transactions goroutine")
			syncingTicker.Stop()
			reportTicker.Stop()
			c.wg.Done()
		}()

		c.log.Info("Starting syncing goroutine...")

		for {
			select {
			case <-syncingTicker.C:
				var err error
				lastSyncedBlockHash, err = c.sync(lastSyncedBlockHash)
				if err != nil {
					c.log.Errorf("unable to sync: %v", err)
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

// AccountAddress return the deposit address of account.
func (c *Connector) AccountAddress(accountAlias connectors.AccountAlias) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodAccountAddress, c.cfg.Metrics)
	defer m.Finish()

	switch accountAlias {
	case "all", "*":
		return "", errors.Errorf("name of account '%v' is "+
			"reserved for internal usage", accountAlias)
	}

	account := aliasToAccount(accountAlias)
	return c.cfg.AccountStorage.GetLastAccountAddress(account)
}

// CreateAddress is used to create deposit address.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) CreateAddress(accountAlias connectors.AccountAlias) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodCreateAddress, c.cfg.Metrics)
	defer m.Finish()

	switch accountAlias {
	case "all", "*":
		return "", errors.Errorf("name of account '%v' is "+
			"reserved for internal usage", accountAlias)
	}

	var address string
	var err error

	address, err = c.client.PersonalNewAddress(c.cfg.DaemonCfg.Password)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", errors.Errorf("unable to create account: %v", err)
	}

	account := aliasToAccount(accountAlias)
	err = c.cfg.AccountStorage.AddAddressToAccount(address, account)
	if err != nil {
		return "", err
	}

	return address, err
}

// PendingTransactions returns the transactions with confirmation number lower
// the required by payment system.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) PendingTransactions(accountAlias connectors.AccountAlias) (
	[]*connectors.Payment, error) {

	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingTransactions, c.cfg.Metrics)
	defer m.Finish()

	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	account := aliasToAccount(accountAlias)

	var payments []*connectors.Payment
	for _, tx := range c.memPoolTxs[account] {
		payments = append(payments, tx)
	}

	for _, tx := range c.unconfirmedTxs[account] {
		payments = append(payments, tx)
	}

	return payments, nil
}

// CreatePayment generates the payment, but not sends it,
// instead returns the payment id and waits for it to be approved.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) CreatePayment(toAddress, amountStr string) (
	*connectors.Payment, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodCreatePayment, c.cfg.Metrics)
	defer m.Finish()

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return nil, errors.Errorf("unable parse amount: %v", err)
	}

	// If we send transaction too frequently ethereum transaction counter
	// for that reason we use internal nonce counter.
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
		PaymentID: generatePaymentID(details.TxID, toAddress, connectors.Outgoing),
		UpdatedAt: connectors.NowInMilliSeconds(),
		Status:    connectors.Waiting,
		Direction: connectors.Outgoing,
		Receipt:   toAddress,
		Asset:     connectors.Asset(c.cfg.Asset),
		Media:     connectors.Blockchain,
		Amount:    amount,
		MediaFee:  fee,
		MediaID:   details.TxID,
		Detail:    details,
	}

	if err := c.cfg.PaymentStorage.SavePayment(payment); err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable add payment(%v) in store: %v",
			payment.PaymentID, err)
	}

	c.log.Infof("Create payment %v", spew.Sdump(payment))

	return payment, err
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

	requiredFee := decimal.NewFromBigInt(txFee, 0).Div(weiInEth).Round(8)
	return &connectors.GeneratedTxDetails{
		RawTx: []byte(rawTxStr),
		TxID:  tx.Hash,
	}, requiredFee, nil
}

// SendPayment sends created previously payment to the
// blockchain network.
//
// NOTE: Part of the connectors.BlockchainConnector interface.
func (c *Connector) SendPayment(paymentID string) (*connectors.Payment, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSendPayment, c.cfg.Metrics)
	defer m.Finish()

	// We should be able to receive payment which we putter in storage
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

	err = c.cfg.PaymentStorage.SavePayment(payment)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		c.log.Errorf("unable update payment(%v) status: %v", paymentID, err)
	}

	// Increase transaction nonce only after transaction is sent.
	// TODO(andrew.shvv) what if multi-thread access to rpc send payment?
	nonce, err := c.cfg.AccountStorage.DefaultAddressNonce()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to get default nonce: %v", err)
	}

	err = c.cfg.AccountStorage.PutDefaultAddressNonce(nonce + 1)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable to get default nonce: %v", err)
	}

	c.log.Infof("Send payment %v", spew.Sdump(payment))

	return payment, nil
}

// ConfirmedBalance returns number of funds available under control of
// connector.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ConfirmedBalance(accountAlias connectors.AccountAlias) (decimal.Decimal, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodConfirmedBalance, c.cfg.Metrics)
	defer m.Finish()

	balance := decimal.Zero
	var addresses []string
	var err error

	account := aliasToAccount(accountAlias)

	if account == allAccounts {
		// Iterate over every accounts and later for every address
		// belonging to this accounts, summing balances from all of them.
		addresses, err = c.cfg.AccountStorage.AllAddresses()
		if err != nil {
			return decimal.Zero, err
		}

	} else {
		addresses, err = c.cfg.AccountStorage.GetAddressesByAccount(account)
		if err != nil {
			return decimal.Zero, err
		}
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
func (c *Connector) PendingBalance(accountAlias connectors.AccountAlias) (decimal.Decimal,
	error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingBalance, c.cfg.Metrics)
	defer m.Finish()

	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	account := aliasToAccount(accountAlias)

	var amount decimal.Decimal
	if account == allAccounts {
		for _, accounts := range c.memPoolTxs {
			for _, payment := range accounts {
				amount = amount.Add(payment.Amount)
			}
		}

		for _, accounts := range c.unconfirmedTxs {
			for _, payment := range accounts {
				amount = amount.Add(payment.Amount)
			}
		}
	} else {
		for _, tx := range c.memPoolTxs[account] {
			amount = amount.Add(tx.Amount)
		}

		for _, tx := range c.unconfirmedTxs[account] {
			amount = amount.Add(tx.Amount)
		}
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

			if account == defaultAccount {
				payment.PaymentID = generatePaymentID(tx.Hash, tx.To, connectors.Internal)
				payment.Direction = connectors.Internal
				payment.MediaFee = decimal.Zero
			} else {
				payment.PaymentID = generatePaymentID(tx.Hash, tx.To, connectors.Incoming)
				payment.Direction = connectors.Incoming
				payment.MediaFee = decimal.Zero
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

		if account == defaultAccount {
			payment.PaymentID = generatePaymentID(tx.Hash, tx.To,
				connectors.Internal)
			payment.Direction = connectors.Internal
			payment.MediaFee = decimal.Zero
		} else {
			payment.PaymentID = generatePaymentID(tx.Hash, tx.To,
				connectors.Incoming)
			payment.Direction = connectors.Incoming
			payment.MediaFee = decimal.Zero
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
					sender = "internal"
				}

				switch receiver {
				case "default":
					receiver = "default"
				case "":
					receiver = "unknown"
				default:
					receiver = "internal"
				}

				return fmt.Sprintf("%v => %v", sender, receiver)
			}

			var (
				needSaveInternal bool
				needSaveIncoming bool
				needSaveOutgoing bool
				needRedirect     bool
			)

			d := makeDirection(senderAccount, receiverAccount)
			switch d {
			case "default => default":
				needSaveInternal = true

			case "default => internal":
				needSaveIncoming = true
				needSaveOutgoing = true
				needRedirect = true

			case "default => unknown":
				needSaveOutgoing = true

			case "internal => internal",
				"internal => unknown":
				// We are not sending payment from internal,
				// this should be unexpected behaviour.
				return nil, errors.Errorf("unexpected behavior, received tx("+
					"%v) from one of the internal accounts", confirmedTx.Hash)

			case "unknown => unknown":
				// This should has been handled previously.
				// We shouldn't handle payment which do not touches our
				// system.
				continue

			case "internal => default":
				// In this case we receive previously redirected payment
				// on our default address.
				needSaveInternal = true

			case "unknown => default":
				needSaveInternal = true

			case "unknown => internal":
				needSaveIncoming = true
				needRedirect = true
			}

			c.log.Infof("Handling %v transaction(%v)", d, confirmedTx.Hash)

			// Convert amount from wei representation to Ethereum.
			amount := decimal.NewFromBigInt(&confirmedTx.Value, 0).Div(weiInEth)
			gas := decimal.New(int64(confirmedTx.Gas), 0)
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

			if needSaveInternal {
				internalPayment := payment
				internalPayment.Direction = connectors.Internal
				internalPayment.MediaFee = decimal.Zero
				internalPayment.PaymentID = generatePaymentID(
					confirmedTx.Hash, confirmedTx.To,
					internalPayment.Direction)

				if err := c.cfg.PaymentStorage.SavePayment(&internalPayment); err != nil {
					return nil, errors.Errorf("unable to add payment to storage: %v",
						internalPayment.PaymentID)
				}

				c.log.Infof("Confirmed internal payment(%v)",
					spew.Sdump(internalPayment))
			}

			if needSaveOutgoing {
				outgoingPayment := payment
				outgoingPayment.PaymentID = generatePaymentID(confirmedTx.Hash,
					confirmedTx.To, connectors.Outgoing)
				outgoingPayment.Direction = connectors.Outgoing

				if err := c.cfg.PaymentStorage.SavePayment(&outgoingPayment); err != nil {
					return nil, errors.Errorf("unable to add payment to storage: %v",
						outgoingPayment.PaymentID)
				}

				c.log.Infof("Confirm outgoing payment(%v)",
					spew.Sdump(outgoingPayment))
			}

			if needSaveIncoming {
				incomingPayment := payment
				incomingPayment.Direction = connectors.Incoming
				incomingPayment.MediaFee = decimal.Zero
				incomingPayment.PaymentID = generatePaymentID(
					confirmedTx.Hash, confirmedTx.To,
					incomingPayment.Direction)

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
						return nil, errors.Errorf("unable to make payment(%v) "+
							"redirection: %v", incomingPayment.PaymentID, err)
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
		PaymentID: generatePaymentID(aggregateTx.TxID, c.defaultAddress, connectors.Internal),
		UpdatedAt: connectors.NowInMilliSeconds(),
		Status:    connectors.Waiting,
		Direction: connectors.Internal,
		Account:   defaultAccount,
		Receipt:   c.defaultAddress,
		Asset:     c.cfg.Asset,
		Media:     connectors.Blockchain,
		Amount:    amount.Sub(fee),
		MediaFee:  fee,
		MediaID:   aggregateTx.TxID,
		Detail:    aggregateTx,
	}

	if err := c.cfg.PaymentStorage.SavePayment(aggregatePayment); err != nil {
		return errors.Errorf("unable to add payment to storage: %v",
			aggregateTx.TxID)
	}

	c.log.Infof("Send redirect payment(%v)", spew.Sdump(aggregatePayment))

	if _, err = c.SendPayment(aggregatePayment.PaymentID); err != nil {
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

	c.cfg.StateStorage.PutLastSyncedHash([]byte(block.Hash))
	if err != nil {
		return "", errors.Errorf("unable to put best block in db: %v", err)
	}

	return block.Hash, nil
}

// fetchDefaultAddress...
func (c *Connector) fetchDefaultAddress() (string, error) {
	defaultAddress, err := c.AccountAddress(connectors.DefaultAccount)
	if err != nil && err != ErrAccountAddressNotFound {
		return "", errors.Errorf("unable to get default address: %v", err)
	}

	if defaultAddress == "" {
		c.log.Info("Unable to find default address in db, generating it...")
		defaultAddress, err = c.CreateAddress(connectors.DefaultAccount)
		if err != nil {
			return "", errors.Errorf("unable to generate default address: %v", err)
		}
	}

	return defaultAddress, nil
}

func (c *Connector) sync(lastSyncedBlockHash string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSync, c.cfg.Metrics)
	defer m.Finish()

	bestBlockNumber, err := c.client.EthBlockNumber()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return lastSyncedBlockHash, errors.Errorf("unable to fetch best block number: %v", err)
	}

	lastSyncedBlock, err := c.client.EthGetBlockByHash(lastSyncedBlockHash, false)
	if err != nil {
		// TODO(andrew.shvv) Check reoginizations
		m.AddError(metrics.HighSeverity)
		return lastSyncedBlockHash, errors.Errorf("unable to get last sync block from daemon: %v", err)
	}

	// Sync block below minimum confirmations threshold
	lastSyncedBlock, err = c.syncConfirmed(bestBlockNumber, lastSyncedBlock)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return lastSyncedBlockHash, errors.Errorf("unable to process blocks: %v", err)
	}
	lastSyncedBlockHash = lastSyncedBlock.Hash

	// Report last synchronised block number from daemon point of view.
	m.BlockNumber(int64(lastSyncedBlock.Number))

	// Sync block above minimum confirmation threshold and
	// populate unconfirmed pending map with transactions.
	unconfirmedTxs, err := c.syncUnconfirmed(bestBlockNumber,
		lastSyncedBlock.Number)
	if err != nil {
		m.AddError(metrics.MiddleSeverity)
		return lastSyncedBlockHash, errors.Errorf("unable to sync unconfirmed txs: %v", err)
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
		return lastSyncedBlockHash,
			errors.Errorf("unable to fetch mempool txs: %v", err)
	}

	c.pendingLock.Lock()
	c.memPoolTxs.merge(memPoolTxs, func(tx *connectors.Payment) {
		c.log.Infof("Mempool tx(%v) were added, "+
			"account(%v), amount(%v)", tx.PaymentID, tx.Account, tx.Amount)
	})
	c.pendingLock.Unlock()

	// Check number of funds available and track this metric in metric
	// backend for farther analysis.
	balance, err := c.ConfirmedBalance(connectors.DefaultAccount)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return lastSyncedBlockHash, errors.Errorf("unable to "+
			"get available funds: %v", err)
	}

	c.log.Infof("Asset(%v), media(blockchain), available funds(%v)",
		c.cfg.Asset, balance.Round(8).String())

	f, _ := balance.Float64()
	m.CurrentFunds(f)

	return lastSyncedBlockHash, nil
}

// ValidateAddress validates given blockchain address.
//
// NOTE: Part of the connectors.Connector interface.
func (c *Connector) ValidateAddress(address string) error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodValidateAddress, c.cfg.Metrics)
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
		MethodEstimateFee, c.cfg.Metrics)
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
		connectors.Completed, "", connectors.Blockchain)
	if err != nil {
		m.AddError(metrics.LowSeverity)
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

	c.log.Infof("Metrics reported, overall received(%v %v), "+
		"overall sent(%v %v), overall fee(%v %v)", overallReceivedF,
		c.cfg.Asset, overallSentF, c.cfg.Asset, overallFeeF, c.cfg.Asset)

	return nil
}
