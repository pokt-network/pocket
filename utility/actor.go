package utility

import (
	"math"
	"math/big"

	typesGenesis "github.com/pokt-network/pocket/persistence/types"
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

// TODO(andrew): Make sure the `er` value in all the functions here is used. E.g. It is not used in `GetMinimumPauseBlocks`.
// TODO(andrew): Remove code that is unnecessarily repeated in this file. E.g. The number of times `store.GetHeight()` can be reduced in the entire file.

// setters

func (u *UtilityContext) SetActorStakedTokens(actorType typesUtil.UtilActorType, tokens *big.Int, address []byte) typesUtil.Error {
	var er error
	store := u.Store()

	switch actorType {
	case typesUtil.UtilActorType_App:
		er = store.SetAppStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.UtilActorType_Fish:
		er = store.SetFishermanStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.UtilActorType_Node:
		er = store.SetServiceNodeStakeAmount(address, typesUtil.BigIntToString(tokens))
	case typesUtil.UtilActorType_Val:
		er = store.SetValidatorStakeAmount(address, typesUtil.BigIntToString(tokens))
	}

	if er != nil {
		return typesUtil.ErrSetValidatorStakedTokens(er)
	}

	return nil
}

func (u *UtilityContext) SetActorUnstaking(actorType typesUtil.UtilActorType, unstakingHeight int64, address []byte) typesUtil.Error {
	store := u.Store()
	var er error

	switch actorType {
	case typesUtil.UtilActorType_App:
		er = store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Fish:
		er = store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Node:
		er = store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.UtilActorType_Val:
		er = store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	}

	if er != nil {
		return typesUtil.ErrSetUnstakingHeightAndStatus(er)
	}

	return nil
}

func (u *UtilityContext) DeleteActor(actorType typesUtil.UtilActorType, address []byte) typesUtil.Error {
	var err error
	store := u.Store()

	switch actorType {
	case typesUtil.UtilActorType_App:
		err = store.DeleteApp(address)
	case typesUtil.UtilActorType_Fish:
		err = store.DeleteFisherman(address)
	case typesUtil.UtilActorType_Node:
		err = store.DeleteServiceNode(address)
	case typesUtil.UtilActorType_Val:
		err = store.DeleteValidator(address)
	}

	if err != nil {
		return typesUtil.ErrDelete(err)
	}

	return nil
}

func (u *UtilityContext) SetActorPauseHeight(actorType typesUtil.UtilActorType, address []byte, height int64) typesUtil.Error {
	var err error
	store := u.Store()

	switch actorType {
	case typesUtil.UtilActorType_App:
		err = store.SetAppPauseHeight(address, height)
	case typesUtil.UtilActorType_Fish:
		err = store.SetFishermanPauseHeight(address, height)
	case typesUtil.UtilActorType_Node:
		err = store.SetServiceNodePauseHeight(address, height)
	case typesUtil.UtilActorType_Val:
		err = store.SetValidatorPauseHeight(address, height)
	}

	if err != nil {
		return typesUtil.ErrSetPauseHeight(err)
	}

	return nil
}

// getters

func (u *UtilityContext) GetActorStakedTokens(actorType typesUtil.UtilActorType, address []byte) (*big.Int, typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, typesUtil.ErrGetStakedTokens(er)
	}

	var stakedTokens string
	switch actorType {
	case typesUtil.UtilActorType_App:
		stakedTokens, er = store.GetAppStakeAmount(height, address)
	case typesUtil.UtilActorType_Fish:
		stakedTokens, er = store.GetFishermanStakeAmount(height, address)
	case typesUtil.UtilActorType_Node:
		stakedTokens, er = store.GetServiceNodeStakeAmount(height, address)
	case typesUtil.UtilActorType_Val:
		stakedTokens, er = store.GetValidatorStakeAmount(height, address)
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

func (u *UtilityContext) GetMaxPausedBlocks(actorType typesUtil.UtilActorType) (maxPausedBlocks int, err typesUtil.Error) {
	var paramName string

	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		paramName = modules.AppMaxPauseBlocksParamName
	case typesUtil.UtilActorType_Fish:
		paramName = modules.FishermanMaxPauseBlocksParamName
	case typesUtil.UtilActorType_Node:
		paramName = modules.ServiceNodeMaxPauseBlocksParamName
	case typesUtil.UtilActorType_Val:
		paramName = modules.ValidatorMaxPausedBlocksParamName
	}

	maxPausedBlocks, er = store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	return
}

func (u *UtilityContext) GetMinimumPauseBlocks(actorType typesUtil.UtilActorType) (minPauseBlocks int, err typesUtil.Error) {
	var paramName string

	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er) // TODO(andrew): does this need a custom error?
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		paramName = modules.AppMinimumPauseBlocksParamName
	case typesUtil.UtilActorType_Fish:
		paramName = modules.FishermanMinimumPauseBlocksParamName
	case typesUtil.UtilActorType_Node:
		paramName = modules.ServiceNodeMinimumPauseBlocksParamName
	case typesUtil.UtilActorType_Val:
		paramName = modules.ValidatorMinimumPauseBlocksParamName
	}

	minPauseBlocks, er = store.GetIntParam(paramName, height)
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, er)
	}

	return
}

