package bitcoind

import (
	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/connectors"
	"github.com/bitlum/connector/connectors/rpc"
	"github.com/bitlum/connector/db/sqlite"
	"github.com/bitlum/connector/metrics/crypto"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btclog"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"math"
	"testing"
	"time"
)

type response struct {
	data interface{}
	err  error
}

// ReplayRPCClient...
type ReplayRPCClient struct {
	responses chan response
	invoked   chan struct{}
	t         *testing.T
	delay     time.Duration
}

var _ rpc.Client = (*ReplayRPCClient)(nil)

func NewReplayRPCClient(t *testing.T) (*ReplayRPCClient) {
	return &ReplayRPCClient{
		responses: make(chan response, 100),
		invoked:   make(chan struct{}, 100),
		t:         t,
		delay:     time.Second * 1,
	}
}

func (c *ReplayRPCClient) respond(response response) error {
	select {
	case c.responses <- response:
		return nil
	case <-time.After(time.Second * 2):
		return errors.Errorf("waited more than 2 seconds for rpc call")
	}
}

func (c *ReplayRPCClient) GetBlockChainInfo() (*rpc.BlockChainInfoResp, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(*rpc.BlockChainInfoResp), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) GetBlockVerboseByHash(blockHash *chainhash.Hash) (
	*rpc.BlockVerboseResp, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(*rpc.BlockVerboseResp), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}
func (c *ReplayRPCClient) GetBestBlockHash() (*chainhash.Hash, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(*chainhash.Hash), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) UnlockUnspent() error {
	c.t.Log(common.GetFunctionName())
	return nil
}

func (c *ReplayRPCClient) LockUnspent(input rpc.UnspentInput) error {
	c.t.Log(common.GetFunctionName())
	return nil
}

func (c *ReplayRPCClient) ListUnspentMinMax(minConf, maxConf int) ([]rpc.UnspentInput, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.([]rpc.UnspentInput), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) GetAddressesByLabel(label string) ([]btcutil.Address, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.([]btcutil.Address), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) GetNewAddress(label string) (btcutil.Address, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(btcutil.Address), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) SignRawTransaction(tx *wire.MsgTx) (*wire.MsgTx,
	error) {
	c.t.Log(common.GetFunctionName())

	return tx, nil
}

func (c *ReplayRPCClient) CreateRawTransaction([]rpc.UnspentInput,
	map[btcutil.Address]btcutil.Amount) (*wire.MsgTx, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(*wire.MsgTx), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) SendRawTransaction(tx *wire.MsgTx) error {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.err
	case <-time.After(c.delay):
		return errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) GetBalanceByLabel(label string,
	minConfirms int) (btcutil.Amount, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(btcutil.Amount), resp.err
	case <-time.After(c.delay):
		return 0, errors.Errorf("response delay")
	}

}

func (c *ReplayRPCClient) GetTransaction(txHash *chainhash.Hash) (*rpc.Transaction, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(*rpc.Transaction), resp.err
	case <-time.After(c.delay):
		return nil, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) EstimateFee() (float64, error) {
	c.t.Log(common.GetFunctionName())

	select {
	case resp := <-c.responses:
		return resp.data.(float64), resp.err
	case <-time.After(c.delay):
		return 0, errors.Errorf("response delay")
	}
}

func (c *ReplayRPCClient) DaemonName() string {
	c.t.Log(common.GetFunctionName())
	return "mock"
}

