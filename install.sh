#!/usr/bin/env bash


CURDIR=`pwd`
OLDGOPATH="$GOPATH"
echo $CURDIR
export GOPATH="$CURDIR"

go install main

export GOPATH="$OLDGOPATH"

echo 'finished'