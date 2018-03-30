package crypto

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/bitlum/graphql-go/errors"
	"github.com/bitlum/connector/metrics"
)

const (
	// subsystem is used as the second part in the name of the metric,
	// after the metrics namespace.
	subsystem = "crypto"

	// requestLabel is used to distinguish different RPC requests during the
	// process of metric analysis and alert rule constructing.
	requestLabel = "request"

	// assetLabel is used to distinguish different currency and daemon on the
	// metrics server.
	assetLabel = "asset"

	// severityLabel is used to distinguish different error codes by its
	// level of importance.
	severityLabel = "severity"

	// daemonLabel is used to distinguish different daemon names,
	// and quickly identify the problem if such occurs.
	daemonLabel = "daemon"
)

// MetricsBackend is a system which is responsible for receiving and storing
// the connector metricsBackend.
type MetricsBackend interface {
	AddRequest(daemon, asset, request string)
	AddError(daemon, asset, request, severity string)
	AddPanic(daemon, asset, request string)
	AddRequestDuration(daemon, asset, request string, dur time.Duration)
}

// EmptyBackend is used as an empty metricsBackend backend in order to avoid
type EmptyBackend struct{}

func (b *EmptyBackend) AddRequest(query string)                            {}
func (b *EmptyBackend) AddError(query string, errCode string)              {}
func (b *EmptyBackend) AddPanic(query string)                              {}
func (b *EmptyBackend) AddRequestDuration(query string, dur time.Duration) {}

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
	requestsTotal          *prometheus.CounterVec
	errorsTotal            *prometheus.CounterVec
	panicsTotal            *prometheus.CounterVec
	requestDurationSeconds *prometheus.HistogramVec
}

// AddRequest increases request counter for the given request name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddRequest(daemon, asset, request string) {
	m.requestsTotal.With(
		prometheus.Labels{
			requestLabel: request,
			assetLabel:   asset,
			daemonLabel:  daemon,
		},
	).Add(1)
}

// AddError increases error counter for the given method name.
//
// WARN: Error code name should be taken from limited set.
// Don't use dynamic naming, it may cause dramatic increase of the amount of
// data on the metric server.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddError(daemon, asset, request, severity string) {
	m.errorsTotal.With(
		prometheus.Labels{
			requestLabel:  request,
			severityLabel: severity,
			assetLabel:    asset,
			daemonLabel:   daemon,
		},
	).Add(1)
}

// AddPanic increases panic counter for the given method name.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddPanic(daemon, asset, request string) {
	m.panicsTotal.With(
		prometheus.Labels{
			requestLabel: request,
			assetLabel:   asset,
			daemonLabel:  daemon,
		},
	).Add(1)
}

// AddRequestDuration sends the metric with how much time request has taken
// to proceed.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) AddRequestDuration(daemon, asset, request string,
	dur time.Duration) {
	m.requestDurationSeconds.With(
		prometheus.Labels{
			requestLabel: request,
			assetLabel:   asset,
			daemonLabel:  daemon,
		},
	).Observe(dur.Seconds())
}

// InitMetricsBackend creates subsystem metrics for specified
// net. Creates and tries to register metrics singletons. If register was
// already done, than function not returning error.
func InitMetricsBackend(net string) (MetricsBackend, error) {
	backend := PrometheusBackend{}

	backend.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total requests processed",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.requestsTotal); err != nil {
		// Skip returning error if we re-registered metric.
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return backend, errors.Errorf(
				"unable to register 'requestsTotal' metric:" +
					err.Error())
		}
	}

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
			assetLabel,
			severityLabel,
			daemonLabel,
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

	backend.panicsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "panics_total",
			Help:      "Total requests which processing ended with panic",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.panicsTotal); err != nil {
		// Skip returning error if we re-registered metric.
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return backend, errors.Errorf(
				"unable to register 'panicsTotal' metric: " +
					err.Error())
		}
	}

	backend.requestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Request processing duration in seconds",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			requestLabel,
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.requestDurationSeconds); err != nil {
		// Skip returning error if we re-registered metric.
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return backend, errors.Errorf(
				"unable to register 'requestDurationSeconds' metric: " +
					err.Error())
		}
	}

	return backend, nil
}
