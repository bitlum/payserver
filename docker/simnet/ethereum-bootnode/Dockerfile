FROM golang:1.10.3 AS builder

ARG ETHEREUM_REVISION

WORKDIR /ethereum

RUN curl -L https://github.com/bitlum/go-ethereum/archive/$ETHEREUM_REVISION.tar.gz \
| tar xz --strip 1

RUN make all



FROM ubuntu:18.04

# P2P port
EXPOSE 30301

# Copying required binaries from builder stage
COPY --from=builder /ethereum/build/bin/bootnode /usr/local/bin/

# Entrypoint script used to init datadir if required and for
# starting bootnode daemon
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]