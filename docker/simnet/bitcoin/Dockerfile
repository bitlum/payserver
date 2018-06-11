FROM ubuntu:18.04 AS builder

ARG BITCOIN_VERSION

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

RUN curl https://bitcoin.org/bin/bitcoin-core-$BITCOIN_VERSION/bitcoin-${BITCOIN_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/bitcoind \
*/bin/bitcoin-cli



FROM ubuntu:18.04

# ROLE is bitcoin node role: primary or secondary.
#
# Primary role means that this node will init new blockchain if it not
# exists during deploy or restart.
#
# Secondary rank means that this node will try to connect to primary
# node and use blockchain of latter.
ARG ROLE

# RPC port.
EXPOSE 8332

# P2P port.
EXPOSE 8333

# `zmqpubrawblock` and `zmqpubrawtx` port.
EXPOSE 8334

# Copying required binaries from builder stage.
COPY --from=builder bitcoind bitcoin-cli /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy.
COPY bitcoin.simnet.$ROLE.conf /root/default/bitcoin.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin daemon.
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh