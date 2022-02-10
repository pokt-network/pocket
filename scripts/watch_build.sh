#!/bin.bash

reflex -r '\.go$' -s -- sh -c "go build -v cmd/consensus/main.go"