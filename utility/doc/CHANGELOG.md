# Changelog

All notable changes to this module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2022-07-20

### Code cleanup
- Removed transaction fees from the transaction structure as fees will be enforced at the state level
- Removed actor specific messages (besides DoubleSign) and added actorType field to the struct
- Removed pause messages and functionality as it is out of scope for the current POS iteration
- Removed session and test-scoring as it's out of scope for the current POS iteration
- Consolidated unit test functionality for actors
- Modified pre-persistence to match persistence for Update(actor), 'amountToAdd' is now just 'amount'

## [Unreleased]

## [0.0.0] - 2022-03-15

### Added

- Added minimal 'proof of stake' implementation with few Pocket Specific terminologies and actors
  - Structures
    - Accounts
    - Validators
    - Fishermen
    - Applications
    - Service Nodes
    - Pools
  - Messages
    - Stake
    - Unstake
    - EditStake
    - Pause
    - Unpause
    - Send
- Added initial governance params
