package geth

import (
	"fmt"

	"time"

	"sync"

	"sync/atomic"

	"strings"

	"encoding/binary"

	"math/big"

	"github.com/AndrewSamokhvalov/slice"
	"github.com/bitlum/connector/common"
	core "github.com/bitlum/viabtc_rpc_client"
	"github.com/bitlum/connector/db"
	"github.com/coreos/bbolt"
	"github.com/btcsuite/btclog"
	"github.com/go-errors/errors"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/bitlum/connector/metrics"
)

var (
	// infoBucket is a bucket which stores the information requires for
	// correct continuation of synchronization.
	infoBucket = []byte("infobucket")

	// accountsToAddressesBucket is a bucket which represent account to
	// address mapping.
	accountsToAddressesBucket = []byte("accountsToAddresses")

	// addressesToAccountsBucket is a bucket which represent account to
	// address mapping.
	addressesToAccountsBucket = []byte("addressesToAccounts")

	// txBucketKey is the bucket which contains the last transaction which has
	// been proceeded. With it ff accidental temporary fork happened we would
	// avoid double notification sending.
	txBucketKey = []byte("txs")

	// txNumber is the number of transactions which we store in tx bucket.
	txNumber = 100

	// drainDelay is a time after which transactions which exceed max
	// number will be removed from tx bucket.
	drainDelay = time.Hour

	// lastSyncedBlockHashKey is the key which corresponds to the last block hash
	// which was handled by the processing handler. We store hash rather than
	// block number because of the possibility of blockchain reorganization.
	lastSyncedBlockHashKey = []byte("lastblockhash")

	// weiInEth is a number of wei in the one ethereum.
	weiInEth = decimal.NewFromFloat(10e16)
)

const (
	MethodStart               = "Start"
	MethodAccountAddress      = "AccountAddress"
	MethodCreateAddress       = "CreateAddress"
	MethodPendingTransactions = "PendingTransactions"
	MethodGenerateTransaction = "GenerateTransaction"
	MethodSendTransaction     = "SendTransaction"
	MethodPendingBalance      = "PendingBalance"
	MethodSync                = "Sync"
)

type DaemonConfig struct {
	Name       string
	ServerHost string
	ServerPort int
	Password   string
}

// Config is a connector config.
type Config struct {
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

	// DataDir is the directory where we should bolddb and logs files.
	DataDir string

	// DaemonCfg holds the information about how to connect to the
	// blockchain daemon.
	DaemonCfg *DaemonConfig

	// Asset denotes asset which is represented by this config.
	Asset core.AssetType

	Logger btclog.Logger

	// Metrics is a metric backend which is used to collect metrics from
	// connector. In case of prometheus client they stored locally till
	// they will be collected by prometheus server.
	Metrics crypto.MetricsBackend
}

func (c *Config) validate() error {
	if c.DataDir == "" {
		return errors.New("data dir should be specified")
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
	db     *db.DB

	defaultAddress string

	memPoolTxs     pendingMap
	unconfirmedTxs pendingMap
	pendingLock    sync.Mutex

	notifications chan []*common.Payment

	log *common.NamedLogger
}

// A compile time check to ensure Connector implements the BlockchainConnector
// interface.
var _ common.BlockchainConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Connector{
		cfg:            cfg,
		notifications:  make(chan []*common.Payment),
		quit:           make(chan struct{}),
		memPoolTxs:     make(pendingMap),
		unconfirmedTxs: make(pendingMap),
		log: &common.NamedLogger{
			Name:   string(cfg.Asset),
			Logger: cfg.Logger,
		},
	}, nil
}

