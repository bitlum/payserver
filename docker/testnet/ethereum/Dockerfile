FROM golang:1.10.3 AS builder

ARG ETHEREUM_REVISION

WORKDIR /ethereum

RUN curl -L https://github.com/bitlum/go-ethereum/archive/$ETHEREUM_REVISION.tar.gz \
| tar xz --strip 1

RUN make geth



FROM ubuntu:18.04

# RPC port.
EXPOSE 11332

# RPC-WS port.
EXPOSE 11331

# P2P port.
EXPOSE 11333

# Copying required binaries from builder stage.
COPY --from=builder /ethereum/build/bin/geth /usr/local/bin/

# Default config and genesis used to initalize datadir volume
# at first or cleaned deploy.
COPY ethereum.testnet.conf /root/default/ethereum.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh