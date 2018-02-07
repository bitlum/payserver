// package: crpc
// file: rpc.proto

var jspb = require("google-protobuf");
var rpc_pb = require("./rpc_pb");
var Connector = {
  serviceName: "crpc.Connector"
};
Connector.CreateAddress = {
  methodName: "CreateAddress",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.CreateAddressRequest,
  responseType: rpc_pb.Address
};
Connector.AccountAddress = {
  methodName: "AccountAddress",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.AccountAddressRequest,
  responseType: rpc_pb.Address
};
Connector.PendingBalance = {
  methodName: "PendingBalance",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.PendingBalanceRequest,
  responseType: rpc_pb.Balance
};
Connector.PendingTransactions = {
  methodName: "PendingTransactions",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.PendingTransactionsRequest,
  responseType: rpc_pb.PendingTransactionsResponse
};
Connector.GenerateTransaction = {
  methodName: "GenerateTransaction",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.GenerateTransactionRequest,
  responseType: rpc_pb.GenerateTransactionResponse
};
Connector.SendTransaction = {
  methodName: "SendTransaction",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.SendTransactionRequest,
  responseType: rpc_pb.EmtpyResponse
};
Connector.NetworkInfo = {
  methodName: "NetworkInfo",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.NetworkInfoRequest,
  responseType: rpc_pb.NetworkInfoResponse
};
Connector.CreateInvoice = {
  methodName: "CreateInvoice",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.CreateInvoiceRequest,
  responseType: rpc_pb.Invoice
};
Connector.SendPayment = {
  methodName: "SendPayment",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.SendPaymentRequest,
  responseType: rpc_pb.EmtpyResponse
};
Connector.CheckReachable = {
  methodName: "CheckReachable",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.CheckReachableRequest,
  responseType: rpc_pb.CheckReachableResponse
};
Connector.Estimate = {
  methodName: "Estimate",
  service: Connector,
  requestStream: false,
  responseStream: false,
  requestType: rpc_pb.EstimateRequest,
  responseType: rpc_pb.EstimationResponse
};
module.exports = {
  Connector: Connector,
};

