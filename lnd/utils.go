package lnd

import (
	"github.com/bitlum/btcutil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/metrics/crypto"
	"runtime/debug"
	"strings"
)

func btcToSatoshi(amount string) (int64, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, errors.Errorf("unable to parse amount(%v): %v",
			amount, err)
	}

	a, _ := amt.Float64()
	btcAmount, err := btcutil.NewAmount(a)
	if err != nil {
		return 0, errors.Errorf("unable to parse amount(%v): %v", a, err)
	}

	return int64(btcAmount), nil
}

// finishHandler used as defer in handlers, to ensure that we track panics and
// measure handler time.
func finishHandler(metrics crypto.Metric) {
	metrics.AddRequestDuration()

	if r := recover(); r != nil {
		metrics.AddPanic()
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
