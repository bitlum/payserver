package sqlite

import (
	"gopkg.in/gormigrate.v1"
	"testing"
)

func TestAddPaymentStatusMigration(t *testing.T) {
	db, err := Open("./", "test_db_add_status_field")
	if err != nil {
		t.Fatalf("unable create test db: %v", err)
	}

	tx := db.Begin()
	defer tx.Rollback()

	err = migrate(tx, []*gormigrate.Migration{addPaymentSystemType})
	if err != nil {
		t.Fatalf("unable migrate db: %v", err)
	}
}
