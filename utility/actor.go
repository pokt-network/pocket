package utility

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

func (u *utilityContext) setActorStakeAmount(actorType coreTypes.ActorType, addr []byte, amount *big.Int) typesUtil.Error {
	amountStr := utils.BigIntToString(amount)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.store.SetAppStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.store.SetFishermanStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.store.SetServicerStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.store.SetValidatorStakeAmount(addr, amountStr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetValidatorStakedAmount(err)
	}
	return nil
}

func (u *utilityContext) setActorUnstakingHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	unstakingStatus := int32(coreTypes.StakeStatus_Unstaking)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.store.SetAppUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.store.SetFishermanUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.store.SetServicerUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.store.SetValidatorUnstakingHeightAndStatus(addr, height, unstakingStatus)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(err)
	}
	return nil
}

func (u *utilityContext) setActorPausedHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = u.store.SetAppPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = u.store.SetFishermanPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		err = u.store.SetServicerPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = u.store.SetValidatorPauseHeight(addr, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}
	return nil
}

// Actor getters

func (u *utilityContext) getActorStakeAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakeAmount, err = u.store.GetAppStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakeAmount, err = u.store.GetFishermanStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		stakeAmount, err = u.store.GetServicerStakeAmount(u.height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakeAmount, err = u.store.GetValidatorStakeAmount(u.height, addr)
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

func (u *utilityContext) getMaxAllowedPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

	maxPausedBlocks, err := u.store.GetIntParam(paramName, u.height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}

	return maxPausedBlocks, nil
}

func (u *utilityContext) getMinRequiredPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

	minPausedBlocks, er := u.store.GetIntParam(paramName, u.height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}
	return minPausedBlocks, nil
}

func (u *utilityContext) getPausedHeightIfExists(actorType coreTypes.ActorType, addr []byte) (int64, typesUtil.Error) {
	var pauseHeight int64
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		pauseHeight, err = u.store.GetAppPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		pauseHeight, err = u.store.GetFishermanPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		pauseHeight, err = u.store.GetServicerPauseHeightIfExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		pauseHeight, err = u.store.GetValidatorPauseHeightIfExists(addr, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(err)
	}

	return pauseHeight, nil
}

func (u *utilityContext) getActorStatus(actorType coreTypes.ActorType, addr []byte) (coreTypes.StakeStatus, typesUtil.Error) {
	var status int32
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		status, err = u.store.GetAppStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		status, err = u.store.GetFishermanStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		status, err = u.store.GetServicerStatus(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		status, err = u.store.GetValidatorStatus(addr, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(err)
	}

	if _, ok := coreTypes.StakeStatus_name[status]; !ok {
		return typesUtil.ZeroInt, typesUtil.ErrUnknownStatus(status)
	}

	return coreTypes.StakeStatus(status), nil
}

func (u *utilityContext) getMinRequiredStakeAmount(actorType coreTypes.ActorType) (*big.Int, typesUtil.Error) {
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

	minStake, er := u.store.GetStringParam(paramName, u.height)
	if er != nil {
		return nil, typesUtil.ErrGetParam(paramName, er)
	}

	amount, err := utils.StringToBigInt(minStake)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) getUnbondingHeight(actorType coreTypes.ActorType) (int64, typesUtil.Error) {
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

	unstakingBlocksPeriod, err := u.store.GetIntParam(paramName, u.height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}

	return u.height + int64(unstakingBlocksPeriod), nil
}

func (u *utilityContext) getMaxAllowedChains(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

	maxChains, err := u.store.GetIntParam(paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return maxChains, nil
}

func (u *utilityContext) getActorExists(actorType coreTypes.ActorType, addr []byte) (bool, typesUtil.Error) {
	var exists bool
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		exists, err = u.store.GetAppExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		exists, err = u.store.GetFishermanExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		exists, err = u.store.GetServicerExists(addr, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		exists, err = u.store.GetValidatorExists(addr, u.height)
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
func (u *utilityContext) getActorOutputAddress(actorType coreTypes.ActorType, operator []byte) ([]byte, typesUtil.Error) {
	var outputAddr []byte
	var err error

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		outputAddr, err = u.store.GetAppOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		outputAddr, err = u.store.GetFishermanOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		outputAddr, err = u.store.GetServicerOutputAddress(operator, u.height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		outputAddr, err = u.store.GetValidatorOutputAddress(operator, u.height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, err)

	}
	return outputAddr, nil
}
