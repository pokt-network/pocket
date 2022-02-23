ARG GOLANG_IMAGE_VERSION=golang:1.17-rc-alpine3.14

FROM ${GOLANG_IMAGE_VERSION} AS builder

ENV POCKET_ROOT=/go/src/github.com/pocket-network

WORKDIR $POCKET_ROOT

COPY . .

# Install bash
RUN apk add --no-cache bash

# Hot reloading
RUN go install github.com/cespare/reflex@latest

CMD ["/bin/bash"]