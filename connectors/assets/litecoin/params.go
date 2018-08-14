package litecoin

import (
	"github.com/btcsuite/btcd/chaincfg"
	bwire "github.com/btcsuite/btcd/wire"
	lwire "github.com/ltcsuite/ltcd/wire"
	"github.com/go-errors/errors"
)

var (
	// chainIDPrefix is created to distinguish different chains during
	// the process of registration with btcutil mustRegister function.
	//
	// NOTE: This is needed because of the fact how btcutil DecodeAddress works,
	// it couldn't proper decode address if its networks wasn't previously
	// registered.
	chainIDPrefix bwire.BitcoinNet = 3

	// legacyChainIDPrefix were created specifically because of the
	// legacy P2PKH address in order to achieve proper address validation
	// with DecodeAddress function.
	legacyChainIDPrefix bwire.BitcoinNet = 4
)

var (
	// Mainnet represents the main network.
	Mainnet = bwire.MainNet + chainIDPrefix

	// MainnetLegacy represents legacy main network with legacy P2SH address prefix
	MainnetLegacy = bwire.MainNet + legacyChainIDPrefix

	// TestNet represents the regression network.
	TestNet = bwire.TestNet + chainIDPrefix

	// TestNet4 represents the test network.
	TestNet4 = bwire.BitcoinNet(lwire.TestNet4) + chainIDPrefix
)

// With ScriptHashAddrID=SCRIPT_ADDRESS https://github.com/litecoin-project/litecoin/blob/master/src/chainparams.cpp#L237
// It have new P2SH prefix
var MainNetParams = chaincfg.Params{
	Net:  Mainnet,
	Name: "mainnet",

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "ltc", // always ltc for main net

	PubKeyHashAddrID:        0x30, // starts with L
	ScriptHashAddrID:        0x50, // starts with M
	PrivateKeyID:            0xB0, // starts with 6 (uncompressed) or T (compressed)
	WitnessPubKeyHashAddrID: 0x06, // starts with p2
	WitnessScriptHashAddrID: 0x0A, // starts with 7Xh

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
}

// MainNetParamsLegacy was created to distinguish the two types
// of address litecoin addresses, in this case we use legacy script hash addr
// id, which corresponds to the Bitcoin network. For more information read:
// https://github.com/litecoin-project/litecoin/blob/master/src/chainparams.cpp#L237
// It have legacy P2SH prefix.
var MainNetParamsLegacy = chaincfg.Params{
	Net:  MainnetLegacy,
	Name: "mainnet-legacy",

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "ltc", // always ltc for main net

	PubKeyHashAddrID:        0x30, // starts with L
	ScriptHashAddrID:        0x5,  // starts with 3
	PrivateKeyID:            0xB0, // starts with 6 (uncompressed) or T (compressed)
	WitnessPubKeyHashAddrID: 0x06, // starts with p2
	WitnessScriptHashAddrID: 0x0A, // starts with 7Xh

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
}

var TestNet4Params = chaincfg.Params{
	Net:  TestNet4,
	Name: "testnet4",

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "tltc", // always tb for test net

	// Address encoding magics
	PubKeyHashAddrID:        0x6f, // starts with m or n
	ScriptHashAddrID:        0xc4, // starts with 2
	WitnessPubKeyHashAddrID: 0x52, // starts with QW
	WitnessScriptHashAddrID: 0x31, // starts with T7n
	PrivateKeyID:            0xef, // starts with 9 (uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
}

// RegressionNetParams defines the network parameters for the regression test
// Dash network.  Not to be confused with the test Dash network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = chaincfg.Params{
	Net:  TestNet,
	Name: "regtest",

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "tltc", // always tltc for test net

	// Address encoding magics
	PubKeyHashAddrID: 0x6f, // starts with m or n
	ScriptHashAddrID: 0xc4, // starts with 2
	PrivateKeyID:     0xef, // starts with 9 (uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
}

// mustRegister performs the same function as Register except it panics if there
// is an error.  This should only be called from package init functions.
func mustRegister(params *chaincfg.Params) {
	if err := chaincfg.Register(params); err != nil &&
		err != chaincfg.ErrDuplicateNet {
		panic("failed to register network: " + err.Error())
	}
}

func init() {
	mustRegister(&MainNetParams)
	mustRegister(&MainNetParamsLegacy)
	mustRegister(&TestNet4Params)
	mustRegister(&RegressionNetParams)
}

func GetParams(netName string) (*chaincfg.Params, error) {
	switch netName {
	case "mainnet", "main":
		return &MainNetParams, nil
	case "mainnet-legacy":
		return &MainNetParamsLegacy, nil
	case "regtest", "simnet":
		return &RegressionNetParams, nil
	case "testnet4", "test", "testnet":
		return &TestNet4Params, nil
	}

	return nil, errors.Errorf("network '%s' is "+
		"invalid or unsupported", netName)
}
