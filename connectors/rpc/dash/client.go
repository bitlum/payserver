package dash

import (
	"encoding/json"
	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/bitlum/go-bitcoind-rpc/btcjson"
	"github.com/bitlum/go-bitcoind-rpc/rpcclient"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
)

type ClientConfig bitcoin.ClientConfig
type Client struct {
	*bitcoin.Client
}

// Runtime check to ensure that Client implements rpc.Client interface.
var _ rpc.Client = (*Client)(nil)

func NewClient(cfg ClientConfig) (*Client, error) {
	client, err := bitcoin.NewClient(bitcoin.ClientConfig(cfg))
	return &Client{Client: client}, err
}

type getDashBlockChainInfoResult struct {
	*btcjson.GetBlockChainInfoResult

	// Override initial btcjson field with interface in order to avoid json
	// unmarshal error, because in dash the format of this field is different.
	Bip9SoftForks interface{} `json:"bip9_softforks"`
}

func (c *Client) GetBlockChainInfo() (*rpc.BlockChainInfoResp, error) {
	res := c.Daemon.GetBlockChainInfoAsync()
	info, err := receiveDashInfo(res)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", err)
		return nil, err
	}

	resp := &rpc.BlockChainInfoResp{
		Chain: info.Chain,
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

func receiveDashInfo(r rpcclient.FutureGetBlockChainInfoResult) (
	*getDashBlockChainInfoResult, error) {

	res, err := rpcclient.ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	var chainInfo getDashBlockChainInfoResult
	if err := json.Unmarshal(res, &chainInfo); err != nil {
		return nil, err
	}
	return &chainInfo, nil
}

func (c *Client) EstimateFee() (float64, error) {
	confTarget := uint32(2)
	res, err := c.Daemon.EstimateSmartFeeWithMode(confTarget, "")
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