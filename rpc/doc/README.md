# RPC

This document is meant to be a starting point/placeholder for a full-fledged RPC specification that allows interaction with the nodes.

## Inspiration

Pocket V0 has inspired a lot the first iteration but then we converged towards a spec-centric approach, where the boilerplate code (serialization, routing, etc) is derived from an [OpenAPI 3.0](../v1/openapi.yaml) specification.

This approach will allow us to focus on the features and less on the boilerpate and ultimately to iterate more quickly as we discover the way ahead of us.

## Code generation

The current implementation uses code generation for ease of development.

The source of truth is the the [OpenAPI3.0 yaml file](../v1/openapi.yaml) (also conveniently visible [here](https://editor.swagger.io/?url=https://raw.githubusercontent.com/pokt-network/pocket/main/rpc/v1/openapi.yaml) via the Swagger Editor)

Anytime we make changes to the yaml file, we need to regenerate the boilerplate code by running

```bash
$ make generate_rpc_openapi
```

The compilation errors should guide towards the next steps.

## Transports

Currently, the API is in its simplest form. Basically a **REST API**.

As the codebase matures, we'll consider other transports such as [**JSON RPC 2.0**](https://www.jsonrpc.org/specification) and [**GRPC**](https://grpc.io/).

## Spec

<!-- TODO (deblasis): add link when merged to `main` -->

The Swagger Editor with preview is available here.

This first iteration includes the bare minimum:

### Node related

- Node liveness check (**GET /v1/health**)
- Node version check (**GET /v1/version**)

These are pretty self-explanatory.

### Transaction related

- Sync signed transaction submission (**POST /v1/client/broadcast_tx_sync**)

  #### Payload:

  ```json
  {
    "address": "string",
    "raw_hex_bytes": "string"
  }
  ```

  `address`: specifies the address the transaction originates from.

  `raw_hex_bytes`: hex encoded JSON of a signed transaction.

  #### Return:

  Currently only **OK** (HTTP Status code 200) or **KO** (HTTP Status code 4xx/5xx)

  This API might be extended to return potentially useful information such as the transaction hash which is known at the moment of submission and can be used to query the blockchain.

#### What's next?

Definitely we'll need ways to retrieve transactions as well so we can envisage:

- Get a transaction by hash (**GET /v1/client/tx**)

## Code Organization

```bash
├── client.gen.config.yml    # code generation config for the client
├── client.gen.go            # generated client boilerplate code
├── doc                      # folder containing RPC specific docs
├── handlers.go              # concrete implementation of the HTTP handlers invoked by the server
├── module.go                # RPC module
├── noop_module.go           # noop RPC module (used when the module is disabled)
├── server.gen.config.yml    # code generation config for the server + dtos
├── server.gen.go            # generated server boilerplate code
├── server.go                # RPC server configuration and initialization
├── types
│   ├── proto
│   │   └── rpc_config.proto # protobuf file describing the RPC module configuration
│   └── rpc_config.pb.go     # protoc generated struct and methods for RPC config
└── v1
    └── openapi.yaml         # OpenAPI v3.0 spec (source for the generated files above)
```
