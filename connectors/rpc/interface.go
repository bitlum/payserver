package rpc

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// Client is an interface which represent source of information from
// blockchain daemon. With this interface we could easily track which
// information is really required by connectors, mock client in
// order to test behaviour of connectors, and also hide bitcoin rpc client
// implementation differentiation.
type Client interface {
	// TODO(andrew.shvv) shouldn't belong here
	DaemonName() string

	// GetBestBlockHash returns the hash of the best block in the longest block
	// chain.
	GetBlockChainInfo() (*BlockChainInfoResp, error)

	// GetBlockVerbose returns a data structure from the server with information
	// about a block given its hash.
	GetBlockVerboseByHash(blockHash *chainhash.Hash) (*BlockVerboseResp, error)

	// GetBestBlockHash returns the hash of the best block in the longest block
	// chain.
	GetBestBlockHash() (*chainhash.Hash, error)

	// UnlockUnspent marks all outputs as unlocked.
	UnlockUnspent() error

	// LockUnspent marks outputs as locked.  When locked, the unspent output
	// will not be selected as input for newly created, non-raw
	// transactions, and will not be returned in future, until the output
	// is marked unlocked again.
	LockUnspent(input UnspentInput) error

	// ListUnspentMinMax returns all unspent transaction outputs known to a
	// wallet, using the specified number of minimum and maximum number of
	// confirmations as a filter.
	ListUnspentMinMax(minConf, maxConf int) ([]UnspentInput, error)

	// GetAddressesByAccount returns the list of addresses associated with the
	// passed label.
	GetAddressesByLabel(label string) ([]btcutil.Address, error)

	// GetNewAddress returns a new address.
	GetNewAddress(label string) (btcutil.Address, error)

	// SignRawTransaction signs inputs for the passed transaction and
	// returns the signed transaction.
	SignRawTransaction(tx *wire.MsgTx) (*wire.MsgTx, error)

	// CreateRawTransaction returns a new transaction spending the provided
	// inputs and sending to the provided addresses.
	CreateRawTransaction([]UnspentInput, map[btcutil.Address]btcutil.Amount) (*wire.MsgTx,
		error)

	// SendRawTransaction submits the encoded transaction to the server which
	// will then relay it to the network.
	SendRawTransaction(tx *wire.MsgTx) error

	// TODO(andrew.shvv) should be renamed after adapting bitcoind v18.0 version
	GetBalanceByLabel(label string, minConfirms int) (btcutil.Amount, error)

	// GetTransaction returns detailed information about a wallet transaction.
	GetTransaction(txHash *chainhash.Hash) (*Transaction, error)

	// EstimateFee estimates the approximate fee per kilobyte needed
	// for a transaction in order to be included in the block.
	EstimateFee() (float64, error)
}
type BlockChainInfoResp struct {
	Chain string
}

type BlockVerboseResp struct {
	Hash          string
	Height        int64
	NextHash      string
	Confirmations int64
	PreviousHash  string
	Tx            []string
}

type UnspentInput struct {
	Address       string
	Account       string
	Amount        float64
	Confirmations int64
	TxID          string
	Vout          uint32
}

type Transaction struct {
	Amount        float64
	Fee           float64
	Confirmations int64
	TxID          string
	Details       []TransactionDetails
}

type TransactionDetails struct {
	Account           string
	Address           string
	Amount            float64
	Category          string
	InvolvesWatchOnly bool
	Fee               *float64
	Vout              uint32
}
