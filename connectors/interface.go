package connectors

import (
	"github.com/shopspring/decimal"
	"github.com/lightningnetwork/lnd/lnrpc"
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
	CreateAddress(account string) (string, error)

	// AccountAddress return the deposit address of account.
	AccountAddress(account string) (string, error)

	// ConfirmedBalance return the amount of confirmed funds available for account.
	ConfirmedBalance(account string) (decimal.Decimal, error)

	// PendingBalance return the amount of funds waiting to be confirmed.
	PendingBalance(account string) (decimal.Decimal, error)

	// PendingTransactions return the transactions which has confirmation
	// number lower the required by payment system.
	PendingTransactions(account string) ([]*Payment, error)

	// CreatePayment generates the payment, but not sends it,
	// instead returns the payment id and waits for it to be approved.
	CreatePayment(address, amount string) (*Payment, error)

	// SendPayment sends created previously payment to the
	// blockchain network.
	SendPayment(paymentID string) (*Payment, error)

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
	CreateInvoice(account, amount, description string) (string, error)

	// SendTo is used to send specific amount of money to address within this
	// payment system.
	SendTo(invoice, amount string) (*Payment, error)

	// ConfirmedBalance return the amount of confirmed funds available for account.
	// TODO(andrew.shvv) Implement lightning wallet balance
	ConfirmedBalance(account string) (decimal.Decimal, error)

	// PendingBalance return the amount of funds waiting to be confirmed.
	// TODO(andrew.shvv) Implement lightning wallet balance
	PendingBalance(account string) (decimal.Decimal, error)

	// QueryRoutes returns list of routes from to the given lnd node,
	// and insures the the capacity of the channels is sufficient.
	QueryRoutes(pubKey, amount string, limit int32) ([]*lnrpc.Route, error)

	// ValidateInvoice takes the encoded lightning network invoice and ensure
	// its valid.
	ValidateInvoice(invoice, amount string) error
}
