package connectors

import (
	"io"
	"encoding/json"
	"io/ioutil"
)

// Serializable is an interface which defines a serializable
// object.
type Serializable interface {
	// Decode reads the bytes stream and converts it to the object.
	Decode(io.Reader, uint32) error

	// Encode converts object to the bytes stream and write it into the
	// writer.
	Encode(io.Writer, uint32) error
}

// Runtime check to ensure that BlockchainPendingDetails implements
// Serializable interface.
var _ Serializable = (*BlockchainPendingDetails)(nil)

// Decode reads the bytes stream and converts it to the object.
func (d *BlockchainPendingDetails) Decode(r io.Reader, v uint32) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, d)
}

// Encode converts object to the bytes stream and write it into the
// writer.
func (d *BlockchainPendingDetails) Encode(w io.Writer, v uint32) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// GeneratedTxDetails is the string form of signed blockchain transaction
// which.
type GeneratedTxDetails struct {
	// RawTx byte representation of blockchain transaction.
	RawTx []byte

	// TxID blockchain identification of transaction.
	TxID string
}

// Runtime check to ensure that BlockchainPendingDetails implements
// Serializable interface.
var _ Serializable = (*GeneratedTxDetails)(nil)

// Decode reads the bytes stream and converts it to the object.
func (d *GeneratedTxDetails) Decode(r io.Reader, v uint32) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, d)
}

// Encode converts object to the bytes stream and write it into the
// writer.
func (d *GeneratedTxDetails) Encode(w io.Writer, v uint32) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
