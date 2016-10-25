#!/bin/sh

export GOPATH=$(pwd)
go get golang.org/x/text || exit 1
