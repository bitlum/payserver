package sqlite

import (
	"github.com/bitlum/connector/connectors"
	"time"
)

type ConnectorState struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	Asset    string `gorm:"primary_key"`
	LastHash string
}

type ConnectorStateStorage struct {
	db    *DB
	asset connectors.Asset
}

func NewConnectorStateStorage(asset connectors.Asset,
	db *DB) *ConnectorStateStorage {
	return &ConnectorStateStorage{
		asset: asset,
		db:    db,
	}
}

// Runtime check to ensure that ConnectorStateStorage implements
// connector.StateStorage interface.
var _ connectors.StateStorage = (*ConnectorStateStorage)(nil)

// PutLastSyncedHash is used to save last synchronised block hash.
//
// NOTE: Part of the bitcoind.Storage interface.
func (s *ConnectorStateStorage) PutLastSyncedHash(hash []byte) error {
	return s.db.Save(&ConnectorState{
		Asset:    string(s.asset),
		LastHash: string(hash),
	}).Error
}

// LastSyncedHash is used to retrieve last synchronised block hash.
//
// NOTE: Part of the bitcoind.Storage interface.
func (s *ConnectorStateStorage) LastSyncedHash() ([]byte, error) {
	state := &ConnectorState{}
	if err := s.db.Where("asset = ?", string(s.asset)).
		Find(state).Error; err != nil {
		return nil, err
	}

	return []byte(state.LastHash), nil
}
