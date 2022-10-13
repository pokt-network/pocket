# Development Overview

Please note that this repository is under very active development and breaking changes are likely to occur. If the documentation falls out of date please see our [guide](./../contributing/README.md) on how to contribute!

- [Development Overview](#development-overview)
  - [LFG - Development](#lfg---development)
    - [Install Dependencies](#install-dependencies)
    - [Prepare Local Environment](#prepare-local-environment)
    - [View Available Commands](#view-available-commands)
    - [Running Unit Tests](#running-unit-tests)
    - [Running LocalNet](#running-localnet)
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

## Code Organization

```bash
Pocket
├── app              # Entrypoint to running the Pocket node and clients
|   ├── client       # Entrypoint to running a local Pocket debug client
|   ├── pocket       # Entrypoint to running a local Pocket node
├── bin              # [currently-unused] Destination for compiled pocket binaries
├── build            # Build related source files including Docker, scripts, etc
|   ├── config       # Configuration files for to run nodes in development
|   ├── deployments  # Docker-compose to run different cluster of services for development
|   ├── Docker*      # Various Dockerfile(s)
├── consensus        # Implementation of the Consensus module
├── core             # [currently-unused]
├── docs             # Links to V1 Protocol implementation documentation (excluding the protocol specification)
├── p2p              # Implementation of the P2P module
├── persistence      # Implementation of the Persistence module
├── shared           # [to-be-refactored] Shared types, modules and utils
├── utility          # Implementation of the Utility module
├── Makefile         # [to-be-deleted] The source of targets used to develop, build and test
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
