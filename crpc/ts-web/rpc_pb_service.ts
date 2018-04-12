// package: crpc
// file: rpc.proto

import * as rpc_pb from "./rpc_pb";
export class Connector {
  static serviceName = "crpc.Connector";
}
export namespace Connector {
  export class CreateAddress {
    static readonly methodName = "CreateAddress";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.CreateAddressRequest;
    static readonly responseType = rpc_pb.Address;
  }
  export class AccountAddress {
    static readonly methodName = "AccountAddress";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.AccountAddressRequest;
    static readonly responseType = rpc_pb.Address;
  }
  export class PendingBalance {
    static readonly methodName = "PendingBalance";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.PendingBalanceRequest;
    static readonly responseType = rpc_pb.Balance;
  }
  export class PendingTransactions {
    static readonly methodName = "PendingTransactions";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.PendingTransactionsRequest;
    static readonly responseType = rpc_pb.PendingTransactionsResponse;
  }
  export class GenerateTransaction {
    static readonly methodName = "GenerateTransaction";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.GenerateTransactionRequest;
    static readonly responseType = rpc_pb.GenerateTransactionResponse;
  }
  export class SendTransaction {
    static readonly methodName = "SendTransaction";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.SendTransactionRequest;
    static readonly responseType = rpc_pb.EmtpyResponse;
  }
  export class Info {
    static readonly methodName = "Info";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.InfoRequest;
    static readonly responseType = rpc_pb.InfoResponse;
  }
  export class CreateInvoice {
    static readonly methodName = "CreateInvoice";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.CreateInvoiceRequest;
    static readonly responseType = rpc_pb.Invoice;
  }
  export class SendPayment {
    static readonly methodName = "SendPayment";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.SendPaymentRequest;
    static readonly responseType = rpc_pb.EmtpyResponse;
  }
  export class CheckReachable {
    static readonly methodName = "CheckReachable";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.CheckReachableRequest;
    static readonly responseType = rpc_pb.CheckReachableResponse;
  }
  export class Estimate {
    static readonly methodName = "Estimate";
    static readonly service = Connector;
    static readonly requestStream = false;
    static readonly responseStream = false;
    static readonly requestType = rpc_pb.EstimateRequest;
    static readonly responseType = rpc_pb.EstimationResponse;
  }
}
