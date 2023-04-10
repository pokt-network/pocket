package unit_of_work

// Internal business logic for functionality shared across all `Actors`.
//
// An Actor is any protocol level actor that likely has something-at-stake and interacts with the
// protocol through some sort of on-chain state transitions.

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

// Actor setters

func (u *baseUtilityUnitOfWork) setActorStakeAmount(actorType coreTypes.ActorType, addr []byte, amount *big.Int) typesUtil.Error {
	amountStr := utils.BigIntToString(amount)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.persistenceRWContext.SetFishermanStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorStakeAmount(addr, amountStr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetValidatorStakedAmount(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setActorUnbondingHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.persistenceRWContext.SetFishermanUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setActorPausedHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.persistenceRWContext.SetFishermanPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorPauseHeight(addr, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}
	return nil
}

// Actor getters

func (u *baseUtilityUnitOfWork) getActorStakeAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakeAmount, err = u.persistenceReadContext.GetAppStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakeAmount, err = u.persistenceReadContext.GetFishermanStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		stakeAmount, err = u.persistenceReadContext.GetServicerStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakeAmount, err = u.persistenceReadContext.GetValidatorStakeAmount(u.height, addr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	amount, err := utils.StringToBigInt(stakeAmount)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}

	return amount, nil
}

func (u *baseUtilityUnitOfWork) getMaxAllowedPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMaxPausedBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	maxPausedBlocks, err := u.getIntParam(paramName)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return maxPausedBlocks, nil
}

func (u *baseUtilityUnitOfWork) getMinRequiredPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumPauseBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minPausedBlocks, er := u.getIntParam(paramName)
	if er != nil {
		return 0, typesUtil.ErrGetParam(paramName, er)
	}
	return minPausedBlocks, nil
}

func (u *baseUtilityUnitOfWork) getPausedHeightIfExists(actorType coreTypes.ActorType, addr []byte) (int64, typesUtil.Error) {
	var pauseHeight int64
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		pauseHeight, err = u.persistenceReadContext.GetAppPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		pauseHeight, err = u.persistenceReadContext.GetFishermanPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		pauseHeight, err = u.persistenceReadContext.GetServicerPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		pauseHeight, err = u.persistenceReadContext.GetValidatorPauseHeightIfExists(addr, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return 0, typesUtil.ErrGetPauseHeight(err)
	}

	return pauseHeight, nil
}

func (u *baseUtilityUnitOfWork) getActorStatus(actorType coreTypes.ActorType, addr []byte) (coreTypes.StakeStatus, typesUtil.Error) {
	var status int32
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		status, err = u.persistenceReadContext.GetAppStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		status, err = u.persistenceReadContext.GetFishermanStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		status, err = u.persistenceReadContext.GetServicerStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		status, err = u.persistenceReadContext.GetValidatorStatus(addr, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return coreTypes.StakeStatus_UnknownStatus, typesUtil.ErrGetStatus(err)
	}

	if _, ok := coreTypes.StakeStatus_name[status]; !ok {
		return coreTypes.StakeStatus_UnknownStatus, typesUtil.ErrUnknownStatus(status)
	}

	return coreTypes.StakeStatus(status), nil
}

func (u *baseUtilityUnitOfWork) getMinRequiredStakeAmount(actorType coreTypes.ActorType) (*big.Int, typesUtil.Error) {
	var paramName string

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumStakeParamName
	default:
		return nil, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minStake, er := u.getStringParam(paramName)
	if er != nil {
		return nil, typesUtil.ErrGetParam(paramName, er)
	}

	amount, err := utils.StringToBigInt(minStake)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *baseUtilityUnitOfWork) getUnbondingHeight(actorType coreTypes.ActorType) (int64, typesUtil.Error) {
	var paramName string

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorUnstakingBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	unstakingBlocksPeriod, err := u.getIntParam(paramName)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return u.height + int64(unstakingBlocksPeriod), nil
}

func (u *baseUtilityUnitOfWork) getMaxAllowedChains(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMaxChainsParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	maxChains, err := u.getIntParam(paramName)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return maxChains, nil
}

func (u *baseUtilityUnitOfWork) getActorExists(actorType coreTypes.ActorType, addr []byte) (bool, typesUtil.Error) {
	var exists bool
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		exists, err = u.persistenceReadContext.GetAppExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		exists, err = u.persistenceReadContext.GetFishermanExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		exists, err = u.persistenceReadContext.GetServicerExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		exists, err = u.persistenceReadContext.GetValidatorExists(addr, u.height)
	default:
		return false, typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return false, typesUtil.ErrGetExists(err)
	}

	return exists, nil
}

// IMPROVE: Need to re-evaluate the design of `Output Address` to support things like "rev-share"
// and multiple output addresses.
func (u *baseUtilityUnitOfWork) getActorOutputAddress(actorType coreTypes.ActorType, operator []byte) ([]byte, typesUtil.Error) {
	var outputAddr []byte
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		outputAddr, err = u.persistenceReadContext.GetAppOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		outputAddr, err = u.persistenceReadContext.GetFishermanOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		outputAddr, err = u.persistenceReadContext.GetServicerOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		outputAddr, err = u.persistenceReadContext.GetValidatorOutputAddress(operator, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, err)

	}
	return outputAddr, nil
}
