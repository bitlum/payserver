package sqlite

import (
	"github.com/bitlum/connector/connectors"
	"bytes"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
)

type PaymentsStore struct {
	DB *DB
}

type Payment struct {
	// PaymentID it is unique identificator of the payment generated inside
	// the system.
	PaymentID string `gorm:"primary_key"`

	// UpdatedAt denotes the time when payment object has been last updated.
	UpdatedAt int64

	// Status denotes the stage of the processing the payment.
	Status string

	// Direction denotes the direction of the payment.
	Direction string

	// Receipt is a string which identifies the receiver of the
	// payment. It is address in case of the blockchain media,
	// and lightning network invoice in case lightning media.
	Receipt string

	// Asset is an acronym of the crypto currency.
	Asset string

	// Account caries the additional information about receiver of the payment.
	Account string

	// Media is a type of technology which is used to transport value of
	// underlying asset.
	Media string

	// Amount is the number of funds which receiver gets at the end.
	Amount string

	// MediaFee is the fee which is taken by the blockchain or lightning
	// network in order to propagate the payment.
	MediaFee string

	// MediaID is identificator of the payment inside the media.
	// In case of blockchain media payment id is the transaction id,
	// in case of lightning media it is the payment hash. It is not used as
	// payment identificator because of the reason that it is not unique.
	MediaID string

	// Detail stores all additional information which is needed for this type
	// and status of payment.
	Detail string

	// DetailType is used to identify details type, to decode it properly.
	DetailType int
}

// Runtime check to ensure that PaymentStore implements
// connectors.PaymentsStore interface.
var _ connectors.PaymentsStore = (*PaymentsStore)(nil)

// PaymentByID returns payment by id.
//
// NOTE: Part of the connectors.PaymentsStore interface.
func (s *PaymentsStore) PaymentByID(paymentID string) (*connectors.Payment, error) {
	dbPayment := &Payment{PaymentID: paymentID}
	if err := s.DB.Find(dbPayment).Error; err != nil {
		return nil, err
	}

	return convertPaymentFrom(dbPayment)
}

// PaymentByReceipt returns payment by receipt.
//
// NOTE: Part of the connectors.PaymentsStore interface.
func (s *PaymentsStore) PaymentByReceipt(receipt string) ([]*connectors.Payment, error) {
	var dbPayments []*Payment
	err := s.DB.Where("receipt = ?", receipt).Find(&dbPayments).Error
	if err != nil {
		return nil, err
	}

	var payments []*connectors.Payment
	for _, dbPayment := range dbPayments {
		payment, err := convertPaymentFrom(dbPayment)
		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// SavePayment add payment to the store.
//
// NOTE: Part of the connectors.PaymentsStore interface.
func (s *PaymentsStore) SavePayment(payment *connectors.Payment) error {
	dbPayment, err := convertPaymentTo(payment)
	if err != nil {
		return err
	}

	return s.DB.Save(dbPayment).Error
}

// ListPayments return list of all payments.
//
// NOTE: Part of the connectors.PaymentsStore interface.¬
func (s *PaymentsStore) ListPayments(asset connectors.Asset,
	status connectors.PaymentStatus, direction connectors.PaymentDirection,
	media connectors.PaymentMedia) ([]*connectors.Payment, error) {

	db := s.DB.DB

	if asset != "" {
		db = db.Where("asset = ?", asset)
	}

	if status != "" {
		db = db.Where("status = ?", status)
	}

	if direction != "" {
		db = db.Where("direction = ?", direction)
	}

	if media != "" {
		db = db.Where("media = ?", media)
	}

	var dbPayments []*Payment
	err := db.Find(&dbPayments).Error
	if err != nil {
		return nil, err
	}

	var payments []*connectors.Payment
	for _, dbPayment := range dbPayments {
		payment, err := convertPaymentFrom(dbPayment)
		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

func convertPaymentTo(payment *connectors.Payment) (*Payment, error) {
	var details string
	var detailType int

	if payment.Detail != nil {
		var b bytes.Buffer
		if err := payment.Detail.Encode(&b, 0); err != nil {
			return nil, err
		}

		details = b.String()

		switch payment.Detail.(type) {
		case *connectors.GeneratedTxDetails:
			detailType = 1
		case *connectors.BlockchainPendingDetails:
			detailType = 2
		default:
			return nil, errors.Errorf("unknown details type: %v", payment.Detail)
		}
	}

	dbPayment := &Payment{
		PaymentID:  payment.PaymentID,
		UpdatedAt:  payment.UpdatedAt,
		Status:     string(payment.Status),
		Direction:  string(payment.Direction),
		Receipt:    payment.Receipt,
		Asset:      string(payment.Asset),
		Account:    payment.Account,
		Media:      string(payment.Media),
		Amount:     payment.Amount.String(),
		MediaFee:   payment.MediaFee.String(),
		MediaID:    payment.MediaID,
		Detail:     details,
		DetailType: detailType,
	}

	return dbPayment, nil
}

func convertPaymentFrom(dbPayment *Payment) (*connectors.Payment, error) {
	var detail connectors.Serializable

	if dbPayment.DetailType != 0 {
		switch dbPayment.DetailType {
		case 1:
			detail = &connectors.GeneratedTxDetails{}
		case 2:
			detail = &connectors.BlockchainPendingDetails{}
		default:
			return nil, errors.Errorf("unknown details type: %v", dbPayment.DetailType)
		}

		b := bytes.NewBufferString(dbPayment.Detail)
		if err := detail.Decode(b, 0); err != nil {
			return nil, err
		}
	}

	amount, err := decimal.NewFromString(dbPayment.Amount)
	if err != nil {
		return nil, err
	}

	mediaFee, err := decimal.NewFromString(dbPayment.MediaFee)
	if err != nil {
		return nil, err
	}

	payment := &connectors.Payment{
		PaymentID: dbPayment.PaymentID,
		UpdatedAt: dbPayment.UpdatedAt,
		Status:    connectors.PaymentStatus(dbPayment.Status),
		Direction: connectors.PaymentDirection(dbPayment.Direction),
		Receipt:   dbPayment.Receipt,
		Asset:     connectors.Asset(dbPayment.Asset),
		Account:   dbPayment.Account,
		Media:     connectors.PaymentMedia(dbPayment.Media),
		Amount:    amount,
		MediaFee:  mediaFee,
		MediaID:   dbPayment.MediaID,
		Detail:    detail,
	}

	return payment, nil
}
