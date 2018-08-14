package bitcoincash

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	cashAddr "github.com/schancel/cashaddr-converter/address"
)

// ValidateAddress validates bitcoin cash address according net.
// Supports cashaddr and legacy addresses.
//
// NOTE: Regtest net treated as testnet 3.Ò
func ValidateAddress(address, network string) error {
	netParams, err := GetParams(network)
	if err != nil {
		return errors.Errorf("unable to get net params: %v", err)
	}

	// Check that address is valid, but if it return errors,
	// continue validation because it might be special "Cash Address" type.
	decodedAddress, err := btcutil.DecodeAddress(address, netParams)
	if err == nil && decodedAddress.IsForNet(netParams) {
		return nil
	}

	cashAddress, err := cashAddr.NewFromString(address)
	if err != nil {
		return errors.New("address neither legacy address nor cash addr")
	}

	if cashAddrNetToInt(cashAddress.Network) != bitcoinCashNetToInt(netParams) {
		return errors.New("address is not for specified network")
	}

	return nil
}

func cashAddrNetToInt(networkType cashAddr.NetworkType) int {
	switch networkType {
	case cashAddr.MainNet:
		return 0
	case cashAddr.TestNet:
		return 1
	case cashAddr.RegTest:
		return 1
	}
	return 2
}

func bitcoinCashNetToInt(network *chaincfg.Params) int {
	switch network.Net {
	case Mainnet:
		return 0
	case TestNet3:
		return 1
	case TestNet:
		return 1
	}
	return 2
}
