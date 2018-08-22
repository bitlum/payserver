FROM ubuntu:18.04 AS builder

ARG DASH_VERSION

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

RUN curl -L https://github.com/dashpay/dash/releases/download/v$DASH_VERSION/dashcore-${DASH_VERSION}-x86_64-linux-gnu.tar.gz \
| tar xz --wildcards --strip 2 \
*/bin/dashd \
*/bin/dash-cli



FROM ubuntu:18.04

# RPC port.
EXPOSE 10332

# P2P port.
EXPOSE 10333

# Copying required binaries from builder stage.
COPY --from=builder dashd dash-cli /usr/local/bin/

# Default config used to initalize datadir volume at first or
# cleaned deploy. It will be restored and used after each restart.
COPY dash.testnet.conf /root/default/dash.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon.
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]