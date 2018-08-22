package lnd

import (
	"github.com/btcsuite/btcutil"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"math/big"
	"github.com/bitlum/connector/connectors"
)

var satoshiPerBitcoin = decimal.New(btcutil.SatoshiPerBitcoin, 0)

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

func sat2DecAmount(amount btcutil.Amount) decimal.Decimal {
	amt := decimal.NewFromBigInt(big.NewInt(int64(amount)), 0)
	return amt.Div(satoshiPerBitcoin)
}

func generatePaymentID(invoiceStr, paymentHash string) string {
	return connectors.GeneratePaymentID(invoiceStr, paymentHash)
}
