package sqlite

import (
	"testing"
	"github.com/bitlum/connector/connectors"
)

func TestAddPaymentStatusMigration(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable create test db: %v", err)
	}
	defer clear()

	paymentsStore := PaymentsStore{DB: db}

	if err := paymentsStore.SavePayment(&connectors.Payment{
		Direction: "internal",
	}); err != nil {
		t.Fatalf("unable save payment: %v", err)
	}

	if err := paymentsStore.SavePayment(&connectors.Payment{
		Direction: "internal",
	}); err != nil {
		t.Fatalf("unable save payment: %v", err)
	}

}