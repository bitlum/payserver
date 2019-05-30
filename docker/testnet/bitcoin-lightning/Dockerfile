FROM golang:1.11-alpine as builder

ARG BITCOIN_LIGHTNING_REVISION

# Install dependencies and install/build lnd.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

# Copy from repository to build from.
RUN git clone https://github.com/lightningnetwork/lnd.git /go/src/github.com/lightningnetwork/lnd

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

RUN cd /go/src/github.com/lightningnetwork/lnd \
&&  git checkout $BITCOIN_LIGHTNING_REVISION \
&&  make build-itest \
&&  mv lnd-itest /go/bin/lnd \
&&  mv lncli-itest /go/bin/lncli

# Install delve - debugger for the Go programming language, in order to be
# able to attach to the lnd process and understand why it stuck.
RUN go get -u github.com/derekparker/delve/cmd/dlv

# Start a new, final image to reduce size.
FROM alpine as final

# Expose lnd ports (server, rpc).
EXPOSE 9735 10009

# Add bash.
RUN apk add --no-cache \
    bash

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lnd /usr/local/bin/
COPY --from=builder /go/bin/lncli /usr/local/bin/
COPY --from=builder /go/bin/dlv /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY bitcoin-lightning.testnet.conf /root/default/lnd.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]