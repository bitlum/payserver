package litecoin

import (
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

// DecodeAddress ensures that address is valid and belongs to the given
// network, and return decoded address.
func DecodeAddress(address, netName string) (btcutil.Address, error) {
	netParams, err := GetParams(netName)
	if err != nil {
		return nil, errors.Errorf("unable to get net params: %v", err)
	}

	decodedAddress, err := decode(address, netParams)
	if err != nil {
		// If there is error we shouldn't return it in mainnet straight away,
		// but instead because there is a possibility of address belong to the
		// legacy litecoin address type we should make another check.
		if netName == "mainnet" {
			legacyNetParams, err := GetParams("mainnet-legacy")
			if err != nil {
				return nil, errors.Errorf("unable to get legacy mainnet params: %v", err)
			}

			return decode(address, legacyNetParams)
		}

		return nil, err
	}

	// If validation was successful, than address is valid and we should
	// exit.
	return decodedAddress, nil
}

func decode(address string, network *chaincfg.Params) (btcutil.Address, error) {
	decodedAddress, err := btcutil.DecodeAddress(address, network)
	if err != nil {
		return nil, err
	}

	if !decodedAddress.IsForNet(network) {
		return nil, errors.New("address is not for specified network")
	}

	return decodedAddress, nil
}
