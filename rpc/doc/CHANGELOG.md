# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.11] - 2023-02-10

- Updated RPC to expose the node's address book via GET `/v1/p2p/staked_actors_address_book`
- Updated modules to embed `base_modules.IntegratableModule` and `base_modules.InterruptableModule` for DRYness

## [0.0.0.10] - 2023-02-07

- Added GITHUB_WIKI tags where it was missing

## [0.0.0.9] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.8] - 2023-02-06

- Changed readme to remove $ sign from code blocks

## [0.0.0.7] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.6] - 2023-01-23

- Added `pprof` http server feature flag via build tags

## [0.0.0.5] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method

## [0.0.0.4] - 2023-01-03

- Updated to use the new centralized config and genesis handling

## [0.0.0.3] - 2022-12-14

- Updated to use `GetBus()` instead of `bus` wherever possible

## [0.0.0.2] - 2022-12-06

- Updated `PostV1ClientBroadcastTxSync` to broadcast the transaction it receives
- Avoid creating an unnecessary utility context and use the utility module directly

## [0.0.0.1] - 2022-11-02

- Consensus State endpoint
- Added CORS feature flag and config
- Added dockerized swagger-ui

## [0.0.0.0] - 2022-10-20

- First iteration of the RPC
  - Endpoint: Node liveness
  - Endpoint: Node version
  - Endpoint Synchronous signed transaction broadcast
  - Spec: basic Openapi.yaml
  - Codegen: code generation for the Server + DTOs
  - Codegen: code generation for the Client

<!-- GITHUB_WIKI: changelog/rpc -->
