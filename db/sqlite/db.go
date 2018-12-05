package sqlite

import (
	"github.com/bitlum/graphql-go/errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"os"
	"path/filepath"
	"sync"
)

// DB is the primary datastore.
type DB struct {
	*gorm.DB
	dbPath      string
	globalMutex sync.Mutex
}

// Open opens an existing db. Any necessary schemas migrations due to
// updates will take place as necessary.
func Open(dbPath string, dbName string, shouldMigrate bool) (*DB, error) {
	path := filepath.Join(dbPath, dbName)

	if !fileExists(dbPath) {
		if err := os.MkdirAll(dbPath, 0700); err != nil {
			return nil, err
		}
	}

	gdb, err := gorm.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:     gdb,
		dbPath: dbPath,
	}

	if shouldMigrate {
		if err := db.Migrate(); err != nil {
			return nil, errors.Errorf("unable to migrate: %v", err)
		}
	}

	return db, nil
}

// fileExists returns true if the file exists, and false otherwise.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}
