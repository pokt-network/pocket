package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

//	`Actor` is the consolidated term for common functionality among the following network actors: application, fisherman, servicer, validator, etc.

func (u *utilityContext) setActorStakeAmount(actorType coreTypes.ActorType, addr []byte, amount *big.Int) typesUtil.Error {
	store := u.Store()
	amountStr := converters.BigIntToString(amount)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodeStakeAmount(addr, amountStr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorStakeAmount(addr, amountStr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetValidatorStakedAmount(err)
	}
	return nil
}

func (u *utilityContext) setActorUnstakingHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	store := u.Store()
	unstakingStatus := int32(typesUtil.StakeStatus_Unstaking)

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodeUnstakingHeightAndStatus(addr, height, unstakingStatus)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorUnstakingHeightAndStatus(addr, height, unstakingStatus)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(err)
	}
	return nil
}

func (u *utilityContext) setActorPausedHeight(actorType coreTypes.ActorType, addr []byte, height int64) typesUtil.Error {
	store := u.Store()

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanPauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodePauseHeight(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorPauseHeight(addr, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}
	return nil
}

func (u *utilityContext) getActorStakeAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var stakeAmount string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakeAmount, err = store.GetAppStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakeAmount, err = store.GetFishermanStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		stakeAmount, err = store.GetServiceNodeStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakeAmount, err = store.GetValidatorStakeAmount(height, addr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	amount, err := converters.StringToBigInt(stakeAmount)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}

	return amount, nil
}

func (u *utilityContext) getMaxAllowedPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		paramName = typesUtil.ServiceNodeMaxPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMaxPausedBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	maxPausedBlocks, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}

	return maxPausedBlocks, nil
}

func (u *utilityContext) getMinRequiredPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		paramName = typesUtil.ServiceNodeMinimumPauseBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumPauseBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minPausedBlocks, er := store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}
	return minPausedBlocks, nil
}

func (u *utilityContext) getPausedHeightIfExists(actorType coreTypes.ActorType, addr []byte) (int64, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var pauseHeight int64
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		pauseHeight, err = store.GetAppPauseHeightIfExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		pauseHeight, err = store.GetFishermanPauseHeightIfExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		pauseHeight, err = store.GetServiceNodePauseHeightIfExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		pauseHeight, err = store.GetValidatorPauseHeightIfExists(addr, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(err)
	}

	return pauseHeight, nil
}

func (u *utilityContext) getActorStatus(actorType coreTypes.ActorType, addr []byte) (typesUtil.StakeStatus, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var status int32
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		status, err = store.GetAppStatus(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		status, err = store.GetFishermanStatus(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		status, err = store.GetServiceNodeStatus(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		status, err = store.GetValidatorStatus(addr, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(err)
	}

	if _, ok := typesUtil.StakeStatus_name[status]; !ok {
		return typesUtil.ZeroInt, typesUtil.ErrUnknownStatus(status)
	}

	return typesUtil.StakeStatus(status), nil
}

func (u *utilityContext) getMinRequiredStakeAmount(actorType coreTypes.ActorType) (*big.Int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		paramName = typesUtil.ServiceNodeMinimumStakeParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorMinimumStakeParamName
	default:
		return nil, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minStake, er := store.GetStringParam(paramName, height)
	if er != nil {
		return nil, typesUtil.ErrGetParam(paramName, er)
	}

	amount, err := converters.StringToBigInt(minStake)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) getUnbondingHeight(actorType coreTypes.ActorType) (int64, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	var unstakingBlocksPeriod int
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		paramName = typesUtil.ServiceNodeUnstakingBlocksParamName
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		paramName = typesUtil.ValidatorUnstakingBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	unstakingBlocksPeriod, err = store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}

	return u.height + int64(unstakingBlocksPeriod), nil
}

func (u *utilityContext) getMaxAllowedChains(actorType coreTypes.ActorType) (int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		paramName = typesUtil.AppMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		paramName = typesUtil.FishermanMaxChainsParamName
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		paramName = typesUtil.ServiceNodeMaxChainsParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	maxChains, err := store.GetIntParam(paramName, height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return maxChains, nil
}

func (u *utilityContext) getActorExists(actorType coreTypes.ActorType, addr []byte) (bool, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return false, typesUtil.ErrGetHeight(err)
	}

	var exists bool
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		exists, err = store.GetAppExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		exists, err = store.GetFishermanExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		exists, err = store.GetServiceNodeExists(addr, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		exists, err = store.GetValidatorExists(addr, height)
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
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var outputAddr []byte
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		outputAddr, err = store.GetAppOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		outputAddr, err = store.GetFishermanOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		outputAddr, err = store.GetServiceNodeOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		outputAddr, err = store.GetValidatorOutputAddress(operator, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, err)

	}
	return outputAddr, nil
}
