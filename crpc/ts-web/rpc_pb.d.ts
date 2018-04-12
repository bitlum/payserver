// package: crpc
// file: rpc.proto

import * as jspb from "google-protobuf";

export class EstimateRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getAmount(): string;
  setAmount(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EstimateRequest.AsObject;
  static toObject(includeInstance: boolean, msg: EstimateRequest): EstimateRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: EstimateRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EstimateRequest;
  static deserializeBinaryFromReader(message: EstimateRequest, reader: jspb.BinaryReader): EstimateRequest;
}

export namespace EstimateRequest {
  export type AsObject = {
    asset: string,
    amount: string,
  }
}

export class EstimationResponse extends jspb.Message {
  getUsd(): string;
  setUsd(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EstimationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: EstimationResponse): EstimationResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: EstimationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EstimationResponse;
  static deserializeBinaryFromReader(message: EstimationResponse, reader: jspb.BinaryReader): EstimationResponse;
}

export namespace EstimationResponse {
  export type AsObject = {
    usd: string,
  }
}

export class CreateAddressRequest extends jspb.Message {
  getAccount(): string;
  setAccount(value: string): void;

  getAsset(): string;
  setAsset(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateAddressRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateAddressRequest): CreateAddressRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateAddressRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateAddressRequest;
  static deserializeBinaryFromReader(message: CreateAddressRequest, reader: jspb.BinaryReader): CreateAddressRequest;
}

export namespace CreateAddressRequest {
  export type AsObject = {
    account: string,
    asset: string,
  }
}

export class AccountAddressRequest extends jspb.Message {
  getAccount(): string;
  setAccount(value: string): void;

  getAsset(): string;
  setAsset(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AccountAddressRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AccountAddressRequest): AccountAddressRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AccountAddressRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AccountAddressRequest;
  static deserializeBinaryFromReader(message: AccountAddressRequest, reader: jspb.BinaryReader): AccountAddressRequest;
}

export namespace AccountAddressRequest {
  export type AsObject = {
    account: string,
    asset: string,
  }
}

export class PendingBalanceRequest extends jspb.Message {
  getAccount(): string;
  setAccount(value: string): void;

  getAsset(): string;
  setAsset(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PendingBalanceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PendingBalanceRequest): PendingBalanceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PendingBalanceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PendingBalanceRequest;
  static deserializeBinaryFromReader(message: PendingBalanceRequest, reader: jspb.BinaryReader): PendingBalanceRequest;
}

export namespace PendingBalanceRequest {
  export type AsObject = {
    account: string,
    asset: string,
  }
}

export class PendingTransactionsRequest extends jspb.Message {
  getAccount(): string;
  setAccount(value: string): void;

  getAsset(): string;
  setAsset(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PendingTransactionsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PendingTransactionsRequest): PendingTransactionsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PendingTransactionsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PendingTransactionsRequest;
  static deserializeBinaryFromReader(message: PendingTransactionsRequest, reader: jspb.BinaryReader): PendingTransactionsRequest;
}

export namespace PendingTransactionsRequest {
  export type AsObject = {
    account: string,
    asset: string,
  }
}

export class GenerateTransactionResponse extends jspb.Message {
  getRawTx(): Uint8Array | string;
  getRawTx_asU8(): Uint8Array;
  getRawTx_asB64(): string;
  setRawTx(value: Uint8Array | string): void;

  getTxId(): string;
  setTxId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateTransactionResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateTransactionResponse): GenerateTransactionResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateTransactionResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateTransactionResponse;
  static deserializeBinaryFromReader(message: GenerateTransactionResponse, reader: jspb.BinaryReader): GenerateTransactionResponse;
}

export namespace GenerateTransactionResponse {
  export type AsObject = {
    rawTx: Uint8Array | string,
    txId: string,
  }
}

export class SubcribeOnPaymentsRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getType(): string;
  setType(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubcribeOnPaymentsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubcribeOnPaymentsRequest): SubcribeOnPaymentsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SubcribeOnPaymentsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubcribeOnPaymentsRequest;
  static deserializeBinaryFromReader(message: SubcribeOnPaymentsRequest, reader: jspb.BinaryReader): SubcribeOnPaymentsRequest;
}

export namespace SubcribeOnPaymentsRequest {
  export type AsObject = {
    asset: string,
    type: string,
  }
}

export class BlockchainPendingPayment extends jspb.Message {
  hasPayment(): boolean;
  clearPayment(): void;
  getPayment(): Payment | undefined;
  setPayment(value?: Payment): void;

  getConfirmations(): number;
  setConfirmations(value: number): void;

