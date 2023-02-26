package utility

// Internal business logic containing the lifecycle of Block-related operations

import (
	"encoding/hex"
	"fmt"
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// The ApplyBlock function is the 'main' operation that executes a 'block' object against the state.
// Pocket Network adopt a Tendermint-like lifecycle of BeginBlock -> DeliverTx -> EndBlock in that
// order. Like the name suggests, BeginBlock is an autonomous state operation that executes at the
// beginning of every block DeliverTx individually applies each transaction against the state and
// rolls it back (not yet implemented) if fails. Like BeginBlock, EndBlock is an autonomous state
// operation that executes at the end of every block.

// TODO: Make sure to call `utility.HandleTransaction`, which calls `persistence.TransactionExists`
func (u *utilityContext) CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (stateHash string, txs [][]byte, err error) {
	lastBlockByzantineVals, err := u.getLastBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}
	// begin block lifecycle phase
	if err := u.beginBlock(lastBlockByzantineVals); err != nil {
		return "", nil, err
	}
	txs = make([][]byte, 0)
	totalTxsSizeInBytes := 0
	txIndex := 0

	mempool := u.GetBus().GetUtilityModule().GetMempool()
	for !mempool.IsEmpty() {
		txBytes, err := mempool.PopTx()
		if err != nil {
			return "", nil, err
		}
		tx, err := coreTypes.TxFromBytes(txBytes)
		if err != nil {
			return "", nil, err
		}
		txTxsSizeInBytes := len(txBytes)
		totalTxsSizeInBytes += txTxsSizeInBytes
		if totalTxsSizeInBytes >= maxTransactionBytes {
			// Add back popped tx to be applied in a future block
			err := mempool.AddTx(txBytes)
			if err != nil {
				return "", nil, err
			}
			break // we've reached our max
		}
		txResult, err := u.applyTx(txIndex, tx)
		if err != nil {
			u.logger.Err(err).Msg("Error in ApplyTransaction")
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := u.revertLastSavePoint(); err != nil {
				return "", nil, err
			}
			totalTxsSizeInBytes -= txTxsSizeInBytes
			continue
		}
		if err := u.store.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}

		txs = append(txs, txBytes)
		txIndex++
	}

	if err := u.endBlock(proposer); err != nil {
		return "", nil, err
	}
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err = u.store.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
	}
	u.logger.Info().Msgf("CreateAndApplyProposalBlock - computed state hash: %s", stateHash)

	return stateHash, txs, err
}

// TODO: Make sure to call `utility.HandleTransaction`, which calls `persistence.TransactionExists`
// CLEANUP: code re-use ApplyBlock() for CreateAndApplyBlock()
func (u *utilityContext) ApplyBlock() (string, error) {
	lastByzantineValidators, err := u.getLastBlockByzantineValidators()
	if err != nil {
		return "", err
	}

	// begin block lifecycle phase
	if err := u.beginBlock(lastByzantineValidators); err != nil {
		return "", err
	}

	// deliver txs lifecycle phase
	for index, txProtoBytes := range u.proposalBlockTxs {
		tx, err := coreTypes.TxFromBytes(txProtoBytes)
		if err != nil {
			return "", err
		}
		if err := tx.ValidateBasic(); err != nil {
			return "", err
		}
		// TODO(#346): Currently, the pattern is allowing nil err with an error transaction...
		//             Should we terminate applyBlock immediately if there's an invalid transaction?
		//             Or wait until the entire lifecycle is over to evaluate an 'invalid' block

		// Validate and apply the transaction to the Postgres database
		txResult, err := u.applyTx(index, tx)
		if err != nil {
			return "", err
		}

		if err := u.store.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}

		// TODO: if found, remove transaction from mempool.
		// DISCUSS: What if the context is rolled back or cancelled. Do we add it back to the mempool?
		// if err := u.Mempool.RemoveTx(tx); err != nil {
		// 	return nil, err
		// }
	}

	// end block lifecycle phase
	if err := u.endBlock(u.proposalProposerAddr); err != nil {
		return "", err
	}
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err := u.store.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
		return "", typesUtil.ErrAppHash(err)
	}
	u.logger.Info().Msgf("ApplyBlock - computed state hash: %s", stateHash)

	// return the app hash; consensus module will get the validator set directly
	return stateHash, nil
}

