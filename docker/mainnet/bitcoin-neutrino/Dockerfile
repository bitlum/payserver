FROM golang:1.10.3 AS builder

ARG BITCOIN_NEUTRINO_REVISION

RUN go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/roasbeef/btcd

RUN curl -L https://github.com/roasbeef/btcd/archive/$BITCOIN_NEUTRINO_REVISION.tar.gz \
| tar xz --strip 1

RUN glide install

RUN go install -v . ./cmd/btcctl



FROM ubuntu:18.04

# P2P port.
EXPOSE 18333

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/btcd /go/bin/btcctl /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY bitcoin-neutrino.mainnet.conf /root/default/btcd.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]