package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type baseResponse struct {
	Error *Error `json:"error"`
	ID    int32  `json:"id"`
}

type request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int32         `json:"id"`
}

// MaxLimit maximum available limit value, which can be passed in engine rpc
// call.
const MaxLimit int32 = 100

// AssetType...
type AssetType string

const (
	AssetBTC  AssetType = "BTC"
	AssetBCH  AssetType = "BCH"
	AssetETH  AssetType = "ETH"
	AssetLTC  AssetType = "LTC"
	AssetDASH AssetType = "DASH"
)

type OrderType uint32

var (
	// LimitOrderType...
	LimitOrderType OrderType = 1

	// MarketOrderType...
	MarketOrderType OrderType = 2
)

func NewOrderTypeFromString(s string) OrderType {
	switch strings.ToLower(s) {
	case "limit":
		return LimitOrderType
	case "market":
		return MarketOrderType
	}

	return 0
}

func (t OrderType) String() string {
	switch t {
	case LimitOrderType:
		return "limit"
	case MarketOrderType:
		return "market"
	default:
		return "<unknown>"
	}
}

type MarketType struct {
	Stock AssetType
	Money AssetType
}

// TODO(andrew.shvv) make it better
func NewMarket(market string) MarketType {
	return MarketType{
		Stock: AssetType(market[3:]),
		Money: AssetType(market[:3]),
	}
}

func (t MarketType) String() string {
	return fmt.Sprintf("%v%v", t.Money, t.Stock)
}

func (t *MarketType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	market := NewMarket(s)
	*t = market
	return nil
}

func (t MarketType) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.String() + "\""), nil
}

var (
	MarketBTCETH = MarketType{
		Stock: AssetETH,
		Money: AssetBTC,
	}

	MarketBTCBCH = MarketType{
		Stock: AssetBCH,
		Money: AssetBTC,
	}

	MarketBTCLTC = MarketType{
		Stock: AssetLTC,
		Money: AssetBTC,
	}

	MarketBTCDASH = MarketType{
		Stock: AssetDASH,
		Money: AssetBTC,
	}

	MarketETHLTC = MarketType{
		Stock: AssetLTC,
		Money: AssetETH,
	}
)

type MarketOrderSide uint32

func NewMarketSideFromString(s string) MarketOrderSide {
	switch strings.ToLower(s) {
	case "ask":
		return MarketOrderSideAsk
	case "bid":
		return MarketOrderSideBid
	}

	return 0
}

func (s MarketOrderSide) String() string {
	switch s {
	case MarketOrderSideAsk:
		return "ask"
	case MarketOrderSideBid:
		return "bid"
	default:
		return "<unknown>"
	}
}

const (
	// MarketOrderSideAsk is the type of request from the user who are
	// willing to sell the stock. Seller asks the price is willing to
	// take for on of stock. For example if market is USDBTC, the stock will be
	// BTC and money will be USD.
	MarketOrderSideAsk MarketOrderSide = 1

	// MarketOrderSideBid is the type of request from the user who are willing
	// to buy the stock. Seller bids the price is willing to give for one
	// stock. For example if market is USDBTC, the stock will be BTC and
	// money will be USD.
	MarketOrderSideBid MarketOrderSide = 2
)

// ExchangeRole represents which role was taken by the user's order during it
// execution.
type ExchangeRole uint32

func (r ExchangeRole) String() string {
	switch r {
	case MakerRole:
		return "maker"
	case TakerRole:
		return "taker"
	default:
		return "<unknown>"
	}
}

const (
	// MakerRole is assigned to the order/deal if order already
	// existed in the order table and was fulfilled by incoming one.
	MakerRole ExchangeRole = 1

	// TakerRole is assigned to the order/deal if order was completely
	// fulfilled during its lending in the order table.
	TakerRole ExchangeRole = 2
)

// ActionType is used to indicate the reason of funds change. Change of funds
// might occur due to trading operation, deposit, or withdrawal.
type ActionType string

