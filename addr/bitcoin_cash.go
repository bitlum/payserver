package addr

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/connector/connectors/chains/bitcoincash"
	"github.com/go-errors/errors"
	cashAddr "github.com/schancel/cashaddr-converter/address"
)

// validateBitcoinCash validates bitcoin cash address according net.
// Supports cashaddr and legacy addresses. Regtest net treated as testnet3.
func validateBitcoinCash(address string, network *chaincfg.Params) error {

	// Check that address is valid, but if it return errors,
	// continue validation because it might be special "Cash Address" type.
	decodedAddress, err := btcutil.DecodeAddress(address, network)
	if err == nil && decodedAddress.IsForNet(network) {
		return nil
	}

	cashAddress, err := cashAddr.NewFromString(address)
	if err != nil {
		return errors.New("address neither legacy address nor cash addr")
	}

	if cashAddrNetToInt(cashAddress.Network) != bitcoinCashNetToInt(network) {
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
	case bitcoincash.Mainnet:
		return 0
	case bitcoincash.TestNet3:
		return 1
	case bitcoincash.TestNet:
		return 1
	}
	return 2
}
