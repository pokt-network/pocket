# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.6] - 2023-04-10

- Makes Account sub-command respect non-interactive flag

## [0.0.0.5] - 2023-03-24

- The debug CLI now updates its peerstore mimicking the behavior of the validators via `sendConsensusNewHeightEventToP2PModule`.

## [0.0.0.4] - 2023-03-24

- Updated debug keystore initialization to use an embedded backup instead of the yaml file that has to be rehydrated every time.

## [0.0.0.3] - 2023-03-20

- Adds message routing type field labels to debug CLI actions

## [0.0.0.2] - 2023-03-14

- Simplifies the debug CLI tooling by embedding private-keys.yaml manifest
  into the CLI binary when the debug build tag is present.

## [0.0.0.1] - 2023-02-21

- Rename ServiceNode Actor Type Name to Servicer

<!-- GITHUB_WIKI: app -->