  getConfirmationsLeft(): number;
  setConfirmationsLeft(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BlockchainPendingPayment.AsObject;
  static toObject(includeInstance: boolean, msg: BlockchainPendingPayment): BlockchainPendingPayment.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BlockchainPendingPayment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BlockchainPendingPayment;
  static deserializeBinaryFromReader(message: BlockchainPendingPayment, reader: jspb.BinaryReader): BlockchainPendingPayment;
}

export namespace BlockchainPendingPayment {
  export type AsObject = {
    payment?: Payment.AsObject,
    confirmations: number,
    confirmationsLeft: number,
  }
}

export class Payment extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getAmount(): string;
  setAmount(value: string): void;

  getAccount(): string;
  setAccount(value: string): void;

  getAddress(): string;
  setAddress(value: string): void;

  getType(): string;
  setType(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Payment.AsObject;
  static toObject(includeInstance: boolean, msg: Payment): Payment.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Payment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Payment;
  static deserializeBinaryFromReader(message: Payment, reader: jspb.BinaryReader): Payment;
}

export namespace Payment {
  export type AsObject = {
    id: string,
    amount: string,
    account: string,
    address: string,
    type: string,
  }
}

export class EmtpyResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EmtpyResponse.AsObject;
  static toObject(includeInstance: boolean, msg: EmtpyResponse): EmtpyResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: EmtpyResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EmtpyResponse;
  static deserializeBinaryFromReader(message: EmtpyResponse, reader: jspb.BinaryReader): EmtpyResponse;
}

export namespace EmtpyResponse {
  export type AsObject = {
  }
}

export class Balance extends jspb.Message {
  getData(): string;
  setData(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Balance.AsObject;
  static toObject(includeInstance: boolean, msg: Balance): Balance.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Balance, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Balance;
  static deserializeBinaryFromReader(message: Balance, reader: jspb.BinaryReader): Balance;
}

export namespace Balance {
  export type AsObject = {
    data: string,
  }
}

export class Address extends jspb.Message {
  getData(): string;
  setData(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Address.AsObject;
  static toObject(includeInstance: boolean, msg: Address): Address.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Address, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Address;
  static deserializeBinaryFromReader(message: Address, reader: jspb.BinaryReader): Address;
}

export namespace Address {
  export type AsObject = {
    data: string,
  }
}

export class Invoice extends jspb.Message {
  getData(): string;
  setData(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Invoice.AsObject;
  static toObject(includeInstance: boolean, msg: Invoice): Invoice.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Invoice, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Invoice;
  static deserializeBinaryFromReader(message: Invoice, reader: jspb.BinaryReader): Invoice;
}

export namespace Invoice {
  export type AsObject = {
    data: string,
  }
}

export class CheckReachableRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getIdentityKey(): string;
  setIdentityKey(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CheckReachableRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CheckReachableRequest): CheckReachableRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CheckReachableRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CheckReachableRequest;
  static deserializeBinaryFromReader(message: CheckReachableRequest, reader: jspb.BinaryReader): CheckReachableRequest;
}

export namespace CheckReachableRequest {
  export type AsObject = {
    asset: string,
    identityKey: string,
  }
}

export class PendingTransactionsResponse extends jspb.Message {
  clearPaymentsList(): void;
  getPaymentsList(): Array<BlockchainPendingPayment>;
  setPaymentsList(value: Array<BlockchainPendingPayment>): void;
  addPayments(value?: BlockchainPendingPayment, index?: number): BlockchainPendingPayment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PendingTransactionsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: PendingTransactionsResponse): PendingTransactionsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PendingTransactionsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PendingTransactionsResponse;
  static deserializeBinaryFromReader(message: PendingTransactionsResponse, reader: jspb.BinaryReader): PendingTransactionsResponse;
}

export namespace PendingTransactionsResponse {
  export type AsObject = {
    paymentsList: Array<BlockchainPendingPayment.AsObject>,
  }
}

export class GenerateTransactionRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getReceiverAddress(): string;
  setReceiverAddress(value: string): void;

  getAmount(): string;
  setAmount(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateTransactionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateTransactionRequest): GenerateTransactionRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateTransactionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateTransactionRequest;
  static deserializeBinaryFromReader(message: GenerateTransactionRequest, reader: jspb.BinaryReader): GenerateTransactionRequest;
}

export namespace GenerateTransactionRequest {
  export type AsObject = {
    asset: string,
    receiverAddress: string,
    amount: string,
  }
}

