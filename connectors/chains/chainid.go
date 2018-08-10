package chains

import "github.com/bitlum/btcd/wire"

const (
	// ChainID prefixes were created to distinguish different chance during
	// the process of registration with mustRegister function.
	BitcoinCashChainIDPrefix wire.BitcoinNet = 1
	DashChainIDPrefix        wire.BitcoinNet = 2
	LitecoinChainIDPrefix    wire.BitcoinNet = 3

	// LitecoinLegaceChainIDPrefix were created specifically because of the
	// legacy P2PKH address in order to achieve proper address validation
	// with DecodeAddress function.
	LitecoinLegacyChainIDPrefix wire.BitcoinNet = 4
)
