package bitcoind

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AndrewSamokhvalov/slice"
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcd/chaincfg/chainhash"
	"github.com/bitlum/btcd/wire"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/connector/addr"
	"github.com/bitlum/connector/bitcoind/btcjson"
	"github.com/bitlum/connector/bitcoind/rpcclient"
	"github.com/bitlum/connector/chains/net"
	"github.com/bitlum/connector/common"
	core "github.com/bitlum/viabtc_rpc_client"
	"github.com/bitlum/connector/db"
	"github.com/coreos/bbolt"
	"github.com/btcsuite/btclog"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/metrics/crypto"
)

var (
	// infoBucket is the bucket which holds all information which is needed
	// for client to be informed of state of last synchronization.
	infoBucket = []byte("info")

	// txBucket is the bucket which contains the last transaction which has
	// been arrived, so that if accidental temporary fork happened we haven't
	// send notification twice.
	txBucket = []byte("txs")

	// lastSyncedBlockHashKey is the key which corresponds to the last block hash
	// which was handled by the processing handler.
	lastSyncedBlockHashKey = []byte("lastblockhash")

	// txNumber is the number of transactions which we store locally in db
	// because of the possibility of bitcoin blockchain reorganizations,
	// in order to not apply them twice.
	txNumber = 100

	// drainDelay...
	drainDelay = time.Hour

	// allAccounts denotes all accounts in rpc response to bitcoind.
	allAccounts = "*"

	// defaultAccount denotes default account of the bitcoind wallet.
	defaultAccount = ""
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
	User       string
	Password   string
}

// Config is a bitcoind config.
type Config struct {
	// MinConfirmations is a minimum number of confirmations which is needed
	// to treat transaction as confirmed.
	MinConfirmations int

	// SyncLoopDelay is how much processing loop should sleep before
	// trying to update the information about
	SyncLoopDelay int

	// LastSyncedBlockHash is the hash of block which were proceeded last.
	// In this field is specified than, hash will be initialized form it,
	// rather than form database.
	LastSyncedBlockHash string

	// DataDir is the datadir to store bolddb and logs files.
	DataDir string

	// DaemonCfg holds the information about how to connect to the daemon
	// which interact with the payment system network.
	DaemonCfg *DaemonConfig

	// Asset is an asset with which this connector is working.
	Asset core.AssetType

	// FeePerUnit fee per unit, where in bitcoin and litecoin unit is weight,
	// because of the weight, in dash it is byte.
	// TODO(andrew.shvv) Create subsystem to return current fee per unit
	FeePerUnit int

	Logger btclog.Logger

	// Metric is an metrics backend which is used for tracking the metrics of
	// connector.
	Metrics crypto.MetricsBackend
}

func (c *Config) validate() error {
	if c.MinConfirmations <= 0 {
		return errors.New("min confirmations shouldn't be greater than zero")
	}

	if c.DataDir == "" {
		return errors.New("data dir should be specified")
	}

	if c.DaemonCfg == nil {
		return errors.New("daemon config should be specified")
	}

	if c.Logger == nil {
		return errors.New("logger should be specified")
	}

	if c.SyncLoopDelay == 0 {
		c.SyncLoopDelay = 5
	}

	if c.Asset == "" {
		return errors.New("asset should be specified")
	}

	if c.FeePerUnit == 0 {
		return errors.New("fee per unit should be specified")
	}

	if c.Metrics == nil {
		return errors.New("metrics backend should be specified")
	}

	return nil
}

// Connector implements common.BlockchainConnector interface for bitcoind
// client.
type Connector struct {
	started  int32
	shutdown int32
	wg       sync.WaitGroup
	quit     chan struct{}

	cfg    *Config
	client *ExtendedRPCClient
	db     *db.DB

	// pending is a map of blockhain pending payments,
	// which hasn't been confirmed from connector point of view.
	pending map[string][]*common.BlockchainPendingPayment

	notifications chan []*common.Payment

	lastSyncedBlockHash *chainhash.Hash
	netParams           *chaincfg.Params
	log                 *common.NamedLogger

	coinSelectMtx sync.Mutex

	// unspent is used to store btc uxto set locally,
	// in order to craft transactions faster.
	unspent map[string]btcjson.ListUnspentResult

	// unspentSyncMtx is used to lock the utxo local map during is
	// usage/population.
	unspentSyncMtx sync.Mutex
}

// A compile time check to ensure Connector implements the BlockchainConnector
// interface.
var _ common.BlockchainConnector = (*Connector)(nil)