func (c *Connector) Start() error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		c.log.Warn("client already started")
		return nil
	}

	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodStart, c.cfg.Metrics)
	defer m.Finish()

	c.log.Info("Creating RPC client...")
	url := fmt.Sprintf("http://%v:%v", c.cfg.DaemonCfg.ServerHost,
		c.cfg.DaemonCfg.ServerPort)
	c.client = &ExtendedEthRpc{ethrpc.NewEthRPC(url)}

	c.log.Info("Opening BoltDB database...")
	database, err := db.Open(c.cfg.DataDir, strings.ToLower(string(c.cfg.Asset)))
	if err != nil {
		m.AddError(errToSeverity(ErrOpenDatabase))
		return errors.Errorf("unable to open db: %v", err)
	}
	c.db = database

	c.log.Info("Getting last synced block hash...")
	var lastSyncedBlockHash string
	if c.cfg.LastSyncedBlockHash != "" {
		lastSyncedBlockHash = c.cfg.LastSyncedBlockHash
	} else {
		lastSyncedBlockHash, err = c.fetchLastSyncedBlockHash()
		if err != nil {
			m.AddError(errToSeverity(ErrInitLastSyncedBlock))
			return errors.Errorf("unable to fetch last block synced "+
				"hash: %v", err)
		}
	}
	c.log.Infof("Last synced block hash(%v)", lastSyncedBlockHash)

	defaultAddress, err := c.fetchDefaultAddress()
	if err != nil {
		m.AddError(errToSeverity(ErrGetDefaultAddress))
		return errors.Errorf("unable to fetch default address: %v", err)
	}
	c.log.Infof("Default address %v", defaultAddress)
	c.defaultAddress = defaultAddress

	c.wg.Add(1)
	go func() {
		delay := time.Duration(c.cfg.SyncTickDelay) * time.Second
		syncingTicker := time.NewTicker(delay)

		defer func() {
			c.log.Info("Quit syncing transactions goroutine")
			syncingTicker.Stop()
			c.wg.Done()
		}()

		c.log.Info("Starting syncing goroutine...")

		for {
			select {
			case <-time.After(drainDelay):
				if err := c.drainTransactions(); err != nil {
					m.AddError(errToSeverity(ErrOpenDatabase))
					c.log.Errorf("unable to drain old applied transactions"+
						": %v", err)
				}
			case <-syncingTicker.C:
				var err error
				lastSyncedBlockHash, err = c.sync(lastSyncedBlockHash)
				if err != nil {
					c.log.Errorf("unable to sync: %v", err)
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

	close(c.notifications)
	c.log.Info("client shutdown")
}

func (c *Connector) WaitShutDown() <-chan struct{} {
	return c.quit
}

// AccountAddress return the deposit address of account.
func (c *Connector) AccountAddress(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodAccountAddress, c.cfg.Metrics)
	defer m.Finish()

	var address string
	err := c.db.Update(func(tx *bolt.Tx) error {
		accountsBucket, err := tx.CreateBucketIfNotExists(accountsToAddressesBucket)
		if err != nil {
			m.AddError(errToSeverity(ErrDatabase))
			return errors.Errorf("unable to get accounts bucket: %v", err)
		}

		if data := accountsBucket.Get([]byte(account)); data != nil {
			address = string(data)
		}
		return nil
	})

	return address, err
}

// CreateAddress is used to create deposit address.
func (c *Connector) CreateAddress(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodCreateAddress, c.cfg.Metrics)
	defer m.Finish()

	var address string
	err := c.db.Update(func(tx *bolt.Tx) error {
		accountsBucket, err := tx.CreateBucketIfNotExists(accountsToAddressesBucket)
		if err != nil {
			m.AddError(errToSeverity(ErrDatabase))
			return errors.Errorf("unable to get accounts bucket: %v", err)
		}

		if data := accountsBucket.Get([]byte(account)); data != nil {
			address = string(data)
			return nil
		}

		address, err = c.client.PersonalNewAccount(c.cfg.DaemonCfg.Password)
		if err != nil {
			m.AddError(errToSeverity(ErrCreateAddress))
			return errors.Errorf("unable to create account: %v", err)
		}

		addressesBucket, err := tx.CreateBucketIfNotExists(addressesToAccountsBucket)
		if err != nil {
			m.AddError(errToSeverity(ErrDatabase))
			return errors.Errorf("unable to get addresses bucket: %v", err)
		}

		err = addressesBucket.Put([]byte(address), []byte(account))
		if err != nil {
			m.AddError(errToSeverity(ErrDatabase))
			return errors.Errorf("unable to create address <-> account link"+
				": %v", err)
		}

		err = accountsBucket.Put([]byte(account), []byte(address))
		if err != nil {
			m.AddError(errToSeverity(ErrDatabase))
			return errors.Errorf("unable to create account <-> address link"+
				": %v", err)
		}

		return nil
	})

	return address, err
}

// PendingTransactions returns the transactions with confirmation number lower
// the required by payment system.
func (c *Connector) PendingTransactions(account string) (
	[]*common.BlockchainPendingPayment, error) {

	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingTransactions, c.cfg.Metrics)
	defer m.Finish()

	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	var transactions []*common.BlockchainPendingPayment
	for _, tx := range c.memPoolTxs[account] {
		transactions = append(transactions, tx)
	}

	for _, tx := range c.unconfirmedTxs[account] {
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GenerateTransaction generates raw blockchain transaction.
func (c *Connector) GenerateTransaction(to, amount string) (
	common.GeneratedTransaction, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodGenerateTransaction, c.cfg.Metrics)
	defer m.Finish()

	tx, err := c.generateTransaction(c.defaultAddress, to, amount, false)
	if err != nil {
		m.AddError(errToSeverity(ErrCraftTx))
		return nil, err
	}

	return tx, err
}

func (c *Connector) generateTransaction(from, to, amount string, includeFee bool) (
	common.GeneratedTransaction, error) {

	a, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, errors.Errorf("unable parse amount: %v", err)
	}

	// Fetch suggested by the daemon gas price.
	gp, err := c.client.EthGasPrice()
	if err != nil {
		return nil, err
	}

	gasPrice := big.NewInt(0)
	gasPrice, _ = gasPrice.SetString(gp, 0)

	weiAmount := big.NewInt(0)
	weiAmount.SetString(a.Mul(weiInEth).String(), 0)

	gas := big.NewInt(21000)
	txFee := new(big.Int).Mul(gas, gasPrice)

	// If transaction is redirected to the default account than we should use
	// "send all the available money" model.
	// TODO(andrew.shvv) what if tx fee is greater than tx amount in case of
	// redirection?
	txAmount := weiAmount
	if includeFee {
		txAmount = new(big.Int).Sub(txAmount, txFee)
	}

	_, err = c.client.PersonalUnlockAccount(from, c.cfg.DaemonCfg.Password, 2)
	if err != nil {
		return nil, errors.Errorf("unable to unlock sender account: %v", err)
	}

	tx, rawTx, err := c.client.EthSignTransaction(ethrpc.T{
		From:     from,
		To:       to,
		Gas:      int(gas.Int64()),
		GasPrice: gasPrice,
		Value:    txAmount,
		Data:     "",
		Nonce:    0,
	})
	if err != nil {
		return nil, errors.Errorf("unable to sign tx: %v", err)
	}

	return &GeneratedTransaction{
		rawTx: rawTx,
		hash:  tx.Hash,
	}, nil
}

