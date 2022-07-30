package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

/*
This 'block' file contains all the lifecycle block operations.
The ApplyBlock function is the 'main' operation that executes a 'block' object against the state
Pocket Network adpots a Tendermint-like lifecycle of BeginBlock -> DeliverTx -> EndBlock in that order
Like the name suggests, BeginBlock is an autonomous state operation that executes at the beginning of every block
DeliverTx individually applys each transaction against the state and rolls it back (not yet implemented) if fails.
like BeginBlock, EndBlock is an autonomous state oepration that executes at the end of every block.
*/

func (u *UtilityContext) ApplyBlock(latestHeight int64, proposerAddress []byte, transactions [][]byte, lastBlockByzantineValidators [][]byte) ([]byte, error) {
	u.LatestHeight = latestHeight
	// begin block lifecycle phase
	if err := u.BeginBlock(lastBlockByzantineValidators); err != nil {
		return nil, err
	}
	// deliver txs lifecycle phase
	for _, transaction := range transactions {
		tx, err := typesUtil.TransactionFromBytes(transaction)
		if err != nil {
			return nil, err
		}
		if err := tx.ValidateBasic(); err != nil {
			return nil, err
		}
		if err := u.ApplyTransaction(tx); err != nil {
			return nil, err
		}
		// TODO: if found, remove transaction from mempool
		// if err := u.Mempool.DeleteTransaction(tx); err != nil {
		// 	return nil, err
		// }
	}
	// end block lifecycle phase
	if err := u.EndBlock(proposerAddress); err != nil {
		return nil, err
	}
	// return the app hash (consensus module will get the validator set directly
	return u.GetAppHash()
}

func (u *UtilityContext) BeginBlock(previousBlockByzantineValidators [][]byte) types.Error {
	if err := u.HandleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	return nil
}

func (u *UtilityContext) EndBlock(proposer []byte) types.Error {
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

func (u *UtilityContext) GetAppHash() ([]byte, types.Error) {
	// Get the root hash of the merkle state tree for state consensus integrity
	appHash, er := u.Context.AppHash()
	if er != nil {
		return nil, types.ErrAppHash(er)
	}
	return appHash, nil
}

// HandleByzantineValidators handles the validators who either didn't sign at all or disagreed with the 2/3+ majority
func (u *UtilityContext) HandleByzantineValidators(lastBlockByzantineValidators [][]byte) types.Error {
	latestBlockHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
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
			if err = u.PauseValidatorAndSetMissedBlocks(address, latestBlockHeight, typesUtil.HeightNotUsed); err != nil {
				return err
			}
			// burn validator for missing blocks
			burnPercentage, err := u.GetMissedBlocksBurnPercentage()
			if err != nil {
				return err
			}
			if err = u.BurnActor(typesUtil.ActorType_Val, burnPercentage, address); err != nil {
				return err
			}
		} else if err := u.SetValidatorMissedBlocks(address, numberOfMissedBlocks); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) UnstakeActorsThatAreReady() (err types.Error) {
	var er error
	store := u.Store()
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	for _, actorType := range typesUtil.ActorTypes {
		var readyToUnstake []*types.UnstakingActor
		poolName := actorType.GetActorPoolName()
		switch actorType {
		case typesUtil.ActorType_App:
			readyToUnstake, er = store.GetAppsReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.ActorType_Fish:
			readyToUnstake, er = store.GetFishermenReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.ActorType_Node:
			readyToUnstake, er = store.GetServiceNodesReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		case typesUtil.ActorType_Val:
			readyToUnstake, er = store.GetValidatorsReadyToUnstake(latestHeight, typesUtil.UnstakingStatus)
		}
		if er != nil {
			return types.ErrGetReadyToUnstake(er)
		}
		for _, actor := range readyToUnstake {
			if err = u.SubPoolAmount(poolName, actor.GetStakeAmount()); err != nil {
				return err
			}
			if err = u.AddAccountAmountString(actor.GetOutputAddress(), actor.GetStakeAmount()); err != nil {
				return err
			}
			if err = u.DeleteActor(actorType, actor.GetAddress()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *UtilityContext) BeginUnstakingMaxPaused() (err types.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return err
	}
	for _, actorType := range typesUtil.ActorTypes {
		maxPausedBlocks, err := u.GetMaxPausedBlocks(actorType)
		if err != nil {
			return err
		}
		beforeHeight := latestHeight - int64(maxPausedBlocks)
		// genesis edge case
		if beforeHeight < 0 {
			beforeHeight = 0
		}
		if err := u.UnstakeActorPausedBefore(beforeHeight, actorType); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) UnstakeActorPausedBefore(pausedBeforeHeight int64, actorType typesUtil.ActorType) (err types.Error) {
	var er error
	store := u.Store()
	unstakingHeight, err := u.GetUnstakingHeight(actorType)
	if err != nil {
		return err
	}
	switch actorType {
	case typesUtil.ActorType_App:
		er = store.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Fish:
		er = store.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Node:
		er = store.SetServiceNodeStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Val:
		er = store.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unstakingHeight, typesUtil.UnstakingStatus)
	}
	if er != nil {
		return types.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

func (u *UtilityContext) HandleProposalRewards(proposer []byte) types.Error {
	feesAndRewardsCollected, err := u.GetPoolAmount(typesGenesis.FeePoolName)
	if err != nil {
		return err
	}
	if err := u.SetPoolAmount(typesGenesis.FeePoolName, big.NewInt(0)); err != nil {
		return err
	}
	proposerCutPercentage, err := u.GetProposerPercentageOfFees()
	if err != nil {
		return err
	}
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 || daoCutPercentage > 100 {
		return types.ErrInvalidProposerCutPercentage()
	}
	amountToProposerFloat := new(big.Float).SetInt(feesAndRewardsCollected)
	amountToProposerFloat.Mul(amountToProposerFloat, big.NewFloat(float64(proposerCutPercentage)))
	amountToProposerFloat.Quo(amountToProposerFloat, big.NewFloat(100))
	amountToProposer, _ := amountToProposerFloat.Int(nil)
	amountToDAO := feesAndRewardsCollected.Sub(feesAndRewardsCollected, amountToProposer)
	if err := u.AddAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err := u.AddPoolAmount(typesGenesis.DAOPoolName, amountToDAO); err != nil {
		return err
	}
	return nil
}

// GetValidatorMissedBlocks gets the total blocks that a validator has not signed a certain window of time denominated by blocks
func (u *UtilityContext) GetValidatorMissedBlocks(address []byte) (int, types.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetMissedBlocks(er)
	}
	missedBlocks, er := store.GetValidatorMissedBlocks(address, height)
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetMissedBlocks(er)
	}
	return missedBlocks, nil
}

func (u *UtilityContext) PauseValidatorAndSetMissedBlocks(address []byte, pauseHeight int64, missedBlocks int) types.Error {
	store := u.Store()
	if err := store.SetValidatorPauseHeightAndMissedBlocks(address, pauseHeight, missedBlocks); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *UtilityContext) SetValidatorMissedBlocks(address []byte, missedBlocks int) types.Error {
	store := u.Store()
	er := store.SetValidatorMissedBlocks(address, missedBlocks)
	if er != nil {
		return types.ErrSetMissedBlocks(er)
	}
	return nil
}
