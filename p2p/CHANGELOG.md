# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.11] - 2022-12-15

- `ValidatorMapToAddrBook` renamed to `ActorToAddrBook`
- `ValidatorToNetworkPeer` renamed to `ActorToNetworkPeer`

## [0.0.0.10] - 2022-12-14

- mempool cap is now configurable via P2PConfig. Tests implement the mock accordingly.
- Introduced the concept of a `addrbookProvider` that abstracts the fetching and the mapping from `Actor` to `NetworkPeer`
- Temporary hack to allow access to the `addrBook` to the debug client (will be removed in an upcoming PR already in the works for issues [#203](https://github.com/pokt-network/pocket/issues/203) and [#331](https://github.com/pokt-network/pocket/issues/331))
- Transport related functions are now in the `transport` package
- Updated tests to source the `addrBook` from the `addrbookProvider` and therefore `Persistence`
- Updated Raintree network constructur with dependency injection
- Updated stdNetwork constructur with dependency injection
- Improved documentation for the `peersManager`

## [0.0.0.9] - 2022-12-04

- Raintree mempool cannot grow unbounded anymore. It's now bounded by a constant limit and when new nonces are inserted the oldest ones are removed.
- Raintree is now capable of fetching the address book for a previous height and to instantiate an ephemeral `peersManager` with it.

## [0.0.0.8] - 2022-11-14

- Removed topic from messaging

## [0.0.0.7] - 2022-10-24

- Updated README to reference the python simulator as a learning references and unit test generation tool
- Added a RainTree unit test for 12 nodes using the simulator in https://github.com/pokt-network/rain-tree-sim/blob/main/python

## [0.0.0.6] - 2022-10-20

- Add a telemetry `send` event within the context `RainTree` network module that is triggered during network writes
- Change the `RainTree` testing framework counting method to simulate real reads/writes from the network
- Improve documentation related to the `RainTree` testing framework & how the counters are computed

## [0.0.0.5] - 2022-10-12

### [#235](https://github.com/pokt-network/pocket/pull/235) Config and genesis handling

- Updated to use `RuntimeMgr`
- Updated tests and mocks
- Removed some cross-module dependencies

## [0.0.0.4] - 2022-10-06

- Don't ignore the exit code of `m.Run()` in the unit tests

## [0.0.0.3] - 2022-09-15

**[TECHDEBT] AddrBook management optimization and refactoring** [#246](github.com/pokt-network/pocket/issues/246)

- Added `peersManager` and `target` in order to abstract away and eliminate redundant computations
- Refactored debug logging in `getTarget` to print first and second target on the same line
- Refactored `AddPeerToAddrBook` to use an event-driven approach in order to leverage sorted data structures
- Added `RemovePeerToAddrBook` making use of the same pattern
- Improved performance of `AddPeerToAddrBook` and `RemovePeerToAddrBook` by making the implementations O(n)
- Updated `stdnetwork` to use a map instead of a slice

## [0.0.0.2] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Ensured proto structures implement shared interfaces
- `P2PConfig` uses shared interfaces in order to accept `MockP2PConfig` in `test_artifacts`
- Moved connection_type to bool for simplicity (need to figure out how to do Enums without sharing the structure)

## [0.0.0.1] - 2022-07-26

- Deprecated old p2p for pre2p raintree

## [0.0.0.0] - 2022-06-16

- RainTree first iteration in Pre2P module (no cleanup or redundancy)
