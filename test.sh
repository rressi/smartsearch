#!/bin/sh

export GOPATH=$(pwd)
go fmt  github.com/rressi/smartsearch     || exit 1
go test  github.com/rressi/smartsearch -v || exit 1
go fmt   src/makeindex.go                 || exit 1
go build src/makeindex.go                 || exit 1
go fmt   src/searchservice.go             || exit 1
go build src/searchservice.go             || exit 1

num_failed=0
for test_script in test/test_*.py
do
    echo
    echo "---------------------------------------------------------------------"
    echo "Running $test_script..."
    python3 ${test_script} || num_failed=$(expr ${num_failed} + 1)
done

echo "---------------------------------------------------------------------"
if [ "${num_failed}" -eq "0" ]; then
   echo "Success!";
else
   echo "${num_failed} failed tests";
fi