package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *UtilityContext) HandleMessageChangeParameter(message *utilTypes.MessageChangeParameter) types.Error {
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
	case utilTypes.BlocksPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodesPerSessionParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSession(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxAppChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamAppMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppBaselineStakeRateParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBaselineAppStakeRate(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppStakingAdjustmentParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetStakingAdjustment(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetAppMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamServiceNodeMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodeMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamFishermanMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMaxChainsParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxChains(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMaxPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetFishermanMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMinimumStakeParamName:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetParamValidatorMinimumStake(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorUnstakingBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorUnstakingBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMinimumPauseBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMinimumPauseBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaxPausedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaxPausedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaximumMissedBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetValidatorMaximumMissedBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ProposerPercentageOfFeesParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetProposerPercentageOfFees(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaxEvidenceAgeInBlocksParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMaxEvidenceAgeInBlocks(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MissedBlocksBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMissedBlocksBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.DoubleSignBurnPercentageParamName:
		i, ok := value.(*wrapperspb.Int32Value)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetDoubleSignBurnPercentage(int(i.Value))
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AclOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAclOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.BlocksPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetBlocksPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodesPerSessionOwner:
		i, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetServiceNodesPerSessionOwner(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxAppChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppBaselineStakeRateOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetBaselineAppOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppStakingAdjustmentOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetStakingAdjustmentOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.AppMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetAppMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamServiceNodeMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxServiceNodeChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ServiceNodeMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetServiceNodeMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMaxChainsOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxFishermanChainsOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.FishermanMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetFishermanMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMinimumStakeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetParamValidatorMinimumStakeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorUnstakingBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorUnstakingBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMinimumPauseBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMinimumPauseBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaxPausedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaxPausedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaximumMissedBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetValidatorMaximumMissedBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ProposerPercentageOfFeesOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetProposerPercentageOfFeesOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.ValidatorMaxEvidenceAgeInBlocksOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMaxEvidenceAgeInBlocksOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MissedBlocksBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMissedBlocksBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.DoubleSignBurnPercentageOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetDoubleSignBurnPercentageOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil

	case utilTypes.MessageSendFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageSendFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseFishermanFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseFishermanFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageFishermanPauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageFishermanPauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageProveTestScoreFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageProveTestScoreFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseAppFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseAppFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseValidatorFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseValidatorFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageEditStakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnstakeServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessagePauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseServiceNodeFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageUnpauseServiceNodeFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageChangeParameterFeeOwner:
		owner, ok := value.(*wrapperspb.BytesValue)
		if !ok {
			return types.ErrInvalidParamValue(value, owner)
		}
		err := store.SetMessageChangeParameterFeeOwner(owner.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageSendFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageSendFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseFishermanFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseFishermanFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageFishermanPauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageFishermanPauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageProveTestScoreFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageProveTestScoreFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseAppFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseAppFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseValidatorFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseValidatorFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageEditStakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageEditStakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnstakeServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnstakeServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessagePauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessagePauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageUnpauseServiceNodeFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageUnpauseServiceNodeFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageChangeParameterFee:
		i, ok := value.(*wrapperspb.StringValue)
		if !ok {
			return types.ErrInvalidParamValue(value, i)
		}
		err := store.SetMessageChangeParameterFee(i.Value)
		if err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case utilTypes.MessageDoubleSignFeeOwner:
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
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	appMininimumStake, err := store.GetParamAppMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(utilTypes.AppMinimumStakeParamName, err)
	}
	return types.StringToBigInt(appMininimumStake)
}

func (u *UtilityContext) GetAppMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetMaxAppChains()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, types.Error) {
	store := u.Store()
	baselineRate, err := store.GetBaselineAppStakeRate()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppBaselineStakeRateParamName, err)
	}
	return baselineRate, nil
}

func (u *UtilityContext) GetStakingAdjustment() (int, types.Error) {
	store := u.Store()
	adjustment, err := store.GetStakingAdjustment()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppStakingAdjustmentParamName, err)
	}
	return adjustment, nil
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetAppUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetAppMinimumPauseBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetAppMaxPausedBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.AppMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	ServiceNodeMininimumStake, err := store.GetParamServiceNodeMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(utilTypes.ServiceNodeMinimumStakeParamName, err)
	}
	return types.StringToBigInt(ServiceNodeMininimumStake)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetServiceNodeMaxChains()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ServiceNodeMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetServiceNodeUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ServiceNodeUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ServiceNodeMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetServiceNodeMaxPausedBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ServiceNodeMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	ValidatorMininimumStake, err := store.GetParamValidatorMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(utilTypes.ValidatorMinimumStakeParamName, err)
	}
	return types.StringToBigInt(ValidatorMininimumStake)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetValidatorUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ValidatorUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetValidatorMinimumPauseBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ValidatorMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetValidatorMaxPausedBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ValidatorMaxPausedBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err types.Error) {
	store := u.Store()
	proposerPercentage, er := store.GetProposerPercentageOfFees()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ProposerPercentageOfFeesParamName, er)
	}
	return proposerPercentage, nil
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetValidatorMaximumMissedBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ValidatorMaximumMissedBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err types.Error) {
	store := u.Store()
	maxMissedBlocks, er := store.GetMaxEvidenceAgeInBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.ValidatorMaxEvidenceAgeInBlocksParamName, er)
	}
	return maxMissedBlocks, nil
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetDoubleSignBurnPercentage()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.DoubleSignBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err types.Error) {
	store := u.Store()
	burnPercentage, er := store.GetMissedBlocksBurnPercentage()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.MissedBlocksBurnPercentageParamName, er)
	}
	return burnPercentage, nil
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, types.Error) {
	store := u.Store()
	FishermanMininimumStake, err := store.GetParamFishermanMinimumStake()
	if err != nil {
		return nil, types.ErrGetParam(utilTypes.FishermanMinimumStakeParamName, err)
	}
	return types.StringToBigInt(FishermanMininimumStake)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, types.Error) {
	store := u.Store()
	maxChains, err := store.GetFishermanMaxChains()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.FishermanMaxChainsParamName, err)
	}
	return maxChains, nil
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, types.Error) {
	store := u.Store()
	unstakingHeight, err := store.GetFishermanUnstakingBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.FishermanUnstakingBlocksParamName, err)
	}
	return int64(unstakingHeight), nil
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, types.Error) {
	store := u.Store()
	minPauseBlocks, err := store.GetFishermanMinimumPauseBlocks()
	if err != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.FishermanMinimumPauseBlocksParamName, err)
	}
	return minPauseBlocks, nil
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	store := u.Store()
	maxPausedBlocks, er := store.GetFishermanMaxPausedBlocks()
	if er != nil {
		return utilTypes.ZeroInt, types.ErrGetParam(utilTypes.FishermanMaxPauseBlocksParamName, er)
	}
	return maxPausedBlocks, nil
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageDoubleSignFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageDoubleSignFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageSendFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageSendFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageStakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageEditStakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnstakeFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessagePauseFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseFishermanFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnpauseFishermanFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageFishermanPauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageFishermanPauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageTestScoreFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageProveTestScoreFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageProveTestScoreFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageStakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageEditStakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeAppFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnstakeAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessagePauseAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseAppFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnpauseAppFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageStakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageEditStakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnstakeValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessagePauseValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseValidatorFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnpauseValidatorFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageStakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageEditStakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageEditStakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnstakeServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnstakeServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessagePauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessagePauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageUnpauseServiceNodeFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageUnpauseServiceNodeFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, types.Error) {
	store := u.Store()
	fee, er := store.GetMessageChangeParameterFee()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.MessageChangeParameterFee, er)
	}
	return types.StringToBigInt(fee)
}

