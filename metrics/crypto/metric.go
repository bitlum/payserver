package crypto

import "time"

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
func (m Metric) AddError(severity string) {
	m.backend.AddError(m.daemon, m.asset, m.requestName, severity)
}

// AddPanic adds panic metric
func (m Metric) AddPanic() {
	m.backend.AddPanic(m.daemon, m.asset, m.requestName)
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
