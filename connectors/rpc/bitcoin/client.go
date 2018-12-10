package bitcoin

import (
	"fmt"
	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/bitlum/go-bitcoind-rpc/btcjson"
	"github.com/bitlum/go-bitcoind-rpc/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btclog"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
)

type ClientConfig struct {
	Logger   btclog.Logger
	Asset    connectors.Asset
	Name     string
	RPCHost  string
	RPCPort  int
	User     string
	Password string
}

// Client bitcoind implementation of rpc.Client interface.
type Client struct {
	Daemon     *rpcclient.Client
	Logger     common.NamedLogger
	daemonName string
}

// Runtime check to ensure that Client implements rpc.Client interface.
var _ rpc.Client = (*Client)(nil)

func NewClient(cfg ClientConfig) (*Client, error) {
	host := fmt.Sprintf("%v:%v", cfg.RPCHost, cfg.RPCPort)

	rpcCfg := &rpcclient.ConnConfig{
		Host: host,
		User: cfg.User,
		Pass: cfg.Password,
		// TODO(andrew.shvv) switch on production
		DisableTLS:   true,
		HTTPPostMode: true,
	}

	// Create RPC client in order to talk with cryptocurrency Daemon.
	rpcClient, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, errors.Errorf("unable to create RPC client: %v", err)
	}

	return &Client{
		Daemon:     rpcClient,
		daemonName: cfg.Name,
		Logger: common.NamedLogger{
			Logger: cfg.Logger,
			Name:   string(cfg.Asset),
		},
	}, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetBlockChainInfo() (*rpc.BlockChainInfoResp, error) {
	daemonResp, err := c.Daemon.GetBlockChainInfo()
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	resp := &rpc.BlockChainInfoResp{
		Chain: daemonResp.Chain,
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetBlockVerboseByHash(blockHash *chainhash.Hash) (
	*rpc.BlockVerboseResp, error) {

	daemonResp, err := c.Daemon.GetBlockVerbose(blockHash)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	resp := &rpc.BlockVerboseResp{
		Hash:          daemonResp.Hash,
		Height:        daemonResp.Height,
		NextHash:      daemonResp.NextHash,
		Confirmations: daemonResp.Confirmations,
		PreviousHash:  daemonResp.PreviousHash,
		Tx:            daemonResp.Tx,
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetBestBlockHash() (*chainhash.Hash, error) {
	resp, err := c.Daemon.GetBestBlockHash()
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, err
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) UnlockUnspent() error {
	err := c.Daemon.LockUnspent(true, nil)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		"empty")

	return nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) LockUnspent(input rpc.UnspentInput) error {
	hash, err := chainhash.NewHashFromStr(input.TxID)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return err
	}

	outputs := []*wire.OutPoint{{Hash: *hash, Index: input.Vout}}

	if err := c.Daemon.LockUnspent(false, outputs); err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(),
			err)
		return err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		"empty")

	return nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) ListUnspentMinMax(minConf, maxConf int) ([]rpc.UnspentInput,
	error) {

	unspent, err := c.Daemon.ListUnspentMinMax(minConf, maxConf)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	resp := make([]rpc.UnspentInput, 0)
	for _, u := range unspent {
		resp = append(resp, rpc.UnspentInput{
			Address:       u.Address,
			Account:       u.Account,
			Amount:        u.Amount,
			Confirmations: u.Confirmations,
			TxID:          u.TxID,
			Vout:          u.Vout,
		})
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetAddressesByLabel(label string) ([]btcutil.Address, error) {
	addresses, err := c.Daemon.GetAddressesByAccount(label)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(addresses))

	return addresses, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetNewAddress(label string) (btcutil.Address, error) {
	address, err := c.Daemon.GetNewAddress(label)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(address))

	return address, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetNewRawChangeAddress(label string) (btcutil.Address, error) {
	address, err := c.Daemon.GetRawChangeAddress(label)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(address))

	return address, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) SignRawTransaction(tx *wire.MsgTx) (*wire.MsgTx, error) {
	signedTx, isSigned, err := c.Daemon.SignRawTransaction(tx)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	if !isSigned {
		err := errors.Errorf("unable to sign all inputs")
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(signedTx))

	return signedTx, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) CreateRawTransaction(inputs []rpc.UnspentInput,
	outputs map[btcutil.Address]btcutil.Amount) (*wire.MsgTx, error) {
	lockTime := int64(0)

	txInputs := make([]btcjson.TransactionInput, len(inputs))
	for i, input := range inputs {
		txInputs[i] = btcjson.TransactionInput{
			Txid: input.TxID,
			Vout: input.Vout,
		}
	}

	tx, err := c.Daemon.CreateRawTransaction(txInputs, outputs, &lockTime)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(tx))

	return tx, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) SendRawTransaction(tx *wire.MsgTx) error {
	_, err := c.Daemon.SendRawTransaction(tx, false)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		"empty")

	return nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetBalanceByLabel(label string,
	minConfirms int) (btcutil.Amount, error) {

	amount, err := c.Daemon.GetBalanceMinConf(label, minConfirms)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(amount))

	return amount, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetTransaction(txHash *chainhash.Hash) (
	*rpc.Transaction, error) {

	tx, err := c.Daemon.GetTransaction(txHash)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return nil, err
	}

	details := make([]rpc.TransactionDetails, len(tx.Details))
	for i, detail := range tx.Details {
		details[i] = rpc.TransactionDetails{
			Account:           detail.Account,
			Address:           detail.Address,
			Amount:            detail.Amount,
			Category:          detail.Category,
			InvolvesWatchOnly: detail.InvolvesWatchOnly,
			Fee:               detail.Fee,
			Vout:              detail.Vout,
		}
	}

	resp := &rpc.Transaction{
		Amount:        tx.Amount,
		Fee:           tx.Fee,
		Confirmations: tx.Confirmations,
		TxID:          tx.TxID,
		Details:       details,
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) EstimateFee() (float64, error) {
	confTarget := uint32(2)
	res, err := c.Daemon.EstimateSmartFeeWithMode(confTarget,
		btcjson.ConservativeEstimateMode)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	if res.Errors != nil {
		err := errors.New((*res.Errors)[0])
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	if res.FeeRate == nil {
		err := errors.Errorf("fee rate is nil")
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	if res.Blocks != int(confTarget) {
		err := errors.New("not enough data to make an estimation")
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	feeRate := *res.FeeRate
	if feeRate <= 0 {
		err := errors.New("not enough data to make an estimation")
		c.Logger.Tracef("method: %v, error: %v", common.GetFunctionName(), err)
		return 0, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(feeRate))

	return feeRate, nil
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) DaemonName() string {
	return c.daemonName
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) SendToAddress(address btcutil.Address,
	amount btcutil.Amount) (*chainhash.Hash, error) {
	return c.Daemon.SendToAddress(address, amount)
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) ListTransactionByLabel(label string, count, from int) (
	[]btcjson.ListTransactionsResult, error) {
	return c.Daemon.ListTransactionsCountFrom(label, count, from)
}

// NOTE: Part of the rpc.Client interface. For more info look in
// the interface description.
func (c *Client) GetTransactionByHash(hash *chainhash.Hash) (
	*rpc.Transaction, error) {
	tx, err := c.Daemon.GetTransaction(hash)
	if err != nil {
		return nil, err
	}

	details := make([]rpc.TransactionDetails, len(tx.Details))
	for i, detail := range tx.Details {
		details[i] = rpc.TransactionDetails{
			Account:           detail.Account,
			Address:           detail.Address,
			Amount:            detail.Amount,
			Category:          detail.Category,
			InvolvesWatchOnly: detail.InvolvesWatchOnly,
			Fee:               detail.Fee,
			Vout:              detail.Vout,
		}
	}

	return &rpc.Transaction{
		Amount:        tx.Amount,
		Fee:           tx.Fee,
		Confirmations: tx.Confirmations,
		TxID:          tx.TxID,
		Details:       details,
	}, nil
}
