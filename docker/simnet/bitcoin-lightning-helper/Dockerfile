FROM golang:1.10.3 AS builder

ARG BITCOIN_LIGHTNING_REVISION

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/lightningnetwork/lnd

RUN git clone https://github.com/lightningnetwork/lnd.git .

RUN git checkout $BITCOIN_LIGHTNING_REVISION

RUN dep ensure -v

RUN make install



FROM python:2.7

EXPOSE 80

# Copying required binaries from builder stage.
COPY --from=builder /go/bin/lncli /usr/local/bin/

RUN pip install --no-cache-dir Flask

COPY http-server.py /

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["python", "http-server.py"]