func initConnector(
	client *ReplayRPCClient,
	paymentsStore connectors.PaymentsStore,
	stateStore connectors.StateStorage,
) (*Connector, error) {
	c, err := NewConnector(&Config{
		Net:                  "regtest",
		MinConfirmations:     1,
		SyncLoopDelay:        math.MaxInt16,
		InputReorgLoopDelay:  math.MaxInt16,
		SyncUnspentLoopDelay: math.MaxInt16,
		Asset:                connectors.BTC,
		Logger:               btclog.Disabled,
		Metrics:              crypto.DisabledBackend,
		LastSyncedBlockHash:  "",
		PaymentStore:         paymentsStore,
		StateStorage:         stateStore,
		FeePerByte:           1,
		RPCClient:            client,
	})
	if err != nil {
		return nil, errors.Errorf("unable create connector: %v", err)
	}

	// Return chain on getting information on start of connector
	{
		if err := client.respond(response{
			data: &rpc.BlockChainInfoResp{
				Chain: "regtest",
			},
			err: nil,
		}); err != nil {
			return nil, errors.Errorf("unable to respond: %v", err)
		}
	}

	// Return last synced block hash and last synced block on start of connector
	lastBlockHashStr := "00000000000000000023486b9d95a729418b2be569f35372b698bd8b308d1178"
	{
		lastBlockHash, err := chainhash.NewHashFromStr(lastBlockHashStr)
		if err != nil {
			return nil, errors.Errorf("non-correct hash: %v", err)
		}

		if err := client.respond(response{
			data: lastBlockHash,
			err:  nil,
		}); err != nil {
			return nil, errors.Errorf("unable to respond: %v", err)
		}

		if err := client.respond(response{
			data: &rpc.BlockVerboseResp{
				Hash:          lastBlockHashStr,
				Height:        1,
				NextHash:      "",
				Confirmations: 0,
				PreviousHash:  "",
				Tx:            []string{},
			},
			err: nil,
		}); err != nil {
			return nil, errors.Errorf("unable to respond: %v", err)
		}
	}

	// Return default account address on start of connector
	{
		address, err := decodeAddress(connectors.BTC,
			"2N6yeKrmeMFNkMWohbQCThx3djEfNhj3WYM", "regtest")
		if err != nil {
			return nil, errors.Errorf("unable to decode address: %v", err)
		}

		if err := client.respond(response{
			data: []btcutil.Address{address},
			err:  nil,
		}); err != nil {
			return nil, errors.Errorf("unable to respond: %v", err)
		}
	}

	if err := c.Start(); err != nil {
		return nil, errors.Errorf("unable to start connector: %v", err)
	}

	return c, nil
}

