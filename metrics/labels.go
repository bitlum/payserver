package metrics

// Severity defines how severe the problem we faced, depending on this we
// farther react on the problem differently. For example in Prometheus
// metric backend implementation we setup the alert rules on the server
// which notifies as about the problem.
type Severity string

var (
	// HighSeverity high level of error importance.
	HighSeverity Severity = "high"

	// MiddleSeverity middle level of err importance.
	MiddleSeverity Severity = "middle"

	// LowSeverity low level of error importance.
	LowSeverity Severity = "low"
)
