#! /usr/bin/env bash

# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
protoc --go_out=../../ *.proto
