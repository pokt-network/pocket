package unit_of_work

// Internal business logic containing the lifecycle of Block-related operations

import (
	"encoding/hex"
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	moduleTypes "github.com/pokt-network/pocket/shared/modules/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func (u *baseUtilityUnitOfWork) beginBlock(previousBlockByzantineValidators [][]byte) typesUtil.Error {
	if err := u.handleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	// INCOMPLETE: Identify what else needs to be done in the begin block lifecycle phase
	return nil
}

func (u *baseUtilityUnitOfWork) endBlock(proposer []byte) typesUtil.Error {
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

func (u *baseUtilityUnitOfWork) handleProposerRewards(proposer []byte) typesUtil.Error {
	feePoolAddress := coreTypes.Pools_POOLS_FEE_COLLECTOR.Address()
	feesAndRewardsCollected, err := u.getPoolAmount(feePoolAddress)
	if err != nil {
		return err
	}

	// Nullify the rewards pool
	if err := u.setPoolAmount(feePoolAddress, big.NewInt(0)); err != nil {
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
	if err := u.addPoolAmount(coreTypes.Pools_POOLS_DAO.Address(), amountToDAO); err != nil {
		return err
	}
	return nil
}

func (u *baseUtilityUnitOfWork) unbondUnstakingActors() (err typesUtil.Error) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		var readyToUnbond []*moduleTypes.UnstakingActor
		var poolAddress []byte

		var er error
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			readyToUnbond, er = u.persistenceReadContext.GetAppsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_APP_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			readyToUnbond, er = u.persistenceReadContext.GetFishermenReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_FISHERMAN_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			readyToUnbond, er = u.persistenceReadContext.GetServicersReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_SERVICER_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			readyToUnbond, er = u.persistenceReadContext.GetValidatorsReadyToUnstake(u.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_VALIDATOR_STAKE.Address()
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

			if err := u.subPoolAmount(poolAddress, stakeAmount); err != nil {
				return err
			}
			if err := u.addAccountAmount(outputAddrBz, stakeAmount); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *baseUtilityUnitOfWork) beginUnstakingMaxPausedActors() (err typesUtil.Error) {
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

func (u *baseUtilityUnitOfWork) beginUnstakingActorsPausedBefore(pausedBeforeHeight int64, actorType coreTypes.ActorType) (err typesUtil.Error) {
	unbondingHeight, err := u.getUnbondingHeight(actorType)
	if err != nil {
		return err
	}

	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = u.persistenceRWContext.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = u.persistenceRWContext.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		er = u.persistenceRWContext.SetServicerStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = u.persistenceRWContext.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	}
	if er != nil {
		return typesUtil.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

// TODO: Need to design & document this business logic.
func (u *baseUtilityUnitOfWork) prevBlockByzantineValidators() ([][]byte, error) {
	return nil, nil
}

// TODO: This has not been tested or investigated in detail
func (u *baseUtilityUnitOfWork) revertLastSavePoint() typesUtil.Error {
	// TODO(@deblasis): Implement this
	// if len(u.savePointsSet) == 0 {
	// 	return typesUtil.ErrEmptySavePoints()
	// }
	// var key []byte
	// popIndex := len(u.savePointsList) - 1
	// key, u.savePointsList = u.savePointsList[popIndex], u.savePointsList[:popIndex]
	// delete(u.savePointsSet, hex.EncodeToString(key))
	// if err := u.store.RollbackToSavePoint(key); err != nil {
	// 	return typesUtil.ErrRollbackSavePoint(err)
	// }
	return nil
}

//nolint:unused // TODO: This has not been tested or investigated in detail
func (u *baseUtilityUnitOfWork) newSavePoint(txHashBz []byte) typesUtil.Error {
	// TODO(@deblasis): Implement this
	// if err := u.store.NewSavePoint(txHashBz); err != nil {
	// 	return typesUtil.ErrNewSavePoint(err)
	// }
	// txHash := hex.EncodeToString(txHashBz)
	// if _, exists := u.savePointsSet[txHash]; exists {
	// 	return typesUtil.ErrDuplicateSavePoint()
	// }
	// u.savePointsList = append(u.savePointsList, txHashBz)
	// u.savePointsSet[txHash] = struct{}{}
	return nil
}
