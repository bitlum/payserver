package bitcoind

import (
	"github.com/bitlum/connector/metrics"
)

const (
	ErrCreateRPCClient     = iota
	ErrGetBlockchainInfo
	ErrGetNetParams
	ErrValidateAddress
	ErrOpenDatabase
	ErrInitLastSyncedBlock
	ErrGetDefaultAddress
	ErrDrainTransactions
	ErrProceedNextBlock
	ErrSync
	ErrSyncUnspent
	ErrGetAddress
	ErrCreateAddress
	ErrDecodeAddress
	ErrDecodeAmount
	ErrCraftTx
	ErrSignTx
	ErrSerialiseTx
	ErrDeserialiseTx
	ErrSendTx
)


var errToSeverityMap = map[int]metrics.Severity{
	ErrCreateRPCClient:     metrics.HighSeverity,
	ErrGetBlockchainInfo:   metrics.HighSeverity,
	ErrGetNetParams:        metrics.HighSeverity,
	ErrValidateAddress:     metrics.LowSeverity,
	ErrOpenDatabase:        metrics.HighSeverity,
	ErrInitLastSyncedBlock: metrics.HighSeverity,
	ErrGetDefaultAddress:   metrics.HighSeverity,
	ErrDrainTransactions:   metrics.MiddleSeverity,
	ErrProceedNextBlock:    metrics.MiddleSeverity,
	ErrSync:                metrics.MiddleSeverity,
	ErrSyncUnspent:         metrics.MiddleSeverity,
	ErrGetAddress:          metrics.MiddleSeverity,
	ErrCreateAddress:       metrics.HighSeverity,
	ErrDecodeAddress:       metrics.HighSeverity,
	ErrSerialiseTx:         metrics.HighSeverity,
	ErrDecodeAmount:        metrics.LowSeverity,
	ErrCraftTx:             metrics.HighSeverity,
	ErrSignTx:              metrics.HighSeverity,
	ErrDeserialiseTx:       metrics.HighSeverity,
	ErrSendTx:              metrics.HighSeverity,
}

func errToSeverity(err int) string {
	severity := metrics.LowSeverity

	if s, ok := errToSeverityMap[err]; ok {
		severity = s
	}

	return string(severity)
}
