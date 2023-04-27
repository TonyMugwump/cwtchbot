#!/bin/sh

echo "Staticcheck..."
staticcheck ./...

echo "Formatting..."
gofmt -l -s -w .

# ineffassign (https://github.com/gordonklaus/ineffassign)
echo "Checking for ineffectual assignment of errors (unchecked errors...)"
ineffassign ./..

# misspell (https://github.com/client9/misspell/cmd/misspell)
echo "Checking for misspelled words..."
misspell . | grep -v "testing/" | grep -v "vendor/" | grep -v "go.sum" | grep -v ".idea"
