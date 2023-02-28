package utility

// Internal business logic containing the lifecycle of Block-related operations

import (
	"encoding/hex"
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// CreateAndApplyProposalBlock implements the exposed functionality of the shared UtilityContext interface.
func (u *utilityContext) CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (stateHash string, txs [][]byte, err error) {
	prevBlockByzantineVals, err := u.prevBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}

	// begin block lifecycle phase
	if err := u.beginBlock(prevBlockByzantineVals); err != nil {
		return "", nil, err
	}
	txs = make([][]byte, 0)
	txsTotalBz := 0
	txIdx := 0

	mempool := u.GetBus().GetUtilityModule().GetMempool()
	for !mempool.IsEmpty() {
		// NB: In order for transactions to have entered the mempool, `HandleTransaction` must have
		// been called which handles basic checks & validation.
		txBz, err := mempool.PopTx()
		if err != nil {
			return "", nil, err
		}

		tx, err := coreTypes.TxFromBytes(txBz)
		if err != nil {
			return "", nil, err
		}

		txBzSize := len(txBz)
		txsTotalBz += txBzSize

		// Exceeding maximum transaction bytes to be added in this block
		if txsTotalBz >= maxTransactionBytes {
			// Add back popped tx to be applied in a future block
			if err := mempool.AddTx(txBz); err != nil {
				return "", nil, err
			}
			break // we've reached our max
		}

		txResult, err := u.hydrateTxResult(tx, txIdx)
		if err != nil {
			u.logger.Err(err).Msg("Error in ApplyTransaction")
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := u.revertLastSavePoint(); err != nil {
				return "", nil, err
			}
			txsTotalBz -= txBzSize
			continue
		}

		// Index the transaction
		if err := u.store.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): The transaction can by hydrated but not indexed. Crash the process for now: %v\n", err)
		}

		txs = append(txs, txBz)
		txIdx++
	}

	if err := u.endBlock(proposer); err != nil {
		return "", nil, err
	}

	// Compute & return the new state hash
	stateHash, err = u.store.ComputeStateHash()
	if err != nil {
		u.logger.Fatal().Err(err).Msg("Updating the app hash failed. TODO: Look into roll-backing the entire commit...")
	}
	u.logger.Info().Str("state_hash", stateHash).Msgf("CreateAndApplyProposalBlock finished successfully")

	return stateHash, txs, err
}

// CLEANUP: code re-use ApplyBlock() for CreateAndApplyBlock()
func (u *utilityContext) ApplyBlock() (string, error) {
	lastByzantineValidators, err := u.prevBlockByzantineValidators()
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
		txResult, err := u.hydrateTxResult(tx, index)
		if err != nil {
			return "", err
		}

		if err := u.store.IndexTransaction(txResult); err != nil {
			u.logger.Fatal().Err(err).Msgf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}

		// TODO: if found, remove transaction from mempool.
		// DISCUSS: What if the context is rolled back or cancelled. Do we add it back to the mempool?
		// if err := mempool.RemoveTx(tx.Bytes()); err != nil {
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
	// INCOMPLETE: Identify what else needs to be done in the begin block lifecycle phase
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

	// INCOMPLETE: Identify what else needs to be done in the begin block lifecycle phase
	return nil
}

func (u *utilityContext) handleProposerRewards(proposer []byte) typesUtil.Error {
	feePoolName := coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName()
	feesAndRewardsCollected, err := u.getPoolAmount(feePoolName)
	if err != nil {
		return err
	}

	// Nullify the rewards pool
	if err := u.setPoolAmount(feePoolName, big.NewInt(0)); err != nil {
		return err
	}

	//
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

func (u *utilityContext) unbondUnstakingActors() (err typesUtil.Error) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		var readyToUnbond []*moduleTypes.UnstakingActor
		var poolName string

		var er error
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			readyToUnbond, er = u.store.GetAppsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_APP_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			readyToUnbond, er = u.store.GetFishermenReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_FISHERMAN_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			readyToUnbond, er = u.store.GetServicersReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_SERVICER_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			readyToUnbond, er = u.store.GetValidatorsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED:
			continue
		}
		if er != nil {
			return typesUtil.ErrGetReadyToUnstake(er)
		}

		// Loop through all unstaking actors and unbond those that have reached the waiting period,
		// move their stake from the pool back to the corresponding account.
		for _, actor := range readyToUnbond {
			stakeAmount, err := utils.StringToBigInt(actor.StakeAmount)
			if err != nil {
				return typesUtil.ErrStringToBigInt(err)
			}

			outputAddrBz, err := hex.DecodeString(actor.OutputAddress)
			if err != nil {
				return typesUtil.ErrHexDecodeFromString(err)
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

// TODO: Need to design & document this business logic.
func (u *utilityContext) prevBlockByzantineValidators() ([][]byte, error) {
	return nil, nil
}