// SendTransaction sens the given transaction to the blockchain network.
func (c *Connector) SendTransaction(rawTx []byte) error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSendTransaction, c.cfg.Metrics)
	defer m.Finish()

	_, err := c.client.EthSendRawTransaction(string(rawTx))
	if err != nil {
		m.AddError(errToSeverity(ErrSendTx))
		return errors.Errorf("unable to execute send tx rpc call: %v", err)
	}

	return nil
}

// PendingBalance return the amount of funds waiting to be confirmed.
func (c *Connector) PendingBalance(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingBalance, c.cfg.Metrics)
	defer m.Finish()

	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	var amount decimal.Decimal
	for _, tx := range c.memPoolTxs[account] {
		amount = amount.Add(tx.Amount)
	}

	for _, tx := range c.unconfirmedTxs[account] {
		amount = amount.Add(tx.Amount)
	}

	return amount.Round(8).String(), nil
}

// ReceivedPayments returns channel with transactions which
// passed the minimum threshold required by the bitcoin client to treat the
// transactions as confirmed.
func (c *Connector) ReceivedPayments() <-chan []*common.Payment {
	return c.notifications
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
			return nil, errors.Errorf("unable to get last sync block from daemon"+
				": %v", err)
		}

		if err := c.db.Update(func(dbTx *bolt.Tx) error {
			accountsBucket, err := dbTx.CreateBucketIfNotExists(addressesToAccountsBucket)
			if err != nil {
				return errors.Errorf("unable to get addresses bucket: %v", err)
			}

			for _, tx := range block.Transactions {
				data := accountsBucket.Get([]byte(tx.To))
				if data == nil {
					continue
				}

				account := string(data)
				amount := decimal.NewFromBigInt(&tx.Value, 0).Div(weiInEth)

				unconfirmedTxs.add(&common.BlockchainPendingPayment{
					Payment: common.Payment{
						ID:      tx.Hash,
						Amount:  amount,
						Account: account,
						Address: tx.To,
						Type:    common.Blockchain,
					},
					Confirmations:     confirmations,
					ConfirmationsLeft: int64(c.cfg.MinConfirmations) - confirmations,
				})
			}

			return nil
		}); err != nil {
			return nil, err
		}

		lastSyncedBlockNumber = nextBlockNumber
	}
}

