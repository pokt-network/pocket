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

const (
	IgnoreProposalBlockCheckHash = "0100000000000000000000000000000000000000000000000000000000000010"
)

func (uow *baseUtilityUnitOfWork) beginBlock() typesUtil.Error {
	log := uow.logger.With().Fields(map[string]interface{}{
		"source": "beginBlock",
	}).Logger()

	log.Debug().Bool("TODO", true).Msg("determining prevBlockByzantineValidators")
	previousBlockByzantineValidators, err := uow.prevBlockByzantineValidators()
	if err != nil {
		return typesUtil.ErrGetPrevBlockByzantineValidators(err)
	}

	log.Info().Msg("handling byzantine validators")
	if err := uow.handleByzantineValidators(previousBlockByzantineValidators); err != nil {
		return err
	}
	// INCOMPLETE: Identify what else needs to be done in the begin block lifecycle phase
	return nil
}

func (uow *baseUtilityUnitOfWork) endBlock(proposer []byte) typesUtil.Error {
	log := uow.logger.With().Fields(map[string]interface{}{
		"proposer": hex.EncodeToString(proposer),
		"source":   "endBlock",
	}).Logger()

	log.Info().Msg("handling proposer rewards")
	// reward the block proposer
	if err := uow.handleProposerRewards(proposer); err != nil {
		return err
	}

	log.Info().Msg("handling unstaking actors")
	// unstake actors that have been 'unstaking' for the <Actor>UnstakingBlocks
	if err := uow.unbondUnstakingActors(); err != nil {
		return err
	}

	log.Info().Msg("handling unstaking paused actors")
	// begin unstaking the actors who have been paused for MaxPauseBlocks
	if err := uow.beginUnstakingMaxPausedActors(); err != nil {
		return err
	}

	// INCOMPLETE: Identify what else needs to be done in the begin block lifecycle phase
	return nil
}

func (uow *baseUtilityUnitOfWork) handleProposerRewards(proposer []byte) typesUtil.Error {
	feePoolAddress := coreTypes.Pools_POOLS_FEE_COLLECTOR.Address()
	feesAndRewardsCollected, err := uow.getPoolAmount(feePoolAddress)
	if err != nil {
		return err
	}

	// Nullify the rewards pool
	if err := uow.setPoolAmount(feePoolAddress, big.NewInt(0)); err != nil {
		return err
	}

	proposerCutPercentage, err := getGovParam[int](uow, typesUtil.ProposerPercentageOfFeesParamName)
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
	if err := uow.addAccountAmount(proposer, amountToProposer); err != nil {
		return err
	}
	if err := uow.addPoolAmount(coreTypes.Pools_POOLS_DAO.Address(), amountToDAO); err != nil {
		return err
	}
	return nil
}

func (uow *baseUtilityUnitOfWork) unbondUnstakingActors() (err typesUtil.Error) {
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
			readyToUnbond, er = uow.persistenceReadContext.GetAppsReadyToUnstake(uow.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_APP_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			readyToUnbond, er = uow.persistenceReadContext.GetFishermenReadyToUnstake(uow.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_FISHERMAN_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			readyToUnbond, er = uow.persistenceReadContext.GetServicersReadyToUnstake(uow.height, int32(coreTypes.StakeStatus_Unstaking))
			poolAddress = coreTypes.Pools_POOLS_SERVICER_STAKE.Address()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			readyToUnbond, er = uow.persistenceReadContext.GetValidatorsReadyToUnstake(uow.height, int32(coreTypes.StakeStatus_Unstaking))
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

			if err := uow.subPoolAmount(poolAddress, stakeAmount); err != nil {
				return err
			}
			if err := uow.addAccountAmount(outputAddrBz, stakeAmount); err != nil {
				return err
			}
		}
	}

	return nil
}

func (uow *baseUtilityUnitOfWork) beginUnstakingMaxPausedActors() (err typesUtil.Error) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		if actorType == coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED {
			continue
		}
		maxPausedBlocks, err := uow.getMaxAllowedPausedBlocks(actorType)
		if err != nil {
			return err
		}
		maxPauseHeight := uow.height - int64(maxPausedBlocks)
		if maxPauseHeight < 0 { // genesis edge case
			maxPauseHeight = 0
		}
		if err := uow.beginUnstakingActorsPausedBefore(maxPauseHeight, actorType); err != nil {
			return err
		}
	}
	return nil
}

func (uow *baseUtilityUnitOfWork) beginUnstakingActorsPausedBefore(pausedBeforeHeight int64, actorType coreTypes.ActorType) (err typesUtil.Error) {
	unbondingHeight, err := uow.getUnbondingHeight(actorType)
	if err != nil {
		return err
	}

	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = uow.persistenceRWContext.SetAppStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = uow.persistenceRWContext.SetFishermanStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		er = uow.persistenceRWContext.SetServicerStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = uow.persistenceRWContext.SetValidatorsStatusAndUnstakingHeightIfPausedBefore(pausedBeforeHeight, unbondingHeight, int32(coreTypes.StakeStatus_Unstaking))
	}
	if er != nil {
		return typesUtil.ErrSetStatusPausedBefore(er, pausedBeforeHeight)
	}
	return nil
}

// TODO: Need to design & document this business logic.
func (uow *baseUtilityUnitOfWork) prevBlockByzantineValidators() ([][]byte, error) {
	return nil, nil
}

// TODO: This has not been tested or investigated in detail
func (uow *baseUtilityUnitOfWork) revertLastSavePoint() typesUtil.Error {
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
func (uow *baseUtilityUnitOfWork) newSavePoint(txHashBz []byte) typesUtil.Error {
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
