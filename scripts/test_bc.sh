#!/bin/bash
set -e

if [ $# -lt 1 ]
then
    printf "private key must be specified as the first argument\n"
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null 2>&1 && pwd )"
TEMP_DIR=$(mktemp -d)

# Create the key file.
echo $1 | xxd -r -p - $TEMP_DIR/key

ls -ls $TEMP_DIR

# Build ubkom-cli.
cd $SCRIPT_DIR/../cmd/ubikom-cli
go build -o $TEMP_DIR/ubikom-cli

# Deploy the contract.
$TEMP_DIR/ubikom-cli bc deploy registry --key=$TEMP_DIR/key --node-url=http://127.0.0.1:7545

# Register a name.

# Lookup the name.

# Create a connector.

# Lookup the connector.

# Update price.

# Buy.

# Transfer ownership.

rm -rf $TEMP_DIR
