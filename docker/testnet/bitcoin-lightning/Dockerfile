FROM golang:1.10.3 AS builder

ARG BITCOIN_LIGHTNING_REVISION

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

RUN git clone https://github.com/lightningnetwork/lnd.git .

RUN git checkout $BITCOIN_LIGHTNING_REVISION

RUN dep ensure -v

RUN make install



FROM ubuntu:18.04

# RPC port.
EXPOSE 10009

# P2P port.
EXPOSE 9735

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lnd /go/bin/lncli /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY bitcoin-lightning.testnet.conf /root/default/lnd.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh