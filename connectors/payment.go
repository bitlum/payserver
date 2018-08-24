package connectors

import (
	"github.com/shopspring/decimal"
	"hash/fnv"
	"strconv"
	"time"
)

// Asset is the list of a trading assets which are available in the exchange
// platform.
type Asset string

var (
	BTC  Asset = "BTC"
	BCH  Asset = "BCH"
	ETH  Asset = "ETH"
	LTC  Asset = "LTC"
	DASH Asset = "DASH"
)

// Media is a list of possible media types. Media is a type of technology which
// is used to transport value of underlying asset.
type PaymentMedia string

var (
	// Blockchain means that blockchain direct used for making the payments.
	Blockchain PaymentMedia = "Blockchain"

	// Lightning means that second layer on top of the blockchain is used for
	// making the payments.
	Lightning PaymentMedia = "Lightning"
)

// PaymentStatus denotes the stage of the processing the payment.
type PaymentStatus string

var (
	// Waiting means that payment has been created and waiting to be approved
	// for sending.
	Waiting PaymentStatus = "Waiting"

	// Pending means that service is seeing the payment, but it not yet approved
	// from the its POV.
	Pending PaymentStatus = "Pending"

	// Completed in case of outgoing/incoming payment this means that we
	// sent/received the transaction in/from the network and it was confirmed
	// number of times service believe sufficient. In case of the forward
	// transaction it means that we successfully routed it through and
	// earned fee for that.
	Completed PaymentStatus = "Completed"

	// Failed means that services has tried to send payment for couple of
	// times, but without success, and now service gave up.
	Failed PaymentStatus = "Failed"
)

// PaymentDirection denotes the direction of the payment.
type PaymentDirection string

var (
	// Internal type of payment which service has made to itself,
	// for the purpose of stabilisation of system. In lightning it might
	// rebalancing, in ethereum send on default address, in bitcoin dust
	// aggregation.
	Internal PaymentDirection = "Internal"

	// Incoming type of payment which service has received from someone else
	// in the media.
	Incoming PaymentDirection = "Incoming"

	//
	// Outgoing type of payment which service has sent to someone else in the
	// media.
	Outgoing PaymentDirection = "Outgoing"
)

type Payment struct {
	// PaymentID it is unique identificator of the payment generated inside
	// the system.
	PaymentID string

	// UpdatedAt denotes the time when payment object has been last updated.
	UpdatedAt int64

	// Status denotes the stage of the processing the payment.
	Status PaymentStatus

	// Direction denotes the direction of the payment.
	Direction PaymentDirection

	// Receipt is a string which identifies the receiver of the
	// payment. It is address in case of the blockchain media,
	// and lightning network invoice in case lightning media.
	Receipt string

	// Asset is an acronym of the crypto currency.
	Asset Asset

	// Account caries the additional information about receiver of the payment.
	Account string

	// Media is a type of technology which is used to transport value of
	// underlying asset.
	Media PaymentMedia

	// Amount is the number of funds which receiver gets at the end.
	Amount decimal.Decimal

	// MediaFee is the fee which is taken by the blockchain or lightning
	// network in order to propagate the payment.
	MediaFee decimal.Decimal

	// MediaID is identificator of the payment inside the media.
	// In case of blockchain media payment id is the transaction id,
	// in case of lightning media it is the payment hash. It is not used as
	// payment identificator because of the reason that it is not unique.
	MediaID string

	// Detail stores all additional information which is needed for this type
	// and status of payment.
	Detail Serializable
}

// BlockchainPendingDetails is the information about pending blockchain
// transaction.
type BlockchainPendingDetails struct {
	// Confirmations is the number of confirmations.
	Confirmations int64

	// ConfirmationsLeft is the number of confirmations left in order to
	// interpret the transaction as confirmed.
	ConfirmationsLeft int64
}

// GeneratePaymentID generates payment id based of the which is uniqie for
// the given connector.
func GeneratePaymentID(parts ...string) string {
	uniqueString := ""
	for _, part := range parts {
		uniqueString += ":" + part
	}

	algorithm := fnv.New64a()
	algorithm.Write([]byte(uniqueString))
	return strconv.FormatUint(algorithm.Sum64(), 10)
}

func NowInMilliSeconds() int64 {
	return int64(time.Now().Nanosecond() / 1000)
}
