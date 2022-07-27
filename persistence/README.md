# Persistence Module <!-- omit in toc -->

This document is meant to be a supplement to the living protocol specification at [1.0 Pocket's Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) primarily focused on the implementation, details related to the design of the codebase, and information related to testing and development.

- [Database Migrations](#database-migrations)
- [Node Configuration](#node-configuration)
- [Debugging & Development](#debugging--development)
  - [Code Structure](#code-structure)
  - [Makefile Helpers](#makefile-helpers)
    - [Admin View - db_admin](#admin-view---db_admin)
    - [Benchmarking - db_bench](#benchmarking---db_bench)
- [Testing](#testing)
  - [Unit Tests](#unit-tests)
  - [Dependencies](#dependencies)
  - [Setup](#setup)
    - [Setup Issue - Docker Daemon is not Running](#setup-issue---docker-daemon-is-not-running)
    - [Setup Issue - Docker Daemon is not Running](#setup-issue---docker-daemon-is-not-running-1)
- [Implementation FAQ](#implementation-faq)
- [Implementation TODOs](#implementation-todos)

## Database Migrations

**TODO**: https://github.com/pokt-network/pocket/issues/77

## Node Configuration

The persistence specific configuration within a node's `config.json` looks like this:

```
  "persistence": {
    "postgres_url": "postgres://postgres:postgres@pocket-db:5432/postgres",
    "schema": "node1"
  }
```

See [config.go](./shared/config/config.go) for the specification and [config1.json](./build/config/config1.json) for an example.

**IMPORTANT**: The `schema` parameter **MUST** be unique for each node associated with the same Postgres instance, and there is currently no check or validation for it.

## Debugging & Development

### Code Structure

```bash
persistence         # Directly contains the persistence module interface for each actor
├── CHANGELOG.md    # Persistence module changelog
├── README.md       # Persistence module README
├── account.go
├── application.go
├── block.go
├── db.go           # Helpers to connect and initialize the Postgres database
├── fisherman.go
├── gov.go
├── module.go       # Implementation of the persistence module interface
├── service_node.go
├── shared_sql.go   # Database implementation helpers shared across all protocol actors
└── validator.go
├── docs
├── schema         # Directly contains the SQL schema and SQL query builders used by the files above
│   ├── account.go
│   ├── application.go
│   ├── base_actor.go      # Implementation of the `protocol_actor.go` interface shared across all actors
│   ├── block.go
│   ├── fisherman.go
│   ├── gov.go
│   ├── migrations
│   ├── protocol_actor.go   # Interface definition for the schema shared across all actors
│   ├── service_node.go
│   ├── shared_sql.go       # Query building implementation helpers shared across all protocol actors
│   └── validator.go
└── test  # Unit & fuzzing tests
```

### Makefile Helpers

If you run `make` from the root of the `pocket` repo, there will be several targets prefixed with `db_` that can help with design & development of the infrastructure associated with this module

We only explain a subset of these in the list below.

#### Admin View - db_admin

When you run `db_admin`, the following will be echoed to your screen

```
echo "Open http://0.0.0.0:5050 and login with 'pgadmin4@pgadmin.org' and 'pgadmin4'.\n The password is 'postgres'"
```

After logging in, you can view the tables within each schema by following the following screenshot.

![](./docs/pgadmin.png "pgadmin view")

#### Benchmarking - db_bench

// TODO(olshansky)

## Testing

_Note: There are many TODO's in the testing environment including thread safety. It's possible that running the tests in parallel may cause tests to break so it is recommended to use `-p 1` flag_

### Unit Tests

Unit tests can be executed with:

```bash
$ make test_persistence
```

### Dependencies

We use a library called [dockertest](https://github.com/ory/dockertest), along with `TestMain` (learn more [here](https://medium.com/goingogo/why-use-testmain-for-testing-in-go-dafb52b406bc]), to use the local Docker Daemon for unit testing.

### Setup

Make sure you have a Docker daemon running. See the [Development Guide](docs/development/README.md) for more references and links.

#### Setup Issue - Docker Daemon is not Running

If you see an issue similar to the one below, make sure your Docker Daemon is running.

```
not start resource: : dial unix /var/run/docker.sock: connect: no such file or directory
```

For example, on macOS, you can run `open /Applications/Docker.app` to start it up.

#### Setup Issue - Docker Daemon is not Running

If you see an issue similar to the one below, make sure you don't already have a Postgres docker container running or one running on your host machine.

```
Bind for 0.0.0.0:5432 failed: port is already allocated
```

For example, on macOS, you can check for this with `lsof -i:5432` and kill the appropriate process if one exists.

## Implementation FAQ

**Q**: Why do `Get` methods (e.g. `GetAccountAmount`) not return 0 by default?
**A**: This was done intentionally to differentiate between accounts with a history and without a history. Since accounts are just a proxy into a public key, they all "exist by default" in some senses.

**Q**: Why are amounts strings?
**A**: A lesson from Tendermint in order to enforce the use of BigInts throughout and avoid floating point issues when storing data on disk.

**Q**: Why not use an ORM?
**A**: We are trying to keep the module small and lean initially but are ope

## Implementation TODOs

These are major TODOs spanning the entire repo so they are documented in one place instead.

Short-term (i.e. simpler starter) tasks:

- [ ] DOCUMENT: Need to do a better job at documenting the process of paused apps being turned into unstaking apps.
- [ ] CLEANUP: Remove unused parameters from `the PostgresContext` interface (i.e. see where \_ is used in the implementation such as in `InsertFisherman`)
- [ ] IMPROVE: Consider converting all address params from bytes to string to avoid unnecessary encoding
- [ ] CLEANUP(https://github.com/pokt-network/pocket/issues/76): Review all the `gov_*.go` related files and simplify the code
- [ ] REFACTOR/DISCUSS: Should we prefix the functions in the `PersistenceModule` with the Param / Actor it's impacting to make autocomplete in implementation better?
- [ ] DISCUSS: Consider removing all `Set` methods (e.g. `SetAccountAmount`) and replace with `Add` (e.g. `AddAccountAmount`) by having it leverage a "default zero".
- [ ] REFACTOR(https://github.com/pokt-network/pocket/issues/102): Split `account` and `pool` into a shared actor (e.g. like fisherman/validator/serviceNode/application) and simplify the code in half
- [ ] CLEANUP: Remove `tokens` or `stakedTokens` in favor of using `amount` everywhere since the denomination is not clear. As a follow up. Consider a massive rename to make the denomination explicit.

Mid-term (i.e. new feature or major refactor) tasks:

- [ ] IMPROVE: Consider using prepare statements and/or a proper query builder
- [ ] TODO(https://github.com/pokt-network/pocket/issues/77): Implement proper DB SQL migrations
- [ ] INVESTIGATE: Benchmark the queries (especially the ones that need to do sorting)
- [ ] DISCUSS: Look into `address` is being computed (string <-> hex) and determine if we could/should avoid it
-

Long-term (i.e. design) tasks

- [ ] INVESTIGATE: Expand the existing fuzzing approach to push random changes in state transitions to its limit.
- [ ] INVESTIGATE: Use a DSL-like approach to design complex "user stories" for state transitions between protocol actors in different situations.
