package crpc

import "github.com/bitlum/connector/metrics"


var errToSeverityMap = map[int]metrics.Severity{
	ErrAssetNotSupported:   metrics.LowSeverity,
	ErrInvalidArgument:     metrics.LowSeverity,
	ErrNetworkNotSupported: metrics.LowSeverity,
}

func errMetricsInfo(err int) string {
	severity := metrics.LowSeverity

	if s, ok := errToSeverityMap[err]; ok {
		severity = s
	}

	return string(severity)
}
