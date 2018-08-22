package sqlite

import (
	"io/ioutil"
	"os"
)

// MakeTestDB creates a new instance of the ChannelDB for testing purposes. A
// callback which cleans up the created temporary directories is also returned
// and intended to be executed after the test completes.
func MakeTestDB() (*DB, func(), error) {
	// First, create a temporary directory to be used for the duration of
	// this test.
	tempDirName, err := ioutil.TempDir("", "db")
	if err != nil {
		return nil, nil, err
	}

	db, err := Open(tempDirName, "sqlite.db")
	if err != nil {
		return nil, nil, err
	}

	cleanUp := func() {
		db.Close()
		os.RemoveAll(tempDirName)
	}

	return db, cleanUp, nil
}

