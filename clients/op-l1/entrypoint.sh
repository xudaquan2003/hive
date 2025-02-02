#!/bin/sh
set -exu

# Regenerate the L1 genesis file
date +%s | xargs printf "0x%x" > /hive/genesis_timestamp
GENESIS_TIMESTAMP=$(cat /hive/genesis_timestamp); jq ". | .timestamp = \"$GENESIS_TIMESTAMP\" " < ./genesis.json > /hive/genesis-l1.json

VERBOSITY=${GETH_VERBOSITY:-3}
GETH_DATA_DIR=/db
GETH_CHAINDATA_DIR="$GETH_DATA_DIR/geth/chaindata"
GETH_KEYSTORE_DIR="$GETH_DATA_DIR/keystore"
CHAIN_ID=$(cat /hive/genesis-l1.json | jq -r .config.chainId)
BLOCK_SIGNER_PRIVATE_KEY="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
BLOCK_SIGNER_ADDRESS="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

if [ ! -d "$GETH_KEYSTORE_DIR" ]; then
	echo "$GETH_KEYSTORE_DIR missing, running account import"
	echo -n "pwd" > "$GETH_DATA_DIR"/password
	echo -n "$BLOCK_SIGNER_PRIVATE_KEY" | sed 's/0x//' > "$GETH_DATA_DIR"/block-signer-key
	geth account import \
		--datadir="$GETH_DATA_DIR" \
		--password="$GETH_DATA_DIR"/password \
		"$GETH_DATA_DIR"/block-signer-key
else
	echo "$GETH_KEYSTORE_DIR exists."
fi

if [ ! -d "$GETH_CHAINDATA_DIR" ]; then
	echo "$GETH_CHAINDATA_DIR missing, running init"
	echo "Initializing genesis."
	geth --verbosity="$VERBOSITY" init \
		--datadir="$GETH_DATA_DIR" \
		--state.scheme=hash \
		"/hive/genesis-l1.json"
else
	echo "$GETH_CHAINDATA_DIR exists."
fi

# Warning: Archive mode is required, otherwise old trie nodes will be
# pruned within minutes of starting the devnet.

exec geth \
    --networkid=$CHAIN_ID \
	--datadir="$GETH_DATA_DIR" \
	--verbosity="$VERBOSITY" \
	--http \
	--http.corsdomain="*" \
	--http.vhosts="*" \
	--http.addr=0.0.0.0 \
	--http.port=$HIVE_L1_HTTP_PORT \
	--http.api=admin,engine,net,eth,web3,debug \
	--ws \
	--ws.addr=0.0.0.0 \
	--ws.port=$HIVE_L1_WS_PORT \
	--ws.origins="*" \
	--ws.api=admin,engine,net,eth,web3,debug \
    --syncmode=full \
    --nodiscover \
    --maxpeers=1 \
    --unlock=$BLOCK_SIGNER_ADDRESS \
	--dev \
	--dev.period=3 \
	--miner.etherbase=$BLOCK_SIGNER_ADDRESS \
	--password="$GETH_DATA_DIR"/password \
    --allow-insecure-unlock \
    --rpc.allow-unprotected-txs \
    --authrpc.addr=0.0.0.0 \
    --authrpc.port=8551 \
    --authrpc.vhosts="*" \
    --gcmode=archive \
	"$@"

