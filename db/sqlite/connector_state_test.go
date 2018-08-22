package sqlite

import (
	"testing"
	"bytes"
	"github.com/bitlum/connector/connectors"
)

func TestPutLastHash(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	btcStateStorage := NewConnectorStateStorage(connectors.BTC, db)
	err = btcStateStorage.PutLastSyncedHash([]byte("btc_hash"))
	if err != nil {
		t.Fatalf("unable to put hash: %v", err)
	}

	err = btcStateStorage.PutLastSyncedHash([]byte("btc_hash_next"))
	if err != nil {
		t.Fatalf("unable to put hash: %v", err)
	}

	btcHash, err := btcStateStorage.LastSyncedHash()
	if err != nil {
		t.Fatalf("unable to get hash: %v", err)
	}

	if !bytes.Equal(btcHash, []byte("btc_hash_next")) {
		t.Fatalf("wrong hash")
	}

	ethStateStorage := NewConnectorStateStorage(connectors.ETH, db)
	err = ethStateStorage.PutLastSyncedHash([]byte("eth_hash"))
	if err != nil {
		t.Fatalf("unable to put hash: %v", err)
	}

	err = ethStateStorage.PutLastSyncedHash([]byte("eth_hash_next"))
	if err != nil {
		t.Fatalf("unable to put hash: %v", err)
	}

	ethHash, err := ethStateStorage.LastSyncedHash()
	if err != nil {
		t.Fatalf("unable to get hash: %v", err)
	}

	if !bytes.Equal(ethHash, []byte("eth_hash_next")) {
		t.Fatalf("wrong hash")
	}
}
