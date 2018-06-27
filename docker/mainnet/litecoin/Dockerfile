FROM ubuntu:18.04 AS builder

ARG LITECOIN_VERSION

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

RUN curl https://download.litecoin.org/litecoin-$LITECOIN_VERSION/linux/litecoin-${LITECOIN_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/litecoind \
*/bin/litecoin-cli



FROM ubuntu:18.04

# RPC port.
EXPOSE 12332

# P2P port.
EXPOSE 12333

# Copying required binaries from builder stage.
COPY --from=builder litecoind litecoin-cli /usr/local/bin/

# Default config used to initalize datadir volume
# at first or cleaned deploy.
COPY litecoin.mainnet.conf /root/default/litecoin.conf

# Entrypoint script used to init datadir if required and for
# starting litecoin daemon.
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]