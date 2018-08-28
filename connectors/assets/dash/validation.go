package dash

import (
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcutil"
)

// DecodeAddress ensures that address is valid and belongs to the given
// network, returns decoded address.
func DecodeAddress(address, netName string) (btcutil.Address, error) {
	netParams, err := GetParams(netName)
	if err != nil {
		return nil, errors.Errorf("unable  to get net params: %v", err)
	}

	decodedAddress, err := btcutil.DecodeAddress(address, netParams)
	if err != nil {
		return nil, err
	}

	if !decodedAddress.IsForNet(netParams) {
		return nil, errors.New("address is not for specified network")
	}

	return decodedAddress, nil
}