func NewConnector(cfg *Config) (*Connector, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Connector{
		cfg:           cfg,
		notifications: make(chan []*common.Payment),
		quit:          make(chan struct{}),
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

	host := fmt.Sprintf("%v:%v", c.cfg.DaemonCfg.ServerHost,
		c.cfg.DaemonCfg.ServerPort)
	cfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         c.cfg.DaemonCfg.User,
		Pass:         c.cfg.DaemonCfg.Password,
		DisableTLS:   true, // TODO(andrew.shvv) switch on production
		HTTPPostMode: true,
	}

	// Create RPC client in order to talk with cryptocurrency daemon.
	c.log.Info("Creating RPC client...")
	client, err := rpcclient.New(cfg, nil)
	if err != nil {
		m.AddError(errToSeverity(ErrCreateRPCClient))
		return errors.Errorf("unable to create RPC client: %v", err)
	}
	c.client = &ExtendedRPCClient{
		Client: client,
	}

	var chain string
	if c.cfg.Asset == core.AssetDASH {
		resp, err := c.client.GetDashBlockChainInfo()
		if err != nil {
			m.AddError(errToSeverity(ErrGetBlockchainInfo))
			return errors.Errorf("unable to get type of network: %v", err)
		}
		chain = resp.Chain
	} else {
		resp, err := c.client.GetBlockChainInfo()
		if err != nil {
			m.AddError(errToSeverity(ErrGetBlockchainInfo))
			return errors.Errorf("unable to get type of network: %v", err)
		}
		chain = resp.Chain
	}

	c.netParams, err = net.GetParams(string(c.cfg.Asset), chain)
	if err != nil {
		m.AddError(errToSeverity(ErrGetNetParams))
		return errors.Errorf("failed to get net params: %v", err)
	}

	// Create or open database file to host the last state of synchronization.
	c.log.Info("Opening BoltDB database...")

	database, err := db.Open(c.cfg.DataDir, strings.ToLower(string(c.cfg.Asset)))
	if err != nil {
		m.AddError(errToSeverity(ErrOpenDatabase))
		return err
	}
	c.db = database

	// Initialize cache with the last synced block hash.
	c.log.Info("Getting last synced block hash...")
	if c.cfg.LastSyncedBlockHash != "" {
		c.lastSyncedBlockHash, err = chainhash.NewHashFromStr(c.cfg.LastSyncedBlockHash)
		if err != nil {
			m.AddError(errToSeverity(ErrInitLastSyncedBlock))
			return errors.Errorf("unable to decode hash: %v", err)
		}
	} else {
		c.lastSyncedBlockHash, err = c.fetchLastSyncedBlockHash()
		if err != nil {
			m.AddError(errToSeverity(ErrInitLastSyncedBlock))
			return errors.Errorf("unable to fetch last block synced "+
				"hash: %v", err)
		}
	}

	c.log.Infof("Last synced block hash(%v)", c.lastSyncedBlockHash)

	defaultAddress, err := c.fetchDefaultAddress()
	if err != nil {
		m.AddError(errToSeverity(ErrGetDefaultAddress))
		return errors.Errorf("unable to fetch default address: %v", err)
	}
	c.log.Infof("Default address %v", defaultAddress)

	c.wg.Add(1)
	go func() {
		defer func() {
			c.log.Info("Quit syncing blocks goroutine")
			c.wg.Done()
		}()

		c.log.Info("Starting syncing goroutine...")

		syncDelay := time.Second * time.Duration(c.cfg.SyncLoopDelay)
		for {
			select {
			case <-time.After(drainDelay):
				if err := c.drainTransactions(); err != nil {
					m.AddError(errToSeverity(ErrDrainTransactions))
					c.log.Errorf("unable to drain old applied transactions"+
						": %v", err)
				}
			case <-time.After(syncDelay):
				if err := c.sync(); err != nil {
					c.log.Error(err)
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
			c.log.Info("Quit syncing unspent goroutine")
			c.wg.Done()
		}()

		for {

			if err := c.syncUnspent(); err != nil {
				m.AddError(errToSeverity(ErrSyncUnspent))
				c.log.Errorf("unable to main sync unspent: %v", err)
			}

			select {
			case <-time.After(time.Minute):
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
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) AccountAddress(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodAccountAddress, c.cfg.Metrics)
	defer m.Finish()

	addresses, err := c.client.GetAddressesByAccount(account)
	if err != nil {
		m.AddError(errToSeverity(ErrGetAddress))
		return "", err
	}

	if len(addresses) == 0 {
		return "", err
	}
	address := addresses[0].String()

	return address, nil
}

// CreateAddress is used to create deposit address.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) CreateAddress(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodCreateAddress, c.cfg.Metrics)
	defer m.Finish()

	address, err := c.client.GetAccountAddress(account)
	if err != nil {
		m.AddError(errToSeverity(ErrCreateAddress))
		return "", err
	}

	return address.String(), nil
}

// PendingTransactions return the transactions which has confirmation
// number lower the required by payment system.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) PendingTransactions(account string) (
	[]*common.BlockchainPendingPayment, error) {

	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingTransactions, c.cfg.Metrics)
	defer m.Finish()

	transactions := make([]*common.BlockchainPendingPayment, len(c.pending[account]))
	for i, tx := range c.pending[account] {
		transactions[i] = tx
	}

	return transactions, nil
}

// GenerateTransaction generates raw blockchain transaction.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) GenerateTransaction(address string, amount string) (common.GeneratedTransaction, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodGenerateTransaction, c.cfg.Metrics)
	defer m.Finish()

	err := addr.Validate(string(c.cfg.Asset), c.netParams.Name, address)
	if err != nil {
		m.AddError(errToSeverity(ErrValidateAddress))
		return nil, errors.Errorf("invalid address: %v", err)
	}

	decodedAddress, err := btcutil.DecodeAddress(address, c.netParams)
	if err != nil {
		m.AddError(errToSeverity(ErrDecodeAddress))

		return nil, errors.Errorf("unable to decode address: %v", err)
	}

	amt, err := decimal.NewFromString(amount)
	if err != nil {
		m.AddError(errToSeverity(ErrDecodeAmount))
		return nil, errors.Errorf("unable to decode amount: %v", err)
	}

	tx, _, err := c.craftTransaction(uint64(c.cfg.FeePerUnit), decAmount2Sat(amt), decodedAddress)
	if err != nil {
		m.AddError(errToSeverity(ErrCraftTx))
		return nil, errors.Errorf("unable to generate new transaction: %v", err)
	}

	signedTx, isSigned, err := c.client.SignRawTransaction(tx)
	if err != nil {
		m.AddError(errToSeverity(ErrSignTx))
		return nil, errors.Errorf("unable to sign generated transaction: %v", err)
	}

	if !isSigned {
		m.AddError(errToSeverity(ErrSignTx))
		return nil, errors.Errorf("unable to sign all generated transaction inputs: %v", err)
	}

	var rawTx bytes.Buffer
	if err := signedTx.Serialize(&rawTx); err != nil {
		m.AddError(errToSeverity(ErrSerialiseTx))
		return nil, errors.Errorf("unable serialize signed tx: %v", err)
	}

	return &GeneratedTransaction{
		rawTx: rawTx.Bytes(),
		txID:  signedTx.TxHash().String(),
	}, nil
}

