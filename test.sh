#!/bin/sh

export GOPATH=$(pwd)
go test github.com/rressi/smartsearch || echo "Failed" || false
go build src/makeindex.go || echo "Failed" || false
go build src/searchservice.go || echo "Failed" || false
python3 test/functional_test.py || echo "Failed" || false
