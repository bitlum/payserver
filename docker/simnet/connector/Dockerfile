FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# This implies that service has to be built locally first, and putted
# in the docker directory, for running docker build.
COPY bin/connector /usr/local/bin

# Default config used to initalize datadir volume at first or
# cleaned deploy.
COPY connector.simnet.conf /root/default/connector.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

# We are using exec syntax to enable gracefull shutdown. Check
# http://veithen.github.io/2014/11/16/sigterm-propagation.html.
ENTRYPOINT ["bash", "/root/entrypoint.sh"]