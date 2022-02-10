#!/bin/sh

config=$1

reflex -r '\.go$' -s -- sh -c "go run cmd/consensus/main.go --config=$config"

