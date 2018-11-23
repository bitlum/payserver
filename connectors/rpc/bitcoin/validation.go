package bitcoin

import (
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

func DecodeAddress(address, netName string) (btcutil.Address, error) {
	netParams, err := GetParams(netName)
	if err != nil {
		return nil, errors.Errorf("unable to get net params: %v", err)
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
