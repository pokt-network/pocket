# ARG GOLANG_IMAGE_VERSION=golang:1.18-rc-alpine3.14
ARG GOLANG_IMAGE_VERSION=golang:1.16-alpine3.14

FROM ${GOLANG_IMAGE_VERSION} AS builder

ENV POCKET_ROOT=/go/src/github.com/pocket-network/

WORKDIR $POCKET_ROOT

COPY . .

# Hot reloading.
RUN go install github.com/cespare/reflex@latest

RUN apk add build-base
# Debugging.
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Needed to make `go install dlv` and `dlbb debug` work...
RUN apk update && apk add --no-cache gcc musl-dev
RUN go get github.com/go-delve/delve/cmd/dlv@latest
