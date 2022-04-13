#!/bin/bash

if builtin type -P "reflex"
then
  reflex -r '\.go$' -s -- sh -c "go build -v app/pocket/main.go"
else
    echo "reflex not found. Install with `go install github.com/cespare/reflex@latest`"
fi
