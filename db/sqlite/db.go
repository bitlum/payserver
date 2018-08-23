package sqlite

import (
	"github.com/jinzhu/gorm"
	"path/filepath"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// DB is the primary datastore.
type DB struct {
	*gorm.DB
	dbPath string
}

// Open opens an existing db. Any necessary schemas migrations due to
// updates will take place as necessary.
func Open(dbPath string, dbName string) (*DB, error) {
	path := filepath.Join(dbPath, dbName)

	gdb, err := gorm.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = gdb.AutoMigrate(
		&ConnectorState{},
		&EthereumAddress{},
		&Payment{},
	).Error
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:     gdb,
		dbPath: dbPath,
	}

	return db, nil
}
