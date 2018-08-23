package connectors

import (
	"testing"
	"bytes"
	"reflect"
)

func TestBlockchainPendingDetailsEncodeDecode(t *testing.T) {
	d := &BlockchainPendingDetails{
		Confirmations:     1,
		ConfirmationsLeft: 3,
	}

	var b bytes.Buffer
	if err := d.Encode(&b, 0); err != nil {
		t.Fatalf("unable to encode details: %v", err)
	}

	d1 := &BlockchainPendingDetails{}
	if err := d1.Decode(&b, 0); err != nil {
		t.Fatalf("unable to decode details: %v", err)
	}

	if !reflect.DeepEqual(d1, d) {
		t.Fatal("objects are different")
	}
}

func TestGeneratedTxDetailsEncodeDecode(t *testing.T) {
	d := &GeneratedTxDetails{
		RawTx: []byte("rawtx"),
		TxID:  "txid",
	}

	var b bytes.Buffer
	if err := d.Encode(&b, 0); err != nil {
		t.Fatalf("unable to encode details: %v", err)
	}

	d1 := &GeneratedTxDetails{}
	if err := d1.Decode(&b, 0); err != nil {
		t.Fatalf("unable to encode details: %v", err)
	}

	if !reflect.DeepEqual(d1, d) {
		t.Fatal("objects are different")
	}
}
