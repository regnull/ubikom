#!/bin/bash
#set -e

if [ $# -lt 2 ]
then
    printf "private key must be specified as the first argument\n"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null 2>&1 && pwd )"
TEMP_DIR=$(mktemp -d)
NODE_URL=http://127.0.0.1:7545

# Create the key files.
echo $1 | xxd -r -p - $TEMP_DIR/key1
echo $2 | xxd -r -p - $TEMP_DIR/key2

# Build ubkom-cli.
cd $SCRIPT_DIR/../cmd/ubikom-cli
go build -o $TEMP_DIR/ubikom-cli || exit 1

# Get addresses.
ADDRESS1=$($TEMP_DIR/ubikom-cli get ethereum-address --key=$TEMP_DIR/key1) || exit 1
ADDRESS2=$($TEMP_DIR/ubikom-cli get ethereum-address --key=$TEMP_DIR/key2) || exit 1

# Deploy the contract.
REG_RES=$($TEMP_DIR/ubikom-cli bc deploy registry --key=$TEMP_DIR/key1 --node-url=$NODE_URL)
if [ "$?" -ne 0 ];
then
    exit 1
fi 
echo $REG_RES
CONTRACT_ADDRESS=$(echo $REG_RES| jq -r ".Address")

# Create an encryption key.
#$TEMP_DIR/ubikom-cli create key --out=$TEMP_DIR/pub_key --skip-passphrase || exit 1

# Try to register an invalid name.
$TEMP_DIR/ubikom-cli bc register name "$$$" --key=$TEMP_DIR/key1 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL && exit 1
$TEMP_DIR/ubikom-cli bc register name "_" --key=$TEMP_DIR/key1 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL && exit 1
$TEMP_DIR/ubikom-cli bc register name "-foo" --key=$TEMP_DIR/key1 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL && exit 1

# Register a name.
$TEMP_DIR/ubikom-cli bc register name foo --key=$TEMP_DIR/key1 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Lookup the name.
$TEMP_DIR/ubikom-cli bc lookup name foo --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Update the public key.
$TEMP_DIR/ubikom-cli create key --out=$TEMP_DIR/pub_key1 --skip-passphrase || exit 1
$TEMP_DIR/ubikom-cli bc update public-key foo --key=$TEMP_DIR/key1 --pub-key=$TEMP_DIR/pub_key1 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Update ownership.
$TEMP_DIR/ubikom-cli bc update owner foo --key=$TEMP_DIR/key1 --new-owner-address=$ADDRESS2 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Update price.
$TEMP_DIR/ubikom-cli bc update price foo --key=$TEMP_DIR/key2 --price=1230000000000000000 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1
$TEMP_DIR/ubikom-cli bc lookup name foo --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Buy.
$TEMP_DIR/ubikom-cli bc buy name foo --key=$TEMP_DIR/key1 --value=1230000000000000000 --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1

# Update config.
$TEMP_DIR/ubikom-cli bc update config foo --key=$TEMP_DIR/key1 --config-name=myconfigname --config-value=myconfigvalue --contract-address=$CONTRACT_ADDRESS --node-url=$NODE_URL || exit 1
$TEMP_DIR/ubikom-cli bc lookup config foo --contract-address=$CONTRACT_ADDRESS --config-name=myconfigname --node-url=$NODE_URL || exit 1

# Lookup connector.

rm -rf $TEMP_DIR
