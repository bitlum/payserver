package bitcoind

import (
	"encoding/json"

	"github.com/bitlum/connector/connectors/daemons/bitcoind/btcjson"
	"github.com/bitlum/connector/connectors/daemons/bitcoind/rpcclient"
)

type GetDashBlockChainInfoResult struct {
	*btcjson.GetBlockChainInfoResult

	// Override initial btcjson field with interface in order to avoid json
	// unmarshal error, because in dash the format of this field is different.
	Bip9SoftForks interface{} `json:"bip9_softforks"`
}

type ExtendedRPCClient struct {
	*rpcclient.Client
}

func (c *ExtendedRPCClient) GetDashBlockChainInfo() (*GetDashBlockChainInfoResult,
	error) {
	res := c.GetBlockChainInfoAsync()
	return ReceiveDashInfo(res)
}

func ReceiveDashInfo(r rpcclient.FutureGetBlockChainInfoResult) (
	*GetDashBlockChainInfoResult, error) {

	res, err := rpcclient.ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	var chainInfo GetDashBlockChainInfoResult
	if err := json.Unmarshal(res, &chainInfo); err != nil {
		return nil, err
	}
	return &chainInfo, nil
}
