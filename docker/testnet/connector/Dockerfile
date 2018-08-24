FROM golang:1.10.3 AS builder

ARG CONNECTOR_REVISION

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/bitlum/connector/

RUN curl -L https://github.com/bitlum/connector/archive/$CONNECTOR_REVISION.tar.gz \
| tar xz --strip 1

RUN dep ensure -v

RUN go install . ./cmd/...



FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

# This implies that service has to be built locally first, and putted
# in the docker directory, for running docker build.
COPY --from=builder /go/bin/connector /usr/local/bin
COPY --from=builder /go/bin/pscli /usr/local/bin

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY connector.testnet.conf /root/default/connector.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]
