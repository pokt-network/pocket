# Persistence Module

This document is meant to be a supplement to the living specification of [1.0 Pocket's Persistence Specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/persistence) primarily focused on the implementation, and additional details related to the design of the codebase and information related to development. D

## Database Migrations

### Configuration

The persistence specific configuratin within `config.json` looks like this:

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
