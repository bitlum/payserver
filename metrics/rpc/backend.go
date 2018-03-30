package rpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/bitlum/graphql-go/errors"
	"github.com/bitlum/connector/metrics"
)

const (
	// subsystem is used as the second part in the name of the metric,
	// after the metrics namespace.
	subsystem = "rpc"

	// requestLabel is used to distinguish different RPC requests during the
	// process of metric analysis and alert rule constructing.
	requestLabel = "request"

	// severityLabel is used to distinguish different error codes by its
	// level of importance.
	severityLabel = "severity"
)

// MetricsBackend is a system which is responsible for receiving and storing
// the connector metrics.
type MetricsBackend interface {
	AddError(request, severity string)
}

// EmptyBackend is used as an empty metrics backend in order to avoid
type EmptyBackend struct{}

func (b *EmptyBackend) AddError(request, severity string) {}

// PrometheusBackend is the main subsystem metrics implementation. Uses
// prometheus metrics singletons defined above.
//
// WARN: Method name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
type PrometheusBackend struct {
	errorsTotal *prometheus.CounterVec
}

// AddError increases error counter for the given method name.
//
// WARN: Error code name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddError(method, severity string) {
	m.errorsTotal.With(
		prometheus.Labels{
			requestLabel:  method,
			severityLabel: severity,
		},
	).Add(1)
}

// InitMetricsBackend creates subsystem metrics for specified
// net. Creates and tries to register metrics singletons. If register was
// already done, than function not returning error.
func InitMetricsBackend(net string) (MetricsBackend, error) {
	backend := PrometheusBackend{}

	backend.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Total requests which processing ended with error",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
			severityLabel,
		},
	)

	if err := prometheus.Register(backend.errorsTotal); err != nil {
		// Skip returning error if we re-registered metric.
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return backend, errors.Errorf(
				"unable to register 'errorsTotal' metric: " +
					err.Error())
		}
	}

	return backend, nil
}
