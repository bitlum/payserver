FROM golang:1.10.3 AS builder

ARG BITCOIN_LIGHTNING_REVISION

# Install dep for installing lightning network daemon dependencies
RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

# Instead of cloning lightningnetwork/lnd temproray use ourw version of lnd
# daemon, but put in lightningnetwork/lnd directory so that all installation
# scripts could work without being changed.
RUN git clone https://github.com/bitlum/lnd.git .

# Ensure that we running proper revision
RUN git checkout $BITCOIN_LIGHTNING_REVISION

# Install dependencies
RUN dep ensure -v

# Build lightning network daemon
RUN make install

# Install delve - debugger for the Go programming language, in order to be
# able to attach to the lnd process and understand why it stuck.
RUN go get -u github.com/derekparker/delve/cmd/dlv


FROM ubuntu:18.04

# RPC port.
EXPOSE 10009

# P2P port.
EXPOSE 9735

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lnd /go/bin/lncli /go/bin/dlv /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY bitcoin-lightning.testnet.conf /root/default/lnd.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]