func (u *utilityContext) beginBlock(previousBlockByzantineValidators [][]byte) typesUtil.Error {
	if err := u.handleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) endBlock(proposer []byte) typesUtil.Error {
	// reward the block proposer
	if err := u.handleProposerRewards(proposer); err != nil {
		return err
	}
	// unstake actors that have been 'unstaking' for the <Actor>UnstakingBlocks
	if err := u.unbondUnstakingActors(); err != nil {
		return err
	}
	// begin unstaking the actors who have been paused for MaxPauseBlocks
	if err := u.beginUnstakingMaxPausedActors(); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) unbondUnstakingActors() (err typesUtil.Error) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		var readyToUnstake []*moduleTypes.UnstakingActor
		var poolName string
		var er error

		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			readyToUnstake, er = u.store.GetAppsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_APP_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			readyToUnstake, er = u.store.GetFishermenReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_FISHERMAN_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			readyToUnstake, er = u.store.GetServicersReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_SERVICER_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			readyToUnstake, er = u.store.GetValidatorsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED:
			continue
		}
		if er != nil {
			return typesUtil.ErrGetReadyToUnstake(er)
		}
		for _, actor := range readyToUnstake {
			if poolName == coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName() {
				fmt.Println("unstaking validator", actor.StakeAmount)
			}
			stakeAmount, er := utils.StringToBigInt(actor.StakeAmount)
			if er != nil {
				return typesUtil.ErrStringToBigInt(er)
			}
			outputAddrBz, er := hex.DecodeString(actor.OutputAddress)
			if er != nil {
				return typesUtil.ErrHexDecodeFromString(er)
			}

			if err := u.subPoolAmount(poolName, stakeAmount); err != nil {
				return err
			}
			if err := u.addAccountAmount(outputAddrBz, stakeAmount); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *utilityContext) beginUnstakingMaxPausedActors() (err typesUtil.Error) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		if actorType == coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED {
			continue
		}
		maxPausedBlocks, err := u.getMaxAllowedPausedBlocks(actorType)
		if err != nil {
			return err
		}
		maxPauseHeight := u.height - int64(maxPausedBlocks)
		if maxPauseHeight < 0 { // genesis edge case
			maxPauseHeight = 0
		}
		if err := u.beginUnstakingActorsPausedBefore(maxPauseHeight, actorType); err != nil {
			return err
		}
	}
	return nil
}

func (u *utilityContext) beginUnstakingActorsPausedBefore(pausedBeforeHeight int64, actorType coreTypes.ActorType) (err typesUtil.Error) {
	unbondingHeight, err := u.getUnbondingHeight(actorType)
	if err != nil {
		return err
	}

	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = u.store.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = u.store.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		er = u.store.SetServicerStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = u.store.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	}
	if er != nil {
		return typesUtil.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *utilityContext) handleProposerRewards(proposer []byte) typesUtil.Error {
	feePoolName := coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName()
	feesAndRewardsCollected, err := u.getPoolAmount(feePoolName)
	if err != nil {
		return err
	}
	if err := u.setPoolAmount(feePoolName, big.NewInt(0)); err != nil {
		return err
	}
	proposerCutPercentage, err := u.getProposerPercentageOfFees()
	if err != nil {
		return err
	}
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 || daoCutPercentage > 100 {
		return typesUtil.ErrInvalidProposerCutPercentage()
	}
	amountToProposerFloat := new(big.Float).SetInt(feesAndRewardsCollected)
	amountToProposerFloat.Mul(amountToProposerFloat, big.NewFloat(float64(proposerCutPercentage)))
	amountToProposerFloat.Quo(amountToProposerFloat, big.NewFloat(100))
	amountToProposer, _ := amountToProposerFloat.Int(nil)
	amountToDAO := feesAndRewardsCollected.Sub(feesAndRewardsCollected, amountToProposer)
	if err := u.addAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_DAO.FriendlyName(), amountToDAO); err != nil {
		return err
	}
	return nil
}

// TODO: Need to design & document this business logic.
func (u *utilityContext) getLastBlockByzantineValidators() ([][]byte, error) {
	return nil, nil
}
