package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/go-errors/errors"
)

func GetParams(netName string) (*chaincfg.Params, error) {
	switch netName {
	case "mainnet", "main":
		return &chaincfg.MainNetParams, nil
	case "regtest", "simnet":
		return &chaincfg.RegressionNetParams, nil
	case "testnet3", "test", "testnet":
		return &chaincfg.TestNet3Params, nil
	}

	return nil, errors.Errorf("network '%s' is invalid or unsupported",
		netName)
}
