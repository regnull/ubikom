#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null 2>&1 && pwd )"
TEMP_DIR=$(mktemp -d)
GOOS="linux"
GOARCH="amd64"
WIN_EXE=".exe"
BIN_NAME=daily-report
MAIN_DIR="$SCRIPT_DIR/../cmd/$BIN_NAME"
pushd $MAIN_DIR > /dev/null
for OS in $GOOS
do
    for ARCH in $GOARCH
    do
        echo Building $BIN_NAME for $OS $ARCH...
        SUFFIX=""
        if [ $OS == windows ]
        then
            SUFFIX=$WIN_EXE
        fi
        GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 GO_EXTLINK_ENABLED=0 go build -v -o $TEMP_DIR/$BIN_NAME-$OS-$ARCH $MAIN_DIR/main.go
        mkdir -p $SCRIPT_DIR/../build/$OS-$ARCH > /dev/null 2>&1
        cp $TEMP_DIR/$BIN_NAME-$OS-$ARCH $SCRIPT_DIR/../build/$OS-$ARCH/$BIN_NAME$SUFFIX
        cp $SCRIPT_DIR/../config/*.conf $SCRIPT_DIR/../build/$OS-$ARCH/
    done
done
popd > /dev/null
rm -rf $TEMP_DIR

scp $SCRIPT_DIR/../build/linux-amd64/daily-report lg@skytree:~/ubikom
