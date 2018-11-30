package bitcoind_simple

// StateStorage is used to keep data which is needed for connector to
// properly synchronise and track transactions.
//
// NOTE: This storage should be persistent.
type StateStorage interface {
	// PutLastSyncedTxCounter is used to save last synchronised confirmed tx
	// counter.
	PutLastSyncedTxCounter(counter int) error

	// LastTxCounter is used to retrieve last synchronised confirmed tx
	// counter.
	LastTxCounter() (int, error)
}
