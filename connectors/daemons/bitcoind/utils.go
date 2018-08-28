package bitcoind

import (
	"math/big"

	"github.com/btcsuite/btcutil"
	"github.com/shopspring/decimal"
	"github.com/bitlum/connector/connectors/assets/bitcoin"
	"github.com/bitlum/connector/connectors/assets/litecoin"
	"github.com/bitlum/connector/connectors/assets/bitcoincash"
	"github.com/bitlum/connector/connectors/assets/dash"
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/bitlum/connector/connectors"
)

var satoshiPerBitcoin = decimal.New(btcutil.SatoshiPerBitcoin, 0)

func decAmount2Sat(amount decimal.Decimal) btcutil.Amount {
	// If we would try to convert amount in float representation than it
	// could lead to precious error, for that reason convert in manually rather
	// than using btcutil.NewAmount().
	return btcutil.Amount(amount.Mul(satoshiPerBitcoin).IntPart())
}

func sat2DecAmount(amount btcutil.Amount) decimal.Decimal {
	amt := decimal.NewFromBigInt(big.NewInt(int64(amount)), 0)
	return amt.Div(satoshiPerBitcoin)
}

func printAmount(a btcutil.Amount) string {
	return decimal.NewFromFloat(a.ToBTC()).Round(8).String()
}

func isProperNet(desiredNet, actualNet string) bool {
	// Handle the case of different simulation networks names
	if desiredNet == "simnet" && actualNet == "regtest" {
		return true
	}

	// Handle the case of different testnet networks names
	if desiredNet == "testnet" && actualNet == "test" {
		return true
	}

	// Handle the case of different mainnet networks names
	if desiredNet == "mainnet" && actualNet == "main" {
		return true
	}

	return desiredNet == actualNet
}

func validateAddress(asset connectors.Asset, address, network string) error {
	switch asset {
	case connectors.BTC:
		return bitcoin.ValidateAddress(address, network)
	case connectors.LTC:
		return litecoin.ValidateAddress(address, network)
	case connectors.BCH:
		return bitcoincash.ValidateAddress(address, network)
	case connectors.DASH:
		return dash.ValidateAddress(address, network)
	default:
		return errors.Errorf("unsupported asset asset(%v)", asset)
	}
}

func getParams(asset connectors.Asset, network string) (*chaincfg.Params, error) {
	switch asset {
	case connectors.BTC:
		return bitcoin.GetParams(network)
	case connectors.LTC:
		return litecoin.GetParams(network)
	case connectors.BCH:
		return bitcoincash.GetParams(network)
	case connectors.DASH:
		return dash.GetParams(network)
	default:
		return nil, errors.Errorf("unsupported asset asset(%v)", asset)
	}
}

func accountToAlias(account string) string {
	switch account {
	case defaultAccount:
		return "default"
	case allAccounts:
		return "all"
	default:
		return account
	}
}

func aliasToAccount(alias string) string {
	switch alias {
	case "default":
		return defaultAccount
	case "all":
		return allAccounts
	default:
		return alias
	}
}

// generatePaymentID generates unique string based on the tx id and receive
// address, which are together
//
// NOTE: Direction is needed to have a distinction between circular payments,
// i.e. the payment which are going from our wallet to our wallet. Because this
// transaction would have the same address and txid, but should be tracked
// distinctly.
func generatePaymentID(txID, receiveAddress string,
	direction connectors.PaymentDirection) string {
	return connectors.GeneratePaymentID(txID, receiveAddress, string(direction))
}
