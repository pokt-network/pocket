package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
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

// setters

func (u *utilityContext) SetActorStakedTokens(actorType coreTypes.ActorType, tokens *big.Int, address []byte) typesUtil.Error {
	store := u.Store()

	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = store.SetAppStakeAmount(address, converters.BigIntToString(tokens))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = store.SetFishermanStakeAmount(address, converters.BigIntToString(tokens))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		er = store.SetServiceNodeStakeAmount(address, converters.BigIntToString(tokens))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = store.SetValidatorStakeAmount(address, converters.BigIntToString(tokens))
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ErrSetValidatorStakedTokens(er)
	}

	return nil
}

func (u *utilityContext) SetActorUnstaking(actorType coreTypes.ActorType, unstakingHeight int64, address []byte) typesUtil.Error {
	store := u.Store()

	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		er = store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		er = store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		er = store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		er = store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(er)
	}

	return nil
}

func (u *utilityContext) SetActorPauseHeight(actorType coreTypes.ActorType, address []byte, height int64) typesUtil.Error {
	store := u.Store()

	var err error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		err = store.SetAppPauseHeight(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		err = store.SetFishermanPauseHeight(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		err = store.SetServiceNodePauseHeight(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		err = store.SetValidatorPauseHeight(address, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}

	return nil
}

// getters

func (u *utilityContext) GetActorStakedTokens(actorType coreTypes.ActorType, address []byte) (*big.Int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var stakedTokens string
	var er error
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakedTokens, er = store.GetAppStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakedTokens, er = store.GetFishermanStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		stakedTokens, er = store.GetServiceNodeStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakedTokens, er = store.GetValidatorStakeAmount(height, address)
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return nil, typesUtil.ErrGetStakedTokens(er)
	}

	amount, err := converters.StringToBigInt(stakedTokens)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) GetMaxPausedBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

func (u *utilityContext) GetMinimumPauseBlocks(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

	minPauseBlocks, er := store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}
	return minPauseBlocks, nil
}

func (u *utilityContext) GetPauseHeight(actorType coreTypes.ActorType, address []byte) (int64, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var pauseHeight int64
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		pauseHeight, err = store.GetAppPauseHeightIfExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		pauseHeight, err = store.GetFishermanPauseHeightIfExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		pauseHeight, err = store.GetServiceNodePauseHeightIfExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		pauseHeight, err = store.GetValidatorPauseHeightIfExists(address, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(err)
	}

	return pauseHeight, nil
}

func (u *utilityContext) GetActorStatus(actorType coreTypes.ActorType, address []byte) (int32, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var status int32
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		status, err = store.GetAppStatus(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		status, err = store.GetFishermanStatus(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		status, err = store.GetServiceNodeStatus(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		status, err = store.GetValidatorStatus(address, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(err)
	}

	return status, nil
}

func (u *utilityContext) GetMinimumStake(actorType coreTypes.ActorType) (*big.Int, typesUtil.Error) {
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

func (u *utilityContext) GetStakeAmount(actorType coreTypes.ActorType, address []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		stakeAmount, err = store.GetAppStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		stakeAmount, err = store.GetFishermanStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		stakeAmount, err = store.GetServiceNodeStakeAmount(height, address)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		stakeAmount, err = store.GetValidatorStakeAmount(height, address)
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

func (u *utilityContext) GetUnstakingHeight(actorType coreTypes.ActorType) (int64, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}

	var paramName string
	var unstakingBlocks int
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

	var er error
	unstakingBlocks, er = store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	return u.CalculateUnstakingHeight(int64(unstakingBlocks))
}

func (u *utilityContext) GetMaxChains(actorType coreTypes.ActorType) (int, typesUtil.Error) {
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

func (u *utilityContext) GetActorExists(actorType coreTypes.ActorType, address []byte) (bool, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return false, typesUtil.ErrGetHeight(err)
	}

	var exists bool
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		exists, err = store.GetAppExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		exists, err = store.GetFishermanExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		exists, err = store.GetServiceNodeExists(address, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		exists, err = store.GetValidatorExists(address, height)
	default:
		return false, typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return false, typesUtil.ErrGetExists(err)
	}

	return exists, nil
}

func (u *utilityContext) GetActorOutputAddress(actorType coreTypes.ActorType, operator []byte) ([]byte, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}

	var output []byte
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		output, err = store.GetAppOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		output, err = store.GetFishermanOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		output, err = store.GetServiceNodeOutputAddress(operator, height)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		output, err = store.GetValidatorOutputAddress(operator, height)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, err)

	}
	return output, nil
}

// calculators

func (u *utilityContext) BurnActor(actorType coreTypes.ActorType, percentage int, address []byte) typesUtil.Error {
	tokens, err := u.GetActorStakedTokens(actorType, address)
	if err != nil {
		return err
	}
	zeroBigInt := big.NewInt(0)
	tokensFloat := new(big.Float).SetInt(tokens)
	tokensFloat.Mul(tokensFloat, big.NewFloat(float64(percentage)))
	tokensFloat.Quo(tokensFloat, big.NewFloat(100))
	truncatedTokens, _ := tokensFloat.Int(nil)
	if truncatedTokens.Cmp(zeroBigInt) == -1 {
		truncatedTokens = zeroBigInt
	}
	newTokensAfterBurn := big.NewInt(0).Sub(tokens, truncatedTokens)
	// remove from pool
	if err := u.subPoolAmount(coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName(), converters.BigIntToString(truncatedTokens)); err != nil {
		return err
	}
	// remove from actor
	if err := u.SetActorStakedTokens(actorType, newTokensAfterBurn, address); err != nil {
		return err
	}
	// check to see if they fell below minimum stake
	minStake, err := u.GetValidatorMinimumStake()
	if err != nil {
		return err
	}
	// fell below minimum stake
	if minStake.Cmp(truncatedTokens) == 1 {
		unstakingHeight, err := u.GetUnstakingHeight(actorType)
		if err != nil {
			return err
		}
		if err := u.SetActorUnstaking(actorType, unstakingHeight, address); err != nil {
			return err
		}
	}
	return nil
}

func (u *utilityContext) CalculateAppRelays(stakedTokens string) (string, typesUtil.Error) {
	tokens, er := converters.StringToBigInt(stakedTokens)
	if er != nil {
		return typesUtil.EmptyString, typesUtil.ErrStringToBigInt(er)
	}
	// The constant integer adjustment that the DAO may use to move the stake. The DAO may manually
	// adjust an application's MaxRelays at the time of staking to correct for short-term fluctuations
	// in the price of POKT, which may not be reflected in ParticipationRate
	// When this parameter is set to 0, no adjustment is being made.
	stabilityAdjustment, err := u.GetStabilityAdjustment()
	if err != nil {
		return typesUtil.EmptyString, err
	}
	baseRate, err := u.GetBaselineAppStakeRate()
	if err != nil {
		return typesUtil.EmptyString, err
	}
	// convert tokens to float64
	tokensFloat64 := big.NewFloat(float64(tokens.Int64()))
	// get the percentage of the baseline stake rate (can be over 100%)
	basePercentage := big.NewFloat(float64(baseRate) / float64(100))
	// multiply the two
	// DISCUSS evaluate whether or not we should use micro denomination or not
	baselineThroughput := basePercentage.Mul(basePercentage, tokensFloat64)
	// adjust for uPOKT
	baselineThroughput.Quo(baselineThroughput, big.NewFloat(typesUtil.MillionInt))
	// add staking adjustment (can be negative)
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

func (u *utilityContext) CheckAboveMinStake(actorType coreTypes.ActorType, amountStr string) (*big.Int, typesUtil.Error) {
	minStake, err := u.GetMinimumStake(actorType)
	if err != nil {
		return nil, err
	}
	amount, er := converters.StringToBigInt(amountStr)
	if er != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	if converters.BigIntLessThan(amount, minStake) {
		return nil, typesUtil.ErrMinimumStake()
	}
	return amount, nil
}

func (u *utilityContext) CheckBelowMaxChains(actorType coreTypes.ActorType, chains []string) typesUtil.Error {
	// validators don't have chains field

	if actorType == coreTypes.ActorType_ACTOR_TYPE_VAL {
		return nil
	}

	maxChains, err := u.GetMaxChains(actorType)
	if err != nil {
		return err
	}
	if len(chains) > maxChains {
		return typesUtil.ErrMaxChains(maxChains)
	}
	return nil
}

func (u *utilityContext) GetLastBlockByzantineValidators() ([][]byte, error) {
	// TODO(#271): Need to retrieve byzantine validators from the persistence module
	return nil, nil
}

func (u *utilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (int64, typesUtil.Error) {
	latestHeight, err := u.GetLatestBlockHeight()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}

// util

func (u *utilityContext) BytesToPublicKey(publicKey []byte) (crypto.PublicKey, typesUtil.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(publicKey)
	if er != nil {
		return nil, typesUtil.ErrNewPublicKeyFromBytes(er)
	}
	return pk, nil
}
