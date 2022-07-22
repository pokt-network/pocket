package utility

import (
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"math"
	"math/big"
)

// 'Actor' is the consolidated term for common functionality among the following network actors: app, fish, node, val
// Within this file, there are all the state based CRUD operations that are shared between these abstractions.
// The ideology of the separation of the actors is based on the expectation of actor divergence in the near future.
// The current implementation attempts to simplify code footprint and complexity while enabling future divergence.
// It is important to note, that as production approaches, the idea is to attempt more consolidation at an architectural
// multi-module level. Until then, it's a fine line to walk.

// setters

func (u *UtilityContext) SetActorStakedTokens(address []byte, tokens *big.Int, actorType typesUtil.ActorType) types.Error {
	var er error
	store := u.Store()
	switch actorType {
	case typesUtil.ActorType_App:
		er = store.SetAppStakeAmount(address, types.BigIntToString(tokens))
	case typesUtil.ActorType_Fish:
		er = store.SetFishermanStakeAmount(address, types.BigIntToString(tokens))
	case typesUtil.ActorType_Node:
		er = store.SetServiceNodeStakeAmount(address, types.BigIntToString(tokens))
	case typesUtil.ActorType_Val:
		er = store.SetValidatorStakeAmount(address, types.BigIntToString(tokens))
	}
	if er != nil {
		return types.ErrSetValidatorStakedTokens(er)
	}
	return nil
}

func (u *UtilityContext) SetActorUnstaking(address []byte, unstakingHeight int64, actorType typesUtil.ActorType) types.Error {
	store := u.Store()
	var er error
	switch actorType {
	case typesUtil.ActorType_App:
		er = store.SetAppUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Fish:
		er = store.SetFishermanUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Node:
		er = store.SetServiceNodeUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	case typesUtil.ActorType_Val:
		er = store.SetValidatorUnstakingHeightAndStatus(address, unstakingHeight, typesUtil.UnstakingStatus)
	}
	if er != nil {
		return types.ErrSetUnstakingHeightAndStatus(er)
	}
	return nil
}

func (u *UtilityContext) DeleteActor(address []byte, actorType typesUtil.ActorType) types.Error {
	var err error
	store := u.Store()
	switch actorType {
	case typesUtil.ActorType_App:
		err = store.DeleteApp(address)
	case typesUtil.ActorType_Fish:
		err = store.DeleteFisherman(address)
	case typesUtil.ActorType_Node:
		err = store.DeleteServiceNode(address)
	case typesUtil.ActorType_Val:
		err = store.DeleteValidator(address)
	}
	if err != nil {
		return types.ErrDelete(err)
	}
	return nil
}

func (u *UtilityContext) SetActorPauseHeight(address []byte, actorType typesUtil.ActorType, height int64) types.Error {
	var err error
	store := u.Store()
	if err := store.SetAppPauseHeight(address, height); err != nil {
		return types.ErrSetPauseHeight(err)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		err = store.SetAppPauseHeight(address, height)
	case typesUtil.ActorType_Fish:
		err = store.SetFishermanPauseHeight(address, height)
	case typesUtil.ActorType_Node:
		err = store.SetServiceNodePauseHeight(address, height)
	case typesUtil.ActorType_Val:
		err = store.SetValidatorPauseHeight(address, height)
	}
	if err != nil {
		return types.ErrSetPauseHeight(err)
	}
	return nil
}

// getters

func (u *UtilityContext) GetActorStakedTokens(address []byte, actorType typesUtil.ActorType) (*big.Int, types.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, types.ErrGetStakedTokens(er)
	}
	var stakedTokens string
	switch actorType {
	case typesUtil.ActorType_App:
		stakedTokens, er = store.GetAppStakeAmount(height, address)
	case typesUtil.ActorType_Fish:
		stakedTokens, er = store.GetFishermanStakeAmount(height, address)
	case typesUtil.ActorType_Node:
		stakedTokens, er = store.GetServiceNodeStakeAmount(height, address)
	case typesUtil.ActorType_Val:
		stakedTokens, er = store.GetValidatorStakeAmount(height, address)
	}
	if er != nil {
		return nil, types.ErrGetStakedTokens(er)
	}
	i, err := types.StringToBigInt(stakedTokens)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (u *UtilityContext) GetMaxPausedBlocks(actorType typesUtil.ActorType) (maxPausedBlocks int, err types.Error) {
	var er error
	var paramName string
	store := u.Store()
	switch actorType {
	case typesUtil.ActorType_App:
		maxPausedBlocks, er = store.GetAppMaxPausedBlocks()
		paramName = types.AppMaxPauseBlocksParamName
	case typesUtil.ActorType_Fish:
		maxPausedBlocks, er = store.GetFishermanMaxPausedBlocks()
		paramName = types.FishermanMaxPauseBlocksParamName
	case typesUtil.ActorType_Node:
		maxPausedBlocks, er = store.GetServiceNodeMaxPausedBlocks()
		paramName = types.ServiceNodeMaxPauseBlocksParamName
	case typesUtil.ActorType_Val:
		maxPausedBlocks, er = store.GetValidatorMaxPausedBlocks()
		paramName = types.ValidatorMaxPausedBlocksParamName
	}
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, er)
	}
	return
}

