The `rpc` package is using  `gRPC` - open source remote procedure call (RPC)
system initially developed at Google. It uses HTTP/2 for transport, Protocol
Buffers as the interface description language, and provides features such as
authentication, bidirectional streaming and flow control, blocking or
non-blocking bindings, and cancellation and timeouts.

With `rpc.proto` API schema defined we generate cross-platform clients for many
languages, if API had changed, the generated clients also will be changed, and
 what you have to do is to update API by swapping old files with new one.

If you want to connect to exchange server via `gRPC` from web browser, for
example if you are developing browser extension which connects to our
exchange, than you should use, `grpc-web-client` and one of the generated
clients:
* `js-web` - client generated for making requests to exchange server from
browser using javascript. [Usage example]()
* `ts-web` - client generated for making requests to exchange server from
browser using typescript. [Usage example]()

If you want to connect to exchange server via `gRPC` from your own server,
for example if you are developing trading bot, you should use:
* `go` - client generated for usage in golang. [Usage example]()
* `js-node` - client generated for usage in nodejs. [Usage example]()
