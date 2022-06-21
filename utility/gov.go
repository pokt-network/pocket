package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *UtilityContext) HandleMessageChangeParameter(message *typesUtil.MessageChangeParameter) types.Error {
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
	case typesUtil.BlocksPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodesPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxAppChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamAppMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppBaselineStakeRateParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBaselineAppStakeRate(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppStabilityAdjustmentParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetStakingAdjustment(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamServiceNodeMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamFishermanMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamValidatorMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaximumMissedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ProposerPercentageOfFeesParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetProposerPercentageOfFees(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxEvidenceAgeInBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MissedBlocksBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMissedBlocksBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.DoubleSignBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetDoubleSignBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AclOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAclOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.BlocksPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodesPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxAppChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppBaselineStakeRateOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetBaselineAppOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppStakingAdjustmentOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetStakingAdjustmentOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.AppMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamServiceNodeMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxServiceNodeChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ServiceNodeMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxFishermanChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.FishermanMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamValidatorMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaximumMissedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ProposerPercentageOfFeesOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetProposerPercentageOfFeesOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxEvidenceAgeInBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MissedBlocksBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMissedBlocksBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.DoubleSignBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetDoubleSignBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil

	case typesUtil.MessageSendFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageSendFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageFishermanPauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageFishermanPauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageProveTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageProveTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageChangeParameterFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageChangeParameterFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageSendFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageSendFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageFishermanPauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageFishermanPauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageProveTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageProveTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageEditStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnstakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessagePauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageUnpauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageChangeParameterFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageChangeParameterFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case typesUtil.MessageDoubleSignFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageDoubleSignFeeOwner(owner.Value)
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
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	appMininimumStake, err := store.GetParamAppMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(typesUtil.AppMinimumStakeParamName, err)
	}
	return types.StringToBigInt(appMininimumStake)
}

func (u *UtilityContext) GetAppMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetMaxAppChains()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, types.Error) {
	store := u.Store()
	baselineRate, err := store.GetBaselineAppStakeRate()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppBaselineStakeRateParamName, err)
	}
	return baselineRate, nil
}

