package net

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/connector/connectors/chains/bitcoincash"
	"github.com/bitlum/connector/connectors/chains/dash"
	"github.com/bitlum/connector/connectors/chains/litecoin"
	"github.com/go-errors/errors"
)

func GetParams(asset string, netName string) (*chaincfg.Params, error) {
	switch asset {
	case "BTC":
		switch netName {
		case "mainnet", "main":
			return &chaincfg.MainNetParams, nil
		case "regtest", "simnet":
			return &chaincfg.RegressionNetParams, nil
		case "testnet3", "test", "testnet":
			return &chaincfg.TestNet3Params, nil
		}

	case "BCH":
		switch netName {
		case "mainnet", "main":
			return &bitcoincash.MainNetParams, nil
		case "regtest", "simnet":
			return &bitcoincash.RegressionNetParams, nil
		case "testnet3", "test", "testnet":
			return &bitcoincash.TestNet3Params, nil
		}

	case "LTC":
		switch netName {
		case "mainnet", "main":
			return &litecoin.MainNetParams, nil
		case "mainnet-legacy":
			return &litecoin.MainNetParamsLegacy, nil
		case "regtest", "simnet":
			return &litecoin.RegressionNetParams, nil
		case "testnet4", "test", "testnet":
			return &litecoin.TestNet4Params, nil
		}

	case "DASH":
		switch netName {
		case "mainnet", "main":
			return &dash.MainNetParams, nil
		case "regtest", "simnet":
			return &dash.RegressionNetParams, nil
		case "testnet3", "test", "testnet":
			return &dash.TestNet3Params, nil
		}

	default:
		return nil, errors.Errorf("asset '%v' is invalid or unsupported",
			asset)
	}

	return nil, errors.Errorf("asset's network '%s' is "+
		"invalid or unsupported", asset, netName)
}
