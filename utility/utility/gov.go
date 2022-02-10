package utility

import (
	"pocket/utility/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math/big"
)

func (u *UtilityContext) HandleMessageChangeParameter(message *types.MessageChangeParameter) types.Error {
	cdc := u.Codec()
	v, err := cdc.FromAny(message.ParameterValue)
	if err != nil {
		return types.ErrProtoFromAny(err)
	}
	return u.UpdateParam(message.ParameterKey, v)
}

func (u *UtilityContext) UpdateParam(paramName string, value interface{}) types.Error {
	store := u.Store()
	switch paramName {
	case types.BlocksPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodesPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxAppChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamAppMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppBaselineStakeRateParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBaselineAppStakeRate(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppStakingAdjustmentParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetStakingAdjustment(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamServiceNodeMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamFishermanMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamValidatorMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaximumMissedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaximumMissedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ProposerPercentageOfFeesParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetProposerPercentageOfFees(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaxEvidenceAgeInBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxEvidenceAgeInBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MissedBlocksBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMissedBlocksBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.DoubleSignBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetDoubleSignBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ACLOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetACLOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.BlocksPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodesPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxAppChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppBaselineStakeRateOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetBaselineAppOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppStakingAdjustmentOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetStakingAdjustmentOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.AppMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamServiceNodeMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxServiceNodeChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ServiceNodeMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ParamFishermanMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxFishermanChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.FishermanMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamValidatorMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaximumMissedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaximumMissedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ProposerPercentageOfFeesOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetProposerPercentageOfFeesOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.ValidatorMaxEvidenceAgeInBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxEvidenceAgeInBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MissedBlocksBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMissedBlocksBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.DoubleSignBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetDoubleSignBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil

	case types.MessageSendFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageSendFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageFishermanPauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageFishermanPauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageProveTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageProveTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageChangeParameterFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageChangeParameterFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageSendFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageSendFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageFishermanPauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageFishermanPauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageProveTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageProveTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageEditStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnstakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessagePauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageUnpauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case types.MessageChangeParameterFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageChangeParameterFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	default:
		return types.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetBlocksPerSession() (int, types.Error) {
	store := u.Store()
	blocksPerSession, err := store.GetBlocksPerSession()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	appMininimumStake, err := store.GetParamAppMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(types.AppMinimumStakeParamName, err)
	}
	return StringToBigInt(appMininimumStake)
}

func (u *UtilityContext) GetAppMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetMaxAppChains()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, types.Error) {
	store := u.Store()
	baselineRate, err := store.GetBaselineAppStakeRate()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppBaselineStakeRateParamName, err)
	}
	return baselineRate, nil
}

func (u *UtilityContext) GetStakingAdjustment() (int, types.Error) {
	store := u.Store()
	adjustment, err := store.GetStakingAdjustment()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppStakingAdjustmentParamName, err)
	}
	return adjustment, nil
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetAppUnstakingBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetAppMinimumPauseBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetAppMaxPausedBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.AppMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	ServiceNodeMininimumStake, err := store.GetParamServiceNodeMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(types.ServiceNodeMinimumStakeParamName, err)
	}
	return StringToBigInt(ServiceNodeMininimumStake)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetServiceNodeMaxChains()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.ServiceNodeMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.ServiceNodeUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.ServiceNodeMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetServiceNodeMaxPausedBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.ServiceNodeMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	ValidatorMininimumStake, err := store.GetParamValidatorMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(types.ValidatorMinimumStakeParamName, err)
	}
	return StringToBigInt(ValidatorMininimumStake)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetValidatorUnstakingBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.ValidatorUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetValidatorMinimumPauseBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.ValidatorMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetValidatorMaxPausedBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.ValidatorMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err types.Error) {
	store := u.Store()
	proposerPercentage, er := store.GetProposerPercentageOfFees()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.ProposerPercentageOfFeesParamName, er)
	}
	return proposerPercentage, nil
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetValidatorMaximumMissedBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.ValidatorMaximumMissedBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetMaxEvidenceAgeInBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.ValidatorMaxEvidenceAgeInBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetDoubleSignBurnPercentage()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.DoubleSignBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetMissedBlocksBurnPercentage()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.MissedBlocksBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	FishermanMininimumStake, err := store.GetParamFishermanMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(types.FishermanMinimumStakeParamName, err)
	}
	return StringToBigInt(FishermanMininimumStake)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetFishermanMaxChains()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.FishermanMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetFishermanUnstakingBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.FishermanUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetFishermanMinimumPauseBlocks()
	if err != nil {
		return types.ZeroInt, types.ErrGetParam(types.FishermanMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetFishermanMaxPausedBlocks()
	if er != nil {
		return types.ZeroInt, types.ErrGetParam(types.FishermanMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageDoubleSignFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageDoubleSignFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageSendFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageSendFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageStakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageEditStakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnstakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessagePauseFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnpauseFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageFishermanPauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageFishermanPauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageTestScoreFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageProveTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageProveTestScoreFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageStakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageEditStakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnstakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessagePauseAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnpauseAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageStakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageEditStakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnstakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessagePauseValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnpauseValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageStakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageEditStakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnstakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessagePauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageUnpauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageChangeParameterFee()
	if er != nil {
		return nil, types.ErrGetParam(types.MessageChangeParameterFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	store := u.Store()
	switch paramName {
	case types.ACLOwner:
		return store.GetACLOwner()
	case types.BlocksPerSessionParamName:
		return store.GetBlocksPerSessionOwner()
	case types.AppMaxChainsParamName:
		return store.GetMaxAppChainsOwner()
	case types.AppMinimumStakeParamName:
		return store.GetAppMinimumStakeOwner()
	case types.AppBaselineStakeRateParamName:
		return store.GetBaselineAppOwner()
	case types.AppStakingAdjustmentParamName:
		return store.GetStakingAdjustmentOwner()
	case types.AppUnstakingBlocksParamName:
		return store.GetAppUnstakingBlocksOwner()
	case types.AppMinimumPauseBlocksParamName:
		return store.GetAppMinimumPauseBlocksOwner()
	case types.AppMaxPauseBlocksParamName:
		return store.GetAppMaxPausedBlocksOwner()
	case types.ServiceNodesPerSessionParamName:
		return store.GetServiceNodesPerSessionOwner()
	case types.ServiceNodeMinimumStakeParamName:
		return store.GetParamServiceNodeMinimumStakeOwner()
	case types.ServiceNodeMaxChainsParamName:
		return store.GetServiceNodeMaxChainsOwner()
	case types.ServiceNodeUnstakingBlocksParamName:
		return store.GetServiceNodeUnstakingBlocksOwner()
	case types.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetServiceNodeMinimumPauseBlocksOwner()
	case types.ServiceNodeMaxPauseBlocksParamName:
		return store.GetServiceNodeMaxPausedBlocksOwner()
	case types.FishermanMinimumStakeParamName:
		return store.GetFishermanMinimumStakeOwner()
	case types.FishermanMaxChainsParamName:
		return store.GetMaxFishermanChainsOwner()
	case types.FishermanUnstakingBlocksParamName:
		return store.GetFishermanUnstakingBlocksOwner()
	case types.FishermanMinimumPauseBlocksParamName:
		return store.GetFishermanMinimumPauseBlocksOwner()
	case types.FishermanMaxPauseBlocksParamName:
		return store.GetFishermanMaxPausedBlocksOwner()
	case types.ValidatorMinimumStakeParamName:
		return store.GetParamValidatorMinimumStakeOwner()
	case types.ValidatorUnstakingBlocksParamName:
		return store.GetValidatorUnstakingBlocksOwner()
	case types.ValidatorMinimumPauseBlocksParamName:
		return store.GetValidatorMinimumPauseBlocksOwner()
	case types.ValidatorMaxPauseBlocksParamName:
		return store.GetValidatorMaxPausedBlocksOwner()
	case types.ValidatorMaximumMissedBlocksParamName:
		return store.GetValidatorMaximumMissedBlocksOwner()
	case types.ProposerPercentageOfFeesParamName:
		return store.GetProposerPercentageOfFeesOwner()
	case types.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetMaxEvidenceAgeInBlocksOwner()
	case types.MissedBlocksBurnPercentageParamName:
		return store.GetMissedBlocksBurnPercentageOwner()
	case types.DoubleSignBurnPercentageParamName:
		return store.GetDoubleSignBurnPercentageOwner()
	case types.MessageDoubleSignFee:
		return store.GetMessageDoubleSignFeeOwner()
	case types.MessageSendFee:
		return store.GetMessageSendFeeOwner()
	case types.MessageStakeFishermanFee:
		return store.GetMessageStakeFishermanFeeOwner()
	case types.MessageEditStakeFishermanFee:
		return store.GetMessageEditStakeFishermanFeeOwner()
	case types.MessageUnstakeFishermanFee:
		return store.GetMessageUnstakeFishermanFeeOwner()
	case types.MessagePauseFishermanFee:
		return store.GetMessagePauseFishermanFeeOwner()
	case types.MessageUnpauseFishermanFee:
		return store.GetMessageUnpauseFishermanFeeOwner()
	case types.MessageFishermanPauseServiceNodeFee:
		return store.GetMessageFishermanPauseServiceNodeFeeOwner()
	case types.MessageTestScoreFee:
		return store.GetMessageTestScoreFeeOwner()
	case types.MessageProveTestScoreFee:
		return store.GetMessageProveTestScoreFeeOwner()
	case types.MessageStakeAppFee:
		return store.GetMessageStakeAppFeeOwner()
	case types.MessageEditStakeAppFee:
		return store.GetMessageEditStakeAppFeeOwner()
	case types.MessageUnstakeAppFee:
		return store.GetMessageUnstakeAppFeeOwner()
	case types.MessagePauseAppFee:
		return store.GetMessagePauseAppFeeOwner()
	case types.MessageUnpauseAppFee:
		return store.GetMessageUnpauseAppFeeOwner()
	case types.MessageStakeValidatorFee:
		return store.GetMessageStakeValidatorFeeOwner()
	case types.MessageEditStakeValidatorFee:
		return store.GetMessageEditStakeValidatorFeeOwner()
	case types.MessageUnstakeValidatorFee:
		return store.GetMessageUnstakeValidatorFeeOwner()
	case types.MessagePauseValidatorFee:
		return store.GetMessagePauseValidatorFeeOwner()
	case types.MessageUnpauseValidatorFee:
		return store.GetMessageUnpauseValidatorFeeOwner()
	case types.MessageStakeServiceNodeFee:
		return store.GetMessageStakeServiceNodeFeeOwner()
	case types.MessageEditStakeServiceNodeFee:
		return store.GetMessageEditStakeServiceNodeFeeOwner()
	case types.MessageUnstakeServiceNodeFee:
		return store.GetMessageUnstakeServiceNodeFeeOwner()
	case types.MessagePauseServiceNodeFee:
		return store.GetMessagePauseServiceNodeFeeOwner()
	case types.MessageUnpauseServiceNodeFee:
		return store.GetMessageUnpauseServiceNodeFeeOwner()
	case types.MessageChangeParameterFee:
		return store.GetMessageChangeParameterFeeOwner()
	case types.BlocksPerSessionOwner:
		return store.GetACLOwner()
	case types.AppMaxChainsOwner:
		return store.GetACLOwner()
	case types.AppMinimumStakeOwner:
		return store.GetACLOwner()
	case types.AppBaselineStakeRateOwner:
		return store.GetACLOwner()
	case types.AppStakingAdjustmentOwner:
		return store.GetACLOwner()
	case types.AppUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types.AppMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types.AppMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types.ServiceNodeMinimumStakeOwner:
		return store.GetACLOwner()
	case types.ServiceNodeMaxChainsOwner:
		return store.GetACLOwner()
	case types.ServiceNodeUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types.ServiceNodeMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types.ServiceNodesPerSessionOwner:
		return store.GetACLOwner()
	case types.ParamFishermanMinimumStakeOwner:
		return store.GetACLOwner()
	case types.FishermanMaxChainsOwner:
		return store.GetACLOwner()
	case types.FishermanUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types.FishermanMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types.FishermanMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types.ValidatorMinimumStakeOwner:
		return store.GetACLOwner()
	case types.ValidatorUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types.ValidatorMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types.ValidatorMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types.ValidatorMaximumMissedBlocksOwner:
		return store.GetACLOwner()
	case types.ProposerPercentageOfFeesOwner:
		return store.GetACLOwner()
	case types.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetACLOwner()
	case types.MissedBlocksBurnPercentageOwner:
		return store.GetACLOwner()
	case types.DoubleSignBurnPercentageOwner:
		return store.GetACLOwner()
	case types.MessageSendFeeOwner:
		return store.GetACLOwner()
	case types.MessageStakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types.MessageEditStakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnstakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types.MessagePauseFishermanFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnpauseFishermanFeeOwner:
		return store.GetACLOwner()
	case types.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessageTestScoreFeeOwner:
		return store.GetACLOwner()
	case types.MessageProveTestScoreFeeOwner:
		return store.GetACLOwner()
	case types.MessageStakeAppFeeOwner:
		return store.GetACLOwner()
	case types.MessageEditStakeAppFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnstakeAppFeeOwner:
		return store.GetACLOwner()
	case types.MessagePauseAppFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnpauseAppFeeOwner:
		return store.GetACLOwner()
	case types.MessageStakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types.MessageEditStakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnstakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types.MessagePauseValidatorFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnpauseValidatorFeeOwner:
		return store.GetACLOwner()
	case types.MessageStakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessageEditStakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnstakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessagePauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessageUnpauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types.MessageChangeParameterFeeOwner:
		return store.GetACLOwner()
	default:
		return nil, types.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg types.Message) (amount *big.Int, err types.Error) {
	switch x := msg.(type) {
	case *types.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *types.MessageSend:
		return u.GetMessageSendFee()
	case *types.MessageStakeFisherman:
		return u.GetMessageStakeFishermanFee()
	case *types.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanFee()
	case *types.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanFee()
	case *types.MessagePauseFisherman:
		return u.GetMessagePauseFishermanFee()
	case *types.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanFee()
	case *types.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeFee()
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreFee()
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreFee()
	case *types.MessageStakeApp:
		return u.GetMessageStakeAppFee()
	case *types.MessageEditStakeApp:
		return u.GetMessageEditStakeAppFee()
	case *types.MessageUnstakeApp:
		return u.GetMessageUnstakeAppFee()
	case *types.MessagePauseApp:
		return u.GetMessagePauseAppFee()
	case *types.MessageUnpauseApp:
		return u.GetMessageUnpauseAppFee()
	case *types.MessageStakeValidator:
		return u.GetMessageStakeValidatorFee()
	case *types.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorFee()
	case *types.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorFee()
	case *types.MessagePauseValidator:
		return u.GetMessagePauseValidatorFee()
	case *types.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorFee()
	case *types.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeFee()
	case *types.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeFee()
	case *types.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeFee()
	case *types.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeFee()
	case *types.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeFee()
	case *types.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *types.MessageChangeParameter) ([][]byte, types.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, types.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}