// SendTransaction sends the given transaction to the blockchain network.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) SendTransaction(rawTx []byte) error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSendTransaction, c.cfg.Metrics)
	defer m.Finish()

	wireTx := new(wire.MsgTx)
	r := bytes.NewBuffer(rawTx)

	if err := wireTx.Deserialize(r); err != nil {
		m.AddError(errToSeverity(ErrDeserialiseTx))
		return errors.Errorf("unable to deserialize raw tx: %v", err)
	}

	_, err := c.client.SendRawTransaction(wireTx, true)
	if err != nil {
		m.AddError(errToSeverity(ErrSendTx))
		return errors.Errorf("unable to send transaction: %v", err)
	}

	return nil
}

// PendingBalance return the amount of funds waiting ro be confirmed.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) PendingBalance(account string) (string, error) {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodPendingBalance, c.cfg.Metrics)
	defer m.Finish()

	var amount decimal.Decimal
	for _, tx := range c.pending[account] {
		amount = amount.Add(tx.Amount)
	}

	return amount.Round(8).String(), nil
}

// ReceivedPayments returns channel with transactions which
// passed the minimum threshold required by the bitcoin client to treat the
// transactions as confirmed.
//
// NOTE: Part of the common.BlockchainConnector interface.
func (c *Connector) ReceivedPayments() <-chan []*common.Payment {
	return c.notifications
}

