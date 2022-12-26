#!/bin/bash

if command -v reflex >/dev/null
then
  reflex -r '\.go$' -s -- sh -c "go build -v cmd/pocket/main.go"
else
    echo "reflex not found. Install with `go install github.com/cespare/reflex@latest`"
fi
