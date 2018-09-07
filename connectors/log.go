package connectors

import (
	"fmt"

	"github.com/btcsuite/btclog"
)

// NamedLogger is an extension of btclog.Logger which adds additional name to
// the logging.
type NamedLogger struct {
	Logger btclog.Logger
	Name   string
}

func (l *NamedLogger) convert(format string, params []interface{}) (string,
	[]interface{}) {
	format = fmt.Sprintf("(%v) %v", "%v", format)
	var nparams []interface{}
	nparams = append(nparams, l.Name)
	nparams = append(nparams, params...)
	return format, nparams
}

// Tracef formats message according to format specifier and writes to
// to log with LevelTrace.
func (l *NamedLogger) Tracef(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Tracef(format, params...)
}

// Debugf formats message according to format specifier and writes to
// log with LevelDebug.
func (l *NamedLogger) Debugf(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Debugf(format, params...)
}

// Infof formats message according to format specifier and writes to
// log with LevelInfo.
func (l *NamedLogger) Infof(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Infof(format, params...)
}

// Warnf formats message according to format specifier and writes to
// to log with LevelWarn.
func (l *NamedLogger) Warnf(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Warnf(format, params...)
}

// Errorf formats message according to format specifier and writes to
// to log with LevelError.
func (l *NamedLogger) Errorf(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Errorf(format, params...)
}

// Criticalf formats message according to format specifier and writes to
// log with LevelCritical.
func (l *NamedLogger) Criticalf(format string, params ...interface{}) {
	format, params = l.convert(format, params)
	l.Logger.Criticalf(format, params...)
}

// Trace formats message using the default formats for its operands
// and writes to log with LevelTrace.
func (l *NamedLogger) Trace(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Trace(k...)
}

// Debug formats message using the default formats for its operands
// and writes to log with LevelDebug.
func (l *NamedLogger) Debug(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Debug(k...)
}

// Info formats message using the default formats for its operands
// and writes to log with LevelInfo.
func (l *NamedLogger) Info(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Info(k...)
}

// Warn formats message using the default formats for its operands
// and writes to log with LevelWarn.
func (l *NamedLogger) Warn(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Warn(k...)
}

// Error formats message using the default formats for its operands
// and writes to log with LevelError.
func (l *NamedLogger) Error(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Error(k...)
}

// Critical formats message using the default formats for its operands
// and writes to log with LevelCritical.
func (l *NamedLogger) Critical(v ...interface{}) {
	var k []interface{}
	k = append(k, fmt.Sprintf("(%v)", l.Name))
	k = append(k, v...)
	l.Logger.Critical(k...)
}