const (
	ActionDeposit    ActionType = "deposit"
	ActionWithdrawal ActionType = "withdrawal"
	ActionTrade      ActionType = "trade"
)

// Kline is an information about the market during the specified interval of
// time.
type Kline struct {
	// Time is the start if the interval.
	Time float64

	// OpenPrice is the start price of the interval.
	OpenPrice string

	// ClosePrice is the close price of the interval.
	ClosePrice string

	// HighestPrice is the maximum price during interval.
	HighestPrice string

	// LowestPrice is the minimum price during interval.
	LowestPrice string

	// Amount of orders which occurred within preset interval of time.
	Amount string

	// Volume of all orders which we executed during specified in request
	// interval of time.
	Volume string

	Market MarketType
}

// Depth is an amount of funds which associated with particular price. It is
// used to understand how much of stocks we could buy/sell for this price.
type Depth struct {
	Volume string
	Price  string
}

// UnixTime is used
type UnixTime time.Time

// DealDetail represent the detailed information about the deal. Deal is an
// result of execution of two orders.
type DealDetail struct {
	DealID int32        `json:"id"`
	Time   float64      `json:"time"`
	Role   ExchangeRole `json:"role"`
	Amount string       `json:"amount"`
	UserID uint32       `json:"user"`
	Fee    string       `json:"fee"`

	// Price corresponds to deal price, if this it the limit order than
	// the price should be the same for all orders, if it it market order
	// the price will differ.
	Price string `json:"price"`

	// Deal is the number of stocks which was handled in this deal. This
	// number if less or equal to the overall amount of money putter in
	// order.
	Deal string `json:"deal"`

	// DealOrderID corresponds to the order with which this deal was made.
	DealOrderID int32 `json:"deal_order_id"`
}

// OrderDetailedInfo represent the detailed information about user order.
type OrderDetailedInfo struct {
	OrderID      int32           `json:"id"`
	UserID       uint32          `json:"user"`
	Amount       string          `json:"amount"`
	Price        string          `json:"price"`
	Side         MarketOrderSide `json:"side"`
	Type         OrderType       `json:"type"`
	Market       MarketType      `json:"market"`
	Source       string          `json:"source"`
	TakerFeeRate string          `json:"taker_fee"`
	MakerFeeRate string          `json:"maker_fee"`

	// DealStock is the amount of stock which was involved in the order
	// immediate execution.
	DealStock string `json:"deal_stock"`

	// DealStock is the amount of money which was involved in the order
	// immediate execution.
	DealMoney string `json:"deal_money"`

	// DealFee is the amount of fee expressed in money which was taken from
	// order originator during the order immediate execution.
	DealFee string `json:"deal_fee"`

	// CTime is the time of order entity creation within engine
	// core service.
	CTime float64 `json:"ctime"`

	// MTime is the time when order have been updated last time. By
	// update we mean that it has been matched with another order and its left
	// amount has been changed.
	MTime float64 `json:"mtime, omitempty"`

	// FTime is the time when order have been finished
	FTime float64 `json:"ftime, omitempty"`

	// Left the amount of funds left in the market without being handled.
	Left string `json:"left"`
}

type BalanceQueryRequest struct {
	UserID uint32
	Assets []AssetType
}

type BalanceInfo struct {
	// Available is the funds which can be used to in trading.
	Available string `json:"available"`

	// Freeze is the funds which currently occupied in some process, for
	// example in trades.
	Freeze string `json:"freeze"`
}

type BalanceQueryResponse map[AssetType]BalanceInfo

type BalanceUpdateRequest struct {
	UserID uint32
	Asset  AssetType

	ActionType ActionType

	// ActionID is used to not apply the same action of funds change twice.
	ActionID int32

	Change string

	// Detail is used to store any additional information about
	// balance change which might be helpful, and used later.
	Detail map[string]interface{}
}

type BalanceUpdateResponse struct {
	Status string `json:"status"`
}

