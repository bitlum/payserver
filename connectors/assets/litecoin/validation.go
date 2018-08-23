package litecoin

import (
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
)

// ValidateAddress ensures that address is valid and belongs to the given
// network.
func ValidateAddress(address, netName string) error {
	netParams, err := GetParams(netName)
	if err != nil {
		return errors.Errorf("unable to get net params: %v", err)
	}

	err = validate(address, netParams)
	if err != nil {
		// If there is error we shouldn't return it in mainnet straight away,
		// but instead because there is a possibility of address belong to the
		// legacy litecoin address type we should make another check.
		if netName == "mainnet" {
			legacyNetParams, err := GetParams("mainnet-legacy")
			if err != nil {
				return errors.Errorf("unable to get legacy mainnet params: %v", err)
			}

			return validate(address, legacyNetParams)
		}

		return err
	}

	// If validation was successful, than address is valid and we should
	// exit.
	return nil
}

func validate(address string, network *chaincfg.Params) error {
	decodedAddress, err := btcutil.DecodeAddress(address, network)
	if err != nil {
		return err
	}

	if !decodedAddress.IsForNet(network) {
		return errors.New("address is not for specified network")
	}

	return nil
}
