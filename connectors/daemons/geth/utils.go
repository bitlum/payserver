package geth

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/bitlum/connector/connectors"
	"github.com/go-errors/errors"
)

// pendingMap stores the information about pending transactions corresponding
// to accounts.
type pendingMap map[string]map[string]*connectors.Payment

func (m pendingMap) add(tx *connectors.Payment) {
	if _, ok := m[tx.Account]; !ok {
		m[tx.Account] = make(map[string]*connectors.Payment)
	}

	m[tx.Account][tx.PaymentID] = tx
}

// merge merges two pending maps and invokes handler with new entry populated
// as an argument.
func (m pendingMap) merge(m2 pendingMap,
	newEntryHandler func(tx *connectors.Payment)) {
	for account, txs := range m2 {
		// Add all txs if there is no transaction for this
		// account and continue.
		if _, ok := m[account]; !ok {
			m[account] = txs

			for _, tx := range txs {
				newEntryHandler(tx)
			}

			continue
		}

		// If account exist that we should populate it
		// with transactions which aren't there yet.
		for txid, tx := range txs {
			if _, ok := m[account][txid]; !ok {
				m[account][txid] = tx
				newEntryHandler(tx)
			}
		}
	}

	for account, txs := range m {
		if _, ok := m2[account]; !ok {
			delete(m, account)
			continue
		}

		for txid, _ := range txs {
			if _, ok := m[account][txid]; !ok {
				delete(m[account], txid)
			}
		}
	}
}

func convertVersion(actualNet string) string {
	net := "simnet"

	switch actualNet {
	case params.RinkebyChainConfig.ChainID.String():
		net = "testnet"
	case params.MainnetChainConfig.ChainID.String():
		net = "mainnet"
	}

	return net
}

func checkAlias(accountAlias string) error {
	switch accountAlias {
	case "all", "default", "*":
		return errors.Errorf("name of account '%v' is reserved for internal"+
			" usage", accountAlias)
	default:
		return nil
	}
}

// generatePaymentID generates unique string based on the tx id and receive
// address, which are together
func generatePaymentID(txID, receiveAddress string, direction connectors.PaymentDirection) string {
	return connectors.GeneratePaymentID(txID, receiveAddress, string(direction))
}
