#!/bin/sh

export GOPATH=$(pwd)
go build src/makeindex.go
go build src/searchservice.go
