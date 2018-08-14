package bitcoin

import (
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

func ValidateAddress(address, netName string) error {
	netParams, err := GetParams(netName)
	if err != nil {
		return errors.Errorf("unable to get net params: %v", err)
	}

	decodedAddress, err := btcutil.DecodeAddress(address, netParams)
	if err != nil {
		return err
	}

	if !decodedAddress.IsForNet(netParams) {
		return errors.New("address is not for specified network")
	}

	return nil
}