func (u *UtilityContext) GetPauseHeight(actorType typesUtil.UtilActorType, address []byte) (pauseHeight int64, err typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(er)
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		pauseHeight, er = store.GetAppPauseHeightIfExists(address, height)
	case typesUtil.UtilActorType_Fish:
		pauseHeight, er = store.GetFishermanPauseHeightIfExists(address, height)
	case typesUtil.UtilActorType_Node:
		pauseHeight, er = store.GetServiceNodePauseHeightIfExists(address, height)
	case typesUtil.UtilActorType_Val:
		pauseHeight, er = store.GetValidatorPauseHeightIfExists(address, height)
	}

	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetPauseHeight(er)
	}

	return
}

func (u *UtilityContext) GetActorStatus(actorType typesUtil.UtilActorType, address []byte) (status int, err typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(er)
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		status, er = store.GetAppStatus(address, height)
	case typesUtil.UtilActorType_Fish:
		status, er = store.GetFishermanStatus(address, height)
	case typesUtil.UtilActorType_Node:
		status, er = store.GetServiceNodeStatus(address, height)
	case typesUtil.UtilActorType_Val:
		status, er = store.GetValidatorStatus(address, height)
	}

	if er != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStatus(er)
	}

	return status, nil
}

func (u *UtilityContext) GetMinimumStake(actorType typesUtil.UtilActorType) (*big.Int, typesUtil.Error) {
	var paramName string

	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}

	var minStake string
	switch actorType {
	case typesUtil.UtilActorType_App:
		paramName = modules.AppMinimumStakeParamName
	case typesUtil.UtilActorType_Fish:
		paramName = modules.FishermanMinimumStakeParamName
	case typesUtil.UtilActorType_Node:
		paramName = modules.ServiceNodeMinimumStakeParamName
	case typesUtil.UtilActorType_Val:
		paramName = modules.ValidatorMinimumStakeParamName
	}

	minStake, err = store.GetStringParam(paramName, height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}

	return typesUtil.StringToBigInt(minStake)
}

func (u *UtilityContext) GetStakeAmount(actorType typesUtil.UtilActorType, address []byte) (*big.Int, typesUtil.Error) {
	var stakeAmount string
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		stakeAmount, err = store.GetAppStakeAmount(height, address)
	case typesUtil.UtilActorType_Fish:
		stakeAmount, err = store.GetFishermanStakeAmount(height, address)
	case typesUtil.UtilActorType_Node:
		stakeAmount, err = store.GetServiceNodeStakeAmount(height, address)
	case typesUtil.UtilActorType_Val:
		stakeAmount, err = store.GetValidatorStakeAmount(height, address)
	}

	if err != nil {
		return nil, typesUtil.ErrGetStakeAmount(err)
	}

	return typesUtil.StringToBigInt(stakeAmount)
}

