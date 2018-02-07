package core

import (
	"testing"

	"math/rand"
	"time"

	"sync"

	"github.com/pkg/errors"
)

func asyncOrderCancel(engine *Engine, orders []OrderDetailedInfo) {
	var wg sync.WaitGroup

	limitChan := make(chan struct{}, 50)
	for i := 0; i < 50; i++ {
		limitChan <- struct{}{}
	}

	for _, order := range orders {
		<-limitChan
		wg.Add(1)
		go func(order OrderDetailedInfo) {
			defer func() {
				wg.Done()
				limitChan <- struct{}{}
			}()

			req := &OrderCancelRequest{
				UserID:  order.UserID,
				Market:  order.Market.String(),
				OrderID: order.OrderID,
			}

			if _, err := engine.OrderCancel(req); err != nil {
				return
			}
		}(order)
	}

	wg.Wait()
}

// getEmptyProfile is a helper function which search user which haven't been
// used in the system yet.
func getEmptyProfile(engine *Engine) (uint32, error) {
	var userID uint32
	for {
		userID = rand.Uint32()
		req := &BalanceQueryRequest{
			UserID: userID,
		}

		if _, err := engine.BalanceQuery(req); err != nil {
			if err, ok := err.(*Error); ok {
				continue
			} else {
				return 0, errors.Errorf("unable to get user balance: %s",
					err)
			}
		}

		break
	}

	return userID, nil
}

// cancelOrders is a helper function which is used during test environment
// setup which cancels all orders on the exchange.
func cancelOrders(engine *Engine) error {
	cancelOrders := func(market MarketType, side MarketOrderSide) error {
		offset := int32(0)
		var orders []OrderDetailedInfo

		for {
			resp, err := engine.OrderBook(&OrderBookRequest{
				Market: market.String(),
				Side:   side,
				Offset: offset,
				Limit:  MaxLimit,
			})
			if err != nil {
				return err
			}

			orders = append(orders, resp.Orders...)

			l := int32(len(resp.Orders))
			offset += l
			if l < MaxLimit {
				break
			}
		}

		asyncOrderCancel(engine, orders)
		return nil
	}

	for _, market := range AllMarkets {
		cancelOrders(market, MarketOrderSideAsk)
		cancelOrders(market, MarketOrderSideBid)
	}

	return nil
}

// setUp is a helper function which is used to setup integration test
// environment. With this function we get two users which not exist in system
// yet, thereby using clear environment. Also in order to not interfere with
// other users activity we have to cancel all orders.
func setUp(engine *Engine) (*testContext, error) {
	var err error

	firstUserID, err := getEmptyProfile(engine)
	if err != nil {
		return nil, err
	}

	secondUserID, err := getEmptyProfile(engine)
	if err != nil {
		return nil, err
	}

	if firstUserID == secondUserID {
		return nil, errors.New("user ids shouldn't be equal")
	}

	if err := cancelOrders(engine); err != nil {
		return nil, errors.Errorf("unable to cancel orders: %v", err)
	}

	return &testContext{
		engine:       engine,
		firstUserID:  firstUserID,
		secondUserID: secondUserID,
	}, nil
}

type testContext struct {
	firstUserID  uint32
	secondUserID uint32
	engine       *Engine
}

// testUpdate is an integration test which checks the ability of exchange engine
// update user balances.
func testUpdate(t *testing.T, ctx *testContext) {
	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      AssetBTC,
			ActionType: ActionDeposit,
			ActionID:   0,
			Change:     "10",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to deposit: %s", err)
		}
	}

	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      AssetBTC,
			ActionType: ActionWithdrawal,
			ActionID:   0,
			Change:     "-10",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to withdraw: %s", err)
		}
	}

	{
		req := &BalanceQueryRequest{
			UserID: ctx.firstUserID,
		}

		resp, err := ctx.engine.BalanceQuery(req)
		if err != nil {
			t.Fatalf("unable to get user balance: %s", err)
		}

		if resp[AssetBTC].Available != "0" {
			t.Fatal("balance is wrong")
		}
	}
}

