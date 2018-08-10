package geth

import (
	"encoding/json"

	"github.com/onrik/ethrpc"
)

type ExtendedEthRpc struct {
	*ethrpc.EthRPC
}

func (c *ExtendedEthRpc) PersonalNewAccount(pass string) (string, error) {
	var res string
	err := c.call("personal_newAccount", &res, pass)
	return res, err
}

func (c *ExtendedEthRpc) PersonalUnlockAccount(address,
	pass string, delay int) (bool, error) {
	var res bool
	err := c.call("personal_unlockAccount", &res, address, pass, delay)
	return res, err
}

func (c *ExtendedEthRpc) EthSignTransaction(t ethrpc.T) (*ethrpc.Transaction,
	string, error) {
	var resp struct {
		Tx  *ethrpc.Transaction
		Raw string
	}
	err := c.call("eth_signTransaction", &resp, t)
	if err != nil {
		return nil, "", err
	}

	return resp.Tx, resp.Raw, err
}

// EthGetPendingTxs returns transactions from daemon mempool which are
// belongs to one of our accounts.
//
// NOTE: Return both incoming pending transaction because we use
// changed version of the geth client.
func (c *ExtendedEthRpc) EthGetPendingTxs() ([]ethrpc.Transaction, error) {
	var txs []ethrpc.Transaction
	err := c.call("eth_pendingTransactions", &txs)
	return txs, err
}

// EthGasPrice returns the current price per gas in wei.
func (c *ExtendedEthRpc) EthGasPrice() (string, error) {
	var response string
	if err := c.call("eth_gasPrice", &response); err != nil {
		return "", err
	}

	return response, nil
}

func (c *ExtendedEthRpc) call(method string, target interface{},
	params ...interface{}) error {
	result, err := c.Call(method, params...)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	if err := json.Unmarshal(result, target); err != nil {
		return err
	}

	return nil
}
