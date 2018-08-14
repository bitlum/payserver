package bitcoind

import (
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
)

const (
	// CommitWeight is the weight of the base commitment transaction which
	// includes: one p2wsh input, out p2wkh output, and one p2wsh output.
	CommitWeight int64 = 724
)

const (
	// witnessScaleFactor determines the level of "discount" witness data
	// receives compared to "base" data. A scale factor of 4, denotes that
	// witness data is 1/4 as cheap as regular non-witness data. Value copied
	// here for convenience.
	witnessScaleFactor = blockchain.WitnessScaleFactor

	// The weight(weight), which is different from the !size! (see BIP-141),
	// is calculated as:
	// Weight = 4 * BaseSize + WitnessSize (weight).
	// BaseSize - size of the transaction without witness data (bytes).
	// WitnessSize - witness size (bytes).
	// Weight - the metric for determining the weight of the transaction.

	// P2WPKHSize 22 bytes
	//	- OP_0: 1 byte
	//	- OP_DATA: 1 byte (PublicKeyHASH160 length)
	//	- PublicKeyHASH160: 20 bytes
	P2WPKHSize = 1 + 1 + 20

	// P2WSHSize 34 bytes
	//	- OP_0: 1 byte
	//	- OP_DATA: 1 byte (WitnessScriptSHA256 length)
	//	- WitnessScriptSHA256: 32 bytes
	P2WSHSize = 1 + 1 + 32

	// P2PKHOutputSize 34 bytes
	//      - value: 8 bytes
	//      - var_int: 1 byte (pkscript_length)
	//      - pkscript (p2pkh): 25 bytes
	P2PKHOutputSize = 8 + 1 + 25

	// P2WKHOutputSize 31 bytes
	//      - value: 8 bytes
	//      - var_int: 1 byte (pkscript_length)
	//      - pkscript (p2wpkh): 22 bytes
	P2WKHOutputSize = 8 + 1 + P2WPKHSize

	// P2WSHOutputSize 43 bytes
	//      - value: 8 bytes
	//      - var_int: 1 byte (pkscript_length)
	//      - pkscript (p2wsh): 34 bytes
	P2WSHOutputSize = 8 + 1 + P2WSHSize

	// P2SHOutputSize 32 bytes
	//      - value: 8 bytes
	//      - var_int: 1 byte (pkscript_length)
	//      - pkscript (p2sh): 23 bytes
	P2SHOutputSize = 8 + 1 + 23

	// P2PKHScriptSigSize 108 bytes
	//      - OP_DATA: 1 byte (signature length)
	//      - signature
	//      - OP_DATA: 1 byte (pubkey length)
	//      - pubkey
	P2PKHScriptSigSize = 1 + 73 + 1 + 33

	// P2WKHWitnessSize 109 bytes
	//      - number_of_witness_elements: 1 byte
	//      - signature_length: 1 byte
	//      - signature
	//      - pubkey_length: 1 byte
	//      - pubkey
	P2WKHWitnessSize = 1 + 1 + 73 + 1 + 33

	// InputSize 41 bytes
	//	- PreviousOutPoint:
	//		- Hash: 32 bytes
	//		- Index: 4 bytes
	//	- OP_DATA: 1 byte (ScriptSigLength)
	//	- ScriptSig: 0 bytes
	//	- Witness <----	we use "Witness" instead of "ScriptSig" for
	// 			transaction validation, but "Witness" is stored
	// 			separately and weight for it size is smaller. So
	// 			we separate the calculation of ordinary data
	// 			from witness data.
	//	- Sequence: 4 bytes
	InputSize = 32 + 4 + 1 + 4

	// WitnessHeaderSize 2 bytes
	//	- Flag: 1 byte
	//	- Marker: 1 byte
	WitnessHeaderSize = 1 + 1

	// BaseTxSize 8 bytes
	// - Version: 4 bytes
	// - LockTime: 4 bytes
	BaseTxSize = 4 + 4
)

// TxWeightEstimator is able to calculate weight estimates for transactions
// based on the input and output types. For purposes of estimation, all
// signatures are assumed to be of the maximum possible size, 73 bytes.
type TxWeightEstimator struct {
	hasWitness       bool
	inputCount       uint32
	outputCount      uint32
	inputSize        int
	inputWitnessSize int
	outputSize       int
}

// AddP2PKHInput updates the weight estimate to account for an additional input
// spending a P2PKH output.
func (twe *TxWeightEstimator) AddP2PKHInput() {
	twe.inputSize += InputSize + P2PKHScriptSigSize
	twe.inputWitnessSize++
	twe.inputCount++
}

// AddP2WKHInput updates the weight estimate to account for an additional input
// spending a native P2PWKH output.
func (twe *TxWeightEstimator) AddP2WKHInput() {
	twe.AddWitnessInput(P2WKHWitnessSize)
}

// AddWitnessInput updates the weight estimate to account for an additional
// input spending a native pay-to-witness output. This accepts the total size
// of the witness as a parameter.
func (twe *TxWeightEstimator) AddWitnessInput(witnessSize int) {
	twe.inputSize += InputSize
	twe.inputWitnessSize += witnessSize
	twe.inputCount++
	twe.hasWitness = true
}

// AddNestedP2WKHInput updates the weight estimate to account for an additional
// input spending a P2SH output with a nested P2WKH redeem script.
func (twe *TxWeightEstimator) AddNestedP2WKHInput() {
	twe.inputSize += InputSize + P2WPKHSize
	twe.inputWitnessSize += P2WKHWitnessSize
	twe.inputSize++
	twe.hasWitness = true
}

// AddNestedP2WSHInput updates the weight estimate to account for an additional
// input spending a P2SH output with a nested P2WSH redeem script.
func (twe *TxWeightEstimator) AddNestedP2WSHInput(witnessSize int) {
	twe.inputSize += InputSize + P2WSHSize
	twe.inputWitnessSize += witnessSize
	twe.inputSize++
	twe.hasWitness = true
}

// AddP2PKHOutput updates the weight estimate to account for an additional P2PKH
// output.
func (twe *TxWeightEstimator) AddP2PKHOutput() {
	twe.outputSize += P2PKHOutputSize
	twe.outputCount++
}

// AddP2WKHOutput updates the weight estimate to account for an additional
// native P2WKH output.
func (twe *TxWeightEstimator) AddP2WKHOutput() {
	twe.outputSize += P2WKHOutputSize
	twe.outputCount++
}

// AddP2WSHOutput updates the weight estimate to account for an additional
// native P2WSH output.
func (twe *TxWeightEstimator) AddP2WSHOutput() {
	twe.outputSize += P2WSHOutputSize
	twe.outputCount++
}

// AddP2SHOutput updates the weight estimate to account for an additional P2SH
// output.
func (twe *TxWeightEstimator) AddP2SHOutput() {
	twe.outputSize += P2SHOutputSize
	twe.outputCount++
}

// Weight gets the estimated weight of the transaction.
func (twe *TxWeightEstimator) Weight() int {
	txSizeStripped := BaseTxSize +
		wire.VarIntSerializeSize(uint64(twe.inputCount)) + twe.inputSize +
		wire.VarIntSerializeSize(uint64(twe.outputCount)) + twe.outputSize
	weight := txSizeStripped * witnessScaleFactor
	if twe.hasWitness {
		weight += WitnessHeaderSize + twe.inputWitnessSize
	}
	return weight
}
