package utility

import (
	"encoding/hex"
	"log"
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

/*
	This 'block' file contains all the lifecycle block operations.

	The ApplyBlock function is the 'main' operation that executes a 'block' object against the state.

	Pocket Network adopt a Tendermint-like lifecycle of BeginBlock -> DeliverTx -> EndBlock in that
	order. Like the name suggests, BeginBlock is an autonomous state operation that executes at the
	beginning of every block DeliverTx individually applies each transaction against the state and
	rolls it back (not yet implemented) if fails. Like BeginBlock, EndBlock is an autonomous state
	operation that executes at the end of every block.
*/

// TODO: Make sure to call `utility.CheckTransaction`, which calls `persistence.TransactionExists`
func (u *utilityContext) CreateAndApplyProposalBlock(proposer []byte, maxTransactionBytes int) (string, [][]byte, error) {
	lastBlockByzantineVals, err := u.getLastBlockByzantineValidators()
	if err != nil {
		return "", nil, err
	}
	// begin block lifecycle phase
	if err := u.BeginBlock(lastBlockByzantineVals); err != nil {
		return "", nil, err
	}
	transactions := make([][]byte, 0)
	totalTxsSizeInBytes := 0
	txIndex := 0
	for !u.mempool.IsEmpty() {
		txBytes, err := u.mempool.PopTransaction()
		if err != nil {
			return "", nil, err
		}
		transaction, err := typesUtil.TransactionFromBytes(txBytes)
		if err != nil {
			return "", nil, err
		}
		txTxsSizeInBytes := len(txBytes)
		totalTxsSizeInBytes += txTxsSizeInBytes
		if totalTxsSizeInBytes >= maxTransactionBytes {
			// Add back popped transaction to be applied in a future block
			err := u.mempool.AddTransaction(txBytes)
			if err != nil {
				return "", nil, err
			}
			totalTxsSizeInBytes -= txTxsSizeInBytes
			break // we've reached our max
		}
		txResult, err := u.applyTransaction(txIndex, transaction)
		if err != nil {
			// TODO(#327): Properly implement 'unhappy path' for save points
			if err := u.RevertLastSavePoint(); err != nil {
				return "", nil, err
			}
			totalTxsSizeInBytes -= txTxsSizeInBytes
			continue
		}
		if err := u.persistenceContext.IndexTransaction(txResult); err != nil {
			log.Fatalf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}

		transactions = append(transactions, txBytes)
		txIndex++
	}

	if err := u.EndBlock(proposer); err != nil {
		return "", nil, err
	}
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err := u.persistenceContext.ComputeStateHash()
	if err != nil {
		log.Fatalf("Updating the app hash failed: %v. TODO: Look into roll-backing the entire commit...\n", err)
	}
	log.Println("CreateAndApplyProposalBlock - computed state hash:", stateHash)

	return stateHash, transactions, err
}

// TODO: Make sure to call `utility.CheckTransaction`, which calls `persistence.TransactionExists`
// CLEANUP: code re-use ApplyBlock() for CreateAndApplyBlock()
func (u *utilityContext) ApplyBlock() (string, error) {
	lastByzantineValidators, err := u.getLastBlockByzantineValidators()
	if err != nil {
		return "", err
	}

	// begin block lifecycle phase
	if err := u.BeginBlock(lastByzantineValidators); err != nil {
		return "", err
	}

	// deliver txs lifecycle phase
	for index, transactionProtoBytes := range u.proposalBlockTxs {
		tx, err := typesUtil.TransactionFromBytes(transactionProtoBytes)
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
		txResult, err := u.applyTransaction(index, tx)
		if err != nil {
			return "", err
		}

		if err := u.persistenceContext.IndexTransaction(txResult); err != nil {
			log.Fatalf("TODO(#327): We can apply the transaction but not index it. Crash the process for now: %v\n", err)
		}

		// TODO: if found, remove transaction from mempool.
		// DISCUSS: What if the context is rolled back or cancelled. Do we add it back to the mempool?
		// if err := u.Mempool.RemoveTransaction(transaction); err != nil {
		// 	return nil, err
		// }
	}

	// end block lifecycle phase
	if err := u.EndBlock(u.proposalProposerAddr); err != nil {
		return "", err
	}
	// return the app hash (consensus module will get the validator set directly)
	stateHash, err := u.persistenceContext.ComputeStateHash()
	if err != nil {
		log.Fatalf("Updating the app hash failed: %v. TODO: Look into roll-backing the entire commit...\n", err)
		return "", typesUtil.ErrAppHash(err)
	}
	log.Println("ApplyBlock - computed state hash:", stateHash)

	// return the app hash; consensus module will get the validator set directly
	return stateHash, nil
}

func (u *utilityContext) BeginBlock(previousBlockByzantineValidators [][]byte) typesUtil.Error {
	if err := u.HandleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	return nil
}

func (u *utilityContext) EndBlock(proposer []byte) typesUtil.Error {
	// reward the block proposer
	if err := u.HandleProposalRewards(proposer); err != nil {
		return err
	}
	// unstake actors that have been 'unstaking' for the <Actor>UnstakingBlocks
	if err := u.UnstakeActorsThatAreReady(); err != nil {
		return err
	}
	// begin unstaking the actors who have been paused for MaxPauseBlocks
	if err := u.BeginUnstakingMaxPaused(); err != nil {
		return err
	}
	return nil
}

// HandleByzantineValidators handles the validators who either didn't sign at all or disagreed with the 2/3+ majority
func (u *utilityContext) HandleByzantineValidators(lastBlockByzantineValidators [][]byte) typesUtil.Error {
	maxMissedBlocks, err := u.GetValidatorMaxMissedBlocks()
	if err != nil {
		return err
	}
	for _, address := range lastBlockByzantineValidators {
		numberOfMissedBlocks, err := u.GetValidatorMissedBlocks(address)
		if err != nil {
			return err
		}
		// increment missed blocks
		numberOfMissedBlocks++
		// handle if over the threshold
		if numberOfMissedBlocks >= maxMissedBlocks {
			// pause the validator and reset missed blocks
			if err = u.PauseValidatorAndSetMissedBlocks(address, u.height, int(typesUtil.HeightNotUsed)); err != nil {
				return err
			}
			// burn validator for missing blocks
			burnPercentage, err := u.GetMissedBlocksBurnPercentage()
			if err != nil {
				return err
			}
			if err = u.burnValidator(burnPercentage, address); err != nil {
				return err
			}
		} else if err := u.SetValidatorMissedBlocks(address, numberOfMissedBlocks); err != nil {
			return err
		}
	}
	return nil
}

func (u *utilityContext) UnstakeActorsThatAreReady() (err typesUtil.Error) {
	var er error
	store := u.Store()
	for _, actorTypeInt32 := range coreTypes.ActorType_value {
		var readyToUnstake []*moduleTypes.UnstakingActor
		actorType := coreTypes.ActorType(actorTypeInt32)
		var poolName string
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			readyToUnstake, er = store.GetAppsReadyToUnstake(u.height, int32(typesUtil.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_APP_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			readyToUnstake, er = store.GetFishermenReadyToUnstake(u.height, int32(typesUtil.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_FISHERMAN_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
			readyToUnstake, er = store.GetServiceNodesReadyToUnstake(u.height, int32(typesUtil.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_SERVICE_NODE_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			readyToUnstake, er = store.GetValidatorsReadyToUnstake(u.height, int32(typesUtil.StakeStatus_Unstaking))
			poolName = coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName()
		case coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED:
			continue
		}
		if er != nil {
			return typesUtil.ErrGetReadyToUnstake(er)
		}
		for _, actor := range readyToUnstake {
			if err = u.subPoolAmount(poolName, actor.GetStakeAmount()); err != nil {
				return err
			}
			outputAddrBz, er := hex.DecodeString(actor.OutputAddress)
			if er != nil {
				return typesUtil.ErrHexDecodeFromString(er)
			}
			if err = u.addAccountAmountString(outputAddrBz, actor.StakeAmount); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *utilityContext) BeginUnstakingMaxPaused() (err typesUtil.Error) {
	for _, actorTypeInt32 := range coreTypes.ActorType_value {
		actorType := coreTypes.ActorType(actorTypeInt32)
		if actorType == coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED {
			continue
		}
		maxPausedBlocks, err := u.getMaxAllowedPausedBlocks(actorType)
		if err != nil {
			return err
		}
		beforeHeight := u.height - int64(maxPausedBlocks)
		if beforeHeight < 0 { // genesis edge case
			beforeHeight = 0
		}
		if err := u.UnstakeActorPausedBefore(beforeHeight, actorType); err != nil {
			return err
		}
	}
	return nil
}

func (u *utilityContext) UnstakeActorPausedBefore(pausedBeforeHeight int64, actorType coreTypes.ActorType) (err typesUtil.Error) {
	var er error
	store := u.Store()
	unstakingHeight, err := u.getUnstakingHeight(actorType)
	if err != nil {
		return err
	}
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = store.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = store.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		er = store.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = store.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	}
	if er != nil {
		return typesUtil.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *utilityContext) HandleProposalRewards(proposer []byte) typesUtil.Error {
	feePoolName := coreTypes.Pools_POOLS_FEE_COLLECTOR.FriendlyName()
	feesAndRewardsCollected, err := u.getPoolAmount(feePoolName)
	if err != nil {
		return err
	}
	if err := u.setPoolAmount(feePoolName, big.NewInt(0)); err != nil {
		return err
	}
	proposerCutPercentage, err := u.GetProposerPercentageOfFees()
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
	if err = u.addAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err = u.addPoolAmount(coreTypes.Pools_POOLS_DAO.FriendlyName(), amountToDAO); err != nil {
		return err
	}
	return nil
}

// GetValidatorMissedBlocks gets the total blocks that a validator has not signed a certain window of time denominated by blocks
func (u *utilityContext) GetValidatorMissedBlocks(address []byte) (int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}
	missedBlocks, er := store.GetValidatorMissedBlocks(address, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetMissedBlocks(er)
	}
	return missedBlocks, nil
}

func (u *utilityContext) PauseValidatorAndSetMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) typesUtil.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeightAndMissedBlocks(address, pauseHeight, missedBlocks); err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *utilityContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) typesUtil.Error {
	store := u.Store()
	er := store.SetValidatorMissedBlocks(address, missedBlocks)
	if er != nil {
		return typesUtil.ErrSetMissedBlocks(er)
	}
	return nil
}

// TODO(#271): Need to retrieve byzantine validators from the persistence module
func (u *utilityContext) getLastBlockByzantineValidators() ([][]byte, error) {
	return nil, nil
}
