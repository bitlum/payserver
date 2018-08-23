package geth

import (
	"fmt"
	"testing"

	"math/big"

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

	_, _, err = client.EthSignTransaction(ethrpc.T{
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
}
