package crpc

import (
	"github.com/bitlum/connector/metrics/rpc"
	"golang.org/x/net/context"
	"github.com/bitlum/connector/connectors"
	"github.com/go-errors/errors"
	"github.com/bitlum/connector/metrics"
)

// defaultAccount default account which will be used for all request until
// account would be working properly for all assets.
var defaultAccount = "zigzag"

const (
	CreateReceiptReq     = "CreateReceipt"
	ValidateReceiptReq   = "ValidateReceipt"
	BalanceReq           = "Balance"
	EstimateFeeReq       = "EstimateFee"
	SendPaymentReq       = "SendPayment"
	PaymentByIDReq       = "PaymentByID"
	PaymentsByReceiptReq = "PaymentsByReceipt"
	ListPaymentsReq      = "ListPayments"
)

// Server is the gRPC server which implements PayServer interface.
type Server struct {
	net                  string
	blockchainConnectors map[connectors.Asset]connectors.BlockchainConnector
	lightningConnectors  map[connectors.Asset]connectors.LightningConnector
	paymentsStore        connectors.PaymentsStore
	metrics              rpc.MetricsBackend
}

// A compile time check to ensure that Server fully implements the
// PayServer gRPC service.
var _ PayServerServer = (*Server)(nil)

// NewRPCServer creates and returns a new instance of the Server.
func NewRPCServer(net string,
	blockchainConnectors map[connectors.Asset]connectors.BlockchainConnector,
	lightningConnectors map[connectors.Asset]connectors.LightningConnector,
	paymentsStore connectors.PaymentsStore,
	metrics rpc.MetricsBackend) (*Server, error) {
	return &Server{
		blockchainConnectors: blockchainConnectors,
		lightningConnectors:  lightningConnectors,
		paymentsStore:        paymentsStore,
		metrics:              metrics,
		net:                  net,
	}, nil
}

