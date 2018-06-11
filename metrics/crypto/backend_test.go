package crypto

import (
	"testing"
	"time"
)

type testBackend struct {
	panicSent bool
}

func (b *testBackend) CurrentFunds(daemon, asset string, amount float64) {}
func (b *testBackend) AddRequest(daemon, asset, request string)          {}
func (b *testBackend) AddError(daemon, asset, request, severity string)  {}
func (b *testBackend) AddPanic(daemon, asset, request string) {
	b.panicSent = true
}

func (b *testBackend) AddRequestDuration(daemon, asset, request string,
	dur time.Duration) {
}

func TestAddPanic(t *testing.T) {
	b := &testBackend{}

	f := func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()

		metric := NewMetric("lnd", "BTC", "TestMethod", b)
		defer metric.Finish()

		panic("unexpected kek!")
	}

	f()

	if !b.panicSent {
		t.Fatalf("panic haven't been sent")
	}
}