// syncPending creates the in-memory map of transactions which
// are in the memory pool of the blockchain daemon.
func (c *Connector) syncPending() (pendingMap, error) {
	mempoolTxs := make(pendingMap)

	txs, err := c.client.EthGetPendingTxs()
	if err != nil {
		return nil, err
	}

	if err := c.db.Update(func(dbTx *bolt.Tx) error {
		addressesBucket, err := dbTx.CreateBucketIfNotExists(addressesToAccountsBucket)
		if err != nil {
			return errors.Errorf("unable to get addresses bucket: %v", err)
		}

		for _, tx := range txs {
			data := addressesBucket.Get([]byte(tx.To))
			if data == nil {
				continue
			}

			account := string(data)
			txValue := decimal.NewFromBigInt(&tx.Value, 0)
			amount := txValue.Div(weiInEth)

			mempoolTxs.add(&common.BlockchainPendingPayment{
				Payment: common.Payment{
					ID:      tx.Hash,
					Amount:  amount,
					Account: account,
					Address: tx.To,
					Type:    common.Blockchain,
				},
				Confirmations:     0,
				ConfirmationsLeft: int64(c.cfg.MinConfirmations),
			})
		}

		return nil
	}); err != nil {
		return nil, err
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

		if err := c.db.Update(func(dbTx *bolt.Tx) error {
			addressesBucket, err := dbTx.CreateBucketIfNotExists(addressesToAccountsBucket)
			if err != nil {
				return errors.Errorf("unable to get addresses bucket: %v", err)
			}

			var payments []*common.Payment
			for _, tx := range block.Transactions {
				data := addressesBucket.Get([]byte(tx.To))
				if data == nil {
					continue
				}
				account := string(data)

				amount := decimal.NewFromBigInt(&tx.Value, 0).Div(weiInEth)

				payments = append(payments, &common.Payment{
					ID:      tx.Hash,
					Amount:  amount,
					Account: account,
					Address: tx.To,
					Type:    common.Blockchain,
				})
			}

			infoBucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
			if err != nil {
				return errors.Errorf("unable to get info bucket: %v", err)
			}

			// Update database with last synced block hash.
			hash := []byte(block.Hash)
			if err := infoBucket.Put(lastSyncedBlockHashKey, hash); err != nil {
				return errors.Errorf("unable to put block hash in db: %v", err)
			}

			txBucket, err := infoBucket.CreateBucketIfNotExists(txBucketKey)
			if err != nil {
				return errors.Errorf("unable to get txs bucket: %v", err)
			}

			// Ensure that we haven't sent transaction already,
			// which might happens if blockchain reorganization took place
			// earlier.
			var filteredPayments []*common.Payment
			for _, payment := range payments {
				// Skip transaction if this is redirection on the default
				// account.
				if payment.Account == "default" {
					c.log.Infof("Redirect payment(%v) has been confirmed",
						payment.ID)
					continue
				}

				paymentKey := []byte(payment.ID)
				if data := txBucket.Get(paymentKey); data == nil {
					filteredPayments = append(filteredPayments, payment)
				} else {
					c.log.Warnf("Transaction(%v) already has been send, "+
						"possibly because of the reorganization, filter it",
						payment.ID)
				}
			}

			if len(filteredPayments) != 0 {
				select {
				case <-c.quit:
					return errors.Errorf("unable send payment notification, "+
						"block(%v): client shutdown", block.Hash)

				case c.notifications <- filteredPayments:
					for _, payment := range filteredPayments {
						index, err := txBucket.NextSequence()
						if err != nil {
							c.log.Errorf("unable to generate next sequence"+
								": %v", payment.ID)
							continue
						}

						var data [8]byte
						binary.BigEndian.PutUint64(data[:], index)
						err = txBucket.Put([]byte(payment.ID), data[:])
						if err != nil {
							c.log.Errorf("unable to put tx to the bucket")
							continue
						}

						// TODO(andrew.shvv) use persistence task queue on
						// case if fails.
						tx, err := c.generateTransaction(payment.Address,
							c.defaultAddress, payment.Amount.String(), true)
						if err != nil {
							c.log.Errorf("unable to generate transfer tx("+
								"%v): %v", payment.ID, err)
							continue
						}

						if err = c.SendTransaction(tx.Bytes()); err != nil {
							c.log.Errorf("unable to send transfer tx(%v): %v",
								payment.ID, err)
							continue
						}

						c.log.Infof("Payment(%v) has been confirmed", payment.ID)
					}
				case <-time.After(time.Second * 5):
					return errors.Errorf("limit of waiting for sending "+
						"notifications were exceeded, block(%v)",
						block.Hash)
				}
			}

			lastSyncedBlock = block
			return nil
		}); err != nil {
			return nil, err
		}

		// After transaction has been consumed by other subsystem
		// overwrite cache.
		c.log.Infof("Process block hash(%v), number(%v)", block.Hash, block.Number)
	}
}

