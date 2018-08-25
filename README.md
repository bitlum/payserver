Pay Server - is the blockchain microservice which is working on [zigzag.io](zigzag.io), and service as unified API for other microservices to receive and send cryptocurrency.

| State  | Feature |
| ------------- | ------------- |
| implemented  | Unify payment API for BTC, LTC, DASH, ETH, BCH, and Lightning Network  |
| implemented  | Report health statistics about internal state of synchronisation, fees, request delays, sent and received volume, amount of fees spent on payments |
| not implemented | Payment re-try in case of failure |
| not implemented | UTXO re-orginisation |
| not implemented | Lightning Network channel re-balancing |
|not implemented|Support of payment HTLC on addresses|

```
GRPC API:

    // CreateReceipt is used to create blockchain deposit address in
    // case of blockchain media, and lightning network invoice in
    // case of the lightning media, which will be used to receive money from
    // external entity.
    rpc CreateReceipt (CreateReceiptRequest) returns (CreateReceiptResponse);

    // ValidateReceipt is used to validate receipt for given asset and media.
    rpc ValidateReceipt (ValidateReceiptRequest) returns (EmptyResponse);

    // Balance is used to determine balance.
    rpc Balance (BalanceRequest) returns (BalanceResponse);

    // EstimateFee estimates the fee of the payment.
    rpc EstimateFee (EstimateFeeRequest) returns (EstimateFeeResponse);

    // SendPayment sends payment to the given recipient,
    // ensures in the validity of the receipt as well as the
    // account has enough money for doing that.
    rpc SendPayment (SendPaymentRequest) returns (Payment);

    // PaymentByID is used to fetch the information about payment, by the
    // given system payment id.
    rpc PaymentByID (PaymentByIDRequest) returns (Payment);

    // PaymentsByReceipt is used to fetch the information about payment, by the
    // given receipt.
    rpc PaymentsByReceipt (PaymentsByReceiptRequest) returns (PaymentsByReceiptResponse);

    // ListPayments returnes list of payment which were registered by the
    // system.
    rpc ListPayments (ListPaymentsRequest) returns (ListPaymentsResponse);
```
