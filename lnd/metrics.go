package lnd

import (
	"github.com/bitlum/connector/metrics"
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
	ErrUnableQueryRoutes:        metrics.HighSeverity,
	ErrPubkey:                   metrics.LowSeverity,
}

func errToSeverity(err int) string {
	severity := metrics.LowSeverity

	if s, ok := errToSeverityMap[err]; ok {
		severity = s
	}

	return string(severity)
}
