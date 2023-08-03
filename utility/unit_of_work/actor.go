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

func (u *baseUtilityUnitOfWork) setActorStakeAmount(actorType coreTypes.ActorType, addr []byte, amount *big.Int) coreTypes.Error {
	amountStr := utils.BigIntToString(amount)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		err = u.persistenceRWContext.SetWatcherStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorStakeAmount(addr, amountStr)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return coreTypes.ErrSetValidatorStakedAmount(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setActorUnbondingHeight(actorType coreTypes.ActorType, addr []byte, height int64) coreTypes.Error {
	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		err = u.persistenceRWContext.SetWatcherUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorUnstakingHeightAndStatus(addr, height, int32(coreTypes.StakeStatus_Unstaking))
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return coreTypes.ErrSetUnstakingHeightAndStatus(err)
	}
	return nil
}

func (u *baseUtilityUnitOfWork) setActorPausedHeight(actorType coreTypes.ActorType, addr []byte, height int64) coreTypes.Error {
	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.persistenceRWContext.SetAppPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		err = u.persistenceRWContext.SetWatcherPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.persistenceRWContext.SetServicerPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.persistenceRWContext.SetValidatorPauseHeight(addr, height)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return coreTypes.ErrSetPauseHeight(err)
	}
	return nil
}

// Actor getters

func (u *baseUtilityUnitOfWork) getActorStakeAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, coreTypes.Error) {
	var stakeAmount string
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakeAmount, err = u.persistenceReadContext.GetAppStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		stakeAmount, err = u.persistenceReadContext.GetWatcherStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		stakeAmount, err = u.persistenceReadContext.GetServicerStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakeAmount, err = u.persistenceReadContext.GetValidatorStakeAmount(u.height, addr)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, coreTypes.ErrGetStakeAmount(err)
	}

	amount, err := utils.StringToBigInt(stakeAmount)
	if err != nil {
		return nil, coreTypes.ErrStringToBigInt(err)
	}

	return amount, nil
}

func (u *baseUtilityUnitOfWork) getMaxAllowedPausedBlocks(actorType coreTypes.ActorType) (int, coreTypes.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		paramName = typesUtil.WatcherMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMaxPausedBlocksParamName
	default:
		return 0, coreTypes.ErrUnknownActorType(actorType.String())
	}

	maxPausedBlocks, err := u.getIntParam(paramName)
	if err != nil {
		return 0, coreTypes.ErrGetParam(paramName, err)
	}

	return maxPausedBlocks, nil
}

func (u *baseUtilityUnitOfWork) getMinRequiredPausedBlocks(actorType coreTypes.ActorType) (int, coreTypes.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		paramName = typesUtil.WatcherMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumPauseBlocksParamName
	default:
		return 0, coreTypes.ErrUnknownActorType(actorType.String())
	}

	minPausedBlocks, er := u.getIntParam(paramName)
	if er != nil {
		return 0, coreTypes.ErrGetParam(paramName, er)
	}
	return minPausedBlocks, nil
}

func (u *baseUtilityUnitOfWork) getPausedHeightIfExists(actorType coreTypes.ActorType, addr []byte) (int64, coreTypes.Error) {
	var pauseHeight int64
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		pauseHeight, err = u.persistenceReadContext.GetAppPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		pauseHeight, err = u.persistenceReadContext.GetWatcherPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		pauseHeight, err = u.persistenceReadContext.GetServicerPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		pauseHeight, err = u.persistenceReadContext.GetValidatorPauseHeightIfExists(addr, u.height)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return 0, coreTypes.ErrGetPauseHeight(err)
	}

	return pauseHeight, nil
}

func (u *baseUtilityUnitOfWork) getActorStatus(actorType coreTypes.ActorType, addr []byte) (coreTypes.StakeStatus, coreTypes.Error) {
	var status int32
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		status, err = u.persistenceReadContext.GetAppStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		status, err = u.persistenceReadContext.GetWatcherStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		status, err = u.persistenceReadContext.GetServicerStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		status, err = u.persistenceReadContext.GetValidatorStatus(addr, u.height)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return coreTypes.StakeStatus_UnknownStatus, coreTypes.ErrGetStatus(err)
	}

	if _, ok := coreTypes.StakeStatus_name[status]; !ok {
		return coreTypes.StakeStatus_UnknownStatus, coreTypes.ErrUnknownStatus(status)
	}

	return coreTypes.StakeStatus(status), nil
}

func (u *baseUtilityUnitOfWork) getMinRequiredStakeAmount(actorType coreTypes.ActorType) (*big.Int, coreTypes.Error) {
	var paramName string

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		paramName = typesUtil.WatcherMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumStakeParamName
	default:
		return nil, coreTypes.ErrUnknownActorType(actorType.String())
	}

	minStake, er := u.getStringParam(paramName)
	if er != nil {
		return nil, coreTypes.ErrGetParam(paramName, er)
	}

	amount, err := utils.StringToBigInt(minStake)
	if err != nil {
		return nil, coreTypes.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *baseUtilityUnitOfWork) getUnbondingHeight(actorType coreTypes.ActorType) (int64, coreTypes.Error) {
	var paramName string

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		paramName = typesUtil.WatcherUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorUnstakingBlocksParamName
	default:
		return 0, coreTypes.ErrUnknownActorType(actorType.String())
	}

	unstakingBlocksPeriod, err := u.getIntParam(paramName)
	if err != nil {
		return 0, coreTypes.ErrGetParam(paramName, err)
	}

	return u.height + int64(unstakingBlocksPeriod), nil
}

func (u *baseUtilityUnitOfWork) getMaxAllowedChains(actorType coreTypes.ActorType) (int, coreTypes.Error) {
	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		paramName = typesUtil.WatcherMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		paramName = typesUtil.ServicerMaxChainsParamName
	default:
		return 0, coreTypes.ErrUnknownActorType(actorType.String())
	}

	maxChains, err := u.getIntParam(paramName)
	if err != nil {
		return 0, coreTypes.ErrGetParam(paramName, err)
	}

	return maxChains, nil
}

func (u *baseUtilityUnitOfWork) getActorExists(actorType coreTypes.ActorType, addr []byte) (bool, coreTypes.Error) {
	var exists bool
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		exists, err = u.persistenceReadContext.GetAppExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		exists, err = u.persistenceReadContext.GetWatcherExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		exists, err = u.persistenceReadContext.GetServicerExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		exists, err = u.persistenceReadContext.GetValidatorExists(addr, u.height)
	default:
		return false, coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return false, coreTypes.ErrGetExists(err)
	}

	return exists, nil
}

// IMPROVE: Need to re-evaluate the design of `Output Address` to support things like "rev-share"
// and multiple output addresses.
func (u *baseUtilityUnitOfWork) getActorOutputAddress(actorType coreTypes.ActorType, operator []byte) ([]byte, coreTypes.Error) {
	var outputAddr []byte
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		outputAddr, err = u.persistenceReadContext.GetAppOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_WATCHER:
		outputAddr, err = u.persistenceReadContext.GetWatcherOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		outputAddr, err = u.persistenceReadContext.GetServicerOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		outputAddr, err = u.persistenceReadContext.GetValidatorOutputAddress(operator, u.height)
	default:
		err = coreTypes.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, coreTypes.ErrGetOutputAddress(operator, err)

	}
	return outputAddr, nil
}