// какие юзкейсы мы хотим проверить?
// - платёж самому себе
// - реорганизационный платёж
// - отправка платежа
// - получение платежа
// - фии когда нет данных
// - фии когда есть данные
// - вероятно в будущем баланс, который не зависит от использованных аутпутов
// - как быть с балансом?
//	- надо хранить баланс в базе
//	- надо синкать его с настоящим и давать ворнинг если они не совпадают
// 	- при каждом исходящем платеже мы должны отнимать его
// 	- при каждом входящем платеже мы должны пополнять его
//	-
// 	- текущий баланс это количество отправленных
func TestProcessBlockCircularTransaction(t *testing.T) {
	db, clear, err := sqlite.MakeTestDB()
	if err != nil {
		t.Fatalf("unable create test db: %v", err)
	}
	defer clear()

	paymentsStore := &sqlite.PaymentsStore{DB: db}
	stateStore := sqlite.NewConnectorStateStorage(connectors.BTC, db)
	client := NewReplayRPCClient(t)

	c, err := initConnector(client, paymentsStore, stateStore)
	if err != nil {
		t.Fatalf("unabelt init connector: %v", err)
	}
	defer c.Stop("test end")

	// Return last synced block, in order to check confirmation number
	lastBlockHashStr := "00000000000000000023486b9d95a729418b2be569f35372b698bd8b308d1178"
	if err := client.respond(response{
		data: &rpc.BlockVerboseResp{
			Hash:          lastBlockHashStr,
			Height:        1,
			NextHash:      "00000000000000000016645564d1719f59c967975cec21d8c7d2941834aed36a",
			Confirmations: 3,
			Tx:            []string{},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	// Return next block, because confirmation of previous block is higher
	// than minimal required.
	if err := client.respond(response{
		data: &rpc.BlockVerboseResp{
			Hash:          "00000000000000000016645564d1719f59c967975cec21d8c7d2941834aed36a",
			Height:        2,
			NextHash:      "",
			Confirmations: 2,
			Tx: []string{
				"1aef1e640bf5dcca0d922f49afe2f171e1c5bd1517906cdc7507dca466c8e65e",
			},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	// Emulate circular transaction - from our address, to our another address.
	fee := -2.497e-05
	if err := client.respond(response{
		data: &rpc.Transaction{
			Amount:        0,
			Fee:           -2.497e-05,
			Confirmations: 1,
			TxID:          "1aef1e640bf5dcca0d922f49afe2f171e1c5bd1517906cdc7507dca466c8e65e",
			Details: []rpc.TransactionDetails{
				{
					Account:           "",
					Address:           "2N6tEa9BDvgue53LkS8B6BhGBnZEpbJZwhY",
					Amount:            -1,
					Category:          "send",
					InvolvesWatchOnly: false,
					Fee:               &fee,
					Vout:              0,
				},
				{
					Account:           "",
					Address:           "2NEEKkxXPRBGaY55NmxiCG6ksJ7LAxn4TfM",
					Amount:            -8.99997503,
					Category:          "send",
					InvolvesWatchOnly: false,
					Fee:               &fee,
					Vout:              1,
				},
				{
					Account:           "zigzag",
					Address:           "2N6tEa9BDvgue53LkS8B6BhGBnZEpbJZwhY",
					Amount:            1,
					Category:          "receive",
					InvolvesWatchOnly: false,
					Fee:               nil,
					Vout:              0,
				},
				{
					Account:           "",
					Address:           "2NEEKkxXPRBGaY55NmxiCG6ksJ7LAxn4TfM",
					Amount:            8.99997503,
					Category:          "receive",
					InvolvesWatchOnly: false,
					Fee:               nil,
					Vout:              1,
				},
			},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	if err := c.proceedNextBlock(); err != nil {
		t.Fatalf("unable to process blocks: %v", err)
	}

	payments, err := paymentsStore.ListPayments("", "", "", "", "")
	if err != nil {
		t.Fatalf("unable fetch payment: %v", err)
	}

	spew.Dump(payments)
}

func TestProcessBlockOutgoingTransaction(t *testing.T) {
	db, clear, err := sqlite.MakeTestDB()
	if err != nil {
		t.Fatalf("unable create test db: %v", err)
	}
	defer clear()

	paymentsStore := &sqlite.PaymentsStore{DB: db}
	stateStore := sqlite.NewConnectorStateStorage(connectors.BTC, db)
	client := NewReplayRPCClient(t)

	c, err := initConnector(client, paymentsStore, stateStore)
	if err != nil {
		t.Fatalf("unabelt init connector: %v", err)
	}
	defer c.Stop("test end")

	// In order to check behaviour of syncing the outgoing payment,
	// we have to emulate payment send, otherwise necessary database entries
	// will not be created and during block processing this transaction will
	// be skipped.

	// Fee rate client response
	if err := client.respond(response{
		data: 0.0001,
		err:  nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	if err := client.respond(response{
		data: []rpc.UnspentInput{
			{
				Address:       "2NEb8LSh3BLVuCaFSi3iM5PVi4ZhxDoizPV",
				Account:       "zigzag",
				Amount:        3,
				Confirmations: 10,
				TxID:          "7a99cafa3c53cdfaff44bbca4503e510a7c9dfac968d4b05760070881dec8ff8",
				Vout:          0,
			},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	address, err := decodeAddress(connectors.BTC,
		"2N6yeKrmeMFNkMWohbQCThx3djEfNhj3WYM", "regtest")
	if err != nil {
		t.Fatalf("unable to decode address: %v", err)
	}

	if err := client.respond(response{
		data: address,
		err:  nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	if err := client.respond(response{
		data: &wire.MsgTx{
			Version:  1,
			TxIn:     nil,
			TxOut:    nil,
			LockTime: 0,
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	payment, err := c.CreatePayment("2N6tEa9BDvgue53LkS8B6BhGBnZEpbJZwhY", "1")
	if err != nil {
		t.Fatalf("unable create payment: %v", err)
	}

	if _, err := c.SendPayment(payment.PaymentID); err != nil {
		t.Fatalf("unable create payment: %v", err)
	}

	// Return last synced block, in order to check confirmation number
	lastBlockHashStr := "00000000000000000023486b9d95a729418b2be569f35372b698bd8b308d1178"
	if err := client.respond(response{
		data: &rpc.BlockVerboseResp{
			Hash:          lastBlockHashStr,
			Height:        1,
			NextHash:      "00000000000000000016645564d1719f59c967975cec21d8c7d2941834aed36a",
			Confirmations: 3,
			Tx:            []string{},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	// Return next block, because confirmation of previous block is higher
	// than minimal required.
	if err := client.respond(response{
		data: &rpc.BlockVerboseResp{
			Hash:          "00000000000000000016645564d1719f59c967975cec21d8c7d2941834aed36a",
			Height:        2,
			NextHash:      "",
			Confirmations: 2,
			Tx: []string{
				"1aef1e640bf5dcca0d922f49afe2f171e1c5bd1517906cdc7507dca466c8e65e",
			},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	// Emulate circular transaction - from our address, to our another address.
	fee := -2.497e-05
	if err := client.respond(response{
		data: &rpc.Transaction{
			Amount:        -1,
			Fee:           -2.497e-05,
			Confirmations: 1,
			TxID:          "3ef7aee9fecb3a7f546fe2f6ad05899fe67a5017437cec4051c64cea51a717fa",
			Details: []rpc.TransactionDetails{
				{
					Account:           "",
					Address:           "2N28ActFCgieyTdmMeui5GzgQhXapj55Fhe",
					Amount:            -1,
					Category:          "send",
					InvolvesWatchOnly: false,
					Fee:               &fee,
					Vout:              0,
				},
				{
					Account:           "",
					Address:           "2N6yeKrmeMFNkMWohbQCThx3djEfNhj3WYM",
					Amount:            -7.99995006,
					Category:          "send",
					InvolvesWatchOnly: false,
					Fee:               &fee,
					Vout:              1,
				},
				{
					Account:           "",
					Address:           "2N6yeKrmeMFNkMWohbQCThx3djEfNhj3WYM",
					Amount:            7.99995006,
					Category:          "receive",
					InvolvesWatchOnly: false,
					Fee:               nil,
					Vout:              1,
				},
			},
		},
		err: nil,
	}); err != nil {
		t.Fatalf("unable to respond: %v", err)
	}

	if err := c.proceedNextBlock(); err != nil {
		t.Fatalf("unable to process blocks: %v", err)
	}

	payments, err := paymentsStore.ListPayments("", "", "", "", "")
	if err != nil {
		t.Fatalf("unable fetch payment: %v", err)
	}

	// Should be one 2 internal payments, 1 external
	spew.Dump(payments)
}

// (rpc.UnspentInput) {
//  Address: (string) (len=35) "2N6tEa9BDvgue53LkS8B6BhGBnZEpbJZwhY",
//  Account: (string) (len=6) "zigzag",
//  Amount: (float64) 1,
//  Confirmations: (int64) 0,
//  TxID: (string) (len=64) "7a99cafa3c53cdfaff44bbca4503e510a7c9dfac968d4b05760070881dec8ff8",
//  Vout: (uint32) 0
// },
// (rpc.UnspentInput) {
//  Address: (string) (len=35) "2NEEKkxXPRBGaY55NmxiCG6ksJ7LAxn4TfM",
//  Account: (string) "",
//  Amount: (float64) 8.99997503,
//  Confirmations: (int64) 0,
//  TxID: (string) (len=64) "7a99cafa3c53cdfaff44bbca4503e510a7c9dfac968d4b05760070881dec8ff8",
//  Vout: (uint32) 1
// }
//}

// Transaction to someone else (to 2N28ActFCgieyTdmMeui5GzgQhXapj55Fhe)
//
// (rpc.UnspentInput) {
//  Address: (string) (len=35) "2N6yeKrmeMFNkMWohbQCThx3djEfNhj3WYM",
//  Account: (string) "",
//  Amount: (float64) 7.99995006,
//  Confirmations: (int64) 0,
//  TxID: (string) (len=64) "3ef7aee9fecb3a7f546fe2f6ad05899fe67a5017437cec4051c64cea51a717fa",
//  Vout: (uint32) 1
// }
//}

// Transaction to us from someone else:
//
//(*rpc.Transaction)(0xc42073e780)({
// Amount: (float64) 3,
// Fee: (float64) 0,
// Confirmations: (int64) 1,
// TxID: (string) (len=64) "49c3287d6c503934e93529723d5f6cc4459ae1c7ed078d9a095d7a22267bf097",
// Details: ([]rpc.TransactionDetails) (len=1 cap=1) {
//  (rpc.TransactionDetails) {
//   Account: (string) (len=6) "zigzag",
//   Address: (string) (len=35) "2NEb8LSh3BLVuCaFSi3iM5PVi4ZhxDoizPV",
//   Amount: (float64) 3,
//   Category: (string) (len=7) "receive",
//   InvolvesWatchOnly: (bool) false,
//   Fee: (*float64)(<nil>),
//   Vout: (uint32) 1
//  }
// }
//})
//
//
//([]rpc.UnspentInput) (len=1 cap=1) {
// (rpc.UnspentInput) {
//  Address: (string) (len=35) "2NEb8LSh3BLVuCaFSi3iM5PVi4ZhxDoizPV",
//  Account: (string) (len=6) "zigzag",
//  Amount: (float64) 1,
//  Confirmations: (int64) 0,
//  TxID: (string) (len=64) "5d00a0a2d75ddfc36901d0afd95fd16b1f508b060e92ff61fd383e8cde585282",
//  Vout: (uint32) 0
// }
//}

// Transaction
