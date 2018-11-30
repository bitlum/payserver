package bitcoind

import (
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/btcsuite/btcutil"
	"testing"
)

// TestSimpleCreateReorganisationOutputs test creation of outputs without
// specifying fee, to just check general logic.
func TestSimpleCreateReorganisationOutputs(t *testing.T) {
	feeRatePerByte := uint64(0)

	outputValue := 1.0
	inputs := []rpc.UnspentInput{
		{
			Amount: outputValue,
		},
		{
			Amount: outputValue,
		},
	}

	var overallUTXOAmount btcutil.Amount
	for _, input := range inputs {
		amount, _ := btcutil.NewAmount(input.Amount)
		overallUTXOAmount += amount
	}

	utxoValue, _ := btcutil.NewAmount(0.5)
	outputAmounts, fee, err := createReorganisationOutputs(feeRatePerByte,
		inputs, utxoValue)
	if err != nil {
		t.Fatalf("unable craft reoganisation outputs: %v", err)
	}

	overallOutputsAmount := btcutil.Amount(0)
	for _, amount := range outputAmounts {
		overallOutputsAmount += amount
	}

	if (overallUTXOAmount - overallOutputsAmount) != fee {
		t.Fatalf("returned fee is wrong")
	}

	if len(outputAmounts) != 4 {
		t.Fatalf("wrong outputs number")
	}
}

// TestCreateReorganisationOutputsWithFee take real case, and check that
// function will behave properly.
func TestCreateReorganisationOutputsWithFee(t *testing.T) {
	feeRatePerByte := uint64(10)

	outputValue := 0.5
	inputs := []rpc.UnspentInput{
		{
			Amount: outputValue,
		},
		{
			Amount: outputValue,
		},
	}

	var overallUTXOAmount btcutil.Amount
	for _, input := range inputs {
		amount, _ := btcutil.NewAmount(input.Amount)
		overallUTXOAmount += amount
	}

	outputAmounts, fee, err := createReorganisationOutputs(feeRatePerByte,
		inputs, optimalUTXOValue)
	if err != nil {
		t.Fatalf("unable craft reoganisation outputs: %v", err)
	}

	overallOutputsAmount := btcutil.Amount(0)
	for _, amount := range outputAmounts {
		overallOutputsAmount += amount
	}

	if (overallUTXOAmount - overallOutputsAmount) != fee {
		t.Fatalf("returned fee is wrong")
	}

	if len(outputAmounts) != 124 {
		t.Fatalf("wrong number of outputs")
	}

	// Check all outputs except the last one which pays the fees,
	// than they are equal to the optimal value.
	for i := 0; i < len(outputAmounts)-1; i++ {
		if outputAmounts[i] != optimalUTXOValue {
			t.Fatalf("wrong output amount")
		}
	}
}