// testPutOrder is an integration test which checks the ability of exchange
// engine to put order of limit and market types.
func testPutOrder(t *testing.T, ctx *testContext) {

	market := MarketBTCETC

	// Update fist and second user balances in order to be able to put orders.
	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      market.Stock,
			ActionType: ActionDeposit,
			ActionID:   0,
			Change:     "1.0",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to deposit btc: %s", err)
		}
	}
	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.secondUserID,
			Asset:      market.Money,
			ActionType: ActionDeposit,
			ActionID:   0,
			Change:     "2.0",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to deposit ltc: %s", err)
		}
	}

	// Put limit order which specifies the exact price for which users want
	// to sell his BTC and its amount.
	{
		req := &OrderPutLimitRequest{
			UserID:       ctx.firstUserID,
			Market:       market.String(),
			Side:         MarketOrderSideAsk,
			Amount:       "1.0",
			Price:        "2.0",
			TakerFeeRate: "0.1",
			MakerFeeRate: "0.1",
			Source:       "testPutOrder",
		}

		if _, err := ctx.engine.OrderPutLimit(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}

	// Put market order which wil be executed with current price of the
	// market. In this case users specifies the amount of CNY which he want's
	// to spent on buying the BTC.
	{
		req := &OrderPutMarketRequest{
			UserID:       ctx.secondUserID,
			Market:       market.String(),
			Side:         MarketOrderSideBid,
			Amount:       "2.0",
			TakerFeeRate: "0.1",
			Source:       "testPutOrder",
		}

		if _, err := ctx.engine.OrderPutMarket(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}

	// Check that balances of the users have been switched.
	{
		req := &BalanceQueryRequest{
			UserID: ctx.firstUserID,
		}

		resp, err := ctx.engine.BalanceQuery(req)
		if err != nil {
			t.Fatalf("unable to get first user balance: %s", err)
		}

		if resp[market.Stock].Available != "0" {
			t.Fatalf("first user: wrong btc balance: expected %v, "+
				"available: %v", "0", resp[market.Stock].Available)
		}

		if resp[market.Stock].Freeze != "0" {
			t.Fatalf("first user: wrong btc balance: expected %v, "+
				"freeze: %v", "0", resp[market.Stock].Freeze)
		}

		if resp[market.Money].Available != "1.8" {
			t.Fatalf("first user: wrong cny balance: expected %v, "+
				"available: %v", "1.8", resp[market.Money].Available)
		}

		if resp[market.Money].Freeze != "0" {
			t.Fatalf("first user: wrong cny balance: expected %v, "+
				"freeze: %v", "0", resp[market.Money].Freeze)
		}
	}

	{
		req := &BalanceQueryRequest{
			UserID: ctx.secondUserID,
		}

		resp, err := ctx.engine.BalanceQuery(req)
		if err != nil {
			t.Fatalf("unable to get second user balance: %s", err)
		}

		if resp[market.Stock].Available != "0.9" {
			t.Fatalf("second user: wrong btc balance: expected %v, "+
				"available: %v", "0.9", resp[market.Stock].Available)
		}

		if resp[market.Stock].Freeze != "0" {
			t.Fatalf("second user: wrong btc balance: expected %v, "+
				"freeze: %v", "0", resp[market.Stock].Freeze)
		}

		if resp[market.Money].Available != "0" {
			t.Fatalf("second user: wrong cny balance: expected %v, "+
				"available: %v", "0", resp[market.Money].Available)
		}

		if resp[market.Money].Freeze != "0" {
			t.Fatalf("second user: wrong cny balance: expected %v, "+
				"freeze: %v", "0", resp[market.Money].Freeze)
		}
	}
}

// testOrderDepth is an integration test which checks the behavior of the
// request depth interval.
func testOrderDepth(t *testing.T, ctx *testContext) {
	// Update fist and second user balances in order to be able to put orders.
	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      AssetBTC,
			ActionType: ActionDeposit,
			ActionID:   0,
			Change:     "3.0",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to deposit btc: %s", err)
		}
	}

	// Put limit order which specifies the exact price for which users want
	// to sell his BTC and its amount.
	{
		req := &OrderPutLimitRequest{
			UserID:       ctx.firstUserID,
			Market:       MarketBTCETC.String(),
			Side:         MarketOrderSideAsk,
			Amount:       "1.0",
			Price:        "1.0",
			TakerFeeRate: "0.1",
			MakerFeeRate: "0.1",
			Source:       "testOrderDepth",
		}

		if _, err := ctx.engine.OrderPutLimit(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}
	{
		req := &OrderPutLimitRequest{
			UserID:       ctx.firstUserID,
			Market:       MarketBTCETC.String(),
			Side:         MarketOrderSideAsk,
			Amount:       "1.0",
			Price:        "2.0",
			TakerFeeRate: "0.1",
			MakerFeeRate: "0.1",
			Source:       "testOrderDepth",
		}

		if _, err := ctx.engine.OrderPutLimit(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}
	{
		req := &OrderPutLimitRequest{
			UserID:       ctx.firstUserID,
			Market:       MarketBTCETC.String(),
			Side:         MarketOrderSideAsk,
			Amount:       "1.0",
			Price:        "3.0",
			TakerFeeRate: "0.1",
			MakerFeeRate: "0.1",
			Source:       "testOrderDepth",
		}

		if _, err := ctx.engine.OrderPutLimit(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}

	{
		req := &OrderDepthRequest{
			Market:   MarketBTCETC.String(),
			Limit:    MaxLimit,
			Interval: "2",
		}

		resp, err := ctx.engine.OrderDepth(req)
		if err != nil {
			t.Fatalf("unable to put order: %s", err)
		}

		if resp.Asks[0].Price != "2" {
			t.Fatal("price should be altered according to the inverval")
		}

		if resp.Asks[0].Volume != "2" {
			t.Fatal("volume for price lower than 2 should be summirized")
		}

		if resp.Asks[1].Price != "4" {
			t.Fatal("price should be altered according to the inverval")
		}

		if resp.Asks[1].Volume != "1" {
			t.Fatal("volume for price from 2 to 4 should be 1")
		}
	}
}

// testStressUpdate is an integration test which checks the thread
// safeness of exchange engine.
func testStressUpdate(t *testing.T, ctx *testContext) {
	done := make(chan struct{})
	errChan := make(chan error)

	for i := 0; i < 1000; i++ {
		go func(i int) {
			req := &BalanceUpdateRequest{
				UserID:     ctx.firstUserID,
				Asset:      AssetBTC,
				ActionType: ActionDeposit,
				ActionID:   int32(i),
				Change:     "1",
				Detail:     make(map[string]interface{}),
			}

			if _, err := ctx.engine.BalanceUpdate(req); err != nil {
				errChan <- errors.Errorf("unable to deposit: %s", err)
			}

			done <- struct{}{}
		}(i)
	}

	for i := 0; i < 1000; i++ {
		select {
		case <-done:
		case err := <-errChan:
			t.Fatal(err)
		case <-time.Tick(time.Second * 5):
			t.Fatal("timeout")
		}
	}

	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      AssetBTC,
			ActionType: ActionWithdrawal,
			ActionID:   0,
			Change:     "-1000",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to withdraw: %s", err)
		}
	}
}

