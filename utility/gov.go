package utility

import (
	"log"
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *UtilityContext) UpdateParam(paramName string, value any) typesUtil.Error {
	store := u.Store()
	switch t := value.(type) {
	case *wrapperspb.Int32Value:
		if err := store.SetParam(paramName, (int(t.Value))); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.StringValue:
		if err := store.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.BytesValue:
		if err := store.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	default:
		break
	}
	log.Fatalf("unhandled value type %T for %v", value, value)
	return typesUtil.ErrUnknownParam(paramName)
}

func (u *UtilityContext) GetParameter(paramName string, height int64) (any, error) {
	return u.Store().GetParameter(paramName, height)
}

func (u *UtilityContext) GetBlocksPerSession() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.BlocksPerSessionParamName)
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.AppMinimumStakeParamName)
}

func (u *UtilityContext) GetAppMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxChainsParamName)
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppBaselineStakeRateParamName)
}

func (u *UtilityContext) GetStabilityAdjustment() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppStakingAdjustmentParamName)
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.AppUnstakingBlocksParamName)
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ServiceNodeMinimumStakeParamName)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServiceNodeMaxChainsParamName)
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ServiceNodeUnstakingBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ServiceNodeMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ValidatorMinimumStakeParamName)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ValidatorUnstakingBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxPausedBlocksParamName)
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ProposerPercentageOfFeesParamName)
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaximumMissedBlocksParamName)
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.DoubleSignBurnPercentageParamName)
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.MissedBlocksBurnPercentageParamName)
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.FishermanMinimumStakeParamName)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxChainsParamName)
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.FishermanUnstakingBlocksParamName)
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageDoubleSignFee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageSendFee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeFishermanFee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeFishermanFee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeFishermanFee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseFishermanFee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseFishermanFee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageFishermanPauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageTestScoreFee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageProveTestScoreFee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeAppFee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeAppFee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeAppFee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseAppFee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseAppFee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeValidatorFee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeValidatorFee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeValidatorFee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseValidatorFee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseValidatorFee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeServiceNodeFee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageChangeParameterFee)
}