//
// CreateReceipt is used to create blockchain deposit address in
// case of blockchain media, and lightning network invoice in
// case of the lightning media, which will be used to receive money from
// external entity.
func (s *Server) CreateReceipt(ctx context.Context,
	req *CreateReceiptRequest) (*CreateReceiptResponse, error) {

	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	var resp *CreateReceiptResponse

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(CreateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		address, err := c.CreateAddress(connectors.AccountAlias(defaultAccount))
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(CreateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		resp = &CreateReceiptResponse{
			Receipt: address,
		}
	case Media_LIGHTNING:
		c, ok := s.lightningConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(CreateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		// Ensure that even if amount is not specified we treat it as zero
		// value.
		if req.Amount == "" {
			req.Amount = "0"
		}

		invoice, err := c.CreateInvoice(defaultAccount, req.Amount, req.Description)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(CreateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		resp = &CreateReceiptResponse{
			Receipt: invoice,
		}
	default:
		err := errors.Errorf("media(%v) is not supported", req.Media.String())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(CreateReceiptReq, string(metrics.LowSeverity))
		return nil, err
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// ValidateReceipt is used to validate receipt for given asset and media.
func (s *Server) ValidateReceipt(ctx context.Context,
	req *ValidateReceiptRequest) (*EmptyResponse, error) {
	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ValidateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		if err := c.ValidateAddress(req.Receipt); err != nil {
			s.metrics.AddError(ValidateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

	case Media_LIGHTNING:
		c, ok := s.lightningConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ValidateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		if req.Amount == "" {
			req.Amount = "0"
		}

		if err := c.ValidateInvoice(req.Receipt, req.Amount); err != nil {
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ValidateReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

	default:
		err := errors.Errorf("media(%v) is not supported", req.Media.String())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(ValidateReceiptReq, string(metrics.LowSeverity))
		return nil, err
	}

	resp := &EmptyResponse{}
	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// Balance is used to determine balance.
func (s *Server) Balance(ctx context.Context, req *BalanceRequest,
) (*BalanceResponse, error) {
	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	resp := &BalanceResponse{}

	if req.Media == Media_BLOCKCHAIN || req.Media == Media_MEDIA_NONE {
		var cntrs map[connectors.Asset]connectors.BlockchainConnector
		if req.Asset == Asset_ASSET_NONE {
			// If asset wasn't specified return balances for all blockchain
			// assets.
			cntrs = s.blockchainConnectors
		} else {
			c, ok := s.blockchainConnectors[connectors.Asset(req.Asset.String())]
			if !ok {
				err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(BalanceReq, string(metrics.LowSeverity))
				return nil, err
			}
			cntrs = map[connectors.Asset]connectors.BlockchainConnector{
				connectors.Asset(req.Asset.String()): c,
			}
		}

		for asset, c := range cntrs {
			available, err := c.ConfirmedBalance(connectors.SentAccount)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			pending, err := c.PendingBalance(connectors.SentAccount)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			protoAsset, err := convertAssetToProto(asset)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			resp.Balances = append(resp.Balances, &Balance{
				Media:     Media_BLOCKCHAIN,
				Asset:     protoAsset,
				Available: available.String(),
				Pending:   pending.String(),
			})

			// TODO(andrew.shvv) Combine btc balance with lightning btc
			// wallet balance?
		}
	}

	if req.Media == Media_LIGHTNING || req.Media == Media_MEDIA_NONE {
		var cntrs map[connectors.Asset]connectors.LightningConnector
		if req.Asset == Asset_ASSET_NONE {
			// If asset wasn't specified return balances for all blockchain
			// assets.
			cntrs = s.lightningConnectors
		} else {
			c, ok := s.lightningConnectors[connectors.Asset(req.Asset.String())]
			if !ok {
				err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(BalanceReq, string(metrics.LowSeverity))
				return nil, err
			}

			cntrs = map[connectors.Asset]connectors.LightningConnector{
				connectors.Asset(req.Asset.String()): c,
			}
		}

		for asset, c := range cntrs {
			// TODO(andrew.shvv) Show channels balance, problems:
			// * if we don't have distinction between off-chain and on-chain
			// balance it will be unclear for end user how use this balance,
			// otherwise we would ned to have two different rpc methods for that.

			available, err := c.ConfirmedBalance(defaultAccount)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			pending, err := c.PendingBalance(defaultAccount)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			protoAsset, err := convertAssetToProto(asset)
			if err != nil {
				err := newErrInternal(err.Error())
				log.Errorf("command(%v), error: %v", getFunctionName(), err)
				s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
				return nil, err
			}

			resp.Balances = append(resp.Balances, &Balance{
				Media:     Media_LIGHTNING,
				Asset:     protoAsset,
				Available: available.String(),
				Pending:   pending.String(),
			})
		}
	}

	return resp, nil
}

//
// EstimateFee estimates the fee of the outgoing payment.
func (s *Server) EstimateFee(ctx context.Context,
	req *EstimateFeeRequest) (*EstimateFeeResponse, error) {
	log.Tracef("command(%v), request(%v)", getFunctionName(),
		convertProtoMessage(req))

	var resp *EstimateFeeResponse

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
			return nil, err
		}

		if req.Amount == "" {
			req.Amount = "0"
		}

		fee, err := c.EstimateFee(req.Amount)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
			return nil, err
		}

		resp = &EstimateFeeResponse{
			MediaFee: fee.String(),
		}

	case Media_LIGHTNING:
		err := errors.New("fee estimation for lightning is not supported")
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
		return nil, err

	default:
		err := errors.Errorf("media(%v) is not supported", req.Media.String())
		s.metrics.AddError(EstimateFeeReq, string(metrics.LowSeverity))
		return nil, err
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// SendPayment sends payment to the given recipient,
// ensures in the validity of the receipt as well as the
// account has enough money for doing that.
func (s *Server) SendPayment(ctx context.Context,
	req *SendPaymentRequest) (*Payment, error) {

	log.Tracef("command(%v), request(%v)", getFunctionName(), convertProtoMessage(req))

	var (
		resp    *Payment
		payment *connectors.Payment
		err     error
	)

	switch req.Media {
	case Media_BLOCKCHAIN:
		c, ok := s.blockchainConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
			return nil, err
		}

		// TODO(andrew.shvv) generate and send can't be used separately,
		// because of the work of the lock inputs.
		if req.Amount == "" {
			req.Amount = "0"
		}

		payment, err = c.CreatePayment(req.Receipt, req.Amount)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
			return nil, err
		}

		payment, err = c.SendPayment(payment.PaymentID)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
			return nil, err
		}

	case Media_LIGHTNING:
		c, ok := s.lightningConnectors[connectors.Asset(req.Asset.String())]
		if !ok {
			err := newErrAssetNotSupported(req.Asset.String(), req.Media.String())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
			return nil, err
		}

		if req.Amount == "" {
			req.Amount = "0"
		}

		payment, err = c.SendTo(req.Receipt, req.Amount)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
			return nil, err
		}

	default:
		err := errors.Errorf("media(%v) is not supported", req.Media.String())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
		return nil, err
	}

	resp, err = convertPaymentToProto(payment)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(SendPaymentReq, string(metrics.LowSeverity))
		return nil, err
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// PaymentByID is used to fetch the information about payment, by the
// given system payment id.
func (s *Server) PaymentByID(ctx context.Context, req *PaymentByIDRequest) (*Payment,
	error) {
	log.Tracef("command(%v), request(%v)", getFunctionName(), convertProtoMessage(req))

	payment, err := s.paymentsStore.PaymentByID(req.PaymentId)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(PaymentByIDReq, string(metrics.LowSeverity))
		return nil, err
	}

	resp, err := convertPaymentToProto(payment)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(PaymentByIDReq, string(metrics.LowSeverity))
		return nil, err
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, err
}

//
// PaymentsByReceipt is used to fetch the information about payment, by the
// given receipt.
func (s *Server) PaymentsByReceipt(ctx context.Context,
	req *PaymentsByReceiptRequest) (*PaymentsByReceiptResponse, error) {

	log.Tracef("command(%v), request(%v)", getFunctionName(), convertProtoMessage(req))

	payments, err := s.paymentsStore.PaymentByReceipt(req.Receipt)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(PaymentsByReceiptReq, string(metrics.LowSeverity))
		return nil, err
	}

	var protoPayments []*Payment
	for _, payment := range payments {
		protoPayment, err := convertPaymentToProto(payment)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(PaymentsByReceiptReq, string(metrics.LowSeverity))
			return nil, err
		}

		protoPayments = append(protoPayments, protoPayment)
	}

	resp := &PaymentsByReceiptResponse{
		Payments: protoPayments,
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}

//
// ListPayments returns list of payment which were registered by the
// system.
func (s *Server) ListPayments(ctx context.Context,
	req *ListPaymentsRequest) (*ListPaymentsResponse, error) {

	log.Tracef("command(%v), request(%v)", getFunctionName(), convertProtoMessage(req))

	var (
		asset     connectors.Asset
		status    connectors.PaymentStatus
		direction connectors.PaymentDirection
		media     connectors.PaymentMedia
		err       error
	)

	if req.Asset != Asset_ASSET_NONE {
		asset, err = ConvertAssetFromProto(req.Asset)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
			return nil, err
		}
	}

	if req.Direction != PaymentDirection_DIRECTION_NONE {
		direction, err = ConvertPaymentDirectionFromProto(req.Direction)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
			return nil, err
		}
	}

	if req.Status != PaymentStatus_STATUS_NONE {
		status, err = ConvertPaymentStatusFromProto(req.Status)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
			return nil, err
		}
	}

	if req.Media != Media_MEDIA_NONE {
		media, err = ConvertMediaFromProto(req.Media)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
			return nil, err
		}
	}

	payments, err := s.paymentsStore.ListPayments(asset, status, direction, media)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), error: %v", getFunctionName(), err)
		s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
		return nil, err
	}

	var protoPayments []*Payment
	for _, payment := range payments {
		protoPayment, err := convertPaymentToProto(payment)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), error: %v", getFunctionName(), err)
			s.metrics.AddError(ListPaymentsReq, string(metrics.LowSeverity))
			return nil, err
		}

		protoPayments = append(protoPayments, protoPayment)
	}

	resp := &ListPaymentsResponse{
		Payments: protoPayments,
	}

	log.Tracef("command(%v), response(%v)", getFunctionName(),
		convertProtoMessage(resp))

	return resp, nil
}
