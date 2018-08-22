package connectors

import (
	"github.com/go-errors/errors"
)

// PaymentStorage is an external storage for payments, it is used by
// connector to save payment as well as update its state.
type PaymentsStore interface {
	// PaymentByID returns payment by id.
	PaymentByID(paymentID string) (*Payment, error)

	// PaymentByReceipt returns payment by receipt.
	PaymentByReceipt(receipt string) ([]*Payment, error)

	// SavePayment add payment to the store.
	SavePayment(payment *Payment) error

	// ListPayments return list of all payments.
	ListPayments(asset Asset, status PaymentStatus, direction PaymentDirection,
		media PaymentMedia) ([]*Payment, error)
}

var PaymentNotFound = errors.New("payment not found")

// StateStorage is used to keep data which is needed for connector to
// properly synchronise and track transactions.
//
// NOTE: This storage should be persistent.
type StateStorage interface {
	// PutLastSyncedHash is used to save last synchronised block hash.
	PutLastSyncedHash(hash []byte) error

	// LastSyncedHash is used to retrieve last synchronised block hash.
	LastSyncedHash() ([]byte, error)
}
