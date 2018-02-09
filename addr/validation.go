package addr

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/connector/chains/net"
	"github.com/go-errors/errors"
)

// Validate validates asset's addr in specified net. For ETH net is meaningless, you can use "".
func Validate(asset string, netName string, address string) error {
	switch asset {
	case "BTC":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("failed to get net params: %v", err)
		}
		return validateCommon(address, netParams)
	case "BCH":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("failed to get net params: %v", err)
		}
		return validateBitcoinCash(address, netParams)
	case "ETH":
		return validateEthereum(address)
	case "LTC":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("failed to get net params: %v", err)
		}
		err = validateCommon(address, netParams)
		if err == nil || netName != "mainnet" {
			return err
		}
		legacyNetParams, err := net.GetParams(asset, "mainnet-legacy")
		if err != nil {
			return errors.Errorf("failed to get legacy mainnet params: %v", err)
		}
		return validateCommon(address, legacyNetParams)
	case "DASH":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("failed to get net params: %v", err)
		}
		return validateCommon(address, netParams)
	}
	return errors.New("invalid or unsupported asset")
}

func validateCommon(address string, network *chaincfg.Params) error {
	decodedAddress, err := btcutil.DecodeAddress(address, network)
	if err != nil {
		return err
	}
	if !decodedAddress.IsForNet(network) {
		return errors.New("address is not for specified network")
	}
	return nil
}