// syncUnconfirmed initialize the in-memory map of pending transactions which
// correspond to specific account.
func (c *Connector) syncUnconfirmed() error {
	txs, err := c.client.ListUnspentMinMax(0, int(c.cfg.MinConfirmations-1))
	if err != nil {
		return err
	}

	c.pending = make(map[string][]*common.BlockchainPendingPayment)
	for _, tx := range txs {
		c.pending[tx.Account] = append(c.pending[tx.Account],
			&common.BlockchainPendingPayment{
				Payment: common.Payment{
					ID:      tx.TxID,
					Amount:  decimal.NewFromFloat(tx.Amount),
					Account: tx.Account,
					Address: tx.Address,
					Type:    common.Blockchain,
				},
				Confirmations: tx.Confirmations,
				ConfirmationsLeft: int64(c.cfg.MinConfirmations) - tx.
					Confirmations,
			})

		c.log.Infof("Pending transaction(%v),"+
			"confirmations left(%v), account(%v), amount(%v)", tx.TxID,
			int64(c.cfg.MinConfirmations)-tx.Confirmations, tx.Account,
			tx.Amount)
	}

	return nil
}

// findForkBlock is used to find block on which fork has happened,
// at return it, so that syncing could continue.
func (c *Connector) findForkBlock(orphanBlock *btcjson.GetBlockVerboseResult) (
	*btcjson.GetBlockVerboseResult, error) {
	for orphanBlock.Confirmations == -1 {
		prevHash, err := chainhash.NewHashFromStr(orphanBlock.PreviousHash)
		if err != nil {
			return nil, errors.Errorf("unable to decode hash of prev "+
				"orphan block: %v", err)
		}

		orphanBlock, err = c.client.GetBlockVerbose(prevHash)
		if err != nil {
			return nil, errors.Errorf("unable to prev last sync block "+
				"from daemon: %v", err)

		}
	}

	return orphanBlock, nil
}

// proceedNextBlock process new blocks and notify subscribed clients that
// transaction reached the minimum confirmation limit.
func (c *Connector) proceedNextBlock() error {
	lastSyncedBlock, err := c.client.GetBlockVerbose(c.lastSyncedBlockHash)
	if err != nil {
		return errors.Errorf("unable to get last sync block "+
			"from daemon: %v", err)
	}

	// If bitcoind returns negative confirmation number it means that
	// blockchain re-organization happened and we should handle it properly by
	// moving backwards.
	if lastSyncedBlock.Confirmations < 0 {
		c.log.Info("Chain re-organisation has been found, handle it...")

		forkBlock, err := c.findForkBlock(lastSyncedBlock)
		if err != nil {
			return errors.Errorf("unable to handle "+
				"re-organizations: %v", err)
		}

		c.log.Infof("Fork have been detected on block("+
			"%v) using it as last synced block", forkBlock.Hash)
		lastSyncedBlock = forkBlock
	}

	for {
		select {
		case <-c.quit:
			return nil
		default:
		}

		// We should check next block only if there is minimum amount of
		// confirmation above it.
		if lastSyncedBlock.Confirmations < int64(c.cfg.MinConfirmations)+1 {
			return nil
		}

		// This check is a bit redundant, but we should be ensured that
		// next hash exists, otherwise the last synced hash will be overwritten
		// with zero hash.
		if lastSyncedBlock.NextHash == "" {
			c.log.Errorf("unable to continue processing block(%v):"+
				"next hash empty", lastSyncedBlock.Hash)
			return nil
		}

		nextHash, err := chainhash.NewHashFromStr(lastSyncedBlock.NextHash)
		if err != nil {
			return err
		}

		proceededBlock, err := c.client.GetBlockVerbose(nextHash)
		if err != nil {
			return err
		}

		var payments []*common.Payment
		for _, txHashStr := range proceededBlock.Tx {
			txHash, err := chainhash.NewHashFromStr(txHashStr)
			if err != nil {
				c.log.Errorf("unable to decode tx hash(%v)", txHashStr)
				continue
			}

			// Get transaction and if this transaction not correspons to non
			// of our account the error will be returned, in the case skip
			// this transaction.
			tx, err := c.client.GetTransaction(txHash)
			if err != nil {
				continue
			}

			if len(tx.Details) == 0 {
				c.log.Errorf("unable to sync tx(%v), there is "+
					"no details", tx.TxID)
				continue
			}

			for _, detail := range tx.Details {
				// Skip if details describe receive/send tp/from default
				// account.
				if detail.Account == defaultAccount {
					continue
				}

				// We are interested only in incoming payments.
				if detail.Category != "receive" {
					continue
				}

				payments = append(payments, &common.Payment{
					ID:      tx.TxID,
					Amount:  decimal.NewFromFloat(detail.Amount),
					Account: detail.Account,
					Address: detail.Address,
					Type:    common.Blockchain,
				})
			}
		}

		if err := c.db.Update(func(dbTx *bolt.Tx) error {
			infoBucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
			if err != nil {
				return errors.Errorf("unable to get info bucket: %v", err)
			}

			txBucket, err := infoBucket.CreateBucketIfNotExists(txBucket)
			if err != nil {
				return errors.Errorf("unable to get txs bucket: %v", err)
			}

			// Write down last synced hash in database.
			if err := infoBucket.Put(lastSyncedBlockHashKey, nextHash.CloneBytes()); err != nil {
				return errors.Errorf("unable to put block hash in db: %v", err)
			}

			var paymentToSend []*common.Payment
			for _, payment := range payments {
				if data := txBucket.Get([]byte(payment.ID)); data == nil {
					paymentToSend = append(paymentToSend, payment)
				} else {
					c.log.Warnf("Transaction(%v) already has been send, "+
						"possibly because of the reorganization filter it", payment.ID)
				}
			}

			if len(paymentToSend) != 0 {
				select {
				case <-c.quit:
					return errors.Errorf("unable send payment notification, "+
						"block(%v): client shutdown", proceededBlock.Hash)
				case c.notifications <- paymentToSend:
					for _, payment := range paymentToSend {
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
						}

						c.log.Infof("Payment(%v) has been confirmed", payment.ID)
					}
				case <-time.After(time.Second * 5):
					return errors.Errorf("limit of waiting for sending "+
						"notifications were exceeded, block(%v)",
						proceededBlock.Hash)
				}
			}

			c.lastSyncedBlockHash = nextHash
			lastSyncedBlock = proceededBlock
			return nil
		}); err != nil {
			return err
		}

		// After transaction has been consumed by other subsystem
		// overwrite cache.
		c.log.Infof("Process block hash(%v)", proceededBlock.Hash)
	}
}

