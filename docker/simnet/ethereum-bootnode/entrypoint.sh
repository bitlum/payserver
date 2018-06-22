#!/usr/bin/env bash

# Docker container public IP
HOST_ADDR=`awk 'END{print $1}' /etc/hosts`

# Bootnode bind address
BIND_ADDR="$HOST_ADDR:30301"

# This path is expected to be volume to be able to share bootnode keys
DATA_DIR=/bootnode

# This file will contain bootnode key
KEY_FILE=$DATA_DIR/bootnode.key

# This file will contain enode URL
ENODE_FILE=$DATA_DIR/enode.url

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

if [ ! -f $KEY_FILE ]; then
    echo "Generation bootnode key"
    bootnode --genkey=$KEY_FILE
fi

if [ ! -f $ENODE_FILE ]; then
    echo "Computing enode URL"
    ENODE_KEY=`bootnode --nodekey=$KEY_FILE -writeaddress`
    ENODE_URL="enode://$ENODE_KEY@$BIND_ADDR"
    echo "Writing enode URL to file"
    echo $ENODE_URL > $ENODE_FILE
fi

# If external IP defined when we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    EXTERNAL_IP_OPT="--nat extip:$EXTERNAL_IP"
fi

# Start bootnode with computed key file on defined bind address and
# with optional nat external IP.
echo "Starting bootnode"
# We are using `exec` to enable gracefull shutdown of running daemon.
# Check http://veithen.github.io/2014/11/16/sigterm-propagation.html.
exec bootnode --nodekey=$KEY_FILE --addr=$BIND_ADDR $EXTERNAL_IP_OPT