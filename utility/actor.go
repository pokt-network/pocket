package utility

import (
	"math"
	"math/big"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
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

func (u *UtilityContext) SetActorStakedTokens(actorType typesUtil.ActorType, tokens *big.Int, address []byte) typesUtil.Error {
	store := u.Store()

	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		er = store.SetAppStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.ActorType_Fisherman:
		er = store.SetFishermanStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.ActorType_ServiceNode:
		er = store.SetServiceNodeStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.ActorType_Validator:
		er = store.SetValidatorStakeAmount(address, typesUtil.BigIntToString(tokens))
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ErrSetValidatorStakedTokens(er)
	}

	return nil
}

func (u *UtilityContext) SetActorUnstaking(actorType typesUtil.ActorType, unstakingHeight int64, address []byte) typesUtil.Error {
	store := u.Store()

	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		er = store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case typesUtil.ActorType_Fisherman:
		er = store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case typesUtil.ActorType_ServiceNode:
		er = store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	case typesUtil.ActorType_Validator:
		er = store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, int32(typesUtil.StakeStatus_Unstaking))
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(er)
	}

	return nil
}

func (u *UtilityContext) DeleteActor(actorType typesUtil.ActorType, address []byte) typesUtil.Error {
	store := u.Store()

	var err error
	switch actorType {
	case typesUtil.ActorType_App:
		err = store.DeleteApp(address)
	case typesUtil.ActorType_Fisherman:
		err = store.DeleteFisherman(address)
	case typesUtil.ActorType_ServiceNode:
		err = store.DeleteServiceNode(address)
	case typesUtil.ActorType_Validator:
		err = store.DeleteValidator(address)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return typesUtil.ErrDelete(err)
	}

	return nil
}

func (u *UtilityContext) SetActorPauseHeight(actorType typesUtil.ActorType, address []byte, height int64) typesUtil.Error {
	store := u.Store()

	var err error
	switch actorType {
	case typesUtil.ActorType_App:
		err = store.SetAppPauseHeight(address, height)
	case typesUtil.ActorType_Fisherman:
		err = store.SetFishermanPauseHeight(address, height)
	case typesUtil.ActorType_ServiceNode:
		err = store.SetServiceNodePauseHeight(address, height)
	case typesUtil.ActorType_Validator:
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

func (u *UtilityContext) GetActorStakedTokens(actorType typesUtil.ActorType, address []byte) (*big.Int, typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return nil, err
	}

	var stakedTokens string
	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		stakedTokens, er = store.GetAppStakeAmount(height, address)
	case typesUtil.ActorType_Fisherman:
		stakedTokens, er = store.GetFishermanStakeAmount(height, address)
	case typesUtil.ActorType_ServiceNode:
		stakedTokens, er = store.GetServiceNodeStakeAmount(height, address)
	case typesUtil.ActorType_Validator:
		stakedTokens, er = store.GetValidatorStakeAmount(height, address)
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return nil, typesUtil.ErrGetStakedTokens(er)
	}

	i, err := typesUtil.StringToBigInt(stakedTokens)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (u *UtilityContext) GetMaxPausedBlocks(actorType typesUtil.ActorType) (maxPausedBlocks int, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		paramName = modules.AppMaxPauseBlocksParamName
	case typesUtil.ActorType_Fisherman:
		paramName = modules.FishermanMaxPauseBlocksParamName
	case typesUtil.ActorType_ServiceNode:
		paramName = modules.ServiceNodeMaxPauseBlocksParamName
	case typesUtil.ActorType_Validator:
		paramName = modules.ValidatorMaxPausedBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	var er error
	maxPausedBlocks, er = store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	return
}

func (u *UtilityContext) GetMinimumPauseBlocks(actorType typesUtil.ActorType) (minPauseBlocks int, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		paramName = modules.AppMinimumPauseBlocksParamName
	case typesUtil.ActorType_Fisherman:
		paramName = modules.FishermanMinimumPauseBlocksParamName
	case typesUtil.ActorType_ServiceNode:
		paramName = modules.ServiceNodeMinimumPauseBlocksParamName
	case typesUtil.ActorType_Validator:
		paramName = modules.ValidatorMinimumPauseBlocksParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minPauseBlocks, er := store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	return
}

func (u *UtilityContext) GetPauseHeight(actorType typesUtil.ActorType, address []byte) (pauseHeight int64, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		pauseHeight, er = store.GetAppPauseHeightIfExists(address, height)
	case typesUtil.ActorType_Fisherman:
		pauseHeight, er = store.GetFishermanPauseHeightIfExists(address, height)
	case typesUtil.ActorType_ServiceNode:
		pauseHeight, er = store.GetServiceNodePauseHeightIfExists(address, height)
	case typesUtil.ActorType_Validator:
		pauseHeight, er = store.GetValidatorPauseHeightIfExists(address, height)
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(er)
	}

	return
}

