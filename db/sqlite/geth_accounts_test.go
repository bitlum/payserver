package sqlite

import (
	"testing"
	"github.com/davecgh/go-spew/spew"
)

func TestAccountsStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	storage := NewGethAccountsStorage(db)

	err = storage.AddAddressToAccount("address1", "account1")
	if err != nil {
		t.Fatalf("unable to add address: %v", err)
	}

	err = storage.AddAddressToAccount("address2", "account2")
	if err != nil {
		t.Fatalf("unable to add address: %v", err)
	}

	account, err := storage.GetAccountByAddress("address1")
	if err != nil {
		t.Fatalf("unable to get account: %v", err)
	}

	if account != "account1" {
		t.Fatalf("wrong account")
	}

	address, err := storage.GetLastAccountAddress("account1")
	if err != nil {
		t.Fatalf("unable to get address: %v", err)
	}

	if address != "address1" {
		t.Fatalf("wrong address")
	}

	addresses, err := storage.GetAddressesByAccount("account2")
	if err != nil {
		t.Fatalf("unable to get addresses: %v", err)
	}

	if len(addresses) != 1 {
		t.Fatalf("wrong len")
	}

	if addresses[0] != "address2" {
		t.Fatalf("wrong address")
	}

	allAddresses, err := storage.AllAddresses()
	if err != nil {
		t.Fatalf("unable to get all addreses: %v", err)
	}

	if len(allAddresses) != 2 {
		t.Fatalf("wrong len")
	}

	emptyAddress, err := storage.GetAccountByAddress("")
	if err != nil {
		t.Fatalf("unable to get address: %v", err)
	}

	spew.Dump(emptyAddress)
	if emptyAddress != "" {
		t.Fatalf("wrong address")
	}
}
