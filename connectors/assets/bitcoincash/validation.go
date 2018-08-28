package bitcoincash

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	cashAddr "github.com/schancel/cashaddr-converter/address"
)

// DecodeAddress validates bitcoin cash address according net,
// and return btcutil compatible legacy bitcoin address. Supports cashaddr and
// legacy addresses.
//
// NOTE: Regtest net treated as testnet 3.Ò
func DecodeAddress(address, network string) (btcutil.Address, error) {
	netParams, err := GetParams(network)
	if err != nil {
		return nil, errors.Errorf("unable to get net params: %v", err)
	}

	// Check that address is valid, but if it return errors,
	// continue validation because it might be special "Cash Address" type.
	decodedAddress, err := btcutil.DecodeAddress(address, netParams)
	if err == nil {
		if decodedAddress.IsForNet(netParams) {
			return decodedAddress, nil
		} else {
			return nil, errors.New("address is not for specified network")
		}
	}

	cashAddress, err := cashAddr.NewFromString(address)
	if err != nil {
		return nil, errors.New("address neither legacy address nor cash addr")
	}

	legacyAddress, err := cashAddress.Legacy()
	if err != nil {
		return nil, errors.Errorf("unable convert to legacy address: %v", err)
	}

	address, err = legacyAddress.Encode()
	if err != nil {
		return nil, errors.Errorf("unable encode legacy address: %v", err)
	}

	decodedAddress, err = btcutil.DecodeAddress(address, netParams)
	if err == nil {
		if decodedAddress.IsForNet(netParams) {
			return decodedAddress, nil
		} else {
			return nil, errors.New("address is not for specified network")
		}
	}

	return decodedAddress, nil
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
