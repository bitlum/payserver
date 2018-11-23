package common

import (
	"github.com/shopspring/decimal"
	"math/big"
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcutil"
)

var SatoshiPerBitcoin = decimal.New(btcutil.SatoshiPerBitcoin, 0)

func BtcStrToSatoshi(amount string) (int64, error) {
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

func Sat2DecAmount(amount btcutil.Amount) decimal.Decimal {
	amt := decimal.NewFromBigInt(big.NewInt(int64(amount)), 0)
	return amt.Div(SatoshiPerBitcoin)
}
