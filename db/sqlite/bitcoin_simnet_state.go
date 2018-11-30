package sqlite

import (
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/daemons/bitcoind_simple"
	"github.com/jinzhu/gorm"
	"time"
)

type BitcoinSimpleState struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	Asset     string `gorm:"primary_key"`
	TxCounter int
}

type BitcoinSimpleStateStorage struct {
	db    *DB
	asset connectors.Asset
}

func NewBitcoinSimpleStateStorage(asset connectors.Asset,
	db *DB) *BitcoinSimpleStateStorage {
	return &BitcoinSimpleStateStorage{
		asset: asset,
		db:    db,
	}
}

// Runtime check to ensure that ConnectorStateStorage implements
// bitcoind_simple.StateStorage interface.
var _ bitcoind_simple.StateStorage = (*BitcoinSimpleStateStorage)(nil)

// PutLastSyncedTxCounter is used to save last synchronised confirmed tx
// counter.
//
// NOTE: Part of the bitcoind_simple.StateStorage interface.
func (s *BitcoinSimpleStateStorage) PutLastSyncedTxCounter(counter int) error {
	return s.db.Save(&BitcoinSimpleState{
		Asset:     string(s.asset),
		TxCounter: counter,
	}).Error
}

// LastTxCounter is used to retrieve last synchronised confirmed tx
// counter.
//
// NOTE: Part of the bitcoind_simple.StateStorage interface.
func (s *BitcoinSimpleStateStorage) LastTxCounter() (int, error) {
	state := &BitcoinSimpleState{}
	err := s.db.Where("asset = ?", string(s.asset)).Find(state).Error
	if gorm.IsRecordNotFoundError(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return state.TxCounter, nil
}
