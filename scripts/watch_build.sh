#!/bin.bash

reflex -r '\.go$' -s -- sh -c "go build -v cmd/pocket/main.go"