// drainTransactions checks number of stored applied transaction and remove
// the old one if overall number of transaction exceed maximum needed.
func (c *Connector) drainTransactions() error {
	return c.db.Update(func(dbTx *bolt.Tx) error {
		infoBucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return errors.Errorf("unable to get info bucket: %v", err)
		}

		txBucket, err := infoBucket.CreateBucketIfNotExists(txBucketKey)
		if err != nil {
			return errors.Errorf("unable to get txs bucket: %v", err)
		}

		type tx struct {
			index uint64
			id    string
		}

		// Fetch all transaction with their id and sequence index, in order to
		// understand which transaction have been added earlier and which of
		// the have to be removed.
		var txs []tx
		txBucket.ForEach(func(k, v []byte) error {
			if v == nil {
				return nil
			}

			index := binary.BigEndian.Uint64(v)
			txs = append(txs, tx{
				index: index,
				id:    string(k),
			})
			return nil
		})

		if len(txs) < txNumber {
			return nil
		}

		// Sort transaction by sequence number, transaction which has been
		// added earlier will be at the start of slice.
		slice.Sort(txs[:], func(i, j int) bool {
			return txs[i].index < txs[j].index
		})

		transactionToRemove := txs[:txNumber]

		c.log.Infof("Draining %v old applied transactions", len(transactionToRemove))
		for _, tx := range transactionToRemove {
			if err := txBucket.Delete([]byte(tx.id)); err != nil {
				return errors.Errorf("unable to drain transaction(%v): %v",
					tx.id, err)
			}
		}

		return nil
	})
}

// fetchLastSyncedBlockHash returns hash of block which were handled in previous
// cycle of processing.
func (c *Connector) fetchLastSyncedBlockHash() (string, error) {
	var lastHash string
	err := c.db.Update(func(dbTx *bolt.Tx) error {
		bucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return errors.Errorf("unable to get info bucket: %v", err)
		}

		data := bucket.Get(lastSyncedBlockHashKey)
		if data != nil {
			c.log.Info("Restore hash from database...")
			lastHash = string(data)
			return nil
		}

		c.log.Info("Unable to find block in db, fetching best block...")

		bestBlockNumber, err := c.client.EthBlockNumber()
		if err != nil {
			return errors.Errorf("unable to request last best block "+
				"hash: %v", err)
		}

		block, err := c.client.EthGetBlockByNumber(bestBlockNumber, false)
		if err != nil {
			return errors.Errorf("unable to request last best block "+
				"hash: %v", err)
		}

		err = bucket.Put(lastSyncedBlockHashKey, []byte(block.Hash))
		if err != nil {
			return errors.Errorf("unable to put best block in db: %v", err)
		}

		lastHash = string(block.Hash)
		return nil
	})
	if err != nil {
		return "", err
	}

	return lastHash, nil
}

