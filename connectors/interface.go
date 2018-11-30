package connectors

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/shopspring/decimal"
)

type LightningInfo struct {
	Host      string
	Port      string
	MinAmount string
	MaxAmount string
	*lnrpc.GetInfoResponse
}

// BlockchainConnector is an interface which describes the blockchain service
// which is able to connect to blockchain daemon of particular currency and
// operate with transactions, addresses, and also  able to notify other
// subsystems when transaction passes required number of confirmations.
type BlockchainConnector interface {
	// CreateAddress is used to create deposit address.
	CreateAddress() (string, error)

	// ConfirmedBalance return the amount of confirmed funds available for account.
	ConfirmedBalance() (decimal.Decimal, error)

	// PendingBalance return the amount of funds waiting to be confirmed.
	PendingBalance() (decimal.Decimal, error)

	// SendPayment sends payment with given amount to the given address.
	SendPayment(address, amount string) (*Payment, error)

	// ValidateAddress takes the blockchain address and ensure its valid.
	ValidateAddress(address string) error

	// EstimateFee estimate fee for the transaction with the given sending
	// amount.
	EstimateFee(amount string) (decimal.Decimal, error)
}

// LightningConnector is an interface which describes the service
// which is able to connect lightning network daemon of particular currency and
// operate with transactions, addresses, and also  able to notify other
// subsystems when invoice is settled.
type LightningConnector interface {
	// Info returns the information about our lnd node.
	Info() (*LightningInfo, error)

	// CreateInvoice is used to create lightning network invoice.
	CreateInvoice(receipt, amount, description string) (string,
		*zpay32.Invoice, error)

	// SendTo is used to send specific amount of money to address within this
	// payment system.
	SendTo(invoice, amount string) (*Payment, error)

	// ConfirmedBalance return the amount of confirmed funds available for account.
	// TODO(andrew.shvv) Implement lightning wallet balance
	ConfirmedBalance() (decimal.Decimal, error)

	// PendingBalance return the amount of funds waiting to be confirmed.
	// TODO(andrew.shvv) Implement lightning wallet balance
	PendingBalance() (decimal.Decimal, error)

	// QueryRoutes returns list of routes from to the given lnd node,
	// and insures the the capacity of the channels is sufficient.
	QueryRoutes(pubKey, amount string, limit int32) ([]*lnrpc.Route, error)

	// ValidateInvoice takes the encoded lightning network invoice and ensure
	// its valid.
	ValidateInvoice(invoice, amount string) (*zpay32.Invoice, error)

	// EstimateFee estimate fee for the payment with the given sending
	// amount, to the given node.
	EstimateFee(invoice string) (decimal.Decimal, error)
}
