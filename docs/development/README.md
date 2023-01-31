# Development Overview <!-- omit in toc -->

Please note that this repository is under very active development and breaking changes are likely to occur. If the documentation falls out of date please see our [guide](./../contributing/README.md) on how to contribute!

- [LFG - Development](#lfg---development)
  - [Install Dependencies](#install-dependencies)
  - [Prepare Local Environment](#prepare-local-environment)
  - [Pocket Network CLI](#pocket-network-cli)
  - [Swagger UI](#swagger-ui)
  - [View Available Commands](#view-available-commands)
  - [Running Unit Tests](#running-unit-tests)
  - [Running LocalNet](#running-localnet)
  - [Profiling](#profiling)
- [Code Organization](#code-organization)
  - [Linters](#linters)
    - [Installation of golangci-lint](#installation-of-golangci-lint)
    - [Running linters locally](#running-linters-locally)
    - [VSCode Integration](#vscode-integration)
    - [Configuration](#configuration)
    - [Custom linters](#custom-linters)

## LFG - Development

### Install Dependencies

- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)
- Install [Golang](https://go.dev/doc/install)
- `protoc-gen-go`, `protoc-go-inject-tag` and `mockgen` by running `make install_cli_deps`

_Note to the reader: Please update this list if you found anything missing._

Last tested by with:

```bash
$ docker --version
Docker version 20.10.14, build a224086

$ protoc --version
libprotoc 3.19.4

$ which protoc-go-inject-tag && echo "protoc-go-inject-tag Installed"
/your$HOME/go/bin/protoc-go-inject-tag
protoc-go-inject-tag Installed

$ go version
go version go1.18.1 darwin/arm64

$ mockgen --version
v1.6.0

$ system_profiler SPSoftwareDataType
Software:

    System Software Overview:

      System Version: macOS 12.3.1 (21E258)
      Kernel Version: Darwin 21.4.0
```

### Prepare Local Environment

Generate local files

```bash
$ git clone git@github.com:pokt-network/pocket.git && cd pocket
$ make develop_start
```

Optionally activate changelog pre-commit hook

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

_Please note that the Github workflow will still prevent this from merging
unless the CHANGELOG is updated._

### Pocket Network CLI

The Pocket node provides a CLI for interacting with Pocket RPC Server. The CLI can be used for both read & write operations by users and to aid in automation.

In order to build the CLI:

1. Generate local files

```bash
make develop_start
```

2. Build the CLI binary

```bash
make build
```

The cli binary will be available at `bin/p1` and can be used instead of `go run app/client/*.go`

The commands available are listed [here](../../rpc/doc/README.md) or accessible via `bin/p1 --help`

2.1 [OPTIONAL] Add the binary to your `.rc`

You can add the following function so you can run the `p1` from anywhere on your host:

```bash
function p1 {
    pocket_workdir="/Users/olshansky/workspace/pocket/pocket/"
    if [ "$1" = "debug" ]; then
        ${pocket_workdir}/bin/p1 debug --localhost=true --workdir="${pocket_workdir}"
    else
        ${pocket_workdir}/bin/p1 "$@"
    fi
}
```

You can via a demo of it [here](https://user-images.githubusercontent.com/1892194/215901991-076734e5-bc94-4755-9f2a-3d1f3c1e4aef.mov).

### Swagger UI

Swagger UI is available to help during the development process.

In order to spin a local instance of it with the API definition for the Pocket Network Node RPC interface automatically pre-loaded you can run:

```bash
make swagger-ui
```

### View Available Commands

```bash
$ make
```

### Running Unit Tests

```bash
$ make test_all
```

Note that there are a few tests in the library that are prone to race conditions and we are working on improving them. This can be checked with `make test_race`.

### Running LocalNet

![V1 Localnet Demo](./v1_localnet.gif)

1. Delete any previous docker state

```bash
$ make docker_wipe
```

2. In one shell, run the 4 nodes setup:

```bash
$ make compose_and_watch
```

4. In another shell, run the development client:

```bash
$ make client_start && make client_connect
```

4. Check the state of each node:

```bash
✔ PrintNodeState
```

5. Trigger the next view to ensure everything is working:

```bash
✔ TriggerNextView
```

6. Reset the ResetToGenesis if you want to:

```bash
✔ ResetToGenesis
```

7. Set the client to automatic and watch it go:

```bash
✔ TogglePacemakerMode
✔ TriggerNextView
```

8. [Optional] Common manual set of verification steps

```bash
✔ ResetToGenesis
✔ PrintNodeState # Check committed height is 0
✔ TriggerNextView
✔ PrintNodeState # Check committed height is 1
✔ TriggerNextView
✔ PrintNodeState # Check committed height is 2
✔ TogglePacemakerMode # Check that it’s automatic now
✔ TriggerNextView # Let it rip!
```

### Profiling

If you need to profile the node for CPU and/or memory usage, you can use the `pprof` tool.
A quick guide is available [here](./PROFILING.md).

## Code Organization

```bash
Pocket
├── app                               # Entrypoint to running the Pocket node and clients
│   ├── client                        # Entrypoint to running a local Pocket debug client
│   └── pocket                        # Entrypoint to running a local Pocket node
├── bin                               # Destination for compiled pocket binaries
├── build                             # Build related source files including Docker, scripts, etc
│   ├── config                        # Configuration files for to run nodes in development
│   ├── deployments                   # Docker-compose to run different cluster of services for development
│   ├── docs                          # Links to V1 Protocol implementation documentation (excluding the protocol specification)
├── consensus                         # Implementation of the Consensus module
├── docs                              # Links to V1 Protocol implementation documentation (excluding the protocol specification)
├── logger                            # Implementation of the Logger module
├── p2p                               # Implementation of the P2P module
├── persistence                       # Implementation of the Persistence module
├── rpc                               # Implementation of the RPC module
├── runtime                           # Implementation of the Runtime module
│   ├── configs                       # Configuration struct definitions
│   │   └── proto                     # Protobuf representing the specific configuration of the various modules
│   ├── defaults                      # Default values for the configuration structs
│   ├── genesis
│   │   └── proto                     # Protobuf representing the genesis state of the Pocket blockchain
│   └── test_artifacts                # Componentry used for generating test artifacts such as particular genesis states used in testing
├── shared                            # Shared types, modules and utils
│   ├── codec
│   │   └── proto
│   ├── converters
│   ├── core                          # Core types (Actor, Pools, etc.) used throughout the codebase
│   │   └── types
│   │       └── proto                 # Protobuf representing the core types used throughout the codebase
│   ├── crypto
│   ├── docs
│   │   └── flows
│   ├── messaging                     # Messaging structs and functions
│   │   └── proto
│   └── modules                       # Shared modules definitions (interfaces)
├── telemetry                         # Implementation of the Telemetry module
├── utility                           # Implementation of the Utility module
└── Makefile                          # The source of targets used to develop, build and test
```

### Linters

We utilize `golangci-lint` to run the linters. It is a wrapper around a number of linters and is configured to run many at once. The linters are configured to run on every commit and pull request via CI, and all code issues are populated as GitHub annotations to let developers and reviewers easily locate an issue.

#### Installation of golangci-lint

Please follow the instructions on the [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) website.

#### Running linters locally

You can run `golangci-lint` locally against all packages with:

```bash
make go_lint
```

If you need to specify any additional flags, you can run `golangci-lint` directly as it picks up the configuration from the `.golangci.yml` file.

#### VSCode Integration

`golangci-lint` has an integration with VSCode. Per [documentation](https://golangci-lint.run/usage/integrations/), recommended settings are:

```json
"go.lintTool":"golangci-lint",
"go.lintFlags": [
  "--fast"
]
```

#### Configuration

`golangci-lint` is a sophisticated tool including both default and custom linters. The configuration file, which can grow quite large, is located at [`.golangci.yml`](../../.golangci.yml).

The official documentation includes a list of different linters and their configuration options. Please refer to [this page](https://golangci-lint.run/usage/linters/) for more information.

#### Custom linters

We can write custom linters using [`go-ruleguard`](https://go-ruleguard.github.io/). The rules are located in the [`build/linters`](../../build/linters) directory. The rules are written in the [Ruleguard DSL](https://github.com/quasilyte/go-ruleguard/blob/master/_docs/dsl.md), if you've never worked with ruleguard in the past, it makes sense to go through [introduction article](https://quasilyte.dev/blog/post/ruleguard/) and [Ruleguard by example tour](https://go-ruleguard.github.io/by-example/).

Ruleguard is run via `gocritic` linter which is a part of `golangci-lint`, so if you wish to change configuration or debug a particular rule, you can modify the `.golangci.yml` file.
