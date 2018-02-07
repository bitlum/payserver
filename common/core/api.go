package core

import (
	"fmt"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"time"

	"net/rpc"

	"github.com/go-errors/errors"
)

var (
	AllAssets = []AssetType{
		AssetBTC,
		AssetBCH,
		AssetETH,
		AssetLTC,
		AssetDASH,
	}

	// AllMarkets...
	AllMarkets = []MarketType{
		MarketBTCETH,
		MarketBTCBCH,
		MarketBTCLTC,
		MarketBTCDASH,
		MarketETHLTC,
	}

	// AllOrderStatuses...
	AllOrderStatuses = []string{
		"pending",
		"finished",
		"canceled",
	}

	// AllMarketSides...
	AllMarketSides = []string{
		"ask",
		"bid",
	}

	// AllMarketSides...
	AllActionTypes = []ActionType{
		ActionTrade,
		ActionDeposit,
		ActionWithdrawal,
	}
)

var _engine *Engine

type EngineConfig struct {
	// IP is an ip which points out to the main engine server and used to
	// connect to it.
	IP string

	// HTTPPort denotes the port on which engine server is listening for
	// incoming requests.
	HTTPPort int

	WSPort string
}

// Engine is the programmatic connector to the core exchange engine,
// currently is written to interact using http requests to access rpc
// function point, but in future could be rewritten to use unix sockets or
// even use embedded C code.
type Engine struct {
	wsClient   *rpc.Client
	httpClient *http.Client
	url        string
}

func (e *Engine) Stop() error {
	e.wsClient.Close()
	return nil
}

// CreateEngine...
func CreateEngine(cfg *EngineConfig) error {
	if _engine == nil {
		//origin := "http://195.91.204.4"

		//wsUrl := fmt.Sprintf("ws://%v:%v", cfg.IP, cfg.WSPort)
		httpUrl := fmt.Sprintf("http://%v:%v", cfg.IP, cfg.HTTPPort)
		//
		//var dialer = websocket.Dialer{
		//	HandshakeTimeout: time.Second * 10,
		//	Subprotocols:     []string{"chat"},
		//	ReadBufferSize:   1024,
		//	WriteBufferSize:  1024,
		//}
		//
		//_, _, err := dialer.Dial(wsUrl, nil)
		//if err != nil {
		//	return errors.Errorf("ListenAndServe: %v", err)
		//}

		//fmt.Println("done")
		//codec := jsonrpc.NewClient(ws)
		//
		//
		//	fmt.Println("done")
		//	ws, err := websocket.(wsUrl, "", origin)
		//	if err != nil {
		//		return errors.Errorf("unable to connect websocket: %v", err)
		//	}
		//fmt.Println("done")

		_engine = &Engine{
			httpClient: &http.Client{},
			wsClient:   nil,
			url:        httpUrl,
		}
	}

	return nil
}

// GetEngine is used to return singleton instance of the Engine.
func GetEngine() (*Engine, error) {
	if _engine == nil {
		return nil, errors.New("engine isn't created")
	}

	return _engine, nil
}

// RemoveEngine...
func RemoveEngine() error {
	if _engine == nil {
		return _engine.Stop()
	}
	return nil
}

// makeRPCCall is a helper which is used to execute engine remote
// procedure call using http post request, with encoded parameters in a request
// body. On return the rpc response object is populated with data which is
// specific for ever call.
func (e *Engine) makeRPCCall(method string, params interface{},
	rpcResp interface{}) error {

	args, err := extractArguments(params)
	if err != nil {
		return errors.Errorf("unable to extract arguments: %v", err)
	}

	rpcReq := &request{
		Method: method,
		Params: args,
		ID:     int32(time.Now().Unix()),
	}

	data, err := json.Marshal(rpcReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, rpcResp); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("status code: %v", resp.StatusCode)
	}

	return nil
}

