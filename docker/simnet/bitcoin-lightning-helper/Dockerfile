FROM golang:1.10.2 AS builder

ARG LND_REVISION

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

RUN curl -L https://github.com/lightningnetwork/lnd/archive/$LND_REVISION.tar.gz \
| tar xz --strip 1

RUN dep ensure -v

RUN go install -v ./...



FROM python:2.7

EXPOSE 80

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lncli /usr/local/bin/

RUN pip install --no-cache-dir Flask

COPY http-server.py /

ENTRYPOINT python http-server.py