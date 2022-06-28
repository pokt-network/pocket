# Persistence Module <!-- omit in toc -->

This document is meant to be a supplement to the living specification of [1.0 Pocket's Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) primarily focused on the implementation, and additional details related to the design of the codebase and information related to development.

- [Database Migrations](#database-migrations)
  - [Configuration](#configuration)
  - [LocalNet](#localnet)
- [Debugging & Development](#debugging--development)
  - [Makefile Helpers](#makefile-helpers)
    - [Admin View - db_admin](#admin-view---db_admin)
    - [Benchmarking - db_bench](#benchmarking---db_bench)
- [Testing](#testing)
  - [Dependencies](#dependencies)
  - [Setup](#setup)
    - [Setup Issue - Docker Daemon is not Running](#setup-issue---docker-daemon-is-not-running)
    - [Setup Issue - Docker Daemon is not Running](#setup-issue---docker-daemon-is-not-running-1)
  - [Unit Tests](#unit-tests)
- [Implementation FAQ](#implementation-faq)
- [Implementation TODO(team)](#implementation-todoteam)

## Database Migrations

### Configuration

The persistence specific configuration within `config.json` looks like this:

```
  "persistence": {
    "postgres_url": "postgres://postgres:postgres@pocket-db:5432/postgres",
    "schema": "node1"
  }
```

Note that the `schema` parameter must be unique on a per node basis

### LocalNet

For LocalNet, we run a single Postgres instance, that is logically split by node using the `schema` config above. It therefore needs to be unique on a per node basis.

## Debugging & Development

### Makefile Helpers

If you run `make` from the root of the `pocket` repo, there will be several targets prefixed with `db_` that can help with design & development of this module.

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

### Unit Tests

```
$ make test_persistence
```

## Implementation FAQ

**Q**: Why do `Get` methods (e.g. `GetAccountAmount`) not return 0 by default?
**A**: This was

**Q**: Why are amounts strings?
**A**: A lesson from Tendermint in order to enforce the use of BigInts throughout and avoid floating point issues when storing data on disk.

Q: Why not use an ORM?

## Implementation TODO(team)

These are major TODOs spanning the entire repo so they are documented here instead.

- [ ] Remove unused parameters from the PostgresContext interface (see where \_ is used in the implementation e.g. `InsertFisherman`)
- [ ] Benchmark the queries (especially the ones that need to do sorting
- [ ] Consider converting all address params from bytes to string to avoid unnecessary encoding
- [ ] Look into `address` being a computed field in most Postgres tables
- [ ] Figure out how to handle conflicts while having multiple constraints with Postgres
- [ ] Remove all the SQL strings and use a proper query builder
- [ ] Remove all the SQL strings and consider using prepared statements
- [ ] Convert all queries to use 'height' in the interface for historical lookups
- [ ] Consider removing all `Set` methods (e.g. `SetAccountAmount`) and replace with `Add` (e.g. `AddAccountAmount`) by having it leverage a "default zero".
- [ ] Use proper SQL migration methodoligies
