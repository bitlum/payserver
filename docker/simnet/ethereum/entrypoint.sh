#!/usr/bin/env bash

# This directory is expected to be volume to make blockchain and account
# data persistent.
DATA_DIR=/root/.ethereum

BOOTNODE_DIR=/bootnode

# This path is expected to have default data used to init environment
# at first deploy such as config and genesis.
DEFAULTS_DIR=/root/default

CONFIG=$DATA_DIR/ethereum.conf
GENESIS=$DATA_DIR/genesis.conf
ENODE=$BOOTNODE_DIR/enode.url

# If data directory doesn't exists this means that volume is not mounted
# or mounted incorrectly, so we must fail.
if [ ! -d $DATA_DIR ]; then
    echo "Data directory '$DATA_DIR' doesn't exists. Check your volume configuration."
    exit 1
fi

# If bootnode directory doesn't exists this means that volume is not
# mounted or mounted incorrectly, so we must fail.
if [ ! -d $BOOTNODE_DIR ]; then
    echo "Bootnode directory '$BOOTNODE_DIR' doesn't exists. Check your volume configuration."
    exit 2
fi

if [ ! -f $ENODE ]; then
    echo "Bootnode directory '$BOOTNODE_DIR' doesn't contain 'enode.url' file.  Does 'ethereum-bootnode' container successfully started?"
    exit 3
fi

# At first deploy config in datadir should not exists so we
# copying from default config shipped with docker.
if [ ! -f $CONFIG ]; then
    echo "Copying default config"
    cp $DEFAULTS_DIR/ethereum.conf $CONFIG
fi

# At first deploy genesis in datadir should not exists so we
# copying from default genesis shipped with docker.
if [ ! -f $GENESIS ]; then
    echo "Copying default genesis"
    cp $DEFAULTS_DIR/genesis.json $GENESIS
fi

# At first deploy datadir keystore path should not exists which means
# that account is not created, so we are creating it.
KEYSTORE=$DATA_DIR/keystore
if [ ! -d $KEYSTORE ] || [ ! "$(ls -A $KEYSTORE)" ]; then
    echo "Creating new account"
    geth --datadir $DATA_DIR --config $CONFIG account new --password /dev/null
fi

# At first deploy datadir geth path should not exists which means
# that blockchain is not inited, so we are initing it.
GETH=$DATA_DIR/geth
if [ ! -d $GETH ] || [ ! "$(ls -A $GETH)" ]; then
    echo "Initing genesis block"
    geth --datadir $DATA_DIR --config $CONFIG init $GENESIS
fi

# Set mine option to enable blocks mining if required.
if [ $MINE -eq 1 ]; then
    MINE_OPT="--mine"
fi

# If external IP defined when we need to set corresponding run option
if [ ! -z "$EXTERNAL_IP" ]; then
    EXTERNAL_IP_OPT="--nat extip:$EXTERNAL_IP"
fi

geth \
--datadir $DATA_DIR \
--config $CONFIG \
--bootnodes `cat $ENODE` \
$MINE_OPT \
$EXTERNAL_IP_OPT