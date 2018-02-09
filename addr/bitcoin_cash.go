package addr

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/connector/chains/bitcoincash"
	"github.com/go-errors/errors"
	cashAddr "github.com/schancel/cashaddr-converter/address"
)

// validateBitcoinCash validates bitcoin cash address according net. Return error if address is invalid.
// Supports cashaddr and legacy addresses. Regtest net treated as testnet3.
func validateBitcoinCash(address string, network *chaincfg.Params) error {

	decodedAddress, err := btcutil.DecodeAddress(address, network)
	if err == nil && decodedAddress.IsForNet(network) {
		return nil
	}

	cashAddress, err := cashAddr.NewFromString(address)
	if err != nil {
		return errors.New("address neither legacy address nor CashAddr")
	}

	if cashAddrNetToInt(cashAddress.Network) != bitcoincashNetToInt(network) {
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

func bitcoincashNetToInt(network *chaincfg.Params) int {
	switch network.Net {
	case bitcoincash.Mainnet:
		return 0
	case bitcoincash.TestNet3:
		return 1
	case bitcoincash.TestNet: // corresponds to regtest
		return 1
	}
	return 2
}