type BalanceHistoryRequest struct {
	UserID     uint32
	Asset      AssetType
	ActionType ActionType
	StartTime  float64
	EndTime    float64
	Offset     int32
	Limit      int32
}

type BalanceHistoryRecord struct {
	Time       float64    `json:"time"`
	Asset      string     `json:"asset"`
	ActionType ActionType `json:"business"`

	// Change is amount on which balance has been changed.
	Change string `json:"change"`

	// Balance is the final balance of the user.
	Balance string `json:"balance"`

	// Detail is used to store any additional information about
	// balance change which might be helpful, and used later.
	Detail map[string]interface{} `json:"detail"`
}

type BalanceHistoryResponse struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`

	Records []*BalanceHistoryRecord `json:"records"`
}

type AssetListRequest struct{}

type AssetListResponse []struct {
	Name string `json:"name"`

	// Prec is the precious of calculation specified for this asset.
	Prec float64 `json:"prec"`
}

type AssetSummaryRequest []AssetType

type AssetSummaryResponse []struct {
	AssetName string `json:"name"`

	// TotalBalance is an overall balance available for this asset in the
	// engine.
	TotalBalance string `json:"total_balance"`

	// AvailableCount is the number of account which hold this asset with
	// available balance on it.
	AvailableCount int `json:"available_count"`

	// AvailableBalance is the available balance which is not frezed in
	// trades, and might be be withdrawn.
	AvailableBalance string `json:"available_balance"`

	// FreezeCount is the number of accounts with freezed in trades balance.
	FreezeCount int `json:"freeze_count"`

	// FreezeBalance is the overall amount of funds which is freezed in orders.
	FreezeBalance string `json:"freeze_balance"`
}

type OrderPutLimitRequest struct {
	UserID uint32
	Market string
	Side   MarketOrderSide

	// Amount is a number of stock which user is willing the sell/buy.
	Amount string

	// Price is expressed in market money, which seller/buyer is willing to
	// take/give for one of stock.
	Price string

	// TakerFeeRate is an coefficient from [0;1) which is used to determine
	// the percentage of money which will be taken as a fee from total amount
	// of order. This fee coefficient will be applied to the part of order which
	// was executed immediately.
	TakerFeeRate string

	// MakerFeeRate is an coefficient from [0;1) which is used to determine
	// the percentage of money which will be taken as a fee from total amount
	// of order. This fee coefficient will be applied to the part of order which
	// was executed after matching with another order.
	MakerFeeRate string

	// Source designate the origin of the order requests. It is needed to
	// analyze statistics.
	Source string
}

type OrderPutLimitResponse OrderDetailedInfo

type OrderPutMarketRequest struct {
	UserID uint32
	Market string
	Side   MarketOrderSide

	// Amount depending on the side this either the number of stock which user
	// wants to sell and get money (ask) or money which user wants to sell and
	// get stock (bid). On the USDBTC market stock is the BTC and money is USD.
	Amount string

	// TakerFeeRate is an coefficient from [0;1) which is used to determine
	// the percentage of money which will be taken as a fee from total amount
	// of order. This fee coefficient will be applied to the part of order which
	// was executed immediately.
	TakerFeeRate string

	// Source designate the origin of the order requests. It is needed to
	// analyze statistics.
	Source string
}

type OrderPutMarketResponse OrderDetailedInfo

type OrderCancelRequest struct {
	UserID  uint32
	Market  string
	OrderID int32
}

type OrderCancelResponse OrderDetailedInfo

type OrderDealsRequest struct {
	OrderID int32
	Offset  int32
	Limit   int32
}

type OrderDealsResponse struct {
	Offset int32        `json:"offset"`
	Limit  int32        `json:"limit"`
	Deals  []DealDetail `json:"records"`
}

type OrderBookRequest struct {
	Market string
	Side   MarketOrderSide
	Offset int32
	Limit  int32
}

type OrderBookResponse struct {
	Offset int32               `json:"offset"`
	Limit  int32               `json:"limit"`
	Total  int32               `json:"total"`
	Orders []OrderDetailedInfo `json:"orders"`
}

type OrderDepthRequest struct {
	Market string
	Limit  int32

	// Interval is the number which is used to combine volumes of orders if
	// price is lower than specified interval. Instead of giving the exact
	// volume for each price, orders will be combined in the intervals and
	// the volume within those intervals will be summarized.
	Interval string
}

type OrderDepthResponse struct {
	Asks []Depth `json:"asks"`
	Bids []Depth `json:"bids"`
}

type OrderPendingRequest struct {
	UserID uint32
	Market string
	Offset int32
	Limit  int32
}

type OrderPendingResponse struct {
	Offset int32                `json:"offset"`
	Limit  int32                `json:"limit"`
	Total  int32                `json:"total"`
	Orders []*OrderDetailedInfo `json:"records"`
}

type OrderPendingDetailRequest struct {
	Market  string
	OrderID int32
}

type OrderPendingDetailResponse OrderDetailedInfo

type OrderFinishedRequest struct {
	UserID    uint32
	Market    string
	StartTime float64
	EndTime   float64
	Offset    int32
	Limit     int32
	Side      MarketOrderSide
}

type OrderFinishedResponse struct {
	Offset int32                `json:"offset"`
	Limit  int32                `json:"limit"`
	Total  int32                `json:"total"`
	Orders []*OrderDetailedInfo `json:"records"`
}

type OrderFinishedDetailRequest struct {
	OrderID int32
}

type OrderFinishedDetailResponse OrderDetailedInfo

type MarketListRequest struct{}

type MarketListResponse []struct {
	Money      AssetType  `json:"name"`
	Stock      AssetType  `json:"stock"`
	FeePrec    int        `json:"fee_prec"`
	StockPrec  int        `json:"stock_prec"`
	MoneyPrec  int        `json:"money_prec"`
	MinAmount  string     `json:"min_amount"`
	MarketName MarketType `json:"name"`
}

type MarketSummaryRequest []MarketType

type MarketSummaryResponse []struct {
	MarketName MarketType `json:"name"`
	AskCount   int        `json:"ask_count"`
	AskAmount  string     `json:"ask_amount"`
	BidCount   int        `json:"bid_count"`
	BidAmount  string     `json:"bid_amount"`
}

type MarketLastRequest struct {
	Market string
}

type MarketDealsRequest struct {
	Market string
	Limit  int32

	// LastID is used to specify an id till which the deals will be
	// filtered.
	LastID int32
}

type MarketDealsResponse []struct {
	DealID int32   `json:"id"`
	Time   float64 `json:"time"`
	Type   string  `json:"type"`
	Amount string  `json:"amount"`
	Price  string  `json:"price"`
}

type MarketUserDealsRequest struct {
	UserID uint32
	Market string
	Offset int32
	Limit  int32
}

type MarketUserDealsResponse struct {
	Offset int32        `json:"offset"`
	Limit  int32        `json:"limit"`
	Deals  []DealDetail `json:"records"`
}

type MarketKLineRequest struct {
	Market    string
	StartTime float64
	EndTime   float64

	// Interval determines the period of time for which kline should be
	// calculated.
	Interval int32
}

type MarketKLineResponse []Kline

type MarketStatusRequest struct {
	Market string

	// Period is the time within which the market statistic will be calculated.
	// The interval of time is determined as [now-period, now].
	Period int32
}

type MarketStatusResponse struct {
	Period int32  `json:"period"`
	Last   string `json:"last"`
	Open   string `json:"open"`
	Close  string `json:"close"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Volume string `json:"volume"`
}

type MarketStatusTodayRequest struct {
	Market string
}

type MarketStatusTodayResponse struct {
	Open   string `json:"open"`
	Last   string `json:"last"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Deal   string `json:"deal"`
	Volume string `json:"volume"`
}
