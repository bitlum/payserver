package crypto

import (
	"time"
	"runtime/debug"
	"strings"
	"github.com/bitlum/connector/metrics"
)

// Metric is an enhancement of Metric backend, which is more suited for this
// package usage.
type Metric struct {
	// backend is an entity which is responsible for collecting and storing
	// the metricsBackend.
	backend MetricsBackend

	// requestName is name used as value for label when adding metricsBackend.
	// Supposed to be set by `NewMetric` method.
	requestName string

	// startTime is a request start time used to measure request
	// duration in `AddRequestDuration` method. Supposed to be set
	// in `NewMetric` method to now.
	startTime time.Time

	// asset is an asset with which cryptocurrency daemon is working.
	// Used as an additional label in the Metric server.
	asset string

	// daemon is an daemon name with which we interacting, this label would
	// helps as quickly grasp which daemon has a problem if something bas
	// happened.
	daemon string
}

// NewMetric creates new metricsBackend with specified request name, sets start time
// and adds request to metricsBackend.
//
// Note: we use not pointer type receiver so any changes within method
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func NewMetric(daemon, asset, request string, backend MetricsBackend) Metric {
	m := Metric{}

	m.backend = backend
	m.requestName = request
	m.startTime = time.Now()
	m.asset = asset
	m.daemon = daemon

	m.backend.AddRequest(m.daemon, m.asset, request)
	return m
}

// AddError adds error metric with specified error.
//
// Note: we use not pointer type receiver so any changes within method
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func (m Metric) AddError(severity metrics.Severity) {
	m.backend.AddError(m.daemon, m.asset, m.requestName, string(severity))
}

// AddPanic adds panic metric
func (m Metric) AddPanic() {
	m.backend.AddPanic(m.daemon, m.asset, m.requestName)
}

// CurrentFunds set the current amount of funds available.
func (m Metric) CurrentFunds(amount float64) {
	m.backend.CurrentFunds(m.daemon, m.asset, amount)
}

// AddRequestDuration adds request duration metric. Supposed to be
// called after `NewMetric` which defines `startTime`. Calculates
// duration using `startTime` and now as end time.
//
// Note: we use not pointer type receiver so any changes within method
// do not change original struct fields. Each call creates new `metricsBackend`
// with copied fields.
func (m Metric) AddRequestDuration() {
	if m.startTime.Equal(time.Time{}) {
		panic("not initialised request")
	}

	dur := time.Now().Sub(m.startTime)
	m.backend.AddRequestDuration(m.daemon, m.asset, m.requestName, dur)
}

// Finish used as defer in handlers, to ensure that we track panics and
// measure handler time.
func (m Metric) Finish() {
	m.AddRequestDuration()

	if r := recover(); r != nil {
		m.AddPanic()
		panic(stackTrace())
	}
}

func stackTrace() string {
	s := string(debug.Stack())
	ls := strings.Split(s, "\n")
	for i, l := range ls {
		if strings.Index(l, "src/runtime/panic.go") != -1 && i > 0 &&
			strings.Index(ls[i-1], "panic(") == 0 {
			return strings.TrimSpace(strings.Join(ls[i+2:], "\n"))
		}
	}
	return s
}
