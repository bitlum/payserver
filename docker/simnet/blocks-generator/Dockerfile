FROM ubuntu:18.04 AS builder

ARG BITCOIN_VERSION
ARG BITCOIN_CASH_VERSION
ARG DASH_VERSION
ARG LITECOIN_VERSION

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

RUN curl https://bitcoin.org/bin/bitcoin-core-$BITCOIN_VERSION/bitcoin-${BITCOIN_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/bitcoin-cli

RUN mkdir /bitcoin-cash
RUN curl https://download.bitcoinabc.org/$BITCOIN_CASH_VERSION/linux/bitcoin-abc-${BITCOIN_CASH_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz -C /bitcoin-cash --wildcards --strip 2 \
*/bin/bitcoin-cli

RUN curl -L https://github.com/dashpay/dash/releases/download/v$DASH_VERSION/dashcore-${DASH_VERSION}-linux64.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/dash-cli

RUN curl https://download.litecoin.org/litecoin-$LITECOIN_VERSION/linux/litecoin-${LITECOIN_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/litecoin-cli


FROM ubuntu:18.04

# Blocks generation period in seconds
ARG PERIOD=10

# Setting runtime environment variables whhich will be used in `entrypoint.sh`
ENV PERIOD $PERIOD

# Copying CLI binaries from already builded blockchain containers. They
# should be builded already because this container depends on them.
COPY --from=builder bitcoin-cli /usr/local/bin/
COPY --from=builder bitcoin-cash/bitcoin-cli  /usr/local/bin/bitcoin-cash-cli
COPY --from=builder dash-cli /usr/local/bin/
COPY --from=builder litecoin-cli /usr/local/bin/

# Entrypoint script used for periodically blocks generation
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh