#!/usr/bin/env bash

# This path is expected to be volume to make blockchain and accounts
# data persistent.
DATA_DIR=/root/.dashcore

# This path is expected to have default data used to init environment
# at first deploy such as config and genesis.
DEFAULTS_DIR=/root/default

CONFIG=$DATA_DIR/dash.conf

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

# At first deploy config in datadir should not exists so we
# copying from default config shipped with docker.
if [ ! -f $CONFIG ]; then
    echo "Copying default config"
    cp $DEFAULTS_DIR/dash.conf $CONFIG
fi

# If external IP defined when we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    EXTERNAL_IP_OPT="-externalip=$EXTERNAL_IP"
fi

dashd $EXTERNAL_IP_OPT