func (u *UtilityContext) GetActorStatus(actorType typesUtil.ActorType, address []byte) (status int32, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		status, er = store.GetAppStatus(address, height)
	case typesUtil.ActorType_Fisherman:
		status, er = store.GetFishermanStatus(address, height)
	case typesUtil.ActorType_ServiceNode:
		status, er = store.GetServiceNodeStatus(address, height)
	case typesUtil.ActorType_Validator:
		status, er = store.GetValidatorStatus(address, height)
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(er)
	}

	return status, nil
}

func (u *UtilityContext) GetMinimumStake(actorType typesUtil.ActorType) (*big.Int, typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return nil, err
	}

	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		paramName = modules.AppMinimumStakeParamName
	case typesUtil.ActorType_Fisherman:
		paramName = modules.FishermanMinimumStakeParamName
	case typesUtil.ActorType_ServiceNode:
		paramName = modules.ServiceNodeMinimumStakeParamName
	case typesUtil.ActorType_Validator:
		paramName = modules.ValidatorMinimumStakeParamName
	default:
		return nil, typesUtil.ErrUnknownActorType(actorType.String())
	}

	minStake, er := store.GetStringParam(paramName, height)
	if er != nil {
		return nil, typesUtil.ErrGetParam(paramName, er)
	}

	return typesUtil.StringToBigInt(minStake)
}

func (u *UtilityContext) GetStakeAmount(actorType typesUtil.ActorType, address []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}

	var err error
	switch actorType {
	case typesUtil.ActorType_App:
		stakeAmount, err = store.GetAppStakeAmount(height, address)
	case typesUtil.ActorType_Fisherman:
		stakeAmount, err = store.GetFishermanStakeAmount(height, address)
	case typesUtil.ActorType_ServiceNode:
		stakeAmount, err = store.GetServiceNodeStakeAmount(height, address)
	case typesUtil.ActorType_Validator:
		stakeAmount, err = store.GetValidatorStakeAmount(height, address)
	default:
		err = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	return typesUtil.StringToBigInt(stakeAmount)
}

func (u *UtilityContext) GetUnstakingHeight(actorType typesUtil.ActorType) (unstakingHeight int64, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var paramName string
	var unstakingBlocks int
	switch actorType {
	case typesUtil.ActorType_App:
		paramName = modules.AppUnstakingBlocksParamName
	case typesUtil.ActorType_Fisherman:
		paramName = modules.FishermanUnstakingBlocksParamName
	case typesUtil.ActorType_ServiceNode:
		paramName = modules.ServiceNodeUnstakingBlocksParamName
	case typesUtil.ActorType_Validator:
		paramName = modules.ValidatorUnstakingBlocksParamName
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

func (u *UtilityContext) GetMaxChains(actorType typesUtil.ActorType) (maxChains int, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return 0, err
	}

	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		paramName = modules.AppMinimumStakeParamName
	case typesUtil.ActorType_Fisherman:
		paramName = modules.FishermanMinimumStakeParamName
	case typesUtil.ActorType_ServiceNode:
		paramName = modules.ServiceNodeMinimumStakeParamName
	default:
		return 0, typesUtil.ErrUnknownActorType(actorType.String())
	}

	var er error
	maxChains, er = store.GetIntParam(paramName, height)
	if er != nil {
		return 0, typesUtil.ErrGetParam(paramName, er)
	}

	return
}

