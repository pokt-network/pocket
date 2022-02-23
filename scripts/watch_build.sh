#!/bin.bash

reflex -r '\.go$' -s -- sh -c "go build -v cmd/v1/main.go"