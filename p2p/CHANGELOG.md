# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.45] - 2023-04-25

- Added rainTeeFactory type & compile-time enforcement

## [0.0.0.44] - 2023-04-20

- Refactor `mockdns` test helpers

## [0.0.0.43] - 2023-04-17

- Add test to exercise `sortedPeersView#Add()` and `#Remove()`
- Fix raintree add/remove index

## [0.0.0.42] - 2023-04-17

- Added a test which asserts that transport encryption is required (i.e. unencrypted connections are refused)

## [0.0.0.41] - 2023-04-17

- Moved peer & url conversion utils to `p2p/utils` package
- Refactor `getPeerIP` to use `net.DefaultResolver` for easier testing
- Moved & refactor libp2p `host.Host` setup util to `p2p/utils`
- Consolidated Libp2p & P2P `modules.Module` implementations
- Consolidated Libp2p & P2P `stdnetwork` `typesP2P.Network` implementations
- Refactored raintree `typesP2P.Network` implementation to use libp2p
- Moved `shared/p2p` package into `p2p/types` packages
- Removed `Trnasport` interface and implementations
- Removed `ConnectionFactory` type and related members
- Added libp2p `host.Host` mock generator
- Refactor raintree constructor function signature to use new `RainTreeConfig` struct

## [0.0.0.40] - 2023-04-12

- Wrap IPv6 address in square brackets as per RFC3986 §3.2.2

## [0.0.0.39] - 2023-04-12

- Improve URL validation and error handling in Libp2pMultiaddrFromServiceURL function

## [0.0.0.38] - 2023-04-10

- Switched mock generation to use reflect mode for effected interfaces (`modules.ModuleFactoryWithOptions` embedders)

## [0.0.0.37] - 2023-03-30

- Variable name and comment improvements

## [0.0.0.36] - 2023-03-24

- Updated errors on Send from fatal to recoverable
- Updated `PeerstoreProvider` to ignore gracefully peers that are not resolvable/reachable

## [0.0.0.35] - 2023-03-21

- Add log for `StateMachineTransitionEvent`

## [0.0.0.34] - 2023-03-16

- Refactored P2P module to use new P2P interfaces
- Moved `typesP2P.AddrBookMap` to `sharedP2P.PeerAddrMap` and refactor to implement the new `Peerstore` interface
- Factored `SortedPeerManager` out of `raintree.peersManager` and add `peerManager` interface
- Refactored `raintree.peersManager` to use `SortedPeerManager` and implement `PeerManager` interface
- Refactored `stdnetwork.Network` implementation to use P2P interfaces
- Refactored `getAddrBookDelta` to be a member of `PeerList`
- Refactored `AddrBookProvider` to use new P2P interfaces
- Renamed `AddrBookProvider` to `PeerstoreProvider`
- Refactored `typesP2P.Network` to use new P2P interfaces
- Refactored `typesP2P.Transport` to embed `io.ReadWriteCloser`
- Renamed `NetworkPeer#Dialer` to `NetworkPeer#Transport`for readability and consistency
- Refactored `typesP2P.NetworkPeer` to implement the new `Peer` interface

## [0.0.0.33] - 2023-03-03

- Add TECHDEBT comments

## [0.0.0.32] - 2023-03-03

- Added embedded `modules.InitializableModule` to the P2P `AddrBookProvider` interface so that it can be dependency injected as a `modules.Module` via the bus registry.

## [0.0.0.31] - 2023-03-01

- replace `consensus_port` with `port` in P2P config
- update default P2P config `port` to from `8080` to `42069`

## [0.0.0.30] - 2023-02-28

- Renamed package names and parameters to reflect changes in the rest of the codebase

## [0.0.0.29] - 2023-02-24

- Update logger value references with pointers

## [0.0.0.28] - 2023-02-20

- Added basic `bootstrap` nodes support
- Reacting to `ConsensusNewHeightEventType` and `StateMachineTransitionEventType` to update the address book and current height and determine if a bootstrap is needed

## [0.0.0.27] - 2023-02-17

- Deprecated `debugAddressBookProvider`
- Added `rpcAddressBookProvider` to source the address book from the RPC server
- Leveraging `bus` for dependency injection of the `addressBookProvider` and `currentHeightProvider`
- Deprecated `debugCurrentHeightProvider`
- Added `rpcCurrentHeightProvider` to source the current height from the RPC server
- Fixed raintree to use the `currentHeightProvider` instead of consensus (that was what we wanted to avoid in the first place)
- Added `getAddrBookDelta` to calculate changes to the address book between heights and update the internal state and componentry accordingly

## [0.0.0.26] - 2023-02-17

- Modules embed `base_modules.IntegratableModule` and `base_modules.InterruptableModule` for DRYness
- Updated tests

## [0.0.0.25] - 2023-02-09

- Updated logging initialization and passing to the network component instead of using the global logger
- Fixed incorrect use of `bus.GetLoggerModule()` in `stdnetwork.go` since it's never initialized when running the debug CLI

## [0.0.0.24] - 2023-02-06

- Address legacy linter errors from `golangci-lint`

## [0.0.0.23] - 2023-02-04

- Changed log lines to utilize new logger module.

## [0.0.0.22] - 2023-02-03

- Using the generic `mempool.GenericFIFOSet` as a `nonceDeduper`
- Added tests for `nonceDeduper` to ensure that it behaves as expected.

## [0.0.0.21] - 2023-01-30

- Updated `TestRainTreeAddrBookUtilsHandleUpdate` and `testRainTreeMessageTargets` to correct incorrect expected and actual value placements.

## [0.0.0.20] - 2023-01-20

- Updated `P2PConfig#IsEmptyConnectionType` bool to `P2PConfig#ConnectionType` enum

## [0.0.0.19] - 2023-01-19

- Rewrite `interface{}` to `any`

## [0.0.0.18] - 2023-01-11

- Add a lock to the mempool to avoid parallel messages which has caused the node to crash in the past

## [0.0.0.17] - 2023-01-10

- Updated module constructor to accept a `bus` and not a `runtimeMgr` anymore
- Registering module with the `bus` via `RegisterModule` method
- Updated tests and mocks accordingly
- Sorting `validatorIds` in `testRainTreeCalls`

## [0.0.0.16] - 2023-01-09

- Added missing `Close()` call to `persistenceReadContext`

## [0.0.0.15] - 2023-01-03

- Refactored `AddrBookProvider` to support multiple implementations
- Added `CurrentHeightProvider`
- Dependency injection of the aforementioned provider into the module creation (used by the debug-client)
- Updated implementation to use the providers
- Updated tests and mocks

## [0.0.0.14] - 2023-01-03

- `ActorsToAddrBook` now skips actors that are not validators since they don't have a serviceUrl generic parameter

## [0.0.0.13] - 2022-12-21

- Updated to use the new centralized config and genesis handling
- Updated to use the new `Actor` struct under `coreTypes`
- Updated tests and mocks
- Added missing `max_mempool_count` in config (it was causing P2P instabilities in LocalNet)

## [0.0.0.12] - 2022-12-16

- `ValidatorMapToAddrBook` renamed to `ActorToAddrBook`
- `ValidatorToNetworkPeer` renamed to `ActorToNetworkPeer`

## [0.0.0.11] - 2022-12-15

- Bugfix for [[#401](https://github.com/pokt-network/pocket/issues/401)]
- Fixed typo in 'peers_manager.go'

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

<!-- GITHUB_WIKI: changelog/p2p -->
