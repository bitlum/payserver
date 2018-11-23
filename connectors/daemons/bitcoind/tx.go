package bitcoind

import (
	"fmt"

	"math"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"

	"github.com/btcsuite/btcwallet/wallet/txrules"
	"github.com/bitlum/connector/connectors/rpc"
	txsize "github.com/bitlum/btcd/blockchain"
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
	inputsMap map[string]rpc.UnspentInput) (btcutil.Amount,
	[]rpc.UnspentInput, error) {

	var inputs []rpc.UnspentInput
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
	// Find all unlocked unspent outputs with greater than minimum confirmation.
	minConf := int(c.cfg.MinConfirmations)
	maxConf := int(math.MaxInt32)

	var err error
	unspent, err := c.client.ListUnspentMinMax(minConf, maxConf)
	if err != nil {
		return errors.Errorf("unable to list unspent: %v", err)
	}

	var amount decimal.Decimal
	localUnspent := make(map[string]rpc.UnspentInput, len(unspent))
	for _, u := range unspent {
		localUnspent[u.TxID] = u
		a := decimal.NewFromFloat(u.Amount)
		amount = amount.Add(a)
	}

	c.unspentSyncMtx.Lock()
	c.unspent = localUnspent
	c.unspentSyncMtx.Unlock()

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

	c.log.Debugf("Performing coin selection fee rate(%v sat/byte), "+
		"amount(%v)", feeRatePerByte, amtSat)

	// Try to get unspent outputs from local cache,
	// if it is not initialized than sync it.
	if c.unspent == nil {
		if err := c.syncUnspent(); err != nil {
			return nil, 0, errors.Errorf("unable to sync unspent: %v", err)
		}
	}

	// We hold the coin select mutex while querying for outputs, and
	// performing coin selection in order to avoid inadvertent double
	// spends.
	c.unspentSyncMtx.Lock()
	defer c.unspentSyncMtx.Unlock()

	// Perform coin selection over our available, unlocked unspent outputs
	// in order to find enough coins to meet the funding amount
	// requirements.
	selectedInputs, changeAmt, requiredFee, err := coinSelect(feeRatePerByte,
		amtSat, c.unspent)

	if err != nil {
		return nil, 0, errors.Errorf("unable to select inputs: %v", err)
	}

	c.log.Debugf("Selected %v unspent inputs, amount(%v), change(%v), fee(%v)",
		len(selectedInputs), printAmount(amtSat), printAmount(changeAmt),
		printAmount(requiredFee))

	// Lock the selected coins. These coins are now "reserved", this
	// prevents concurrent funding requests from referring to and this
	// double-spending the same set of coins.
	for _, input := range selectedInputs {
		if err = c.client.LockUnspent(input); err != nil {
			return nil, 0, err
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

	tx, err := c.client.CreateRawTransaction(selectedInputs, outputs)
	if err != nil {
		return nil, 0, err
	}

	// Remove unspent utxo from local cache. Otherwise it will be updated only
	// on next cache sync, which might cause inputs re-usage. If transaction
	// will fail, than inputs will be returned on next cache sync.
	for _, input := range selectedInputs {
		delete(c.unspent, input.TxID)
	}

	return tx, requiredFee, nil
}

// coinSelect attempts to select a sufficient amount of coins, including a
// change output to fund amt satoshis, adhering to the specified fee rate. The
// specified fee rate should be expressed in sat/byte for coin selection to
// function properly.
func coinSelect(feeRatePerByte uint64, amtSat btcutil.Amount,
	unspent map[string]rpc.UnspentInput) ([]rpc.UnspentInput,
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
		size := uint64(weightEstimate.Weight() / txsize.WitnessScaleFactor)
		requiredFee := btcutil.Amount(size * feeRatePerByte)

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
