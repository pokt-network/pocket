# Development Overview

Please note that this repository is under very active development and breaking changes are likely to occur. If the documentation falls out of date please see our [guide](./../contributing/CONTRIBUTING.md) on how to contribute!

- [Development Overview](#development-overview)
  - [LFG - Development](#lfg---development)
    - [Install Dependencies](#install-dependencies)
    - [Prepare Local Environment](#prepare-local-environment)
    - [View Available Commands](#view-available-commands)
    - [Running LocalNet](#running-localnet)
    - [Running Tests](#running-tests)
  - [Code Organization](#code-organization)

## LFG - Development

### Install Dependencies

- Install [Docker](https://docs.docker.com/get-docker/)
- Install [Docker Compose](https://docs.docker.com/compose/install/)
- Install [Golang](https://go.dev/doc/install)
- Install [protoc-gen-go](https://pkg.go.dev/google.golang.org/protobuf/cmd/protoc-gen-go)
- Install [mockgen](https://github.com/golang/mock#installation=)

### Prepare Local Environment

Generate local files

```bash
$ make protogen_clean
$ make protogen_local
$ make mockgen
$ go mod vendor && go mod tidy
```

### View Available Commands

```bash
$ make
```

### Running LocalNet

![V1 Localnet Demo](./v1_localnet.gif)

Delete any previous docker state

```bash
$ make docker_wipe
```

In one shell, run:

```bash
$ make compose_and_watch
```

In another shell, run:

```bash
$ make client_start
$ make client_connect

> ResetToGenesis
> PrintNodeState # Check committed height is 0
> TriggerNextView
> PrintNodeState # Check committed height is 1
> TriggerNextView
> PrintNodeState # Check committed height is 2
> TogglePacemakerMode # Check that it’s automatic now
> TriggerNextView # Let it rip!
```

### Running Tests

```bash
$ make test_all
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
|   ├── Docker*      # Various Dockerfiles
├── consensus        # Implementation of the Consensus module
├── core             # [currently-unused]
├── docs             # Links to V1 Protocol documentation (except the protocol specification)
├── p2p              # Implementation of the P2P module
├── persistence      # Implementation of the Persistence module
├── prototype        # [to-be-deleted] A snapshot of the very first v1 prototype
├── shared           # [to-be-refactored] Shared types, modules and utils
├── utility          # Implementation of the Persistence module
├── Makefile         # [to-be-deleted] The source of targets used to develop, build and test
├── mage.go          # [currently-unused] The future source of targets used to develop, build and test
```
