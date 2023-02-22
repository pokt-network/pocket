# Utility Module

This document is meant to be a placeholder to serve as a living representation of the parent 'Utility Module' codebase until the code matches the [1.0 Utility Module specification](https://github.com/pokt-network/pocket-network-protocol/tree/main/utility).

_This document will be archived once the implementation is synonymous with the current 1.0 Utility specification, ._

## Origin Document

Currently, the Utility Module minimally implements the first iteration of an account based, state machine, blockchain protocol.

Synonymous technologies to this current iteration are 'MVP proof of stake' blockchain state machines, that are NOT Pocket Network specific.

The current implementation does add the fundamental Pocket Network 1.0 actors:

- Accounts
- Validators
- Fishermen
- Applications
- Servicers

And implement the basic transaction functionality:

- Send
- Stake
- Unstake
- EditStake
- Pause
- Unpause

Added governance params:

- BlocksPerSessionParamName

- AppMinimumStakeParamName
- AppMaxChainsParamName
- AppBaselineStakeRateParamName
- AppStakingAdjustmentParamName
- AppUnstakingBlocksParamName
- AppMinimumPauseBlocksParamName
- AppMaxPauseBlocksParamName

- ServicerMinimumStakeParamName
- ServicerMaxChainsParamName
- ServicerUnstakingBlocksParamName
- ServicerMinimumPauseBlocksParamName
- ServicerMaxPauseBlocksParamName
- ServicersPerSessionParamName

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
- MessageFishermanPauseServicerFee
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
- MessageStakeServicerFee
- MessageEditStakeServicerFee
- MessageUnstakeServicerFee
- MessagePauseServicerFee
- MessageUnpauseServicerFee
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
- ServicerMinimumStakeOwner
- ServicerMaxChainsOwner
- ServicerUnstakingBlocksOwner
- ServicerMinimumPauseBlocksOwner
- ServicerMaxPausedBlocksOwner
- ServicersPerSessionOwner
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
- MessageFishermanPauseServicerFeeOwner
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
- MessageStakeServicerFeeOwner
- MessageEditStakeServicerFeeOwner
- MessageUnstakeServicerFeeOwner
- MessagePauseServicerFeeOwner
- MessageUnpauseServicerFeeOwner
- MessageChangeParameterFeeOwner

And minimally satisfy the following interface:

```go
CheckTransaction(tx []byte) error
GetProposalTransactions(proposer []byte, maxTransactionBytes int, lastBlockByzantineValidators [][]byte) (txs [][]byte, err error)
ApplyBlock(Height int64, proposer []byte, txs [][]byte, lastBlockByzantineValidators [][]byte) (appHash []byte, err error)
```

## How to build

Utility Module does not come with its own cmd executables.

Rather, it is purposed to be a dependency (i.e. library) of other modules

## How to use

Utility implements the `UtilityModule` and subsequent interface
[`pocket/shared/modules/utility_module.go`](https://github.com/pokt-network/pocket/shared/modules/utility_module.go).

To use, simply initialize a Utility instance using the factory function like so:

```go
utilityMod, err := utility.Create(bus)
```

and use `utilityMod` as desired.

## How to test

```
make test_utility_types && make test_utility_module
```

## Code Organization

```bash
utility
├── account.go     # utility context for accounts & pools
├── actor.go       # utility context for apps, fish, nodes, and validators
├── block.go       # utility context for blocks
├── gov.go         # utility context for dao & parameters
├── module.go      # module implementation and interfaces
├── session.go     # utility context for the session protocol
├── transaction.go # utility context for transactions including handlers
├── doc            # contains the documentation and changelog
├── test           # utility unit tests
├── types          # stateless (without relying on persistence) library of utility types
│   ├── proto          # protobuf3 messages that auto-generate into the types directory
│   ├── actor.go
│   ├── error.go
│   ├── mempool.go
│   ├── message.go     # payloads of transactions
│   ├── transaction.go # the finite unit of the block
│   ├── util.go
│   ├── vote.go        # vote structure for double sign transaction
```

<!-- GITHUB_WIKI: utility/readme -->