export class SendTransactionRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getRawTx(): Uint8Array | string;
  getRawTx_asU8(): Uint8Array;
  getRawTx_asB64(): string;
  setRawTx(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SendTransactionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SendTransactionRequest): SendTransactionRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SendTransactionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SendTransactionRequest;
  static deserializeBinaryFromReader(message: SendTransactionRequest, reader: jspb.BinaryReader): SendTransactionRequest;
}

export namespace SendTransactionRequest {
  export type AsObject = {
    asset: string,
    rawTx: Uint8Array | string,
  }
}

export class InfoRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): InfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: InfoRequest): InfoRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: InfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): InfoRequest;
  static deserializeBinaryFromReader(message: InfoRequest, reader: jspb.BinaryReader): InfoRequest;
}

export namespace InfoRequest {
  export type AsObject = {
  }
}

export class LightningInfo extends jspb.Message {
  getHost(): string;
  setHost(value: string): void;

  getPort(): string;
  setPort(value: string): void;

  getMinAmount(): string;
  setMinAmount(value: string): void;

  getMaxAmount(): string;
  setMaxAmount(value: string): void;

  getIdentityPubkey(): string;
  setIdentityPubkey(value: string): void;

  getAlias(): string;
  setAlias(value: string): void;

  getNumPendingChannels(): number;
  setNumPendingChannels(value: number): void;

  getNumActiveChannels(): number;
  setNumActiveChannels(value: number): void;

  getNumPeers(): number;
  setNumPeers(value: number): void;

  getBlockHeight(): number;
  setBlockHeight(value: number): void;

  getBlockHash(): string;
  setBlockHash(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LightningInfo.AsObject;
  static toObject(includeInstance: boolean, msg: LightningInfo): LightningInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: LightningInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LightningInfo;
  static deserializeBinaryFromReader(message: LightningInfo, reader: jspb.BinaryReader): LightningInfo;
}

export namespace LightningInfo {
  export type AsObject = {
    host: string,
    port: string,
    minAmount: string,
    maxAmount: string,
    identityPubkey: string,
    alias: string,
    numPendingChannels: number,
    numActiveChannels: number,
    numPeers: number,
    blockHeight: number,
    blockHash: string,
  }
}

export class InfoResponse extends jspb.Message {
  getNet(): Net;
  setNet(value: Net): void;

  getTime(): string;
  setTime(value: string): void;

  hasLightingInfo(): boolean;
  clearLightingInfo(): void;
  getLightingInfo(): LightningInfo | undefined;
  setLightingInfo(value?: LightningInfo): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): InfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: InfoResponse): InfoResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: InfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): InfoResponse;
  static deserializeBinaryFromReader(message: InfoResponse, reader: jspb.BinaryReader): InfoResponse;
}

export namespace InfoResponse {
  export type AsObject = {
    net: Net,
    time: string,
    lightingInfo?: LightningInfo.AsObject,
  }
}

export class CreateInvoiceRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getAccount(): string;
  setAccount(value: string): void;

  getAmount(): string;
  setAmount(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateInvoiceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateInvoiceRequest): CreateInvoiceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateInvoiceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateInvoiceRequest;
  static deserializeBinaryFromReader(message: CreateInvoiceRequest, reader: jspb.BinaryReader): CreateInvoiceRequest;
}

export namespace CreateInvoiceRequest {
  export type AsObject = {
    asset: string,
    account: string,
    amount: string,
  }
}

export class SendPaymentRequest extends jspb.Message {
  getAsset(): string;
  setAsset(value: string): void;

  getInvoice(): string;
  setInvoice(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SendPaymentRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SendPaymentRequest): SendPaymentRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SendPaymentRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SendPaymentRequest;
  static deserializeBinaryFromReader(message: SendPaymentRequest, reader: jspb.BinaryReader): SendPaymentRequest;
}

export namespace SendPaymentRequest {
  export type AsObject = {
    asset: string,
    invoice: string,
  }
}

export class CheckReachableResponse extends jspb.Message {
  getIsreachable(): boolean;
  setIsreachable(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CheckReachableResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CheckReachableResponse): CheckReachableResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CheckReachableResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CheckReachableResponse;
  static deserializeBinaryFromReader(message: CheckReachableResponse, reader: jspb.BinaryReader): CheckReachableResponse;
}

export namespace CheckReachableResponse {
  export type AsObject = {
    isreachable: boolean,
  }
}

export enum Asset {
  BTC = 0,
  BCH = 1,
  ETH = 2,
  LTC = 3,
  DASH = 4,
}

export enum Market {
  BTCETH = 0,
  BTCBTH = 1,
  BTCLTC = 2,
  BTCDASH = 3,
  ETHLTC = 4,
}

export enum Net {
  Simnet = 0,
  Testnet = 1,
  Mainnet = 2,
}

