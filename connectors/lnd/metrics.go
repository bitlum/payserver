package lnd

import (
	"github.com/bitlum/connector/metrics"
)

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

var errToSeverityMap = map[int]metrics.Severity{
	ErrSendPayment:              metrics.HighSeverity,
	ErrReadInvoiceStream:        metrics.HighSeverity,
	ErrAddInvoice:               metrics.HighSeverity,
	ErrResubscribeInvoiceStream: metrics.MiddleSeverity,
	ErrSubscribeInvoiceStream:   metrics.MiddleSeverity,
	ErrNonSettledInvoice:        metrics.LowSeverity,
	ErrSendPaymentNotification:  metrics.MiddleSeverity,
	ErrConvertAmount:            metrics.LowSeverity,
	ErrGRPCConnect:              metrics.HighSeverity,
	ErrTLSRead:                  metrics.HighSeverity,
	ErrCreateInvoice:            metrics.HighSeverity,
	ErrGetInfo:                  metrics.HighSeverity,
	ErrUnableQueryRoutes:        metrics.LowSeverity,
	ErrPubkey:                   metrics.LowSeverity,
}

func errToSeverity(err int) string {
	severity := metrics.LowSeverity

	if s, ok := errToSeverityMap[err]; ok {
		severity = s
	}

	return string(severity)
}
