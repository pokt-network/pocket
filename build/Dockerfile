ARG GOLANG_IMAGE_VERSION=golang:1.18-rc-alpine3.14

FROM ${GOLANG_IMAGE_VERSION} AS builder

ENV POCKET_ROOT=/go/src/github.com/pocket-network/

WORKDIR $POCKET_ROOT

COPY . .

# Hot reloading.
RUN go install github.com/cespare/reflex@latest

# Debugging.
RUN go get github.com/go-delve/delve/cmd/dlv@latest