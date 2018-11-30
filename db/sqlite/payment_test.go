package sqlite

import (
	"github.com/bitlum/connector/connectors"
	"github.com/shopspring/decimal"
	"reflect"
	"testing"
)

func TestPaymentsStorage(t *testing.T) {
	db, clear, err := MakeTestDB()
	if err != nil {
		t.Fatalf("unable to create test database: %v", err)
	}
	defer clear()

	store := PaymentsStore{DB: db}

	paymentsBefore := []*connectors.Payment{
		{

			PaymentID: "1",
			UpdatedAt: 1,
			Status:    connectors.Pending,
			System:    connectors.Internal,
			Direction: connectors.Outgoing,
			Receipt:   "receipt",
			Asset:     connectors.BTC,
			Account:   "account",
			Media:     connectors.Blockchain,
			Amount:    decimal.NewFromFloat(1.1),
			MediaFee:  decimal.NewFromFloat(1.1),
			MediaID:   "media_id",
			Detail: &connectors.BlockchainPendingDetails{
				ConfirmationsLeft: 3,
				Confirmations:     1,
			},
		},
		{

			PaymentID: "2",
			UpdatedAt: 2,
			Status:    connectors.Completed,
			Direction: connectors.Incoming,
			System:    connectors.External,
			Receipt:   "receipt",
			Asset:     connectors.ETH,
			Account:   "account",
			Media:     connectors.Lightning,
			Amount:    decimal.NewFromFloat(1.1),
			MediaFee:  decimal.NewFromFloat(1.1),
			MediaID:   "media_id",
			Detail: &connectors.GeneratedTxDetails{
				RawTx: []byte("rawtx"),
				TxID:  "123",
			},
		},
	}

	if err := store.SavePayment(paymentsBefore[0]); err != nil {
		t.Fatalf("unable to save payment: %v", err)
	}

	if err := store.SavePayment(paymentsBefore[1]); err != nil {
		t.Fatalf("unable to save payment: %v", err)
	}

	paymentsAfter, err := store.ListPayments("", "", "", "", "")
	if err != nil {
		t.Fatalf("unable to get payments: %v", err)
	}

	if !reflect.DeepEqual(paymentsBefore, paymentsAfter) {
		t.Fatalf("wrong data")
	}

	{
		payment, err := store.PaymentByID("1")
		if err != nil {
			t.Fatalf("unable to get payment by receipt: %v", err)
		}

		if !reflect.DeepEqual(payment, paymentsAfter[0]) {
			t.Fatalf("wrong data")
		}
	}

	{
		payments, err := store.PaymentByReceipt("receipt")
		if err != nil {
			t.Fatalf("unable to get payment by receipt: %v", err)
		}

		if !reflect.DeepEqual(payments, paymentsAfter) {
			t.Fatalf("wrong data")
		}
	}

	{
		payments, err := store.ListPayments(connectors.ETH, "", "", "", "")
		if err != nil {
			t.Fatalf("unable to list payments: %v", err)
		}

		if !reflect.DeepEqual(payments[0], paymentsAfter[1]) {
			t.Fatalf("wrong data")
		}
	}

	{
		payments, err := store.ListPayments("", connectors.Completed, "", "", "")
		if err != nil {
			t.Fatalf("unable to list payments: %v", err)
		}

		if !reflect.DeepEqual(payments[0], paymentsAfter[1]) {
			t.Fatalf("wrong data")
		}
	}

	{
		payments, err := store.ListPayments("", "", connectors.Outgoing, "", "")
		if err != nil {
			t.Fatalf("unable to list payments: %v", err)
		}

		if !reflect.DeepEqual(payments[0], paymentsAfter[0]) {
			t.Fatalf("wrong data")
		}
	}

	{
		payments, err := store.ListPayments("", "", "", connectors.Lightning, "")
		if err != nil {
			t.Fatalf("unable to list payments: %v", err)
		}

		if !reflect.DeepEqual(payments[0], paymentsAfter[1]) {
			t.Fatalf("wrong data")
		}
	}
}
