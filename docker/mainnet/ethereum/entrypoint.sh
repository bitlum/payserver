#!/usr/bin/env bash

# This directory is expected to be volume to make blockchain and account
# data persistent.
DATA_DIR=/root/.ethereum

# This path is expected to have default data used to init environment
# at first deploy such as config and genesis.
DEFAULTS_DIR=/root/default

CONFIG=$DATA_DIR/ethereum.conf

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

# We always restoring default config shipped with docker.
echo "Restoring default config"
cp $DEFAULTS_DIR/ethereum.conf $CONFIG

# At first deploy datadir keystore path should not exists which means
# that account is not created, so we are creating it.
KEYSTORE=$DATA_DIR/keystore
if [ ! -d $KEYSTORE ] || [ ! "$(ls -A $KEYSTORE)" ]; then
    echo "Creating new account"
    geth --datadir $DATA_DIR --config $CONFIG account new --password /dev/null
fi

# If external IP defined when we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    echo "Setting external IP"
    EXTERNAL_IP_OPT="--nat extip:$EXTERNAL_IP"
fi

# We are using `exec` to enable gracefull shutdown of running daemon.
# Check http://veithen.github.io/2014/11/16/sigterm-propagation.html.
exec geth \
--debug \
--verbosity=4 \
--datadir $DATA_DIR \
--config $CONFIG \
--cache 4096 \
$EXTERNAL_IP_OPT