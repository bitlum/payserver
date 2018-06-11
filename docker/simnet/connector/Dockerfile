FROM ubuntu:18.04

# ROLE is bitcoin node role: primary or secondary.
#
# Primary role means that this node will init new blockchain if it not
# exists during deploy or restart.
#
# Secondary rank means that this node will try to connect to primary
# node and use blockchain of latter.
ARG ROLE

RUN apt-get update && apt-get install -y \
ca-certificates \
curl \
&& rm -rf /var/lib/apt/lists/*

# This implies that service has to be built locally first, and putted
# in the docker directory, for running docker build.
COPY bin/connector /usr/local/bin

# Default config used to initalize datadir volume at first or
# cleaned deploy.
COPY connector.simnet.$ROLE.conf /root/default/connector.conf

# Entrypoint script used to init datadir if required and for
# starting dash daemon
COPY entrypoint.sh /root/

ENTRYPOINT bash /root/entrypoint.sh