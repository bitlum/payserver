package dash

import (
	"github.com/go-errors/errors"
	"github.com/btcsuite/btcutil"
)

// ValidateAddress ensures that address is valid and belongs to the given
// network.
func ValidateAddress(address, netName string) error {
	netParams, err := GetParams(netName)
	if err != nil {
		return errors.Errorf("unable  to get net params: %v", err)
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
