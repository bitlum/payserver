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

# We always restoring default config shipped with docker.
echo "Restoring default config"
cp $DEFAULTS_DIR/dash.conf $CONFIG

# If external IP defined when we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    echo "Setting external IP"
    EXTERNAL_IP_OPT="-externalip=$EXTERNAL_IP"
fi

dashd $EXTERNAL_IP_OPT \
--rpcauth=$DASH_RPC_AUTH