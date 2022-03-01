#!/bin/bash

if builtin type -P "reflex"
then
  reflex -r '\.go$' -s -- sh -c "go build -v cmd/pocket/main.go"
else
    echo "reflex not found. Install with `go install github.com/cespare/reflex@latest`"
fi