func (u *UtilityContext) GetStabilityAdjustment() (int, types.Error) {
	store := u.Store()
	adjustment, err := store.GetStabilityAdjustment()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppStabilityAdjustmentParamName, err)
	}
	return adjustment, nil
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetAppUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetAppMinimumPauseBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetAppMaxPausedBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.AppMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	serviceNodeMininimumStake, err := store.GetParamServiceNodeMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(typesUtil.ServiceNodeMinimumStakeParamName, err)
	}
	return types.StringToBigInt(serviceNodeMininimumStake)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetServiceNodeMaxChains()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ServiceNodeMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ServiceNodeUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetServiceNodeMaxPausedBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	validatorMininimumStake, err := store.GetParamValidatorMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(typesUtil.ValidatorMinimumStakeParamName, err)
	}
	return types.StringToBigInt(validatorMininimumStake)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetValidatorUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ValidatorUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetValidatorMinimumPauseBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ValidatorMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetValidatorMaxPausedBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ValidatorMaxPausedBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err types.Error) {
	store := u.Store()
	proposerPercentage, er := store.GetProposerPercentageOfFees()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ProposerPercentageOfFeesParamName, er)
	}
	return proposerPercentage, nil
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetValidatorMaximumMissedBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ValidatorMaximumMissedBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetMaxEvidenceAgeInBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetDoubleSignBurnPercentage()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.DoubleSignBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetMissedBlocksBurnPercentage()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.MissedBlocksBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	FishermanMininimumStake, err := store.GetParamFishermanMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(typesUtil.FishermanMinimumStakeParamName, err)
	}
	return types.StringToBigInt(FishermanMininimumStake)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetFishermanMaxChains()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.FishermanMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetFishermanUnstakingBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.FishermanUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetFishermanMinimumPauseBlocks()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.FishermanMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetFishermanMaxPausedBlocks()
	if er != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(typesUtil.FishermanMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageDoubleSignFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageDoubleSignFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageSendFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageSendFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageStakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageEditStakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnstakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessagePauseFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnpauseFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageFishermanPauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageFishermanPauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageTestScoreFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageProveTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageProveTestScoreFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageStakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageEditStakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnstakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessagePauseAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnpauseAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageStakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageEditStakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnstakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessagePauseValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnpauseValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageStakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageEditStakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnstakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessagePauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageUnpauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageChangeParameterFee()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.MessageChangeParameterFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetDoubleSignFeeOwner() (owner []byte, err types.Error) {
	store := u.Store()
	owner, er := store.GetMessageDoubleSignFeeOwner()
	if er != nil {
		return nil, types.ErrGetParam(typesUtil.DoubleSignBurnPercentageParamName, er)
	}
	return owner, nil
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	store := u.Store()
	switch paramName {
	case typesUtil.AclOwner:
		return store.GetAclOwner()
	case typesUtil.BlocksPerSessionParamName:
		return store.GetBlocksPerSessionOwner()
	case typesUtil.AppMaxChainsParamName:
		return store.GetMaxAppChainsOwner()
	case typesUtil.AppMinimumStakeParamName:
		return store.GetAppMinimumStakeOwner()
	case typesUtil.AppBaselineStakeRateParamName:
		return store.GetBaselineAppOwner()
	case typesUtil.AppStabilityAdjustmentParamName:
		return store.GetStakingAdjustmentOwner()
	case typesUtil.AppUnstakingBlocksParamName:
		return store.GetAppUnstakingBlocksOwner()
	case typesUtil.AppMinimumPauseBlocksParamName:
		return store.GetAppMinimumPauseBlocksOwner()
	case typesUtil.AppMaxPauseBlocksParamName:
		return store.GetAppMaxPausedBlocksOwner()
	case typesUtil.ServiceNodesPerSessionParamName:
		return store.GetServiceNodesPerSessionOwner()
	case typesUtil.ServiceNodeMinimumStakeParamName:
		return store.GetParamServiceNodeMinimumStakeOwner()
	case typesUtil.ServiceNodeMaxChainsParamName:
		return store.GetServiceNodeMaxChainsOwner()
	case typesUtil.ServiceNodeUnstakingBlocksParamName:
		return store.GetServiceNodeUnstakingBlocksOwner()
	case typesUtil.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetServiceNodeMinimumPauseBlocksOwner()
	case typesUtil.ServiceNodeMaxPauseBlocksParamName:
		return store.GetServiceNodeMaxPausedBlocksOwner()
	case typesUtil.FishermanMinimumStakeParamName:
		return store.GetFishermanMinimumStakeOwner()
	case typesUtil.FishermanMaxChainsParamName:
		return store.GetMaxFishermanChainsOwner()
	case typesUtil.FishermanUnstakingBlocksParamName:
		return store.GetFishermanUnstakingBlocksOwner()
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		return store.GetFishermanMinimumPauseBlocksOwner()
	case typesUtil.FishermanMaxPauseBlocksParamName:
		return store.GetFishermanMaxPausedBlocksOwner()
	case typesUtil.ValidatorMinimumStakeParamName:
		return store.GetParamValidatorMinimumStakeOwner()
	case typesUtil.ValidatorUnstakingBlocksParamName:
		return store.GetValidatorUnstakingBlocksOwner()
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		return store.GetValidatorMinimumPauseBlocksOwner()
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		return store.GetValidatorMaxPausedBlocksOwner()
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		return store.GetValidatorMaximumMissedBlocksOwner()
	case typesUtil.ProposerPercentageOfFeesParamName:
		return store.GetProposerPercentageOfFeesOwner()
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetMaxEvidenceAgeInBlocksOwner()
	case typesUtil.MissedBlocksBurnPercentageParamName:
		return store.GetMissedBlocksBurnPercentageOwner()
	case typesUtil.DoubleSignBurnPercentageParamName:
		return store.GetDoubleSignBurnPercentageOwner()
	case typesUtil.MessageDoubleSignFee:
		return store.GetMessageDoubleSignFeeOwner()
	case typesUtil.MessageSendFee:
		return store.GetMessageSendFeeOwner()
	case typesUtil.MessageStakeFishermanFee:
		return store.GetMessageStakeFishermanFeeOwner()
	case typesUtil.MessageEditStakeFishermanFee:
		return store.GetMessageEditStakeFishermanFeeOwner()
	case typesUtil.MessageUnstakeFishermanFee:
		return store.GetMessageUnstakeFishermanFeeOwner()
	case typesUtil.MessagePauseFishermanFee:
		return store.GetMessagePauseFishermanFeeOwner()
	case typesUtil.MessageUnpauseFishermanFee:
		return store.GetMessageUnpauseFishermanFeeOwner()
	case typesUtil.MessageFishermanPauseServiceNodeFee:
		return store.GetMessageFishermanPauseServiceNodeFeeOwner()
	case typesUtil.MessageTestScoreFee:
		return store.GetMessageTestScoreFeeOwner()
	case typesUtil.MessageProveTestScoreFee:
		return store.GetMessageProveTestScoreFeeOwner()
	case typesUtil.MessageStakeAppFee:
		return store.GetMessageStakeAppFeeOwner()
	case typesUtil.MessageEditStakeAppFee:
		return store.GetMessageEditStakeAppFeeOwner()
	case typesUtil.MessageUnstakeAppFee:
		return store.GetMessageUnstakeAppFeeOwner()
	case typesUtil.MessagePauseAppFee:
		return store.GetMessagePauseAppFeeOwner()
	case typesUtil.MessageUnpauseAppFee:
		return store.GetMessageUnpauseAppFeeOwner()
	case typesUtil.MessageStakeValidatorFee:
		return store.GetMessageStakeValidatorFeeOwner()
	case typesUtil.MessageEditStakeValidatorFee:
		return store.GetMessageEditStakeValidatorFeeOwner()
	case typesUtil.MessageUnstakeValidatorFee:
		return store.GetMessageUnstakeValidatorFeeOwner()
	case typesUtil.MessagePauseValidatorFee:
		return store.GetMessagePauseValidatorFeeOwner()
	case typesUtil.MessageUnpauseValidatorFee:
		return store.GetMessageUnpauseValidatorFeeOwner()
	case typesUtil.MessageStakeServiceNodeFee:
		return store.GetMessageStakeServiceNodeFeeOwner()
	case typesUtil.MessageEditStakeServiceNodeFee:
		return store.GetMessageEditStakeServiceNodeFeeOwner()
	case typesUtil.MessageUnstakeServiceNodeFee:
		return store.GetMessageUnstakeServiceNodeFeeOwner()
	case typesUtil.MessagePauseServiceNodeFee:
		return store.GetMessagePauseServiceNodeFeeOwner()
	case typesUtil.MessageUnpauseServiceNodeFee:
		return store.GetMessageUnpauseServiceNodeFeeOwner()
	case typesUtil.MessageChangeParameterFee:
		return store.GetMessageChangeParameterFeeOwner()
	case typesUtil.BlocksPerSessionOwner:
		return store.GetAclOwner()
	case typesUtil.AppMaxChainsOwner:
		return store.GetAclOwner()
	case typesUtil.AppMinimumStakeOwner:
		return store.GetAclOwner()
	case typesUtil.AppBaselineStakeRateOwner:
		return store.GetAclOwner()
	case typesUtil.AppStakingAdjustmentOwner:
		return store.GetAclOwner()
	case typesUtil.AppUnstakingBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.AppMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.AppMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodeMinimumStakeOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodeMaxChainsOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodeUnstakingBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodeMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ServiceNodesPerSessionOwner:
		return store.GetAclOwner()
	case typesUtil.FishermanMinimumStakeOwner:
		return store.GetAclOwner()
	case typesUtil.FishermanMaxChainsOwner:
		return store.GetAclOwner()
	case typesUtil.FishermanUnstakingBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.FishermanMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorMinimumStakeOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorUnstakingBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.ProposerPercentageOfFeesOwner:
		return store.GetAclOwner()
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetAclOwner()
	case typesUtil.MissedBlocksBurnPercentageOwner:
		return store.GetAclOwner()
	case typesUtil.DoubleSignBurnPercentageOwner:
		return store.GetAclOwner()
	case typesUtil.MessageSendFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageStakeFishermanFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessagePauseFishermanFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageTestScoreFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageProveTestScoreFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageStakeAppFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageEditStakeAppFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnstakeAppFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessagePauseAppFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnpauseAppFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageStakeValidatorFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessagePauseValidatorFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageStakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageEditStakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnstakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessagePauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageUnpauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case typesUtil.MessageChangeParameterFeeOwner:
		return store.GetAclOwner()
	default:
		return nil, types.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg typesUtil.Message) (amount *big.Int, err types.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *typesUtil.MessageSend:
		return u.GetMessageSendFee()
	case *typesUtil.MessageStakeFisherman:
		return u.GetMessageStakeFishermanFee()
	case *typesUtil.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanFee()
	case *typesUtil.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanFee()
	case *typesUtil.MessagePauseFisherman:
		return u.GetMessagePauseFishermanFee()
	case *typesUtil.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanFee()
	case *typesUtil.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeFee()
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreFee()
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreFee()
	case *typesUtil.MessageStakeApp:
		return u.GetMessageStakeAppFee()
	case *typesUtil.MessageEditStakeApp:
		return u.GetMessageEditStakeAppFee()
	case *typesUtil.MessageUnstakeApp:
		return u.GetMessageUnstakeAppFee()
	case *typesUtil.MessagePauseApp:
		return u.GetMessagePauseAppFee()
	case *typesUtil.MessageUnpauseApp:
		return u.GetMessageUnpauseAppFee()
	case *typesUtil.MessageStakeValidator:
		return u.GetMessageStakeValidatorFee()
	case *typesUtil.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorFee()
	case *typesUtil.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorFee()
	case *typesUtil.MessagePauseValidator:
		return u.GetMessagePauseValidatorFee()
	case *typesUtil.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorFee()
	case *typesUtil.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeFee()
	case *typesUtil.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeFee()
	case *typesUtil.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeFee()
	case *typesUtil.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeFee()
	case *typesUtil.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeFee()
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *typesUtil.MessageChangeParameter) ([][]byte, types.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, types.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}