// Accounts returns available and frozen balances of user for every
// supported by engine currency.
func (e *Engine) BalanceQuery(params *BalanceQueryRequest) (
	BalanceQueryResponse, error) {

	type Response struct {
		baseResponse
		Result BalanceQueryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.query", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// BalanceUpdate updates balance of the user, it is used by the
// engine itself to update balances on order matching, and it is used by
// backend subsystem to Deposit and withdraw funds.
func (e *Engine) BalanceUpdate(params *BalanceUpdateRequest) (
	*BalanceUpdateResponse, error) {

	type Response struct {
		baseResponse
		Result *BalanceUpdateResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.update", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// BalanceHistory returns the history of all funds changes we have been done
// with the user chosen asset.
func (e *Engine) BalanceHistory(params *BalanceHistoryRequest) (
	*BalanceHistoryResponse, error) {

	type Response struct {
		baseResponse
		Result *BalanceHistoryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("balance.history", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// AssetList returns the list of assets and its calculation precious, i.e.
// how much decimal places is preserved in engine when operating with asset.
func (e *Engine) AssetList(params *AssetListRequest) (
	*AssetListResponse, error) {

	type Response struct {
		baseResponse
		Result *AssetListResponse
	}

	response := &Response{}
	err := e.makeRPCCall("asset.list", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// AssetSummary returns the aggregated information for all accounts about
// overall available volume corresponding to assets, overall freezed volume,
// and how much accounts have the asset available.
func (e *Engine) AssetSummary(params *AssetSummaryRequest) (
	*AssetSummaryResponse, error) {

	type Response struct {
		baseResponse
		Result *AssetSummaryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("asset.summary", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPutLimit puts the order on the market with fixed price and amount, if
// there is enough volume for specified price the order will be waiting for
// opposite order to come.
func (e *Engine) OrderPutLimit(params *OrderPutLimitRequest) (
	*OrderPutLimitResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPutLimitResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.put_limit", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPutMarket puts the order on the market. As far as price is not fixed
// and the goal of the order is to be fully executed than if market has enough
// volume for executing the order it will be fully handled. But in this case
// the average price might be much lower than the market price.
func (e *Engine) OrderPutMarket(params *OrderPutMarketRequest) (
	*OrderPutMarketResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPutMarketResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.put_market", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderCancel cancels the order of specific user on the market.
func (e *Engine) OrderCancel(params *OrderCancelRequest) (
	*OrderCancelResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderCancelResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.cancel", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderBook by the given market and side returns all available on
// current moment orders.
func (e *Engine) OrderBook(params *OrderBookRequest) (
	*OrderBookResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderBookResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.book", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderDepth returns the overall volume for each available price, also if
// interval is specified than volume within the interval will be combined.
func (e *Engine) OrderDepth(params *OrderDepthRequest) (
	*OrderDepthResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderDepthResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.depth", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPending returns the user's pending orders with their detailed
// information.
func (e *Engine) OrderPending(params *OrderPendingRequest) (
	*OrderPendingResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPendingResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.pending", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderPendingDetail returns the detailed information about specific order.
func (e *Engine) OrderPendingDetail(params *OrderPendingDetailRequest) (
	*OrderPendingDetailResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderPendingDetailResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.pending_detail", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderDeals returns the information about the operations which has been
// applied to order in order to fulfil it. When the order is completely
// fulfilled the number of deals might be [1, inf).
func (e *Engine) OrderDeals(params *OrderDealsRequest) (
	*OrderDealsResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderFinished returns the information about user's finished orders.
func (e *Engine) OrderFinished(params *OrderFinishedRequest) (
	*OrderFinishedResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderFinishedResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.finished", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// OrderFinishedDetail returns the detailed information about specific
// finished order.
func (e *Engine) OrderFinishedDetail(params *OrderFinishedDetailRequest) (
	*OrderFinishedDetailResponse, error) {

	type Response struct {
		baseResponse
		Result *OrderFinishedDetailResponse
	}

	response := &Response{}
	err := e.makeRPCCall("order.finished_detail", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketLast returns last market price.
func (e *Engine) MarketLast(params *MarketLastRequest) (
	*string, error) {

	type Response struct {
		baseResponse
		Result *string
	}

	response := &Response{}
	err := e.makeRPCCall("market.last", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketSummary returns the aggregated information for all accounts about
// overall available orders corresponding to market.
func (e *Engine) MarketSummary(params *MarketSummaryRequest) (
	*MarketSummaryResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketSummaryResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.summary", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketList return the information about the market's calculation precious,
// and minimum amount of order.
func (e *Engine) MarketList(params *MarketListRequest) (
	*MarketListResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketListResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.list", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketDeals returns the information about
func (e *Engine) MarketDeals(params *MarketDealsRequest) (
	MarketDealsResponse, error) {

	type Response struct {
		baseResponse
		Result MarketDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketUserDeals returns the information about deals which were made by
// user. Deal is the result of orders matching.
func (e *Engine) MarketUserDeals(params *MarketUserDealsRequest) (
	*MarketUserDealsResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketUserDealsResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.user_deals", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketKLine returns the information about the market withing preset
// interval of time. The number of requests is determined as (e - s) / i, where
// e - end time, s - start time, i - interval.
func (e *Engine) MarketKLine(params *MarketKLineRequest) (
	MarketKLineResponse, error) {

	type Response struct {
		baseResponse
		Result MarketKLineResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.kline", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketStatus returns the status of the market within given period of time.
func (e *Engine) MarketStatus(params *MarketStatusRequest) (
	*MarketStatusResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketStatusResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.status", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

// MarketStatusToday returns the information about the market within the
// current day.
func (e *Engine) MarketStatusToday(params *MarketStatusTodayRequest) (
	*MarketStatusTodayResponse, error) {

	type Response struct {
		baseResponse
		Result *MarketStatusTodayResponse
	}

	response := &Response{}
	err := e.makeRPCCall("market.status_today", params, response)
	if err != nil {
		return nil, err
	}

	// https://golang.org/doc/faq#nil_error
	if response.Error != nil {
		return nil, response.Error
	}

	return response.Result, nil
}

func (e *Engine) TodayQuery() error {
	return nil
}
