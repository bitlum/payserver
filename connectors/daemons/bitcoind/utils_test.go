package bitcoind

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestDec2Amount(t *testing.T) {
	amt, err := decimal.NewFromString("0.01688044695939197")
	if err != nil {
		t.Fatalf("unable convert amount: %v", err)
	}

	if decAmount2Sat(amt) != 1688044 {
		t.Fatalf("wrong amount")
	}
}