// fetchDefaultAddress...
func (c *Connector) fetchDefaultAddress() (string, error) {
	defaultAddress, err := c.AccountAddress("default")
	if err != nil {
		return "", errors.Errorf("unable to get default address: %v", err)
	}

	if defaultAddress == "" {
		c.log.Info("Unable to find default address in db, generating it...")
		defaultAddress, err = c.CreateAddress("default")
		if err != nil {
			return "", errors.Errorf("unable to generate default address: %v", err)
		}
	}

	return defaultAddress, nil
}

// FundsAvailable returns number of funds available under control of
// connector.
//
// NOTE: Part of the common.Connector interface.
func (c *Connector) FundsAvailable() (decimal.Decimal, error) {
	balance, err := c.client.EthGetBalance(c.defaultAddress, "latest")
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(&balance, 0), nil
}

func (c *Connector) sync(lastSyncedBlockHash string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSync, c.cfg.Metrics)
	defer m.Finish()

	bestBlockNumber, err := c.client.EthBlockNumber()
	if err != nil {
		m.AddError(errToSeverity(ErrGetBlockchainInfo))
		return lastSyncedBlockHash, errors.Errorf("unable to fetch best block number: %v", err)
	}

	lastSyncedBlock, err := c.client.EthGetBlockByHash(lastSyncedBlockHash, false)
	if err != nil {
		// TODO(andrew.shvv) Check reoginizations
		m.AddError(errToSeverity(ErrGetBlockchainInfo))
		return lastSyncedBlockHash, errors.Errorf("unable to get last sync block from daemon: %v", err)
	}

	// Sync block below minimum confirmations threshold and send
	// payment notification.
	lastSyncedBlock, err = c.syncConfirmed(bestBlockNumber, lastSyncedBlock)
	if err != nil {
		m.AddError(errToSeverity(ErrGetBlockchainInfo))
		return lastSyncedBlockHash, errors.Errorf("unable to process blocks: %v", err)
	}
	lastSyncedBlockHash = lastSyncedBlock.Hash

	// Sync block above minimum confirmation threshold and
	// populate unconfirmed pending map with transactions.
	unconfirmedTxs, err := c.syncUnconfirmed(bestBlockNumber,
		lastSyncedBlock.Number)
	if err != nil {
		m.AddError(errToSeverity(ErrSyncUnconfirmed))
		return lastSyncedBlockHash, errors.Errorf("unable to sync unconfirmed txs: %v", err)
	}

	c.pendingLock.Lock()
	c.unconfirmedTxs.merge(unconfirmedTxs,
		func(tx *common.BlockchainPendingPayment) {
			c.log.Infof("Unconfirmed tx(%v) were added, "+
				"account(%v), amount(%v), confirmations(%v), "+
				"left(%v)", tx.ID, tx.Account, tx.Amount,
				tx.Confirmations, tx.ConfirmationsLeft)
		})
	c.pendingLock.Unlock()

	memPoolTxs, err := c.syncPending()
	if err != nil {
		m.AddError(errToSeverity(ErrSyncPending))
		return lastSyncedBlockHash,
			errors.Errorf("unable to fetch mempool txs: %v", err)
	}

	c.pendingLock.Lock()
	c.memPoolTxs.merge(memPoolTxs, func(tx *common.BlockchainPendingPayment) {
		c.log.Infof("Mempool tx(%v) were added, "+
			"account(%v), amount(%v)", tx.ID, tx.Account, tx.Amount)
	})
	c.pendingLock.Unlock()

	balance, err := c.FundsAvailable()
	if err != nil {
		m.AddError(string(metrics.MiddleSeverity))
		return lastSyncedBlockHash, errors.Errorf("unable to "+
			"get available funds: %v", err)
	}

	c.log.Infof("Asset(%v), media(blockchain), available funds(%v)",
		c.cfg.Asset, balance.Round(8).String())

	f, _ := balance.Float64()
	m.CurrentFunds(f)

	return lastSyncedBlockHash, nil
}
