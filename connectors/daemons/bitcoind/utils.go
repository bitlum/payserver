package bitcoind

import (
	"math/big"

	"github.com/btcsuite/btcutil"
	"github.com/shopspring/decimal"
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/rpc/bitcoin"
	"github.com/bitlum/connector/connectors/rpc/litecoin"
	"github.com/bitlum/connector/connectors/rpc/bitcoincash"
	"github.com/bitlum/connector/connectors/rpc/dash"
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

func decodeAddress(asset connectors.Asset, address,
network string) (btcutil.Address, error) {
	switch asset {
	case connectors.BTC:
		return bitcoin.DecodeAddress(address, network)
	case connectors.LTC:
		return litecoin.DecodeAddress(address, network)
	case connectors.BCH:
		return bitcoincash.DecodeAddress(address, network)
	case connectors.DASH:
		return dash.DecodeAddress(address, network)
	default:
		return nil, errors.Errorf("unsupported asset asset(%v)", asset)
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

func accountToAlias(account string) connectors.AccountAlias {
	switch account {
	case defaultAccount:
		return connectors.DefaultAccount

	case allAccounts:
		return connectors.AllAccounts

	default:
		return connectors.AccountAlias(account)
	}
}

func aliasToAccount(alias connectors.AccountAlias) string {
	switch alias {
	case connectors.SentAccount:
		// In bitcoin-like asset you daemons is working in a way that you
		// could use use all the money from all accounts.
		return allAccounts

	case connectors.DefaultAccount:
		return defaultAccount

	case connectors.AllAccounts:
		return allAccounts

	default:
		return string(alias)
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
