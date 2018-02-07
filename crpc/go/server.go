package crpc

import (
	"github.com/bitlum/connector/common"
	"github.com/bitlum/connector/common/core"
	"github.com/bitlum/connector/estimator"
	"github.com/btcsuite/btclog"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/context"
)

// Server...
type Server struct {
	blockchainConnectors map[core.AssetType]common.BlockchainConnector
	lightningConnectors  map[core.AssetType]common.LightningConnector
	estmtr               estimator.USDEstimator
	log                  common.NamedLogger
}

// A compile time check to ensure that Server fully implements the
// ExchangeServer gRPC service.
var _ ConnectorServer = (*Server)(nil)

// NewRPCServer creates and returns a new instance of the Server.
func NewRPCServer(
	blockchainConnectors map[core.AssetType]common.BlockchainConnector,
	lightningConnectors map[core.AssetType]common.LightningConnector,
	estmtr estimator.USDEstimator,
	log btclog.Logger) (*Server, error) {
	return &Server{
		blockchainConnectors: blockchainConnectors,
		lightningConnectors:  lightningConnectors,
		estmtr:               estmtr,
		log: common.NamedLogger{
			Logger: log,
			Name:   "RPC",
		},
	}, nil
}

//
// CreateAddress is used to create deposit address in choosen blockchain
// network.
//
// NOTE: Works only for blockchain daemons.
func (s *Server) CreateAddress(_ context.Context, req *CreateAddressRequest) (*Address,
	error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "create address")
	}

	address, err := c.CreateAddress(req.Account)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &Address{
		Data: address,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// AccountAddress return the deposit address of account.
//
// NOTE: Works only for blockchain daemons.
func (s *Server) AccountAddress(_ context.Context,
	req *AccountAddressRequest) (*Address, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "account address")
	}

	address, err := c.AccountAddress(req.Account)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &Address{
		Data: address,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// PendingBalance return the amount of funds waiting to be confirmed.
//
// NOTE: Works only for blockchain daemons.
func (s *Server) PendingBalance(_ context.Context,
	req *PendingBalanceRequest) (*Balance, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "pending balance")
	}

	balance, err := c.PendingBalance(req.Account)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &Balance{
		Data: balance,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// PendingTransactions return the transactions which has confirmation
// number lower the required by payment system.
//
// NOTE: Works only for blockchain daemons.
func (s *Server) PendingTransactions(_ context.Context,
	req *PendingTransactionsRequest) (*PendingTransactionsResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "pending transactions")
	}

	txs, err := c.PendingTransactions(req.Account)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	payments := make([]*BlockchainPendingPayment, len(txs))
	for i, tx := range txs {
		payments[i] = &BlockchainPendingPayment{
			Payment: &Payment{
				Id:      tx.ID,
				Amount:  tx.Amount.String(),
				Account: tx.Account,
				Address: tx.Address,
				Type:    string(tx.Type),
			},
			Confirmations:     tx.Confirmations,
			ConfirmationsLeft: tx.ConfirmationsLeft,
		}
	}

	resp := &PendingTransactionsResponse{
		Payments: payments,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// GenerateTransaction generates raw blockchain transaction.
//
// NOTE: Blockchain endpoint.
func (s *Server) GenerateTransaction(_ context.Context,
	req *GenerateTransactionRequest) (*GenerateTransactionResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "generate transaction")
	}

	genTx, err := c.GenerateTransaction(req.ReceiverAddress, req.Amount)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &GenerateTransactionResponse{
		RawTx: genTx.Bytes(),
		TxId:  genTx.ID(),
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// SendTransaction send the given transaction to the blockchain network.
//
// NOTE: Works only for blockchain daemons.
func (s *Server) SendTransaction(_ context.Context,
	req *SendTransactionRequest) (*EmtpyResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.blockchainConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "send transaction")
	}

	if err := c.SendTransaction(req.RawTx); err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &EmtpyResponse{}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// NetworkInfo returns information about the daemon and its network,
// depending on the requested
func (s *Server) NetworkInfo(_ context.Context,
	req *NetworkInfoRequest) (*NetworkInfoResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	if req.Type == string(common.Blockchain) {
		return nil, newErrNetworkNotSupported(string(common.Blockchain),
			"network info")
	}

	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "network info")
	}

	info, err := c.Info()
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &NetworkInfoResponse{
		Data: &NetworkInfoResponse_LightingInfo{
			LightingInfo: &LightningInfo{
				Host:               info.Host,
				Port:               info.Port,
				MinAmount:          info.MinAmount,
				MaxAmount:          info.MaxAmount,
				IdentityPubkey:     info.IdentityPubkey,
				Alias:              info.Alias,
				NumPendingChannels: info.NumPendingChannels,
				NumActiveChannels:  info.NumActiveChannels,
				NumPeers:           info.NumPeers,
				BlockHeight:        info.BlockHeight,
				BlockHash:          info.BlockHash,
				SyncedToChain:      info.SyncedToChain,
				Testnet:            info.Testnet,
				Chains:             info.Chains,
			},
		},
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// CreateInvoice creates recept for sender lightning node which contains
// the information about receiver node and
//
// NOTE: Works only for lightning network daemons.
func (s *Server) CreateInvoice(_ context.Context,
	req *CreateInvoiceRequest) (*Invoice, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "create invoice")
	}

	invoice, err := c.CreateInvoice(req.Account, req.Amount)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &Invoice{
		Data: invoice,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// SendPayment is used to send specific amount of money inside lightning
// network.
//
// NOTE: Works only for lightning network daemons.
func (s *Server) SendPayment(_ context.Context,
	req *SendPaymentRequest) (*EmtpyResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "create invoice")
	}

	if err := c.SendTo(req.Invoice); err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &EmtpyResponse{}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// CheckReachable checks that given node can be reached from our
// lightning node.
//
// NOTE: Works only for lightning network daemons.
func (s *Server) CheckReachable(_ context.Context,
	req *CheckReachableRequest) (*CheckReachableResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	c, ok := s.lightningConnectors[core.AssetType(req.Asset)]
	if !ok {
		return nil, newErrAssetNotSupported(req.Asset, "create invoice")
	}

	amount := "0.00000001"
	routes, err := c.QueryRoutes(req.IdentityKey, amount)
	if err != nil {
		// TODO(andrew.shvv) distinguish errors
		return &CheckReachableResponse{
			IsReachable: false,
		}, nil
	}

	if len(routes) != 0 {
		return &CheckReachableResponse{
			IsReachable: true,
		}, nil
	}

	resp := &CheckReachableResponse{
		IsReachable: false,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}

//
// Estimate estimates the dollar price of the choosen asset.
func (s *Server) Estimate(_ context.Context,
	req *EstimateRequest) (*EstimationResponse, error) {

	s.log.Tracef("command(%v), request(%v)", getFunctionName(), spew.Sdump(req))

	usdEstimation, err := s.estmtr.Estimate(req.Asset, req.Amount)
	if err != nil {
		return nil, newErrInternal(err.Error())
	}

	resp := &EstimationResponse{
		Usd: usdEstimation,
	}

	s.log.Tracef("command(%v), response(%v)", getFunctionName(),
		spew.Sdump(resp))

	return resp, nil
}
