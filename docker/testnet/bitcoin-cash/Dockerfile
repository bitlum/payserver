FROM ubuntu:18.04 AS builder

ARG BITCOIN_CASH_VERSION

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

RUN curl https://download.bitcoinabc.org/$BITCOIN_CASH_VERSION/linux/bitcoin-abc-${BITCOIN_CASH_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/bitcoind \
*/bin/bitcoin-cli



FROM ubuntu:18.04

# RPC port.
EXPOSE 9332

# P2P port.
EXPOSE 9333

# Copying required binaries from builder stage.
COPY --from=builder bitcoind  /usr/local/bin/bitcoin-cashd
COPY --from=builder bitcoin-cli /usr/local/bin/bitcoin-cash-cli

# Default config used to initalize datadir volume at first or
# cleaned deploy.
COPY bitcoin.testnet.conf /root/default/bitcoin.conf

# Entrypoint script used to init datadir if required and for
# starting bitcoin cash daemon.
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh