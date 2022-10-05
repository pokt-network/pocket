# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.4] - 2022-10-04

- Don't ignore the exit code of `m.Run()` in the unit tests

## [0.0.0.3] - 2022-09-15

**[TECHDEBT] AddrBook management optimization and refactoring** [#246](github.com/pokt-network/pocket/issues/246)

- Added `peersManager` and `target` in order to abstract away and eliminate redundant computations
- Refactored debug logging in `getTarget` to print first and second target on the same line
- Refactored `AddPeerToAddrBook` to use an event-driven approach in order to leverage sorted data structures
- Added `RemovePeerToAddrBook` making use of the same pattern
- Improved performance of `AddPeerToAddrBook` and `RemovePeerToAddrBook` by making the implementations O(n)
- Updated `stdnetwork` to use a map instead of a slice
  > > > > > > > main

## [0.0.0.2] - 2022-08-25

**Encapsulate structures previously in shared [#163](github.com/pokt-network/pocket/issues/163)**

- Ensured proto structures implement shared interfaces
- `P2PConfig` uses shared interfaces in order to accept `MockP2PConfig` in `test_artifacts`
- Moved connection_type to bool for simplicity (need to figure out how to do Enums without sharing the structure)

## [0.0.0.1] - 2022-07-26

- Deprecated old p2p for pre2p raintree

## [0.0.0.0] - 2022-06-16

- RainTree first iteration in Pre2P module (no cleanup or redundancy)
