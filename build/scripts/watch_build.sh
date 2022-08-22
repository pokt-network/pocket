#!/bin/bash

if command -v reflex >/dev/null
then
  reflex -r '\.go$' -s -- sh -c "go build -v app/pocket/main.go"
else
    echo "reflex not found. Install with `go install github.com/cespare/reflex@latest`"
fi
