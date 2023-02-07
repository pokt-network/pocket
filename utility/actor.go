package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

/*
   `Actor` is the consolidated term for common functionality among the following network actors: app, fish, node, val.

   This file contains all the state based CRUD operations shared between these abstractions.

   The ideology of the separation of the actors is based on the expectation of actor divergence in the near future.
   The current implementation attempts to simplify code footprint and complexity while enabling future divergence.
   It is important to note, that as production approaches, the idea is to attempt more consolidation at an architectural
   multi-module level. Until then, it's a fine line to walk.
*/

func (u *utilityContext) setActorStakedAmount(actorType coreTypes.ActorType, addr []byte, amount *big.Int) typesUtil.Error {
	store := u.Store()

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppStakeAmount(addr, converters.BigIntToString(amount))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanStakeAmount(addr, converters.BigIntToString(amount))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodeStakeAmount(addr, converters.BigIntToString(amount))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorStakeAmount(addr, converters.BigIntToString(amount))
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

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppUnstakingHeightAndStatus(addr, height, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanUnstakingHeightAndStatus(addr, height, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodeUnstakingHeightAndStatus(addr, height, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorUnstakingHeightAndStatus(addr, height, int32(typesUtil.StakeStatus_Unstaking))
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

// getters

func (u *utilityContext) getActorStakedAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var stakedAmount string
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakedAmount, err = store.GetAppStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakedAmount, err = store.GetFishermanStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		stakedAmount, err = store.GetServiceNodeStakeAmount(height, addr)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakedAmount, err = store.GetValidatorStakeAmount(height, addr)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	amount, err := converters.StringToBigInt(stakedAmount)
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

func (u *utilityContext) getCurrentStakeAmount(actorType coreTypes.ActorType, addr []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

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

func (u *utilityContext) getUnstakingHeight(actorType coreTypes.ActorType) (int64, typesUtil.Error) {
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

// calculators

func (u *utilityContext) burnValidator(percentage int, addr []byte) typesUtil.Error {
	// TODO: Will need to extend this to support burning from other actors types & pools when the logic is implemented
	validatorActorType := coreTypes.ActorType_ACTOR_TYPE_VAL
	validatorPool := coreTypes.Pools_POOLS_VALIDATOR_STAKE

	stakeAmount, err := u.getActorStakedAmount(validatorActorType, addr)
	if err != nil {
		return err
	}

	// newStake = currentStake * percentageInt / 100
	burnAmount := new(big.Float).SetInt(stakeAmount)
	burnAmount.Mul(burnAmount, big.NewFloat(float64(percentage)))
	burnAmount.Quo(burnAmount, big.NewFloat(100))
	burnAmountTruncated, _ := burnAmount.Int(nil)

	// Round up to 0 if -ve
	zeroBigInt := big.NewInt(0)
	if burnAmountTruncated.Cmp(zeroBigInt) == -1 {
		burnAmountTruncated = zeroBigInt
	}

	newAmountAfterBurn := big.NewInt(0).Sub(stakeAmount, burnAmountTruncated)

	// remove from pool
	if err := u.subPoolAmount(validatorPool.FriendlyName(), converters.BigIntToString(burnAmountTruncated)); err != nil {
		return err
	}

	// remove from actor
	if err := u.setActorStakedAmount(validatorActorType, addr, newAmountAfterBurn); err != nil {
		return err
	}

	// Need to check if new stake is below min required stake
	minStake, err := u.getValidatorMinimumStake()
	if err != nil {
		return err
	}

	// Check if amount after burn is below the min required stake
	if minStake.Cmp(newAmountAfterBurn) == -1 {
		unstakingHeight, err := u.getUnstakingHeight(validatorActorType)
		if err != nil {
			return err
		}
		if err := u.setActorUnstakingHeight(validatorActorType, addr, unstakingHeight); err != nil {
			return err
		}
	}

	return nil
}

// TODO: Reevaluate the implementation in this function when implementation the Application Protocol
// and rate limiting
func (u *utilityContext) calculateMaxAppRelays(stakeStr string) (string, typesUtil.Error) {
	stakeBigInt, er := converters.StringToBigInt(stakeStr)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigInt(er)
	}

	stabilityAdjustment, err := u.GetStabilityAdjustment()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	baseRate, err := u.GetBaselineAppStakeRate()
	if err != nil {
		return typesUtil.EmptyString, err
	}

	// convert tokens to float64
	stake := big.NewFloat(float64(stakeBigInt.Int64()))
	// get the percentage of the baseline stake rate; can be over 100%
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))
	// multiply the two
	baselineThroughput := basePercentage.Mul(basePercentage, stake)
	// Convert POKT to uPOKT
	baselineThroughput.Quo(baselineThroughput, big.NewFloat(typesUtil.MillionInt))
	// add staking adjustment; can be -ve
	adjusted := baselineThroughput.Add(baselineThroughput, big.NewFloat(float64(stabilityAdjustment)))
	// truncate the integer
	result, _ := adjusted.Int(nil)
	// bounding Max Amount of relays to maxint64
	max := big.NewInt(math.MaxInt64)
	if i := result.Cmp(max); i < -1 {
		result = max
	}
	return converters.BigIntToString(result), nil
}
