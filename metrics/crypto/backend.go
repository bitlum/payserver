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
	OverallSent(daemon, asset string, amount float64)
	OverallReceived(daemon, asset string, amount float64)
	OverallFee(daemon, asset string, amount float64)
	CurrentFunds(daemon, asset string, amount float64)

	AddRequest(daemon, asset, request string)
	AddError(daemon, asset, request, severity string)
	AddPanic(daemon, asset, request string)
	AddRequestDuration(daemon, asset, request string, dur time.Duration)
}

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
	currentFunds           *prometheus.GaugeVec
	overallSentFunds       *prometheus.GaugeVec
	overallReceivedFunds   *prometheus.GaugeVec
	overallFeeFunds        *prometheus.GaugeVec
}

// CurrentFunds sets the number of funds available under control of system.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) CurrentFunds(daemon, asset string, amount float64) {
	m.currentFunds.With(
		prometheus.Labels{
			assetLabel:  asset,
			daemonLabel: daemon,
		},
	).Set(amount)
}

// OverallSent sets the number of funds sent by the connector.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) OverallSent(daemon, asset string, amount float64) {
	m.overallSentFunds.With(
		prometheus.Labels{
			assetLabel:  asset,
			daemonLabel: daemon,
		},
	).Set(amount)
}

// OverallReceived sets the number of funds received by the connector.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) OverallReceived(daemon, asset string, amount float64) {
	m.overallReceivedFunds.With(
		prometheus.Labels{
			assetLabel:  asset,
			daemonLabel: daemon,
		},
	).Set(amount)
}

// OverallFee sets the number of fee funds spent by the connector.
//
// NOTE: Non-pointer receiver made by intent to avoid conflict in the system
// with parallel metrics report.
func (m PrometheusBackend) OverallFee(daemon, asset string, amount float64) {
	m.overallFeeFunds.With(
		prometheus.Labels{
			assetLabel:  asset,
			daemonLabel: daemon,
		},
	).Set(amount)
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
		return backend, errors.Errorf(
			"unable to register 'requestsTotal' metric:" +
				err.Error())

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
		return backend, errors.Errorf(
			"unable to register 'errorsTotal' metric: " +
				err.Error())

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
		return backend, errors.Errorf(
			"unable to register 'panicsTotal' metric: " +
				err.Error())
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
		return backend, errors.Errorf(
			"unable to register 'requestDurationSeconds' metric: " +
				err.Error())
	}

	backend.currentFunds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "current_funds",
			Help:      "Number of funds corresponding to given asset",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.currentFunds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'currentFunds' metric: " +
				err.Error())
	}

	backend.overallSentFunds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "overall_sent",
			Help:      "Number of funds sent by service",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.overallSentFunds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'overallSentFunds' metric: " +
				err.Error())
	}

	backend.overallReceivedFunds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "overall_received",
			Help:      "Number of funds received by service",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.overallReceivedFunds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'overallReceivedFunds' metric: " +
				err.Error())
	}

	backend.overallFeeFunds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.Namespace,
			Subsystem: subsystem,
			Name:      "overall_fee",
			Help:      "Number of funds spent by service",
			ConstLabels: prometheus.Labels{
				metrics.NetLabel: net,
			},
		},
		[]string{
			assetLabel,
			daemonLabel,
		},
	)

	if err := prometheus.Register(backend.overallFeeFunds); err != nil {
		return backend, errors.Errorf(
			"unable to register 'overallFeeFunds' metric: " +
				err.Error())
	}

	return backend, nil
}
