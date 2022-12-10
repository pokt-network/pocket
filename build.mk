OS = $(shell uname | tr A-Z a-z)
GOARCH = $(shell go env GOARCH)
GOLANG_VERSION ?= 1.18

# Build variables
BUILD_DIR ?= bin
BINARY_NAME_client ?= p1
BINARY_NAME_pocket ?= pocket
POST_BUILD_TARGETS = rename-binaries
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE_FMT = +%FT%T%z
ifdef SOURCE_DATE_EPOCH
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u "$(DATE_FMT)")
else
    BUILD_DATE ?= $(shell date "$(DATE_FMT)")
endif
LDFLAGS += -X main.version=${VERSION} -X main.commitHash=${COMMIT_HASH} -X main.buildDate=${BUILD_DATE}
export CGO_ENABLED ?= 0
ifeq (${VERBOSE}, 1)
ifeq ($(filter -v,${GOARGS}),)
	GOARGS += -v
endif
endif

## Clean the build directory
.PHONY: clean
clean: ${CLEAN_TARGETS}
	rm -rf ${BUILD_DIR}/

## Check if the installed Go version is the required one
.PHONY: goversion
goversion:
ifneq (${IGNORE_GOLANG_VERSION}, 1)
	@printf "${GOLANG_VERSION}\n$$(go version | awk '{sub(/^go/, "", $$3);print $$3}')" | sort -t '.' -k 1,1 -k 2,2 -k 3,3 -g | head -1 | grep -q -E "^${GOLANG_VERSION}$$" || (printf "Required Go version is ${GOLANG_VERSION}\nInstalled: `go version`" && exit 1)
endif

## Build a single application in the app directory
.PHONY: build-%
build-%: pre-build
build-%: goversion
ifeq (${VERBOSE}, 1)
	go env
endif

	@mkdir -p ${BUILD_DIR}
	go build ${GOARGS} -trimpath -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/$* ./app/$*

	@${MAKE} post-build

## Build all applications in the app directory
.PHONY: build
build: pre-build
build: goversion
ifeq (${VERBOSE}, 1)
	go env
endif

	@mkdir -p ${BUILD_DIR}
	go build ${GOARGS} -trimpath -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/ ./app/...

	@${MAKE} post-build

## Pre build hook
.PHONY: pre-build
pre-build: ${PRE_BUILD_TARGETS}
	@:

## Post build hook
.PHONY: post-build
post-build: ${POST_BUILD_TARGETS}
	@:

## Run a single application in after building it
.PHONY: run-%
run-%: build-%
	${BUILD_DIR}/$*

## Run all applications in the app directory after building them
.PHONY: run
run: $(patsubst app/%,run-%,$(wildcard app/*)) ## Build and execute all applications

## Rename a binary to the name specified in BINARY_NAME_$* if it exists.
.PHONY: rename-%
rename-%:

## Redirecting stderr to /dev/null to avoid returning an error if the file already exists
	@mv -f ${BUILD_DIR}/$* ${BUILD_DIR}/${BINARY_NAME_$*} 2>/dev/null; true

## Rename all binaries in the bin directory to the name specified in BINARY_NAME_$* if it exists
.PHONY: rename-binaries
rename-binaries: $(patsubst bin/%,rename-%,$(wildcard bin/*)) ## Rename all binaries in the bin directory
