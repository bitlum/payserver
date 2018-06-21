FROM golang:1.10.3 AS builder

ARG BITCOIN_LIGHTNING_REVISION

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

RUN git clone https://github.com/lightningnetwork/lnd.git .

RUN git checkout $BITCOIN_LIGHTNING_REVISION

RUN dep ensure -v

RUN make install



FROM ubuntu:18.04

# ROLE is bitcoin node role: primary or secondary.
#
# Primary role means that this node will init new blockchain if it not
# exists during deploy or restart.
#
# Secondary rank means that this node will try to connect to primary
# node and use blockchain of latter.
ARG ROLE

# P2P port.
EXPOSE 9735

# RPC port.
EXPOSE 10009

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lnd /go/bin/lncli /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy.
COPY bitcoin-lightning.simnet.$ROLE.conf /root/default/lnd.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh