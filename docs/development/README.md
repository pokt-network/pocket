# Development Overview

Please note that this repository is under very active development and breaking changes are likely to occur. If the documentation falls out of date please see our [guide](./../contributing/CONTRIBUTING.md) on how to contribute!

- [Development Overview](#development-overview)
  - [LFG - Development](#lfg---development)
    - [Install Dependencies](#install-dependencies)
    - [Prepare Local Environment](#prepare-local-environment)
    - [View Available Commands](#view-available-commands)
    - [Running Unit Tests](#running-unit-tests)
    - [Running LocalNet](#running-localnet)
  - [Code Organization](#code-organization)

## LFG - Development

### Install Dependencies

- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)
- Install [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go)
    See: https://grpc.io/docs/languages/go/quickstart/
- Install [Golang](https://go.dev/doc/install)
- Install [mockgen](https://github.com/golang/mock)

_Note to the reader: Please update this list if you found anything missing._

Last tested by with:

```bash
$ docker --version
Docker version 20.10.14, build a224086

$ protoc --version
libprotoc 3.19.4

$ protoc --version
libprotoc 3.19.4

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
$ git clone git@github.com:pokt-network/pocket.git  && cd pocket
$ make protogen_clean && make protogen_local
$ make mockgen
$ make go_clean_deps
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
