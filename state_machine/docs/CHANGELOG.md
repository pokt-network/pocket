# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.0.3] - 2023-02-27

- Update state machine states, introduced new state transitions

## [0.0.0.2] - 2023-02-24

- Update logger value references with pointers

## [0.0.0.1] - 2023-02-17

- Introduced this `CHANGELOG.md` and  `README.md`
- Added `StateMachineModule` implementation with a POC of the finite state machine that will be used to manage the node lifecycle
- Added `StateMachine` diagram generator (linked in README.md)
- Integrated the `StateMachine` with the `bus` to propagate `StateMachineTransitionEvent` events whenever they occur

<!-- GITHUB_WIKI: changelog/state_machine -->
