package geth

import (
	"fmt"
	"testing"

	"math/big"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/onrik/ethrpc"
)

func TestClient(t *testing.T) {
	client := ExtendedEthRpc{ethrpc.NewEthRPC("http://165.227.118.113:20306")}

	version, err := client.Web3ClientVersion()
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	number, err := client.EthNewBlockFilter()
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	pversion, err := client.EthProtocolVersion()
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	nversion, err := client.NetVersion()
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	account, err := client.PersonalNewAddress("kek")
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	fmt.Println("version:", version)
	fmt.Println("number:", number)
	fmt.Println("pversion:", pversion)
	fmt.Println("nversion:", nversion)
	fmt.Println("account:", account)
}

func TestSignTransaction(t *testing.T) {
	client := ExtendedEthRpc{ethrpc.NewEthRPC("http://165.227.118.113:20306")}

	address, err := client.PersonalNewAddress("kek")
	if err != nil {
		t.Fatalf("unable to get version: %v", err)
	}

	_, err = client.PersonalUnlockAddress(address, "kek", 2)
	if err != nil {
		t.Fatalf("unable to unlock account: %v", err)
	}

	tx, rawTx, err := client.EthSignTransaction(ethrpc.T{
		From:     address,
		To:       address,
		Gas:      21000,
		GasPrice: big.NewInt(21),
		Value:    big.NewInt(1000),
		Data:     "",
		Nonce:    0,
	})
	if err != nil {
		t.Fatalf("unable to sign tx: %v", err)
	}

	spew.Dump(tx)
	spew.Dump(rawTx)
}

func TestTransactionDecode(t *testing.T) {

	data := common.FromHex(
		"0xf8658015825208948e643bc825bdb44dfecfdeebecbbe3f2a4d5d71b8203e88084687e1699a0224c8101372cb765e1e1c31607d149e6d06c73c97865e84de7ed37de931c5519a020ecd6fbb48368d36320967641d22703b0c649f0320fb78dce4218b6684667fb")

	spew.Dump(data)

	var tx types.Transaction
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		t.Fatal(err)
	}

	spew.Dump(tx.ChainId(), tx)

	chainID := big.NewInt(876546889)
	signer := types.NewEIP155Signer(chainID)
	spew.Dump(types.Sender(signer, &tx))
}
