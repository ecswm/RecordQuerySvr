#!/usr/bin/env bash

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
echo $CURDIR
export GOPATH="$CURDIR"

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

go build main

./main

export GOPATH="$OLDGOPATH"

echo 'finished'