func (u *UtilityContext) GetUnstakingHeight(actorType typesUtil.UtilActorType) (unstakingHeight int64, er typesUtil.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStakeAmount(err)
	}

	var paramName string
	var unstakingBlocks int
	switch actorType {
	case typesUtil.UtilActorType_App:
		paramName = modules.AppUnstakingBlocksParamName
	case typesUtil.UtilActorType_Fish:
		paramName = modules.FishermanUnstakingBlocksParamName
	case typesUtil.UtilActorType_Node:
		paramName = modules.ServiceNodeUnstakingBlocksParamName
	case typesUtil.UtilActorType_Val:
		paramName = modules.ValidatorUnstakingBlocksParamName
	}

	unstakingBlocks, err = store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}

	return u.CalculateUnstakingHeight(int64(unstakingBlocks))
}

func (u *UtilityContext) GetMaxChains(actorType typesUtil.UtilActorType) (maxChains int, er typesUtil.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetStakeAmount(err)
	}

	var paramName string
	switch actorType {
	case typesUtil.UtilActorType_App:
		paramName = modules.AppMinimumStakeParamName
	case typesUtil.UtilActorType_Fish:
		paramName = modules.FishermanMinimumStakeParamName
	case typesUtil.UtilActorType_Node:
		paramName = modules.ServiceNodeMinimumStakeParamName
	}

	maxChains, err = store.GetIntParam(paramName, height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}

	return
}

func (u *UtilityContext) GetActorExists(actorType typesUtil.UtilActorType, address []byte) (bool, typesUtil.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return false, typesUtil.ErrGetExists(err)
	}

	var exists bool
	switch actorType {
	case typesUtil.UtilActorType_App:
		exists, err = store.GetAppExists(address, height)
	case typesUtil.UtilActorType_Fish:
		exists, err = store.GetFishermanExists(address, height)
	case typesUtil.UtilActorType_Node:
		exists, err = store.GetServiceNodeExists(address, height)
	case typesUtil.UtilActorType_Val:
		exists, err = store.GetValidatorExists(address, height)
	}

	if err != nil {
		return false, typesUtil.ErrGetExists(err)
	}

	return exists, nil
}

func (u *UtilityContext) GetActorOutputAddress(actorType typesUtil.UtilActorType, operator []byte) (output []byte, err typesUtil.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, er)
	}

	switch actorType {
	case typesUtil.UtilActorType_App:
		output, er = store.GetAppOutputAddress(operator, height)
	case typesUtil.UtilActorType_Fish:
		output, er = store.GetFishermanOutputAddress(operator, height)
	case typesUtil.UtilActorType_Node:
		output, er = store.GetServiceNodeOutputAddress(operator, height)
	case typesUtil.UtilActorType_Val:
		output, er = store.GetValidatorOutputAddress(operator, height)
	}

	if er != nil {
		return nil, typesUtil.ErrGetOutputAddress(operator, er)

	}
	return output, nil
}

// calculators

func (u *UtilityContext) BurnActor(actorType typesUtil.UtilActorType, percentage int, address []byte) typesUtil.Error {
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
	if err := u.SubPoolAmount(typesGenesis.Pool_Names_ValidatorStakePool.String(), typesUtil.BigIntToString(truncatedTokens)); err != nil {
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
	// TODO (team) evaluate whether or not we should use micro denomination or not
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

func (u *UtilityContext) CheckAboveMinStake(actorType typesUtil.UtilActorType, amount string) (a *big.Int, err typesUtil.Error) {
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

func (u *UtilityContext) CheckBelowMaxChains(actorType typesUtil.UtilActorType, chains []string) typesUtil.Error {
	// validators don't have chains field
	if actorType == typesUtil.UtilActorType_Val {
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
	latestHeight, err := u.GetLatestHeight()
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