// drainTransactions checks number of stored applied transaction and remove
// the old one.
func (c *Connector) drainTransactions() error {
	return c.db.Update(func(dbTx *bolt.Tx) error {
		infoBucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return errors.Errorf("unable to get info bucket: %v", err)
		}

		txBucket, err := infoBucket.CreateBucketIfNotExists(txBucket)
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
func (c *Connector) fetchLastSyncedBlockHash() (*chainhash.Hash, error) {
	var lastHash *chainhash.Hash
	err := c.db.Update(func(dbTx *bolt.Tx) error {
		bucket, err := dbTx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return errors.Errorf("unable to get info bucket: %v", err)
		}

		data := bucket.Get(lastSyncedBlockHashKey)
		if data != nil {
			c.log.Info("Restore hash from database...")

			lastHash, err = chainhash.NewHash(data)
			if err != nil {
				return errors.Errorf("unable initialize hash: %v", err)
			}

			return nil
		}

		c.log.Info("Unable to find block in db, fetching best block...")

		hash, err := c.client.GetBestBlockHash()
		if err != nil {
			return errors.Errorf("unable to request last best block "+
				"hash: %v", err)
		}

		err = bucket.Put(lastSyncedBlockHashKey, hash.CloneBytes())
		if err != nil {
			return errors.Errorf("unable to put best block in db: %v", err)
		}

		lastHash = hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	return lastHash, nil
}

// fetchDefaultAddress...
func (c *Connector) fetchDefaultAddress() (string, error) {
	defaultAddress, err := c.AccountAddress(defaultAccount)
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

func (c *Connector) sync() error {
	m := crypto.NewMetric(c.cfg.DaemonCfg.Name, string(c.cfg.Asset),
		MethodSync, c.cfg.Metrics)
	defer m.Finish()

	if err := c.proceedNextBlock(); err != nil {
		m.AddError(errToSeverity(ErrProceedNextBlock))
		return errors.Errorf("unable to process blocks: %v", err)

	}

	// As far as pending transaction may occur at any time,
	// run it every cycle.
	if err := c.syncUnconfirmed(); err != nil {
		m.AddError(errToSeverity(ErrSync))
		return errors.Errorf("unable to sync unconfirmed txs: %v", err)
	}

	balance, err := c.FundsAvailable()
	if err != nil {
		m.AddError(errToSeverity(ErrSync))
		return errors.Errorf("unable to get available funds: %v", err)
	}

	c.log.Infof("Asset(%v), media(blockchain), available funds(%v)",
		c.cfg.Asset, balance.Round(8).String())

	f, _ := balance.Float64()
	m.CurrentFunds(f)

	return nil
}

// FundsAvailable returns number of funds available under control of
// connector.
//
// NOTE: Part of the common.Connector interface.
func (c *Connector) FundsAvailable() (decimal.Decimal, error) {
	balance, err := c.client.GetBalanceMinConf(allAccounts, c.cfg.MinConfirmations)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.New(int64(balance), 0), nil
}