func (u *UtilityContext) GetMinimumPauseBlocks(actorType typesUtil.ActorType) (minPauseBlocks int, err types.Error) {
	store := u.Store()
	var er error
	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		minPauseBlocks, er = store.GetAppMinimumPauseBlocks()
		paramName = types.AppMinimumPauseBlocksParamName
	case typesUtil.ActorType_Fish:
		minPauseBlocks, er = store.GetFishermanMinimumPauseBlocks()
		paramName = types.FishermanMinimumPauseBlocksParamName
	case typesUtil.ActorType_Node:
		minPauseBlocks, er = store.GetServiceNodeMinimumPauseBlocks()
		paramName = types.ServiceNodeMinimumPauseBlocksParamName
	case typesUtil.ActorType_Val:
		minPauseBlocks, er = store.GetValidatorMinimumPauseBlocks()
		paramName = types.ValidatorMinimumPauseBlocksParamName
	}
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, er)
	}
	return
}

func (u *UtilityContext) GetPauseHeight(address []byte, actorType typesUtil.ActorType) (pauseHeight int64, err types.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetPauseHeight(er)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		pauseHeight, er = store.GetAppPauseHeightIfExists(address, height)
	case typesUtil.ActorType_Fish:
		pauseHeight, er = store.GetFishermanPauseHeightIfExists(address, height)
	case typesUtil.ActorType_Node:
		pauseHeight, er = store.GetServiceNodePauseHeightIfExists(address, height)
	case typesUtil.ActorType_Val:
		pauseHeight, er = store.GetValidatorPauseHeightIfExists(address, height)
	}
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetPauseHeight(er)
	}
	return
}

func (u *UtilityContext) GetActorStatus(address []byte, actorType typesUtil.ActorType) (status int, err types.Error) {
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetStatus(er)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		status, er = store.GetAppStatus(address, height)
	case typesUtil.ActorType_Fish:
		status, er = store.GetFishermanStatus(address, height)
	case typesUtil.ActorType_Node:
		status, er = store.GetServiceNodeStatus(address, height)
	case typesUtil.ActorType_Val:
		status, er = store.GetValidatorStatus(address, height)
	}
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetStatus(er)
	}
	return status, nil
}

func (u *UtilityContext) GetMinimumStake(actorType typesUtil.ActorType) (*big.Int, types.Error) {
	var minStake string
	var err error
	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		minStake, err = u.Store().GetParamAppMinimumStake()
		paramName = types.AppMinimumStakeParamName
	case typesUtil.ActorType_Fish:
		minStake, err = u.Store().GetParamFishermanMinimumStake()
		paramName = types.FishermanMinimumStakeParamName
	case typesUtil.ActorType_Node:
		minStake, err = u.Store().GetParamServiceNodeMinimumStake()
		paramName = types.ServiceNodeMinimumStakeParamName
	case typesUtil.ActorType_Val:
		minStake, err = u.Store().GetParamValidatorMinimumStake()
		paramName = types.ValidatorMinimumStakeParamName
	}
	if err != nil {
		return nil, types.ErrGetParam(paramName, err)
	}
	return types.StringToBigInt(minStake)
}

func (u *UtilityContext) GetStakeAmount(address []byte, actorType typesUtil.ActorType) (*big.Int, types.Error) {
	var stakeAmount string
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, types.ErrGetStakeAmount(err)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		stakeAmount, err = u.Store().GetAppStakeAmount(height, address)
	case typesUtil.ActorType_Fish:
		stakeAmount, err = u.Store().GetFishermanStakeAmount(height, address)
	case typesUtil.ActorType_Node:
		stakeAmount, err = u.Store().GetServiceNodeStakeAmount(height, address)
	case typesUtil.ActorType_Val:
		stakeAmount, err = u.Store().GetValidatorStakeAmount(height, address)
	}
	if err != nil {
		return nil, types.ErrGetStakeAmount(err)
	}
	return types.StringToBigInt(stakeAmount)
}

func (u *UtilityContext) GetUnstakingHeight(actorType typesUtil.ActorType) (unstakingHeight int64, er types.Error) {
	var err error
	var paramName string
	var unstakingBlocks int
	store := u.Store()
	switch actorType {
	case typesUtil.ActorType_App:
		unstakingBlocks, err = store.GetAppUnstakingBlocks()
		paramName = types.AppUnstakingBlocksParamName
	case typesUtil.ActorType_Fish:
		unstakingBlocks, err = store.GetFishermanUnstakingBlocks()
		paramName = types.FishermanUnstakingBlocksParamName
	case typesUtil.ActorType_Node:
		unstakingBlocks, err = store.GetServiceNodeUnstakingBlocks()
		paramName = types.ServiceNodeUnstakingBlocksParamName
	case typesUtil.ActorType_Val:
		unstakingBlocks, err = store.GetValidatorUnstakingBlocks()
		paramName = types.ValidatorUnstakingBlocksParamName
	}
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, err)
	}
	return u.CalculateUnstakingHeight(int64(unstakingBlocks))
}