func (u *UtilityContext) GetDoubleSignFeeOwner() (owner []byte, err types.Error) {
	store := u.Store()
	owner, er := store.GetMessageDoubleSignFeeOwner()
	if er != nil {
		return nil, types.ErrGetParam(utilTypes.DoubleSignBurnPercentageParamName, er)
	}
	return owner, nil
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	store := u.Store()
	switch paramName {
	case utilTypes.AclOwner:
		return store.GetAclOwner()
	case utilTypes.BlocksPerSessionParamName:
		return store.GetBlocksPerSessionOwner()
	case utilTypes.AppMaxChainsParamName:
		return store.GetMaxAppChainsOwner()
	case utilTypes.AppMinimumStakeParamName:
		return store.GetAppMinimumStakeOwner()
	case utilTypes.AppBaselineStakeRateParamName:
		return store.GetBaselineAppOwner()
	case utilTypes.AppStakingAdjustmentParamName:
		return store.GetStakingAdjustmentOwner()
	case utilTypes.AppUnstakingBlocksParamName:
		return store.GetAppUnstakingBlocksOwner()
	case utilTypes.AppMinimumPauseBlocksParamName:
		return store.GetAppMinimumPauseBlocksOwner()
	case utilTypes.AppMaxPauseBlocksParamName:
		return store.GetAppMaxPausedBlocksOwner()
	case utilTypes.ServiceNodesPerSessionParamName:
		return store.GetServiceNodesPerSessionOwner()
	case utilTypes.ServiceNodeMinimumStakeParamName:
		return store.GetParamServiceNodeMinimumStakeOwner()
	case utilTypes.ServiceNodeMaxChainsParamName:
		return store.GetServiceNodeMaxChainsOwner()
	case utilTypes.ServiceNodeUnstakingBlocksParamName:
		return store.GetServiceNodeUnstakingBlocksOwner()
	case utilTypes.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetServiceNodeMinimumPauseBlocksOwner()
	case utilTypes.ServiceNodeMaxPauseBlocksParamName:
		return store.GetServiceNodeMaxPausedBlocksOwner()
	case utilTypes.FishermanMinimumStakeParamName:
		return store.GetFishermanMinimumStakeOwner()
	case utilTypes.FishermanMaxChainsParamName:
		return store.GetMaxFishermanChainsOwner()
	case utilTypes.FishermanUnstakingBlocksParamName:
		return store.GetFishermanUnstakingBlocksOwner()
	case utilTypes.FishermanMinimumPauseBlocksParamName:
		return store.GetFishermanMinimumPauseBlocksOwner()
	case utilTypes.FishermanMaxPauseBlocksParamName:
		return store.GetFishermanMaxPausedBlocksOwner()
	case utilTypes.ValidatorMinimumStakeParamName:
		return store.GetParamValidatorMinimumStakeOwner()
	case utilTypes.ValidatorUnstakingBlocksParamName:
		return store.GetValidatorUnstakingBlocksOwner()
	case utilTypes.ValidatorMinimumPauseBlocksParamName:
		return store.GetValidatorMinimumPauseBlocksOwner()
	case utilTypes.ValidatorMaxPausedBlocksParamName:
		return store.GetValidatorMaxPausedBlocksOwner()
	case utilTypes.ValidatorMaximumMissedBlocksParamName:
		return store.GetValidatorMaximumMissedBlocksOwner()
	case utilTypes.ProposerPercentageOfFeesParamName:
		return store.GetProposerPercentageOfFeesOwner()
	case utilTypes.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetMaxEvidenceAgeInBlocksOwner()
	case utilTypes.MissedBlocksBurnPercentageParamName:
		return store.GetMissedBlocksBurnPercentageOwner()
	case utilTypes.DoubleSignBurnPercentageParamName:
		return store.GetDoubleSignBurnPercentageOwner()
	case utilTypes.MessageDoubleSignFee:
		return store.GetMessageDoubleSignFeeOwner()
	case utilTypes.MessageSendFee:
		return store.GetMessageSendFeeOwner()
	case utilTypes.MessageStakeFishermanFee:
		return store.GetMessageStakeFishermanFeeOwner()
	case utilTypes.MessageEditStakeFishermanFee:
		return store.GetMessageEditStakeFishermanFeeOwner()
	case utilTypes.MessageUnstakeFishermanFee:
		return store.GetMessageUnstakeFishermanFeeOwner()
	case utilTypes.MessagePauseFishermanFee:
		return store.GetMessagePauseFishermanFeeOwner()
	case utilTypes.MessageUnpauseFishermanFee:
		return store.GetMessageUnpauseFishermanFeeOwner()
	case utilTypes.MessageFishermanPauseServiceNodeFee:
		return store.GetMessageFishermanPauseServiceNodeFeeOwner()
	case utilTypes.MessageTestScoreFee:
		return store.GetMessageTestScoreFeeOwner()
	case utilTypes.MessageProveTestScoreFee:
		return store.GetMessageProveTestScoreFeeOwner()
	case utilTypes.MessageStakeAppFee:
		return store.GetMessageStakeAppFeeOwner()
	case utilTypes.MessageEditStakeAppFee:
		return store.GetMessageEditStakeAppFeeOwner()
	case utilTypes.MessageUnstakeAppFee:
		return store.GetMessageUnstakeAppFeeOwner()
	case utilTypes.MessagePauseAppFee:
		return store.GetMessagePauseAppFeeOwner()
	case utilTypes.MessageUnpauseAppFee:
		return store.GetMessageUnpauseAppFeeOwner()
	case utilTypes.MessageStakeValidatorFee:
		return store.GetMessageStakeValidatorFeeOwner()
	case utilTypes.MessageEditStakeValidatorFee:
		return store.GetMessageEditStakeValidatorFeeOwner()
	case utilTypes.MessageUnstakeValidatorFee:
		return store.GetMessageUnstakeValidatorFeeOwner()
	case utilTypes.MessagePauseValidatorFee:
		return store.GetMessagePauseValidatorFeeOwner()
	case utilTypes.MessageUnpauseValidatorFee:
		return store.GetMessageUnpauseValidatorFeeOwner()
	case utilTypes.MessageStakeServiceNodeFee:
		return store.GetMessageStakeServiceNodeFeeOwner()
	case utilTypes.MessageEditStakeServiceNodeFee:
		return store.GetMessageEditStakeServiceNodeFeeOwner()
	case utilTypes.MessageUnstakeServiceNodeFee:
		return store.GetMessageUnstakeServiceNodeFeeOwner()
	case utilTypes.MessagePauseServiceNodeFee:
		return store.GetMessagePauseServiceNodeFeeOwner()
	case utilTypes.MessageUnpauseServiceNodeFee:
		return store.GetMessageUnpauseServiceNodeFeeOwner()
	case utilTypes.MessageChangeParameterFee:
		return store.GetMessageChangeParameterFeeOwner()
	case utilTypes.BlocksPerSessionOwner:
		return store.GetAclOwner()
	case utilTypes.AppMaxChainsOwner:
		return store.GetAclOwner()
	case utilTypes.AppMinimumStakeOwner:
		return store.GetAclOwner()
	case utilTypes.AppBaselineStakeRateOwner:
		return store.GetAclOwner()
	case utilTypes.AppStakingAdjustmentOwner:
		return store.GetAclOwner()
	case utilTypes.AppUnstakingBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.AppMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.AppMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodeMinimumStakeOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodeMaxChainsOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodeUnstakingBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodeMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ServiceNodesPerSessionOwner:
		return store.GetAclOwner()
	case utilTypes.FishermanMinimumStakeOwner:
		return store.GetAclOwner()
	case utilTypes.FishermanMaxChainsOwner:
		return store.GetAclOwner()
	case utilTypes.FishermanUnstakingBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.FishermanMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.FishermanMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorMinimumStakeOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorUnstakingBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorMinimumPauseBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorMaxPausedBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorMaximumMissedBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.ProposerPercentageOfFeesOwner:
		return store.GetAclOwner()
	case utilTypes.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetAclOwner()
	case utilTypes.MissedBlocksBurnPercentageOwner:
		return store.GetAclOwner()
	case utilTypes.DoubleSignBurnPercentageOwner:
		return store.GetAclOwner()
	case utilTypes.MessageSendFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageStakeFishermanFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageEditStakeFishermanFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnstakeFishermanFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessagePauseFishermanFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnpauseFishermanFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageTestScoreFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageProveTestScoreFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageStakeAppFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageEditStakeAppFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnstakeAppFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessagePauseAppFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnpauseAppFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageStakeValidatorFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageEditStakeValidatorFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnstakeValidatorFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessagePauseValidatorFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnpauseValidatorFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageStakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageEditStakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnstakeServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessagePauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageUnpauseServiceNodeFeeOwner:
		return store.GetAclOwner()
	case utilTypes.MessageChangeParameterFeeOwner:
		return store.GetAclOwner()
	default:
		return nil, types.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg utilTypes.Message) (amount *big.Int, err types.Error) {
	switch x := msg.(type) {
	case *utilTypes.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *utilTypes.MessageSend:
		return u.GetMessageSendFee()
	case *utilTypes.MessageStakeFisherman:
		return u.GetMessageStakeFishermanFee()
	case *utilTypes.MessageEditStakeFisherman:
		return u.GetMessageEditStakeFishermanFee()
	case *utilTypes.MessageUnstakeFisherman:
		return u.GetMessageUnstakeFishermanFee()
	case *utilTypes.MessagePauseFisherman:
		return u.GetMessagePauseFishermanFee()
	case *utilTypes.MessageUnpauseFisherman:
		return u.GetMessageUnpauseFishermanFee()
	case *utilTypes.MessageFishermanPauseServiceNode:
		return u.GetMessageFishermanPauseServiceNodeFee()
	//case *types.MessageTestScore:
	//	return u.GetMessageTestScoreFee()
	//case *types.MessageProveTestScore:
	//	return u.GetMessageProveTestScoreFee()
	case *utilTypes.MessageStakeApp:
		return u.GetMessageStakeAppFee()
	case *utilTypes.MessageEditStakeApp:
		return u.GetMessageEditStakeAppFee()
	case *utilTypes.MessageUnstakeApp:
		return u.GetMessageUnstakeAppFee()
	case *utilTypes.MessagePauseApp:
		return u.GetMessagePauseAppFee()
	case *utilTypes.MessageUnpauseApp:
		return u.GetMessageUnpauseAppFee()
	case *utilTypes.MessageStakeValidator:
		return u.GetMessageStakeValidatorFee()
	case *utilTypes.MessageEditStakeValidator:
		return u.GetMessageEditStakeValidatorFee()
	case *utilTypes.MessageUnstakeValidator:
		return u.GetMessageUnstakeValidatorFee()
	case *utilTypes.MessagePauseValidator:
		return u.GetMessagePauseValidatorFee()
	case *utilTypes.MessageUnpauseValidator:
		return u.GetMessageUnpauseValidatorFee()
	case *utilTypes.MessageStakeServiceNode:
		return u.GetMessageStakeServiceNodeFee()
	case *utilTypes.MessageEditStakeServiceNode:
		return u.GetMessageEditStakeServiceNodeFee()
	case *utilTypes.MessageUnstakeServiceNode:
		return u.GetMessageUnstakeServiceNodeFee()
	case *utilTypes.MessagePauseServiceNode:
		return u.GetMessagePauseServiceNodeFee()
	case *utilTypes.MessageUnpauseServiceNode:
		return u.GetMessageUnpauseServiceNodeFee()
	case *utilTypes.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, types.ErrUnknownMessage(x)
	}
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *utilTypes.MessageChangeParameter) ([][]byte, types.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, types.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}