func (u *UtilityContext) GetActorExists(actorType typesUtil.ActorType, address []byte) (bool, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return false, er
	}

	var exists bool
	var err error
	switch actorType {
	case typesUtil.ActorType_App:
		exists, err = store.GetAppExists(address, height)
	case typesUtil.ActorType_Fisherman:
		exists, err = store.GetFishermanExists(address, height)
	case typesUtil.ActorType_ServiceNode:
		exists, err = store.GetServiceNodeExists(address, height)
	case typesUtil.ActorType_Validator:
		exists, err = store.GetValidatorExists(address, height)
	default:
		return false, typesUtil.ErrUnknownActorType(actorType.String())
	}

	if err != nil {
		return false, typesUtil.ErrGetExists(err)
	}

	return exists, nil
}

func (u *UtilityContext) GetActorOutputAddress(actorType typesUtil.ActorType, operator []byte) (output []byte, err typesUtil.Error) {
	store, height, err := u.GetStoreAndHeight()
	if err != nil {
		return nil, err
	}

	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		output, er = store.GetAppOutputAddress(operator, height)
	case typesUtil.ActorType_Fisherman:
		output, er = store.GetFishermanOutputAddress(operator, height)
	case typesUtil.ActorType_ServiceNode:
		output, er = store.GetServiceNodeOutputAddress(operator, height)
	case typesUtil.ActorType_Validator:
		output, er = store.GetValidatorOutputAddress(operator, height)
	default:
		er = typesUtil.ErrUnknownActorType(actorType.String())
	}

	if er != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, er)

	}
	return output, nil
}

// calculators

func (u *UtilityContext) BurnActor(actorType typesUtil.ActorType, percentage int, address []byte) typesUtil.Error {
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
	if err := u.SubPoolAmount(typesUtil.PoolNames_ValidatorStakePool.String(), typesUtil.BigIntToString(truncatedTokens)); err != nil {
		return err
	}
	// remove from validator
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

func (u *UtilityContext) CalculateAppRelays(stakedTokens string) (string, typesUtil.Error) {
	tokens, err := typesUtil.StringToBigInt(stakedTokens)
	if err != nil {
		return typesUtil.EmptyString, err
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
	return typesUtil.BigIntToString(result), nil
}

func (u *UtilityContext) CheckAboveMinStake(actorType typesUtil.ActorType, amount string) (a *big.Int, err typesUtil.Error) {
	minStake, er := u.GetMinimumStake(actorType)
	if er != nil {
		return nil, er
	}
	a, err = typesUtil.StringToBigInt(amount)
	if err != nil {
		return nil, err
	}
	if typesUtil.BigIntLessThan(a, minStake) {
		return nil, typesUtil.ErrMinimumStake()
	}
	return // for convenience this returns amount as a big.Int
}

func (u *UtilityContext) CheckBelowMaxChains(actorType typesUtil.ActorType, chains []string) typesUtil.Error {
	// validators don't have chains field
	if actorType == typesUtil.ActorType_Validator {
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

func (u *UtilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (int64, typesUtil.Error) {
	latestHeight, err := u.GetLatestBlockHeight()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}

// util

func (u *UtilityContext) BytesToPublicKey(publicKey []byte) (crypto.PublicKey, typesUtil.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(publicKey)
	if er != nil {
		return nil, typesUtil.ErrNewPublicKeyFromBytes(er)
	}
	return pk, nil
}
