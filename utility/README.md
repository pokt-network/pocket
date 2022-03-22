# Utility Module

This document is meant to be a placeholder to serve as a living representation of the parent 'Utility Module' codebase until the code matches the [1.0 Utility Module specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/utility).

The spirit of the documentation is to continuously update and inform the reader of the general scope of the Utility Module as breaking, rapid development occurs.

_This document will be archived once the implementation is synonymous with the current 1.0 Utility specification, ._

# Origin Document

Currently, the Utility Module minimally implements the first iteration of an account based, state machine, blockchain protocol.

Synonymous technologies to this current iteration are 'MVP proof of stake' blockchain state machines, that are NOT Pocket Network specific.

The current implementation does add the fundamental Pocket Network 1.0 actors:

- Accounts
- Validators
- Fishermen
- Applications
- Service Nodes

And implement the basic transaction functionality:

- Send-Tx
- Stake
- Unstake
- EditStake
- Pause
- Unpause

And a few additional skeleton implementations for pocket specific transactions:

- FishermanPauseServiceNode [x] Implemented
- TestScore [x] Shell // requires sessions & report card structures
- ProveTestScore [x] Shell // requires sessions & report card structures

Added governance params:

- BlocksPerSessionParamName

- AppMinimumStakeParamName
- AppMaxChainsParamName
- AppBaselineStakeRateParamName
- AppStakingAdjustmentParamName
- AppUnstakingBlocksParamName
- AppMinimumPauseBlocksParamName
- AppMaxPauseBlocksParamName

- ServiceNodeMinimumStakeParamName
- ServiceNodeMaxChainsParamName
- ServiceNodeUnstakingBlocksParamName
- ServiceNodeMinimumPauseBlocksParamName
- ServiceNodeMaxPauseBlocksParamName
- ServiceNodesPerSessionParamName

- FishermanMinimumStakeParamName
- FishermanMaxChainsParamName
- FishermanUnstakingBlocksParamName
- FishermanMinimumPauseBlocksParamName
- FishermanMaxPauseBlocksParamName

- ValidatorMinimumStakeParamName
- ValidatorUnstakingBlocksParamName
- ValidatorMinimumPauseBlocksParamName
- ValidatorMaxPausedBlocksParamName
- ValidatorMaximumMissedBlocksParamName

- ValidatorMaxEvidenceAgeInBlocksParamName
- ProposerPercentageOfFeesParamName
- MissedBlocksBurnPercentageParamName
- DoubleSignBurnPercentageParamName

- MessageDoubleSignFee
- MessageSendFee
- MessageStakeFishermanFee
- MessageEditStakeFishermanFee
- MessageUnstakeFishermanFee
- MessagePauseFishermanFee
- MessageUnpauseFishermanFee
- MessageFishermanPauseServiceNodeFee
- MessageTestScoreFee
- MessageProveTestScoreFee
- MessageStakeAppFee
- MessageEditStakeAppFee
- MessageUnstakeAppFee
- MessagePauseAppFee
- MessageUnpauseAppFee
- MessageStakeValidatorFee
- MessageEditStakeValidatorFee
- MessageUnstakeValidatorFee
- MessagePauseValidatorFee
- MessageUnpauseValidatorFee
- MessageStakeServiceNodeFee
- MessageEditStakeServiceNodeFee
- MessageUnstakeServiceNodeFee
- MessagePauseServiceNodeFee
- MessageUnpauseServiceNodeFee
- MessageChangeParameterFee

- AclOwner
- BlocksPerSessionOwner
- AppMinimumStakeOwner
- AppMaxChainsOwner
- AppBaselineStakeRateOwner
- AppStakingAdjustmentOwner
- AppUnstakingBlocksOwner
- AppMinimumPauseBlocksOwner
- AppMaxPausedBlocksOwner
- ServiceNodeMinimumStakeOwner
- ServiceNodeMaxChainsOwner
- ServiceNodeUnstakingBlocksOwner
- ServiceNodeMinimumPauseBlocksOwner
- ServiceNodeMaxPausedBlocksOwner
- ServiceNodesPerSessionOwner
- FishermanMinimumStakeOwner
- FishermanMaxChainsOwner
- FishermanUnstakingBlocksOwner
- FishermanMinimumPauseBlocksOwner
- FishermanMaxPausedBlocksOwner
- ValidatorMinimumStakeOwner
- ValidatorUnstakingBlocksOwner
- ValidatorMinimumPauseBlocksOwner
- ValidatorMaxPausedBlocksOwner
- ValidatorMaximumMissedBlocksOwner
- ValidatorMaxEvidenceAgeInBlocksOwner
- ProposerPercentageOfFeesOwner
- MissedBlocksBurnPercentageOwner
- DoubleSignBurnPercentageOwner
- MessageDoubleSignFeeOwner
- MessageSendFeeOwner
- MessageStakeFishermanFeeOwner
- MessageEditStakeFishermanFeeOwner
- MessageUnstakeFishermanFeeOwner
- MessagePauseFishermanFeeOwner
- MessageUnpauseFishermanFeeOwner
- MessageFishermanPauseServiceNodeFeeOwner
- MessageTestScoreFeeOwner
- MessageProveTestScoreFeeOwner
- MessageStakeAppFeeOwner
- MessageEditStakeAppFeeOwner
- MessageUnstakeAppFeeOwner
- MessagePauseAppFeeOwner
- MessageUnpauseAppFeeOwner
- MessageStakeValidatorFeeOwner
- MessageEditStakeValidatorFeeOwner
- MessageUnstakeValidatorFeeOwner
- MessagePauseValidatorFeeOwner
- MessageUnpauseValidatorFeeOwner
- MessageStakeServiceNodeFeeOwner
- MessageEditStakeServiceNodeFeeOwner
- MessageUnstakeServiceNodeFeeOwner
- MessagePauseServiceNodeFeeOwner
- MessageUnpauseServiceNodeFeeOwner
- MessageChangeParameterFeeOwner

And minimally satisfy the following interface:

```go
CheckTransaction(tx []byte) error

GetTransactionsForProposal(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (transactions [][]byte, err error)

ApplyBlock(Height int64, proposer []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)
```

## How to build

Utility Module does not come with its own cmd executables.

Rather, it is purposed to be a dependency (i.e. library) of other modules

## How to use

Utility implements the `UtilityModule` and subsequent interface
[`pocket/shared/modules/utility_module.go`](https://github.com/pokt-network/pocket/shared/modules/utility_module.go).

To use, simply initialize a Utility instance using the factory function like so:

```go
utilityMod, err := utility.Create(config)
```

and use `utilityMod` as desired.

## How to test

```
$ make test_utility_module
```