// testOrderPending is an integration test which checks operability of
// pending request.
func testOrderPending(t *testing.T, ctx *testContext) {
	// Update fist and second user balances in order to be able to put orders.
	{
		req := &BalanceUpdateRequest{
			UserID:     ctx.firstUserID,
			Asset:      AssetBTC,
			ActionType: ActionDeposit,
			ActionID:   0,
			Change:     "3.0",
			Detail:     make(map[string]interface{}),
		}

		if _, err := ctx.engine.BalanceUpdate(req); err != nil {
			t.Fatalf("unable to deposit btc: %s", err)
		}
	}

	// Put limit order which specifies the exact price for which users want
	// to sell his BTC and its amount.
	{
		req := &OrderPutLimitRequest{
			UserID:       ctx.firstUserID,
			Market:       MarketBTCLTC.String(),
			Side:         MarketOrderSideAsk,
			Amount:       "1.0",
			Price:        "1.0",
			TakerFeeRate: "0.1",
			MakerFeeRate: "0.1",
			Source:       "testOrderDepth",
		}

		if _, err := ctx.engine.OrderPutLimit(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}

	{
		req := &OrderPendingRequest{
			UserID: ctx.firstUserID,
			Market: MarketBTCLTC.String(),
			Offset: 0,
			Limit:  MaxLimit,
		}

		if _, err := ctx.engine.OrderPending(req); err != nil {
			t.Fatalf("unable to put order: %s", err)
		}
	}
}

// testKline...
func testKline(t *testing.T, ctx *testContext) {
	{
		req := &MarketKLineRequest{
			Market:    MarketBTCLTC.String(),
			StartTime: 2 ^ 64 - 1,
			EndTime:   2 ^ 64 - 1,
			Interval:  3,
		}

		if _, err := ctx.engine.MarketKLine(req); err != nil {
			t.Fatalf("unable to get kline: %s", err)
		}
	}
}

func testWebsocket(t *testing.T, ctx *testContext) {
	//req := &MarketKLineRequest{}
	//for _, asset := range []AssetType{
	//	AssetBTC,
	//	AssetBCH,
	//	AssetETH,
	//	AssetETC,
	//	AssetLTC,
	//	AssetZEC,
	//	AssetDASH,
	//	AssetXRP,
	//} {
	//	req := &BalanceUpdateRequest{
	//		UserID:     2,
	//		Asset:      asset,
	//		ActionType: ActionDeposit,
	//		ActionID:   0,
	//		Change:     "100",
	//		Detail:     make(map[string]interface{}),
	//	}
	//
	//	if _, err := ctx.engine.BalanceUpdate(req); err != nil {
	//		t.Fatalf("unable to deposit: %s", err)
	//	}
	//}
}

var testCases = []struct {
	name string
	test func(t *testing.T, ctx *testContext)
}{
	{
		name: "update user balance once",
		test: testUpdate,
	},
	{
		name: "put limit and market orders",
		test: testPutOrder,
	},
	{
		name: "update user balance stress test",
		test: testStressUpdate,
	},
	{
		name: "check market depth",
		test: testOrderDepth,
	},
	{
		name: "check order pending details",
		test: testOrderPending,
	},
	{
		name: "websocket today query",
		test: testWebsocket,
	},
}

// TestEngine is an main test which is used to run the integration
// exchange engine subtests.
func TestEngine(t *testing.T) {
	rand.Seed(time.Now().Unix())

	cfg := &EngineConfig{
		IP:       "46.101.172.135",
		HTTPPort: 8080,
	}

	if err := CreateEngine(cfg); err != nil {
		t.Fatalf("unable to create engine: %v", err)
	}

	engine, err := GetEngine()
	if err != nil {
		t.Fatalf("unable to get core engine: %s", err)
	}

	t.Logf("Running %v integration tests", len(testCases))
	for _, testCase := range testCases {
		success := t.Run(testCase.name, func(t *testing.T) {
			context, err := setUp(engine)
			if err != nil {
				t.Fatalf("unable to set up context: %v", err)
			}

			testCase.test(t, context)
			t.Logf("Passed(%v)", testCase.name)
		})

		// Stop at the first failure. Mimic behavior of original test
		// framework.
		if !success {
			break
		}
	}
}
