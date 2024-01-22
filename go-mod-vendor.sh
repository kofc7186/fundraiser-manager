#!/bin/sh

# runs 'go mod vendor' from within the directory provided as the first argument
# outputs a valid json array so it can be invoked via terraform "data" "external" module

pushd $1 >/dev/null
rm -rf vendor
go mod vendor && echo '{"time": "'$(date +%s)'", "dir": "'$(pwd)'"}'
popd >/dev/null
