#!/bin/sh

export GOPATH=$(pwd)
go fmt  github.com/rressi/smartsearch     || exit 1
go test  github.com/rressi/smartsearch -v || exit 1
go fmt   src/makeindex.go                 || exit 1
go build src/makeindex.go                 || exit 1
go fmt   src/searchservice.go             || exit 1
go build src/searchservice.go             || exit 1
python   test/functional_test.py          || exit 1
