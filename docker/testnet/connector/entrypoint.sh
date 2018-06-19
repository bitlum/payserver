#!/usr/bin/env bash

# This path is expected to be volume to make connector data persistent.
DATA_DIR=/root/.connector

# This path is expected to have default data used to init environment
# at first deploy such as config.
DEFAULTS_DIR=/root/default

CONFIG=$DATA_DIR/connector.conf

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

# We always restoring default config shipped with docker.
echo "Restoring default config"
cp $DEFAULTS_DIR/connector.conf $CONFIG

# Set exchange notifications disabled
if [ $EXCHANGE_DISABLED -eq 1 ]; then
    echo "Disabling eninge"
    EXCHANGE_DISABLED_OPT="--enginedisabled"
fi

connector --config /root/.connector/connector.conf $EXCHANGE_DISABLED_OPT \
--bitcoin.user=$BITCOIN_RPC_USER \
--bitcoin.password=$BITCOIN_RPC_PASSWORD \
--bitcoincash.user=$BITCOIN_CASH_RPC_USER \
--bitcoincash.password=$BITCOIN_CASH_RPC_PASSWORD \
--dash.user=$DASH_RPC_USER \
--dash.password=$DASH_RPC_PASSWORD \
--litecoin.user=$LITECOIN_RPC_USER \
--litecoin.password=$LITECOIN_RPC_PASSWORD