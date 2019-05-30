FROM golang:1.12 AS builder

WORKDIR $GOPATH/src/github.com/bitlum/connector/

ARG CONNECTOR_REVISION

RUN curl -L https://github.com/bitlum/connector/archive/$CONNECTOR_REVISION.tar.gz \
| tar xz --strip 1

RUN GO111MODULE=on go get
RUN GO111MODULE=on go install . ./cmd/...

FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/connector /usr/local/bin
COPY --from=builder /go/bin/pscli /usr/local/bin

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY connector.mainnet.conf /root/default/connector.conf

# Entrypoint script used to init datadir if required and for
# starting daemon
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]
