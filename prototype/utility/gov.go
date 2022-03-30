package utility

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
	"math/big"
	types2 "pocket/utility/types"
)

func (u *UtilityContext) HandleMessageChangeParameter(message *types2.MessageChangeParameter) types2.Error {
	cdc := u.Codec()
	v, err := cdc.FromAny(message.ParameterValue)
	if err != nil {
		return types2.ErrProtoFromAny(err)
	}
	return u.UpdateParam(message.ParameterKey, v)
}

func (u *UtilityContext) UpdateParam(paramName string, value interface{}) types2.Error {
	store := u.Store()
	switch paramName {
	case types2.BlocksPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSession(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodesPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSession(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxAppChains(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamAppMinimumStake(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppBaselineStakeRateParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetBaselineAppStakeRate(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppStakingAdjustmentParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetStakingAdjustment(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppUnstakingBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamServiceNodeMinimumStake(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxChains(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeUnstakingBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamFishermanMinimumStake(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxChains(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanUnstakingBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamValidatorMinimumStake(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorUnstakingBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaximumMissedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaximumMissedBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ProposerPercentageOfFeesParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetProposerPercentageOfFees(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaxEvidenceAgeInBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxEvidenceAgeInBlocks(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MissedBlocksBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMissedBlocksBurnPercentage(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.DoubleSignBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetDoubleSignBurnPercentage(int(i.Value))
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ACLOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetACLOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.BlocksPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSessionOwner(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodesPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSessionOwner(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxAppChainsOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumStakeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppBaselineStakeRateOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetBaselineAppOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppStakingAdjustmentOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetStakingAdjustmentOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.AppMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamServiceNodeMinimumStakeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxServiceNodeChainsOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ServiceNodeMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ParamFishermanMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumStakeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxFishermanChainsOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.FishermanMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamValidatorMinimumStakeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaximumMissedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaximumMissedBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ProposerPercentageOfFeesOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetProposerPercentageOfFeesOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.ValidatorMaxEvidenceAgeInBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxEvidenceAgeInBlocksOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MissedBlocksBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMissedBlocksBurnPercentageOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.DoubleSignBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetDoubleSignBurnPercentageOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil

	case types2.MessageSendFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageSendFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageFishermanPauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageFishermanPauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageProveTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageProveTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeAppFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseAppFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseAppFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageChangeParameterFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageChangeParameterFeeOwner(owner.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageSendFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageSendFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeFishermanFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeFishermanFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeFishermanFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseFishermanFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseFishermanFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageFishermanPauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageFishermanPauseServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageTestScoreFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageProveTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageProveTestScoreFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeAppFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeAppFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeAppFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseAppFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseAppFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeValidatorFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeValidatorFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeValidatorFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseValidatorFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseValidatorFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageEditStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnstakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessagePauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageUnpauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseServiceNodeFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	case types2.MessageChangeParameterFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types2.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageChangeParameterFee(i.Value)
		if err != nil {
			return types2.ErrUpdateParam(err)
		}
		return nil
	default:
		return types2.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetBlocksPerSession() (int, types2.Error) {
	store := u.Store()
	blocksPerSession, err := store.GetBlocksPerSession()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, types2.Error) {
	store := u.Store()
	appMininimumStake, err := store.GetParamAppMinimumStake()
	if err != nil {
		return nil, types2.ErrGetParam(types2.AppMinimumStakeParamName, err)
	}
	return StringToBigInt(appMininimumStake)
}

func (u *UtilityContext) GetAppMaxChains() (int, types2.Error) {
	store := u.Store()
	maxChains, err := store.GetMaxAppChains()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, types2.Error) {
	store := u.Store()
	baselineRate, err := store.GetBaselineAppStakeRate()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppBaselineStakeRateParamName, err)
	}
	return baselineRate, nil
}

func (u *UtilityContext) GetStakingAdjustment() (int, types2.Error) {
	store := u.Store()
	adjustment, err := store.GetStakingAdjustment()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppStakingAdjustmentParamName, err)
	}
	return adjustment, nil
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, types2.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetAppUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, types2.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetAppMinimumPauseBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err types2.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetAppMaxPausedBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.AppMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, types2.Error) {
	store := u.Store()
	ServiceNodeMininimumStake, err := store.GetParamServiceNodeMinimumStake()
	if err != nil {
		return nil, types2.ErrGetParam(types2.ServiceNodeMinimumStakeParamName, err)
	}
	return StringToBigInt(ServiceNodeMininimumStake)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, types2.Error) {
	store := u.Store()
	maxChains, err := store.GetServiceNodeMaxChains()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ServiceNodeMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, types2.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ServiceNodeUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, types2.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ServiceNodeMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err types2.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetServiceNodeMaxPausedBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ServiceNodeMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, types2.Error) {
	store := u.Store()
	ValidatorMininimumStake, err := store.GetParamValidatorMinimumStake()
	if err != nil {
		return nil, types2.ErrGetParam(types2.ValidatorMinimumStakeParamName, err)
	}
	return StringToBigInt(ValidatorMininimumStake)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, types2.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetValidatorUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ValidatorUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, types2.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetValidatorMinimumPauseBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ValidatorMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err types2.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetValidatorMaxPausedBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ValidatorMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err types2.Error) {
	store := u.Store()
	proposerPercentage, er := store.GetProposerPercentageOfFees()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ProposerPercentageOfFeesParamName, er)
	}
	return proposerPercentage, nil
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err types2.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetValidatorMaximumMissedBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ValidatorMaximumMissedBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err types2.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetMaxEvidenceAgeInBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.ValidatorMaxEvidenceAgeInBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err types2.Error) {
	store := u.Store()
	burnPercentage, er := store.GetDoubleSignBurnPercentage()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.DoubleSignBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err types2.Error) {
	store := u.Store()
	burnPercentage, er := store.GetMissedBlocksBurnPercentage()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.MissedBlocksBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, types2.Error) {
	store := u.Store()
	FishermanMininimumStake, err := store.GetParamFishermanMinimumStake()
	if err != nil {
		return nil, types2.ErrGetParam(types2.FishermanMinimumStakeParamName, err)
	}
	return StringToBigInt(FishermanMininimumStake)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, types2.Error) {
	store := u.Store()
	maxChains, err := store.GetFishermanMaxChains()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.FishermanMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, types2.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetFishermanUnstakingBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.FishermanUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, types2.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetFishermanMinimumPauseBlocks()
	if err != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.FishermanMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err types2.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetFishermanMaxPausedBlocks()
	if er != nil {
		return types2.ZeroInt, types2.ErrGetParam(types2.FishermanMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageDoubleSignFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageDoubleSignFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageSendFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageSendFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeFishermanFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageStakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeFishermanFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageEditStakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeFishermanFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnstakeFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseFishermanFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessagePauseFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseFishermanFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnpauseFishermanFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageFishermanPauseServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageFishermanPauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageTestScoreFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageTestScoreFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageProveTestScoreFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageProveTestScoreFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeAppFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageStakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeAppFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageEditStakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeAppFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnstakeAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseAppFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessagePauseAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseAppFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnpauseAppFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeValidatorFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageStakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeValidatorFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageEditStakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeValidatorFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnstakeValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseValidatorFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessagePauseValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseValidatorFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnpauseValidatorFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageStakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageEditStakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnstakeServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessagePauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseServiceNodeFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageUnpauseServiceNodeFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, types2.Error) {
	store := u.Store()
	fee, er := store.GetMessageChangeParameterFee()
	if er != nil {
		return nil, types2.ErrGetParam(types2.MessageChangeParameterFee, er)
	}
	return StringToBigInt(fee)
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	store := u.Store()
	switch paramName {
	case types2.ACLOwner:
		return store.GetACLOwner()
	case types2.BlocksPerSessionParamName:
		return store.GetBlocksPerSessionOwner()
	case types2.AppMaxChainsParamName:
		return store.GetMaxAppChainsOwner()
	case types2.AppMinimumStakeParamName:
		return store.GetAppMinimumStakeOwner()
	case types2.AppBaselineStakeRateParamName:
		return store.GetBaselineAppOwner()
	case types2.AppStakingAdjustmentParamName:
		return store.GetStakingAdjustmentOwner()
	case types2.AppUnstakingBlocksParamName:
		return store.GetAppUnstakingBlocksOwner()
	case types2.AppMinimumPauseBlocksParamName:
		return store.GetAppMinimumPauseBlocksOwner()
	case types2.AppMaxPauseBlocksParamName:
		return store.GetAppMaxPausedBlocksOwner()
	case types2.ServiceNodesPerSessionParamName:
		return store.GetServiceNodesPerSessionOwner()
	case types2.ServiceNodeMinimumStakeParamName:
		return store.GetParamServiceNodeMinimumStakeOwner()
	case types2.ServiceNodeMaxChainsParamName:
		return store.GetServiceNodeMaxChainsOwner()
	case types2.ServiceNodeUnstakingBlocksParamName:
		return store.GetServiceNodeUnstakingBlocksOwner()
	case types2.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetServiceNodeMinimumPauseBlocksOwner()
	case types2.ServiceNodeMaxPauseBlocksParamName:
		return store.GetServiceNodeMaxPausedBlocksOwner()
	case types2.FishermanMinimumStakeParamName:
		return store.GetFishermanMinimumStakeOwner()
	case types2.FishermanMaxChainsParamName:
		return store.GetMaxFishermanChainsOwner()
	case types2.FishermanUnstakingBlocksParamName:
		return store.GetFishermanUnstakingBlocksOwner()
	case types2.FishermanMinimumPauseBlocksParamName:
		return store.GetFishermanMinimumPauseBlocksOwner()
	case types2.FishermanMaxPauseBlocksParamName:
		return store.GetFishermanMaxPausedBlocksOwner()
	case types2.ValidatorMinimumStakeParamName:
		return store.GetParamValidatorMinimumStakeOwner()
	case types2.ValidatorUnstakingBlocksParamName:
		return store.GetValidatorUnstakingBlocksOwner()
	case types2.ValidatorMinimumPauseBlocksParamName:
		return store.GetValidatorMinimumPauseBlocksOwner()
	case types2.ValidatorMaxPauseBlocksParamName:
		return store.GetValidatorMaxPausedBlocksOwner()
	case types2.ValidatorMaximumMissedBlocksParamName:
		return store.GetValidatorMaximumMissedBlocksOwner()
	case types2.ProposerPercentageOfFeesParamName:
		return store.GetProposerPercentageOfFeesOwner()
	case types2.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetMaxEvidenceAgeInBlocksOwner()
	case types2.MissedBlocksBurnPercentageParamName:
		return store.GetMissedBlocksBurnPercentageOwner()
	case types2.DoubleSignBurnPercentageParamName:
		return store.GetDoubleSignBurnPercentageOwner()
	case types2.MessageDoubleSignFee:
		return store.GetMessageDoubleSignFeeOwner()
	case types2.MessageSendFee:
		return store.GetMessageSendFeeOwner()
	case types2.MessageStakeFishermanFee:
		return store.GetMessageStakeFishermanFeeOwner()
	case types2.MessageEditStakeFishermanFee:
		return store.GetMessageEditStakeFishermanFeeOwner()
	case types2.MessageUnstakeFishermanFee:
		return store.GetMessageUnstakeFishermanFeeOwner()
	case types2.MessagePauseFishermanFee:
		return store.GetMessagePauseFishermanFeeOwner()
	case types2.MessageUnpauseFishermanFee:
		return store.GetMessageUnpauseFishermanFeeOwner()
	case types2.MessageFishermanPauseServiceNodeFee:
		return store.GetMessageFishermanPauseServiceNodeFeeOwner()
	case types2.MessageTestScoreFee:
		return store.GetMessageTestScoreFeeOwner()
	case types2.MessageProveTestScoreFee:
		return store.GetMessageProveTestScoreFeeOwner()
	case types2.MessageStakeAppFee:
		return store.GetMessageStakeAppFeeOwner()
	case types2.MessageEditStakeAppFee:
		return store.GetMessageEditStakeAppFeeOwner()
	case types2.MessageUnstakeAppFee:
		return store.GetMessageUnstakeAppFeeOwner()
	case types2.MessagePauseAppFee:
		return store.GetMessagePauseAppFeeOwner()
	case types2.MessageUnpauseAppFee:
		return store.GetMessageUnpauseAppFeeOwner()
	case types2.MessageStakeValidatorFee:
		return store.GetMessageStakeValidatorFeeOwner()
	case types2.MessageEditStakeValidatorFee:
		return store.GetMessageEditStakeValidatorFeeOwner()
	case types2.MessageUnstakeValidatorFee:
		return store.GetMessageUnstakeValidatorFeeOwner()
	case types2.MessagePauseValidatorFee:
		return store.GetMessagePauseValidatorFeeOwner()
	case types2.MessageUnpauseValidatorFee:
		return store.GetMessageUnpauseValidatorFeeOwner()
	case types2.MessageStakeServiceNodeFee:
		return store.GetMessageStakeServiceNodeFeeOwner()
	case types2.MessageEditStakeServiceNodeFee:
		return store.GetMessageEditStakeServiceNodeFeeOwner()
	case types2.MessageUnstakeServiceNodeFee:
		return store.GetMessageUnstakeServiceNodeFeeOwner()
	case types2.MessagePauseServiceNodeFee:
		return store.GetMessagePauseServiceNodeFeeOwner()
	case types2.MessageUnpauseServiceNodeFee:
		return store.GetMessageUnpauseServiceNodeFeeOwner()
	case types2.MessageChangeParameterFee:
		return store.GetMessageChangeParameterFeeOwner()
	case types2.BlocksPerSessionOwner:
		return store.GetACLOwner()
	case types2.AppMaxChainsOwner:
		return store.GetACLOwner()
	case types2.AppMinimumStakeOwner:
		return store.GetACLOwner()
	case types2.AppBaselineStakeRateOwner:
		return store.GetACLOwner()
	case types2.AppStakingAdjustmentOwner:
		return store.GetACLOwner()
	case types2.AppUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types2.AppMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types2.AppMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types2.ServiceNodeMinimumStakeOwner:
		return store.GetACLOwner()
	case types2.ServiceNodeMaxChainsOwner:
		return store.GetACLOwner()
	case types2.ServiceNodeUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types2.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types2.ServiceNodeMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types2.ServiceNodesPerSessionOwner:
		return store.GetACLOwner()
	case types2.ParamFishermanMinimumStakeOwner:
		return store.GetACLOwner()
	case types2.FishermanMaxChainsOwner:
		return store.GetACLOwner()
	case types2.FishermanUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types2.FishermanMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types2.FishermanMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types2.ValidatorMinimumStakeOwner:
		return store.GetACLOwner()
	case types2.ValidatorUnstakingBlocksOwner:
		return store.GetACLOwner()
	case types2.ValidatorMinimumPauseBlocksOwner:
		return store.GetACLOwner()
	case types2.ValidatorMaxPausedBlocksOwner:
		return store.GetACLOwner()
	case types2.ValidatorMaximumMissedBlocksOwner:
		return store.GetACLOwner()
	case types2.ProposerPercentageOfFeesOwner:
		return store.GetACLOwner()
	case types2.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetACLOwner()
	case types2.MissedBlocksBurnPercentageOwner:
		return store.GetACLOwner()
	case types2.DoubleSignBurnPercentageOwner:
		return store.GetACLOwner()
	case types2.MessageSendFeeOwner:
		return store.GetACLOwner()
	case types2.MessageStakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types2.MessageEditStakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnstakeFishermanFeeOwner:
		return store.GetACLOwner()
	case types2.MessagePauseFishermanFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnpauseFishermanFeeOwner:
		return store.GetACLOwner()
	case types2.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessageTestScoreFeeOwner:
		return store.GetACLOwner()
	case types2.MessageProveTestScoreFeeOwner:
		return store.GetACLOwner()
	case types2.MessageStakeAppFeeOwner:
		return store.GetACLOwner()
	case types2.MessageEditStakeAppFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnstakeAppFeeOwner:
		return store.GetACLOwner()
	case types2.MessagePauseAppFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnpauseAppFeeOwner:
		return store.GetACLOwner()
	case types2.MessageStakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types2.MessageEditStakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnstakeValidatorFeeOwner:
		return store.GetACLOwner()
	case types2.MessagePauseValidatorFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnpauseValidatorFeeOwner:
		return store.GetACLOwner()
	case types2.MessageStakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessageEditStakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnstakeServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessagePauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessageUnpauseServiceNodeFeeOwner:
		return store.GetACLOwner()
	case types2.MessageChangeParameterFeeOwner:
		return store.GetACLOwner()
	default:
		return nil, types2.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg types2.Message) (amount *big.Int, err types2.Error) {
	switch x := msg.(type) {
	case *types2.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *types2.MessageSend:
		return u.GetMessageSendFee()
	case *types2.MessageStakeFisherman:
		return u.GetMessageStakeFishermanFee()
	case *types2.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanFee()
	case *types2.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanFee()
	case *types2.MessagePauseFisherman:
		return u.GetMessagePauseFishermanFee()
	case *types2.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanFee()
	case *types2.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeFee()
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreFee()
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreFee()
	case *types2.MessageStakeApp:
		return u.GetMessageStakeAppFee()
	case *types2.MessageEditStakeApp:
		return u.GetMessageEditStakeAppFee()
	case *types2.MessageUnstakeApp:
		return u.GetMessageUnstakeAppFee()
	case *types2.MessagePauseApp:
		return u.GetMessagePauseAppFee()
	case *types2.MessageUnpauseApp:
		return u.GetMessageUnpauseAppFee()
	case *types2.MessageStakeValidator:
		return u.GetMessageStakeValidatorFee()
	case *types2.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorFee()
	case *types2.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorFee()
	case *types2.MessagePauseValidator:
		return u.GetMessagePauseValidatorFee()
	case *types2.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorFee()
	case *types2.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeFee()
	case *types2.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeFee()
	case *types2.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeFee()
	case *types2.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeFee()
	case *types2.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeFee()
	case *types2.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, types2.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *types2.MessageChangeParameter) ([][]byte, types2.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, types2.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}
