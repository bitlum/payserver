package dash

import (
	"github.com/bitlum/connector/chains"
	"github.com/bitlum/btcd/chaincfg"
	"github.com/bitlum/btcd/wire"
)

var (
	// Mainnet represents the main network.
	Mainnet = wire.MainNet + chains.DashChainID

	// TestNet represents the regression network.
	TestNet = wire.TestNet + chains.DashChainID

	// TestNet3 represents the test network.
	TestNet3 = wire.TestNet3 + chains.DashChainID
)

var MainNetParams = chaincfg.Params{
	Net:              Mainnet,
	PubKeyHashAddrID: 76,  // addresses start with 'X'
	ScriptHashAddrID: 16,  // script addresses start with '7'
	PrivateKeyID:     204, // private keys start with '7' or 'X'

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
}

var TestNet3Params = chaincfg.Params{
	Net:              TestNet3,
	PubKeyHashAddrID: 140, // addresses start with 'y'
	ScriptHashAddrID: 19,  // script addresses start with '8' or '9'
	PrivateKeyID:     239, // private keys start with '9' or 'c' (Bitcoin defaults)

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xCF},
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
}

// RegressionNetParams defines the network parameters for the regression test
// Dash network.  Not to be confused with the test Dash network (version
// 3), this network is sometimes simply called "testnet".
var RegressionNetParams = chaincfg.Params{
	Net:              TestNet,
	PubKeyHashAddrID: 140, // addresses start with 'y'
	ScriptHashAddrID: 19,  // script addresses start with '8' or '9'
	PrivateKeyID:     239, // private keys start with '9' or 'c' (Bitcoin defaults)

	// BIP32 hierarchical deterministic extended key magics
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xCF},
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94},
}
