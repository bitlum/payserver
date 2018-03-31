package geth

import (
	"github.com/bitlum/connector/common"
)

// pendingMap stores the information about pending transactions corresponding
// to accounts.
type pendingMap map[string]map[string]*common.BlockchainPendingPayment

func (m pendingMap) add(tx *common.BlockchainPendingPayment) {
	if _, ok := m[tx.Account]; !ok {
		m[tx.Account] = make(map[string]*common.BlockchainPendingPayment)
	}

	m[tx.Account][tx.ID] = tx
}

func (m pendingMap) merge(m2 pendingMap,
	f func(tx *common.BlockchainPendingPayment)) {
	for account, txs := range m2 {
		// Add all txs if there is no transaction for this
		// account and continue.
		if _, ok := m[account]; !ok {
			m[account] = txs

			for _, tx := range txs {
				f(tx)
			}

			continue
		}

		// If account exist that we should populate it
		// with transactions
		// which aren't there yet.
		for txid, tx := range txs {
			if _, ok := m[account][txid]; !ok {
				m[account][txid] = tx
				f(tx)
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

type GeneratedTransaction struct {
	rawTx string
	hash  string
}

func (t *GeneratedTransaction) ID() string {
	return t.hash
}

func (t *GeneratedTransaction) Bytes() []byte {
	return []byte(t.rawTx)
}

var _ common.GeneratedTransaction = (*GeneratedTransaction)(nil)
