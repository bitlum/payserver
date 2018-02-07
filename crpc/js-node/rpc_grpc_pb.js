// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var rpc_pb = require('./rpc_pb.js');

function serialize_crpc_AccountAddressRequest(arg) {
  if (!(arg instanceof rpc_pb.AccountAddressRequest)) {
    throw new Error('Expected argument of type crpc.AccountAddressRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_AccountAddressRequest(buffer_arg) {
  return rpc_pb.AccountAddressRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_Address(arg) {
  if (!(arg instanceof rpc_pb.Address)) {
    throw new Error('Expected argument of type crpc.Address');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_Address(buffer_arg) {
  return rpc_pb.Address.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_Balance(arg) {
  if (!(arg instanceof rpc_pb.Balance)) {
    throw new Error('Expected argument of type crpc.Balance');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_Balance(buffer_arg) {
  return rpc_pb.Balance.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_CheckReachableRequest(arg) {
  if (!(arg instanceof rpc_pb.CheckReachableRequest)) {
    throw new Error('Expected argument of type crpc.CheckReachableRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_CheckReachableRequest(buffer_arg) {
  return rpc_pb.CheckReachableRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_CheckReachableResponse(arg) {
  if (!(arg instanceof rpc_pb.CheckReachableResponse)) {
    throw new Error('Expected argument of type crpc.CheckReachableResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_CheckReachableResponse(buffer_arg) {
  return rpc_pb.CheckReachableResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_CreateAddressRequest(arg) {
  if (!(arg instanceof rpc_pb.CreateAddressRequest)) {
    throw new Error('Expected argument of type crpc.CreateAddressRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_CreateAddressRequest(buffer_arg) {
  return rpc_pb.CreateAddressRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_CreateInvoiceRequest(arg) {
  if (!(arg instanceof rpc_pb.CreateInvoiceRequest)) {
    throw new Error('Expected argument of type crpc.CreateInvoiceRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_CreateInvoiceRequest(buffer_arg) {
  return rpc_pb.CreateInvoiceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_EmtpyResponse(arg) {
  if (!(arg instanceof rpc_pb.EmtpyResponse)) {
    throw new Error('Expected argument of type crpc.EmtpyResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_EmtpyResponse(buffer_arg) {
  return rpc_pb.EmtpyResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_EstimateRequest(arg) {
  if (!(arg instanceof rpc_pb.EstimateRequest)) {
    throw new Error('Expected argument of type crpc.EstimateRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_EstimateRequest(buffer_arg) {
  return rpc_pb.EstimateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_EstimationResponse(arg) {
  if (!(arg instanceof rpc_pb.EstimationResponse)) {
    throw new Error('Expected argument of type crpc.EstimationResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_EstimationResponse(buffer_arg) {
  return rpc_pb.EstimationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_GenerateTransactionRequest(arg) {
  if (!(arg instanceof rpc_pb.GenerateTransactionRequest)) {
    throw new Error('Expected argument of type crpc.GenerateTransactionRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_GenerateTransactionRequest(buffer_arg) {
  return rpc_pb.GenerateTransactionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_GenerateTransactionResponse(arg) {
  if (!(arg instanceof rpc_pb.GenerateTransactionResponse)) {
    throw new Error('Expected argument of type crpc.GenerateTransactionResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_GenerateTransactionResponse(buffer_arg) {
  return rpc_pb.GenerateTransactionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_Invoice(arg) {
  if (!(arg instanceof rpc_pb.Invoice)) {
    throw new Error('Expected argument of type crpc.Invoice');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_Invoice(buffer_arg) {
  return rpc_pb.Invoice.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_NetworkInfoRequest(arg) {
  if (!(arg instanceof rpc_pb.NetworkInfoRequest)) {
    throw new Error('Expected argument of type crpc.NetworkInfoRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_NetworkInfoRequest(buffer_arg) {
  return rpc_pb.NetworkInfoRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_NetworkInfoResponse(arg) {
  if (!(arg instanceof rpc_pb.NetworkInfoResponse)) {
    throw new Error('Expected argument of type crpc.NetworkInfoResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_NetworkInfoResponse(buffer_arg) {
  return rpc_pb.NetworkInfoResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_PendingBalanceRequest(arg) {
  if (!(arg instanceof rpc_pb.PendingBalanceRequest)) {
    throw new Error('Expected argument of type crpc.PendingBalanceRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_PendingBalanceRequest(buffer_arg) {
  return rpc_pb.PendingBalanceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_PendingTransactionsRequest(arg) {
  if (!(arg instanceof rpc_pb.PendingTransactionsRequest)) {
    throw new Error('Expected argument of type crpc.PendingTransactionsRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_PendingTransactionsRequest(buffer_arg) {
  return rpc_pb.PendingTransactionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_PendingTransactionsResponse(arg) {
  if (!(arg instanceof rpc_pb.PendingTransactionsResponse)) {
    throw new Error('Expected argument of type crpc.PendingTransactionsResponse');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_PendingTransactionsResponse(buffer_arg) {
  return rpc_pb.PendingTransactionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_SendPaymentRequest(arg) {
  if (!(arg instanceof rpc_pb.SendPaymentRequest)) {
    throw new Error('Expected argument of type crpc.SendPaymentRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_SendPaymentRequest(buffer_arg) {
  return rpc_pb.SendPaymentRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_crpc_SendTransactionRequest(arg) {
  if (!(arg instanceof rpc_pb.SendTransactionRequest)) {
    throw new Error('Expected argument of type crpc.SendTransactionRequest');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_crpc_SendTransactionRequest(buffer_arg) {
  return rpc_pb.SendTransactionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var ConnectorService = exports.ConnectorService = {
  //
  // CreateAddress is used to create deposit address in choosen blockchain
  // network.
  //
  // NOTE: Works only for blockchain daemons.
  createAddress: {
    path: '/crpc.Connector/CreateAddress',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.CreateAddressRequest,
    responseType: rpc_pb.Address,
    requestSerialize: serialize_crpc_CreateAddressRequest,
    requestDeserialize: deserialize_crpc_CreateAddressRequest,
    responseSerialize: serialize_crpc_Address,
    responseDeserialize: deserialize_crpc_Address,
  },
  //
  // AccountAddress return the deposit address of account.
  //
  // NOTE: Works only for blockchain daemons.
  accountAddress: {
    path: '/crpc.Connector/AccountAddress',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.AccountAddressRequest,
    responseType: rpc_pb.Address,
    requestSerialize: serialize_crpc_AccountAddressRequest,
    requestDeserialize: deserialize_crpc_AccountAddressRequest,
    responseSerialize: serialize_crpc_Address,
    responseDeserialize: deserialize_crpc_Address,
  },
  //
  // PendingBalance return the amount of funds waiting to be confirmed.
  //
  // NOTE: Works only for blockchain daemons.
  pendingBalance: {
    path: '/crpc.Connector/PendingBalance',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.PendingBalanceRequest,
    responseType: rpc_pb.Balance,
    requestSerialize: serialize_crpc_PendingBalanceRequest,
    requestDeserialize: deserialize_crpc_PendingBalanceRequest,
    responseSerialize: serialize_crpc_Balance,
    responseDeserialize: deserialize_crpc_Balance,
  },
  //
  // PendingTransactions return the transactions which has confirmation
  // number lower the required by payment system.
  //
  // NOTE: Works only for blockchain daemons.
  pendingTransactions: {
    path: '/crpc.Connector/PendingTransactions',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.PendingTransactionsRequest,
    responseType: rpc_pb.PendingTransactionsResponse,
    requestSerialize: serialize_crpc_PendingTransactionsRequest,
    requestDeserialize: deserialize_crpc_PendingTransactionsRequest,
    responseSerialize: serialize_crpc_PendingTransactionsResponse,
    responseDeserialize: deserialize_crpc_PendingTransactionsResponse,
  },
  //
  // GenerateTransaction generates raw blockchain transaction.
  //
  // NOTE: Blockchain endpoint.
  generateTransaction: {
    path: '/crpc.Connector/GenerateTransaction',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.GenerateTransactionRequest,
    responseType: rpc_pb.GenerateTransactionResponse,
    requestSerialize: serialize_crpc_GenerateTransactionRequest,
    requestDeserialize: deserialize_crpc_GenerateTransactionRequest,
    responseSerialize: serialize_crpc_GenerateTransactionResponse,
    responseDeserialize: deserialize_crpc_GenerateTransactionResponse,
  },
  //
  // SendTransaction send the given transaction to the blockchain network.
  //
  // NOTE: Works only for blockchain daemons.
  sendTransaction: {
    path: '/crpc.Connector/SendTransaction',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.SendTransactionRequest,
    responseType: rpc_pb.EmtpyResponse,
    requestSerialize: serialize_crpc_SendTransactionRequest,
    requestDeserialize: deserialize_crpc_SendTransactionRequest,
    responseSerialize: serialize_crpc_EmtpyResponse,
    responseDeserialize: deserialize_crpc_EmtpyResponse,
  },
  //
  // NetworkInfo returns information about the daemon and its network,
  // depending on the requested
  networkInfo: {
    path: '/crpc.Connector/NetworkInfo',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.NetworkInfoRequest,
    responseType: rpc_pb.NetworkInfoResponse,
    requestSerialize: serialize_crpc_NetworkInfoRequest,
    requestDeserialize: deserialize_crpc_NetworkInfoRequest,
    responseSerialize: serialize_crpc_NetworkInfoResponse,
    responseDeserialize: deserialize_crpc_NetworkInfoResponse,
  },
  //
  // CreateInvoice creates recept for sender lightning node which contains
  // the information about receiver node and
  //
  // NOTE: Works only for lightning network daemons.
  createInvoice: {
    path: '/crpc.Connector/CreateInvoice',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.CreateInvoiceRequest,
    responseType: rpc_pb.Invoice,
    requestSerialize: serialize_crpc_CreateInvoiceRequest,
    requestDeserialize: deserialize_crpc_CreateInvoiceRequest,
    responseSerialize: serialize_crpc_Invoice,
    responseDeserialize: deserialize_crpc_Invoice,
  },
  //
  // SendPayment is used to send specific amount of money inside lightning
  // network.
  //
  // NOTE: Works only for lightning network daemons.
  sendPayment: {
    path: '/crpc.Connector/SendPayment',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.SendPaymentRequest,
    responseType: rpc_pb.EmtpyResponse,
    requestSerialize: serialize_crpc_SendPaymentRequest,
    requestDeserialize: deserialize_crpc_SendPaymentRequest,
    responseSerialize: serialize_crpc_EmtpyResponse,
    responseDeserialize: deserialize_crpc_EmtpyResponse,
  },
  //
  // CheckReachable checks that given node can be reached from our
  // lightning node.
  //
  // NOTE: Works only for lightning network daemons.
  checkReachable: {
    path: '/crpc.Connector/CheckReachable',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.CheckReachableRequest,
    responseType: rpc_pb.CheckReachableResponse,
    requestSerialize: serialize_crpc_CheckReachableRequest,
    requestDeserialize: deserialize_crpc_CheckReachableRequest,
    responseSerialize: serialize_crpc_CheckReachableResponse,
    responseDeserialize: deserialize_crpc_CheckReachableResponse,
  },
  //
  // Estimate estimates the dollar price of the choosen asset.
  estimate: {
    path: '/crpc.Connector/Estimate',
    requestStream: false,
    responseStream: false,
    requestType: rpc_pb.EstimateRequest,
    responseType: rpc_pb.EstimationResponse,
    requestSerialize: serialize_crpc_EstimateRequest,
    requestDeserialize: deserialize_crpc_EstimateRequest,
    responseSerialize: serialize_crpc_EstimationResponse,
    responseDeserialize: deserialize_crpc_EstimationResponse,
  },
};

exports.ConnectorClient = grpc.makeGenericClientConstructor(ConnectorService);
