package litecoin

import (
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/bitlum/connector/connectors/rpc"
)

type ClientConfig bitcoin.ClientConfig

// Client is identical to bitcoin implementation.
type Client struct {
	*bitcoin.Client
}

// Runtime check to ensure that Client implements rpc.Client interface.
var _ rpc.Client = (*Client)(nil)

func NewClient(cfg ClientConfig) (*Client, error) {
	client, err := bitcoin.NewClient(bitcoin.ClientConfig(cfg))
	return &Client{Client: client}, err
}
