All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.6] - 2023-03-22

- Improve URL validation and error handling in Libp2pMultiaddrFromServiceURL function

## [0.0.0.5] - 2023-03-21

- Refactored libp2p module to use new P2P interfaces

## [0.0.0.4] - 2023-03-18

- Corrected libp2p module tests
- Improve libp2p code quality in response to post-merge review feedback

## [0.0.0.3] - 2023-03-15

- Added mockdns as a test dependency
- Mocked DNS resolution in url_conversion_test.go
- Added regression tests to url_conversion_test.go for single- and multi-record DNS responses

## [0.0.0.2] - 2023-03-03

- Added a new `modules.P2PModule` implementation to the `libp2p` module directory

## [0.0.0.1] - 2023-03-03

- Added a new `typesP2P.Network` implementation to the `libp2p` module directory
- Added `PoktProtocolID` for use within the libp2p module or by public API consumers

## [0.0.0.0] - 2023-02-23

- prepare pocket repo new libp2p module
- add pocket / libp2p identity helpers
- add url <--> multiaddr conversion helpers for use with libp2p (see: https://github.com/multiformats/go-multiaddr)
- add `Multiaddr` field to `typesP2P.NetworkPeer`

<!-- GITHUB_WIKI: changelog/libp2p -->
