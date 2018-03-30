package lnd

import (
	"github.com/bitlum/btcutil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/metrics/crypto"
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
		panic(r)
	}
}
