#!/bin/sh
set -exu

private_key=$(cat /wallets.json | jq -r '."3151908".batcherPrivateKey')

op-batcher \
    --l2-eth-rpc=$HIVE_L2_HTTP_URL \
    --rollup-rpc=$HIVE_ROLLUP_HTTP_URL \
    --poll-interval=1s \
    --sub-safety-margin=6 \
    --num-confirmations=1 \
    --safe-abort-nonce-too-low-count=3 \
    --resubmission-timeout=30s \
    --rpc.addr=0.0.0.0 \
    --rpc.port=8548 \
    --rpc.enable-admin \
    --max-channel-duration=1 \
    --l1-eth-rpc=$HIVE_L1_HTTP_URL \
    --private-key=$private_key \
    --data-availability-type=blobs \
    --metrics.enabled \
    --metrics.addr=0.0.0.0 \
    --metrics.port=9001



