#!/bin/sh

export GOPATH=$(pwd)
go get "golang.org/x/text" || exit 1
go get "github.com/NYTimes/gziphandler" || exit 1
