package bitcoincash

import (
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/go-errors/errors"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/davecgh/go-spew/spew"
	"github.com/bitlum/connector/common"
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

// NOTE: Part of the rpc.Client interface.
func (c *Client) EstimateFee() (float64, error) {
	// Bitcoin Cash has removed estimatesmartfee in 17.2 version of their
	// client.
	res, err := c.Client.Daemon.EstimateFee(2)
	if err != nil {
		c.Logger.Tracef("method: %v, error: %v", err)
		return 0, err
	}

	if res == nil {
		err := errors.Errorf("result is nil")
		c.Logger.Tracef("method: %v, error: %v", err)
		return 0, err
	}

	feeRate := **res
	if feeRate <= 0 {
		err := errors.New("not enough data to make an estimation")
		c.Logger.Tracef("method: %v, error: %v", err)
		return 0, err
	}

	c.Logger.Tracef("method: %v, response: %v", common.GetFunctionName(),
		spew.Sdump(feeRate))

	return feeRate, nil
}