func (u *UtilityContext) GetDoubleSignFeeOwner() (owner []byte, err typesUtil.Error) {
	return u.getByteArrayParam(typesUtil.MessageDoubleSignFeeOwner)
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	// DISCUSS (@deblasis): here we could potentially leverage the struct tags in gov.proto by specifying an `owner` key
	// eg: `app_minimum_stake` could have `pokt:"owner=app_minimum_stake_owner"`
	// in here we would use that map to point to the owner, removing this switch, centralizing the logic and making it declarative
	switch paramName {
	case typesUtil.AclOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.BlocksPerSessionParamName:
		return u.getByteArrayParam(typesUtil.BlocksPerSessionOwner)
	case typesUtil.AppMaxChainsParamName:
		return u.getByteArrayParam(typesUtil.AppMaxChainsOwner)
	case typesUtil.AppMinimumStakeParamName:
		return u.getByteArrayParam(typesUtil.AppMinimumStakeOwner)
	case typesUtil.AppBaselineStakeRateParamName:
		return u.getByteArrayParam(typesUtil.AppBaselineStakeRateOwner)
	case typesUtil.AppStakingAdjustmentParamName:
		return u.getByteArrayParam(typesUtil.AppStakingAdjustmentOwner)
	case typesUtil.AppUnstakingBlocksParamName:
		return u.getByteArrayParam(typesUtil.AppUnstakingBlocksOwner)
	case typesUtil.AppMinimumPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.AppMinimumPauseBlocksOwner)
	case typesUtil.AppMaxPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.AppMaxPausedBlocksOwner)
	case typesUtil.ServiceNodesPerSessionParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodesPerSessionOwner)
	case typesUtil.ServiceNodeMinimumStakeParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodeMinimumStakeOwner)
	case typesUtil.ServiceNodeMaxChainsParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodeMaxChainsOwner)
	case typesUtil.ServiceNodeUnstakingBlocksParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodeUnstakingBlocksOwner)
	case typesUtil.ServiceNodeMinimumPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodeMinimumPauseBlocksOwner)
	case typesUtil.ServiceNodeMaxPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.ServiceNodeMaxPausedBlocksOwner)
	case typesUtil.FishermanMinimumStakeParamName:
		return u.getByteArrayParam(typesUtil.FishermanMinimumStakeOwner)
	case typesUtil.FishermanMaxChainsParamName:
		return u.getByteArrayParam(typesUtil.FishermanMaxChainsOwner)
	case typesUtil.FishermanUnstakingBlocksParamName:
		return u.getByteArrayParam(typesUtil.FishermanUnstakingBlocksOwner)
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.FishermanMinimumPauseBlocksOwner)
	case typesUtil.FishermanMaxPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.FishermanMaxPausedBlocksOwner)
	case typesUtil.ValidatorMinimumStakeParamName:
		return u.getByteArrayParam(typesUtil.ValidatorMinimumStakeOwner)
	case typesUtil.ValidatorUnstakingBlocksParamName:
		return u.getByteArrayParam(typesUtil.ValidatorUnstakingBlocksOwner)
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		return u.getByteArrayParam(typesUtil.ValidatorMinimumPauseBlocksOwner)
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		return u.getByteArrayParam(typesUtil.ValidatorMaxPausedBlocksOwner)
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		return u.getByteArrayParam(typesUtil.ValidatorMaximumMissedBlocksOwner)
	case typesUtil.ProposerPercentageOfFeesParamName:
		return u.getByteArrayParam(typesUtil.ProposerPercentageOfFeesOwner)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		return u.getByteArrayParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner)
	case typesUtil.MissedBlocksBurnPercentageParamName:
		return u.getByteArrayParam(typesUtil.MissedBlocksBurnPercentageOwner)
	case typesUtil.DoubleSignBurnPercentageParamName:
		return u.getByteArrayParam(typesUtil.DoubleSignBurnPercentageOwner)
	case typesUtil.MessageDoubleSignFee:
		return u.getByteArrayParam(typesUtil.MessageDoubleSignFeeOwner)
	case typesUtil.MessageSendFee:
		return u.getByteArrayParam(typesUtil.MessageSendFeeOwner)
	case typesUtil.MessageStakeFishermanFee:
		return u.getByteArrayParam(typesUtil.MessageStakeFishermanFeeOwner)
	case typesUtil.MessageEditStakeFishermanFee:
		return u.getByteArrayParam(typesUtil.MessageEditStakeFishermanFeeOwner)
	case typesUtil.MessageUnstakeFishermanFee:
		return u.getByteArrayParam(typesUtil.MessageUnstakeFishermanFeeOwner)
	case typesUtil.MessagePauseFishermanFee:
		return u.getByteArrayParam(typesUtil.MessagePauseFishermanFeeOwner)
	case typesUtil.MessageUnpauseFishermanFee:
		return u.getByteArrayParam(typesUtil.MessageUnpauseFishermanFeeOwner)
	case typesUtil.MessageFishermanPauseServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessageFishermanPauseServiceNodeFeeOwner)
	case typesUtil.MessageTestScoreFee:
		return u.getByteArrayParam(typesUtil.MessageTestScoreFeeOwner)
	case typesUtil.MessageProveTestScoreFee:
		return u.getByteArrayParam(typesUtil.MessageProveTestScoreFeeOwner)
	case typesUtil.MessageStakeAppFee:
		return u.getByteArrayParam(typesUtil.MessageStakeAppFeeOwner)
	case typesUtil.MessageEditStakeAppFee:
		return u.getByteArrayParam(typesUtil.MessageEditStakeAppFeeOwner)
	case typesUtil.MessageUnstakeAppFee:
		return u.getByteArrayParam(typesUtil.MessageUnstakeAppFeeOwner)
	case typesUtil.MessagePauseAppFee:
		return u.getByteArrayParam(typesUtil.MessagePauseAppFeeOwner)
	case typesUtil.MessageUnpauseAppFee:
		return u.getByteArrayParam(typesUtil.MessageUnpauseAppFeeOwner)
	case typesUtil.MessageStakeValidatorFee:
		return u.getByteArrayParam(typesUtil.MessageStakeValidatorFeeOwner)
	case typesUtil.MessageEditStakeValidatorFee:
		return u.getByteArrayParam(typesUtil.MessageEditStakeValidatorFeeOwner)
	case typesUtil.MessageUnstakeValidatorFee:
		return u.getByteArrayParam(typesUtil.MessageUnstakeValidatorFeeOwner)
	case typesUtil.MessagePauseValidatorFee:
		return u.getByteArrayParam(typesUtil.MessagePauseValidatorFeeOwner)
	case typesUtil.MessageUnpauseValidatorFee:
		return u.getByteArrayParam(typesUtil.MessageUnpauseValidatorFeeOwner)
	case typesUtil.MessageStakeServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessageStakeServiceNodeFeeOwner)
	case typesUtil.MessageEditStakeServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessageEditStakeServiceNodeFeeOwner)
	case typesUtil.MessageUnstakeServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessageUnstakeServiceNodeFeeOwner)
	case typesUtil.MessagePauseServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessagePauseServiceNodeFeeOwner)
	case typesUtil.MessageUnpauseServiceNodeFee:
		return u.getByteArrayParam(typesUtil.MessageUnpauseServiceNodeFeeOwner)
	case typesUtil.MessageChangeParameterFee:
		return u.getByteArrayParam(typesUtil.MessageChangeParameterFeeOwner)
	case typesUtil.BlocksPerSessionOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppMaxChainsOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppMinimumStakeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppBaselineStakeRateOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppStakingAdjustmentOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppUnstakingBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppMinimumPauseBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.AppMaxPausedBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodeMinimumStakeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodeMaxChainsOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodeUnstakingBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodeMinimumPauseBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodeMaxPausedBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ServiceNodesPerSessionOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.FishermanMinimumStakeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.FishermanMaxChainsOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.FishermanUnstakingBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.FishermanMaxPausedBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorMinimumStakeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorUnstakingBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ProposerPercentageOfFeesOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MissedBlocksBurnPercentageOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.DoubleSignBurnPercentageOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageSendFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageStakeFishermanFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessagePauseFishermanFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageFishermanPauseServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageTestScoreFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageProveTestScoreFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageStakeAppFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageEditStakeAppFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnstakeAppFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessagePauseAppFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnpauseAppFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageStakeValidatorFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessagePauseValidatorFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageStakeServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageEditStakeServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnstakeServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessagePauseServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageUnpauseServiceNodeFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	case typesUtil.MessageChangeParameterFeeOwner:
		return u.getByteArrayParam(typesUtil.AclOwner)
	default:
		return nil, typesUtil.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg typesUtil.Message, actorType coreTypes.ActorType) (amount *big.Int, err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *typesUtil.MessageSend:
		return u.GetMessageSendFee()
	case *typesUtil.MessageStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.GetMessageStakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.GetMessageStakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
			return u.GetMessageStakeServiceNodeFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.GetMessageStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.GetMessageEditStakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.GetMessageEditStakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
			return u.GetMessageEditStakeServiceNodeFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.GetMessageEditStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.GetMessageUnstakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.GetMessageUnstakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
			return u.GetMessageUnstakeServiceNodeFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.GetMessageUnstakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.GetMessageUnpauseAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.GetMessageUnpauseFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
			return u.GetMessageUnpauseServiceNodeFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.GetMessageUnpauseValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, typesUtil.ErrUnknownMessage(x)
	}
	return nil, nil
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *typesUtil.MessageChangeParameter) ([][]byte, typesUtil.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, typesUtil.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}

func (u *UtilityContext) getBigIntParam(paramName string) (*big.Int, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	value, err := store.GetParameter(paramName, height)
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return typesUtil.StringToBigInt(value.(string))
}

func (u *UtilityContext) getIntParam(paramName string) (int, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return 0, er
	}
	value, err := store.GetParameter(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return value.(int), nil
}

func (u *UtilityContext) getInt64Param(paramName string) (int64, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return 0, er
	}
	value, err := store.GetParameter(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value.(int)), nil
}

func (u *UtilityContext) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	value, err := store.GetParameter(paramName, height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return typesUtil.StringToBytes(value.(string))
}
