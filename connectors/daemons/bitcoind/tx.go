package bitcoind

import (
	"fmt"

	"math"

	"github.com/bitlum/connector/connectors/daemons/bitcoind/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"github.com/btcsuite/btcd/blockchain"
)

// ErrInsufficientFunds is a type matching the error interface which is
// returned when coin selection for a new funding transaction fails to due
// having an insufficient amount of confirmed funds.
type ErrInsufficientFunds struct {
	amountNeeded    btcutil.Amount
	amountAvailable btcutil.Amount
}

func (e *ErrInsufficientFunds) Error() string {
	return fmt.Sprintf("not enough outputs to create transaction,"+
		" need %v only have %v  available", printAmount(e.amountNeeded),
		printAmount(e.amountAvailable))
}

// selectInputs selects a slice of inputs necessary to meet the specified
// selection amount. If input selection is unable to succeed to to insufficient
// funds, a non-nil error is returned.
// TODO(andrew.shvv) Probably sort inputs, before selecting them.
// TODO(andrew.shvv) Develop hierstic algorithm of choosing the outputs
// efficiently.
func selectInputs(amt btcutil.Amount,
	inputsMap map[string]btcjson.ListUnspentResult) (btcutil.Amount,
	[]btcjson.ListUnspentResult, error) {

	var inputs []btcjson.ListUnspentResult
	satSelected := btcutil.Amount(0)
	for _, input := range inputsMap {
		amount, err := btcutil.NewAmount(input.Amount)
		if err != nil {
			return 0, nil, err
		}

		inputs = append(inputs, input)
		satSelected += amount
		if satSelected >= amt {
			return satSelected, inputs, nil
		}
	}
	return 0, nil, &ErrInsufficientFunds{amt, satSelected}
}

// syncUnspent populates local map of confirmed from our POV unspent outputs
// so that later we could construct transaction in a fast manner.
// Otherwise construction of transaction might take couple of seconds.
func (c *Connector) syncUnspent() error {
	c.unspentSyncMtx.Lock()
	defer c.unspentSyncMtx.Unlock()

	// Find all unlocked unspent outputs with greater than minimum confirmation.
	minConf := int(c.cfg.MinConfirmations)
	maxConf := int(math.MaxInt32)

	var err error
	unspent, err := c.client.ListUnspentMinMax(minConf, maxConf)
	if err != nil {
		return errors.Errorf("unable to list unspent: %v", err)
	}

	var amount decimal.Decimal
	c.unspent = make(map[string]btcjson.ListUnspentResult, len(unspent))
	for _, u := range unspent {
		c.unspent[u.TxID] = u
		a := decimal.NewFromFloat(u.Amount)
		amount = amount.Add(a)
	}

	c.log.Debugf("Sync %v unspent inputs, with overall %v %v amount",
		len(unspent), amount.String(), c.cfg.Asset)

	return nil
}

