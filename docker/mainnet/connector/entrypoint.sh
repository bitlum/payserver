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


# We are using `exec` to enable gracefull shutdown of running daemon.
# Check http://veithen.github.io/2014/11/16/sigterm-propagation.html.
exec connector --config /root/.connector/connector.conf \
--bitcoin.user=$BITCOIN_RPC_USER \
--bitcoin.password=$BITCOIN_RPC_PASSWORD \
--bitcoin.forcelasthash=$CONNECTOR_BITCOIN_FORCE_HASH \
--bitcoincash.user=$BITCOIN_CASH_RPC_USER \
--bitcoincash.password=$BITCOIN_CASH_RPC_PASSWORD \
--bitcoincash.forcelasthash=$CONNECTOR_BITCOIN_CASH_FORCE_HASH \
--dash.user=$DASH_RPC_USER \
--dash.password=$DASH_RPC_PASSWORD \
--dash.forcelasthash=$CONNECTOR_DASH_FORCE_HASH \
--litecoin.user=$LITECOIN_RPC_USER \
--litecoin.password=$LITECOIN_RPC_PASSWORD \
--litecoin.forcelasthash=$CONNECTOR_LITECOIN_FORCE_HASH \
--etheruem.forcelasthash=$CONNECTOR_ETHEREUM_FORCE_HASH