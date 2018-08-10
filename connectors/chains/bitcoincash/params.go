package bitcoincash

import (
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcd/wire"
	"github.com/bitlum/connector/connectors/chains"
)

var (
	// Mainnet represents the main test network.
	Mainnet = wire.MainNet + chains.BitcoinCashChainIDPrefix

	// TestNet represents the regression test network.
	TestNet = wire.TestNet + chains.BitcoinCashChainIDPrefix

	// TestNet3 represents the test network.
	TestNet3 = wire.TestNet3 + chains.BitcoinCashChainIDPrefix
)

// MainNetParams defines the network parameters for the main network.
var MainNetParams = chaincfg.Params{
	Net:              Mainnet,
	Name:             "mainnet",
	PubKeyHashAddrID: 0,   // addresses start with 'X'
	ScriptHashAddrID: 5,   // script addresses start with '7'
	PrivateKeyID:     128, // private keys start with '7' or 'X'

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
}

// TestNet3Params defines the network parameters for the test network
// (version 3).  Not to be confused with the regression test network, this
// network is sometimes simply called "testnet".
var TestNet3Params = chaincfg.Params{
	Net:              TestNet3,
	Name:             "testnet3",
	PubKeyHashAddrID: 111,
	ScriptHashAddrID: 196,
	PrivateKeyID:     239,

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xCF},
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
}

// RegressionNetParams defines the network parameters for the regression test
// Dash network.  Not to be confused with the test network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = chaincfg.Params{
	Net:              TestNet,
	Name:             "regtest",
	PubKeyHashAddrID: 111,
	ScriptHashAddrID: 196,
	PrivateKeyID:     239,

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xCF},
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
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
	mustRegister(&TestNet3Params)
	mustRegister(&RegressionNetParams)
}
