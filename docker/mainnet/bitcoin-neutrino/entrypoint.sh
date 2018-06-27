#!/usr/bin/env bash

# This path is expected to be volume to make lnd data persistent.
DATA_DIR=/root/.btcd

# This path is expected to have default data used to init environment
# at first deploy such as config.
DEFAULTS_DIR=/root/default

CONFIG=$DATA_DIR/btcd.conf

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

# We always restoring default config shipped with docker.
echo "Restoring default config"
cp $DEFAULTS_DIR/btcd.conf $CONFIG

# If external IP defined we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    echo "Setting external IP"
    EXTERNAL_IP_OPT="--externalip=$EXTERNAL_IP"
fi

RPC_USER_OPT="--bitcoind.rpcuser="

# We are using `exec` to enable gracefull shutdown of running daemon.
# Check http://veithen.github.io/2014/11/16/sigterm-propagation.html.
exec btcd $EXTERNAL_IP_OPT \
--rpcuser=$BITCOIN_NEUTRINO_RPC_USER
--rpcpass=$BITCOIN_NEUTRINO_RPC_PASSWORD