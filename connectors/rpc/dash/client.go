package dash

import (
	"github.com/bitlum/connector/connectors/rpc/bitcoincash"
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/bitlum/go-bitcoind-rpc/btcjson"
	"github.com/bitlum/go-bitcoind-rpc/rpcclient"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/bitlum/connector/common"
)

type ClientConfig bitcoin.ClientConfig
type Client struct {
	*bitcoincash.Client
}

// Runtime check to ensure that Client implements rpc.Client interface.
var _ rpc.Client = (*Client)(nil)

func NewClient(cfg ClientConfig) (*Client, error) {
	client, err := bitcoincash.NewClient(bitcoincash.ClientConfig(cfg))
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