// craftTransaction performs coin selection in order to obtain outputs which sum
// to at least 'numCoins' amount of satoshis. If necessary, a change address will
// also be generated.
func (c *Connector) craftTransaction(feeRatePerByte uint64,
	amtSat btcutil.Amount, address btcutil.Address) (*wire.MsgTx,
	btcutil.Amount, error) {

	// We hold the coin select mutex while querying for outputs, and
	// performing coin selection in order to avoid inadvertent double
	// spends.
	c.coinSelectMtx.Lock()
	defer c.coinSelectMtx.Unlock()

	c.log.Debugf("Performing coin selection fee rate(%v sat/byte), " +
		"amount(%v)", feeRatePerByte, amtSat)

	// TODO(andrew.shvv) what if send two consequent requests? The
	// second one will be using the same outputs and it will lead to issue.

	// First of all unlock all unspent outputs, to exclude the situation where
	// we accidentally locked inputs and server crashed or just forget to
	// unlock them.
	c.log.Debugf("Unlocking unspent inputs...")
	if err := c.client.LockUnspent(true, nil); err != nil {
		return nil, 0, errors.Errorf("unable to unlock unspent outputs")
	}

	// Try to get unspent outputs from local cache,
	// if it is not initialized than sync it.
	c.unspentSyncMtx.Lock()
	if c.unspent == nil {
		c.unspentSyncMtx.Unlock()

		if err := c.syncUnspent(); err != nil {
			return nil, 0, errors.Errorf("unable to sync unspent: %v", err)
		}
	}
	c.unspentSyncMtx.Unlock()

	// Perform coin selection over our available, unlocked unspent outputs
	// in order to find enough coins to meet the funding amount
	// requirements.
	c.unspentSyncMtx.Lock()
	selectedInputs, changeAmt, requiredFee, err := coinSelect(feeRatePerByte,
		amtSat, c.unspent)
	c.unspentSyncMtx.Unlock()

	if err != nil {
		return nil, 0, errors.Errorf("unable to select inputs: %v", err)
	}

	c.log.Debugf("Selected %v unspent inputs, amount(%v), change(%v), fee(%v)",
		len(selectedInputs), printAmount(amtSat), printAmount(changeAmt),
		printAmount(requiredFee))

	// Lock the selected coins. These coins are now "reserved", this
	// prevents concurrent funding requests from referring to and this
	// double-spending the same set of coins.
	inputs := make([]btcjson.TransactionInput, len(selectedInputs))

	for i, input := range selectedInputs {
		txid, err := chainhash.NewHashFromStr(input.TxID)
		if err != nil {
			return nil, 0, err
		}

		outpoint := wire.NewOutPoint(txid, input.Vout)
		err = c.client.LockUnspent(false, []*wire.OutPoint{outpoint})
		if err != nil {
			return nil, 0, err
		}

		inputs[i] = btcjson.TransactionInput{
			Txid: input.TxID,
			Vout: input.Vout,
		}
	}

	// Record any change output(s) generated as a result of the coin
	// selection.
	outputs := make(map[btcutil.Address]btcutil.Amount)
	outputs[address] = amtSat
	if changeAmt != 0 {
		// Create loopback output with remaining amount which point out to the
		// default account of the wallet.
		changeAddr, err := c.client.GetNewAddress(defaultAccount)
		if err != nil {
			return nil, 0, err
		}
		outputs[changeAddr] = changeAmt
	}

	lockTime := int64(0)
	tx, err := c.client.CreateRawTransaction(inputs, outputs, &lockTime)
	if err != nil {
		return nil, 0, err
	}

	c.unspentSyncMtx.Lock()
	for _, input := range selectedInputs {
		delete(c.unspent, input.TxID)
	}
	c.unspentSyncMtx.Unlock()

	return tx, requiredFee, nil
}

// coinSelect attempts to select a sufficient amount of coins, including a
// change output to fund amt satoshis, adhering to the specified fee rate. The
// specified fee rate should be expressed in sat/byte for coin selection to
// function properly.
func coinSelect(feeRatePerByte uint64, amtSat btcutil.Amount,
	unspent map[string]btcjson.ListUnspentResult) ([]btcjson.ListUnspentResult,
	btcutil.Amount, btcutil.Amount, error) {

	amtNeeded := amtSat
	for {
		// First perform an initial round of coin selection to estimate
		// the required fee.
		totalSat, selectedUtxos, err := selectInputs(amtNeeded, unspent)
		if err != nil {
			return nil, 0, 0, err
		}

		var weightEstimate TxWeightEstimator

		// For every input add weight
		for i := 0; i < len(selectedUtxos); i++ {
			weightEstimate.AddP2PKHInput()
		}

		// This is usual transaction and it will contain one P2PKH output to
		// pay to someone else, add weight for it.
		weightEstimate.AddP2PKHOutput()

		// Assume that change output is a P2PKH output.
		weightEstimate.AddP2PKHOutput()

		// The difference between the selected amount and the amount
		// requested will be used to pay fees, and generate a change
		// output with the remaining.
		overShootAmt := totalSat - amtSat

		// Based on the estimated size and fee rate, if the excess
		// amount isn't enough to pay fees, then increase the requested
		// coin amount by the estimate required fee, performing another
		// round of coin selection.
		feeRatePerWeight := feeRatePerByte * blockchain.WitnessScaleFactor
		requiredFee := btcutil.Amount(uint64(weightEstimate.Weight()) * feeRatePerWeight)
		if overShootAmt < requiredFee {
			amtNeeded = amtSat + requiredFee
			continue
		}

		// If the fee is sufficient, then calculate the amount of the
		// change output.
		changeAmt := overShootAmt - requiredFee

		return selectedUtxos, changeAmt, requiredFee, nil
	}
}
