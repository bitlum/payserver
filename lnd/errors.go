package lnd

const (
	ErrSendPayment              = iota
	ErrReadInvoiceStream
	ErrAddInvoice
	ErrResubscribeInvoiceStream
	ErrSubscribeInvoiceStream
	ErrNonSettledInvoice
	ErrSendPaymentNotification
	ErrConvertAmount
	ErrGRPCConnect
	ErrTLSRead
	ErrCreateInvoice
	ErrGetInfo
	ErrUnableQueryRoutes
	ErrPubkey
)
