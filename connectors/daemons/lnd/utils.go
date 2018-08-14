package lnd

import (
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
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
