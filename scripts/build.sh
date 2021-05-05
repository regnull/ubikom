#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null 2>&1 && pwd )"
TEMP_DIR=$(mktemp -d)
for BIN_NAME in ubikom-server ubikom-dump ubikom-cli ubikom-proxy
do
    MAIN_DIR="$SCRIPT_DIR/../cmd/$BIN_NAME"
    pushd $MAIN_DIR > /dev/null
    echo GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO_EXTLINK_ENABLED=0 go build -v -o $TEMP_DIR/$BIN_NAME-linux $MAIN_DIR/main.go
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO_EXTLINK_ENABLED=0 go build -v -o $TEMP_DIR/$BIN_NAME-linux $MAIN_DIR/main.go
    go build -v -o $TEMP_DIR/$BIN_NAME-darwin $MAIN_DIR/main.go
    cp $TEMP_DIR/$BIN_NAME-linux $SCRIPT_DIR/../build/
    cp $TEMP_DIR/$BIN_NAME-darwin $SCRIPT_DIR/../build/
    popd > /dev/null
done
rm -rf $TEMP_DIR