func (u *UtilityContext) GetMaxChains(actorType typesUtil.ActorType) (maxChains int, er types.Error) {
	var err error
	var paramName string
	switch actorType {
	case typesUtil.ActorType_App:
		maxChains, err = u.Store().GetMaxAppChains()
		paramName = types.AppMinimumStakeParamName
	case typesUtil.ActorType_Fish:
		maxChains, err = u.Store().GetFishermanMaxChains()
		paramName = types.FishermanMinimumStakeParamName
	case typesUtil.ActorType_Node:
		maxChains, err = u.Store().GetServiceNodeMaxChains()
		paramName = types.ServiceNodeMinimumStakeParamName
	}
	if err != nil {
		return 0, types.ErrGetParam(paramName, err)
	}
	return
}

func (u *UtilityContext) GetActorExists(address []byte, actorType typesUtil.ActorType) (bool, types.Error) {
	var exists bool
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return false, types.ErrGetExists(err)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		exists, err = store.GetAppExists(address, height)
	case typesUtil.ActorType_Fish:
		exists, err = store.GetFishermanExists(address, height)
	case typesUtil.ActorType_Node:
		exists, err = store.GetServiceNodeExists(address, height)
	case typesUtil.ActorType_Val:
		exists, err = store.GetValidatorExists(address, height)
	}
	if err != nil {
		return false, types.ErrGetExists(err)
	}
	return exists, nil
}

func (u *UtilityContext) GetActorOutputAddress(operator []byte, actorType typesUtil.ActorType) (output []byte, err types.Error) {
	var er error
	store := u.Store()
	height, er := store.GetHeight()
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	switch actorType {
	case typesUtil.ActorType_App:
		output, er = store.GetAppOutputAddress(operator, height)
	case typesUtil.ActorType_Fish:
		output, er = store.GetFishermanOutputAddress(operator, height)
	case typesUtil.ActorType_Node:
		output, er = store.GetServiceNodeOutputAddress(operator, height)
	case typesUtil.ActorType_Val:
		output, er = store.GetValidatorOutputAddress(operator, height)
	}
	if er != nil {
		return nil, types.ErrGetOutputAddress(operator, er)
	}
	return output, nil
}

// calculators

func (u *UtilityContext) BurnActor(address []byte, percentage int, actorType typesUtil.ActorType) types.Error {
	tokens, err := u.GetActorStakedTokens(address, actorType)
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
	if err := u.SubPoolAmount(typesGenesis.ValidatorStakePoolName, types.BigIntToString(truncatedTokens)); err != nil {
		return err
	}
	// remove from validator
	if err := u.SetActorStakedTokens(address, newTokensAfterBurn, actorType); err != nil {
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
		if err := u.SetActorUnstaking(address, unstakingHeight, actorType); err != nil {
			return err
		}
	}
	return nil
}

func (u *UtilityContext) CalculateAppRelays(stakedTokens string) (string, types.Error) {
	tokens, err := types.StringToBigInt(stakedTokens)
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
	return types.BigIntToString(result), nil
}

func (u *UtilityContext) CheckAboveMinStake(amount string, actorType typesUtil.ActorType) (a *big.Int, err types.Error) {
	minStake, er := u.GetMinimumStake(actorType)
	if er != nil {
		return nil, err
	}
	a, err = types.StringToBigInt(amount)
	if err != nil {
		return nil, err
	}
	if types.BigIntLessThan(a, minStake) {
		return nil, types.ErrMinimumStake()
	}
	return // for convenience this returns amount as a big.Int
}

func (u *UtilityContext) CheckBelowMaxChains(chains []string, actorType typesUtil.ActorType) types.Error {
	// validators don't have chains field
	if actorType != typesUtil.ActorType_Val {
		maxChains, err := u.GetMaxChains(actorType)
		if err != nil {
			return err
		}
		if len(chains) > maxChains {
			return types.ErrMaxChains(maxChains)
		}
	}
	return nil
}

func (u *UtilityContext) CalculateUnstakingHeight(unstakingBlocks int64) (int64, types.Error) {
	latestHeight, err := u.GetLatestHeight()
	if err != nil {
		return typesUtil.ZeroInt, err
	}
	return unstakingBlocks + latestHeight, nil
}

// util

func (u *UtilityContext) BytesToPublicKey(publicKey []byte) (crypto.PublicKey, types.Error) {
	pk, er := crypto.NewPublicKeyFromBytes(publicKey)
	if er != nil {
		return nil, types.ErrNewPublicKeyFromBytes(er)
	}
	return pk, nil
}
