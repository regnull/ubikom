#!/bin/bash
set -e

if [ $# -lt 1 ]
then
    printf "private key must be specified as the first argument\n"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null 2>&1 && pwd )"
TEMP_DIR=$(mktemp -d)
NODE_URL=http://127.0.0.1:7545

# Create the key file.
echo $1 | xxd -r -p - $TEMP_DIR/key

# Build ubkom-cli.
cd $SCRIPT_DIR/../cmd/ubikom-cli
go build -o $TEMP_DIR/ubikom-cli

# Deploy the contract.
REG_RES=$($TEMP_DIR/ubikom-cli bc deploy registry --key=$TEMP_DIR/key --node-url=$NODE_URL)
echo $REG_RES
CONTRACT_ADDRESS=$(echo $REG_RES| jq -r ".Address")

# Create an encryption key.
$TEMP_DIR/ubikom-cli create key --out=$TEMP_DIR/enc_key --skip-passphrase

# Register a name.
$TEMP_DIR/ubikom-cli bc register name --name=foo --key=$TEMP_DIR/key --enc-key=$TEMP_DIR/enc_key --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL

# Lookup the name.
$TEMP_DIR/ubikom-cli bc lookup name foo --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL

# Create a connector.

# Lookup the connector.

# Update price.

# Buy.

# Transfer ownership.

rm -rf $TEMP_DIR
