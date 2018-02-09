package addr

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/connector/chains/net"
	"github.com/go-errors/errors"
)

// Validate ensures that address is valid, and belongs to the given network,
// and type.validates asset addr by the given . For ETH net is meaningless,
// you can use "".
func Validate(asset string, netName string, address string) error {
	switch asset {
	case "BTC":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("unable to get net params: %v", err)
		}

		return validateCommon(address, netParams)

	case "BCH":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("unable to get net params: %v", err)
		}

		return validateBitcoinCash(address, netParams)

	case "ETH":
		return validateEthereum(address)

	case "LTC":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("unable to get net params: %v", err)
		}

		err = validateCommon(address, netParams)
		// If there is no error than address is valid, if there is error, it
		// means that probably that address might belongs to the legacy
		// litecoin address type, but legacy address only exist in mainnet.
		if err == nil || netName != "mainnet" {
			return err
		}

		legacyNetParams, err := net.GetParams(asset, "mainnet-legacy")
		if err != nil {
			return errors.Errorf("unable to get legacy mainnet params: %v",
				err)
		}

		return validateCommon(address, legacyNetParams)

	case "DASH":
		netParams, err := net.GetParams(asset, netName)
		if err != nil {
			return errors.Errorf("unable  to get net params: %v", err)
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
