ARG GOLANG_IMAGE_VERSION=golang:1.18-rc-alpine3.14

FROM ${GOLANG_IMAGE_VERSION} AS builder

ENV POCKET_ROOT=/go/src/github.com/pocket-network/

# Badger CLI
RUN cd /bin && wget https://github.com/dgraph-io/badger/archive/refs/tags/v3.2103.2.tar.gz && \
    tar -xvf v3.2103.2.tar.gz && cd badger-3.2103.2/badger go install && \
    cd .. && rm -f v3.2103.2.tar.gz

# Source code
WORKDIR $POCKET_ROOT

COPY . .

# Hot reloading
RUN go install github.com/cespare/reflex@latest

RUN apk add build-base

# Debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest


# Needed to make `go install dlv` and `dlv debug` work...
RUN apk update && apk add --no-cache gcc musl-dev
RUN go get github.com/go-delve/delve/cmd/dlv@latest
