package utility

import (
	"log"
	"math/big"

	typesUtil "github.com/pokt-network/pocket/internal/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *UtilityContext) UpdateParam(paramName string, value interface{}) typesUtil.Error {
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

func (u *UtilityContext) GetBlocksPerSession() (int, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return 0, er
	}
	blocksPerSession, err := store.GetBlocksPerSession(height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(typesUtil.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
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
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	switch paramName {
	case typesUtil.AclOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.BlocksPerSessionParamName:
		return store.GetBytesParam(typesUtil.BlocksPerSessionOwner, height)
	case typesUtil.AppMaxChainsParamName:
		return store.GetBytesParam(typesUtil.AppMaxChainsOwner, height)
	case typesUtil.AppMinimumStakeParamName:
		return store.GetBytesParam(typesUtil.AppMinimumStakeOwner, height)
	case typesUtil.AppBaselineStakeRateParamName:
		return store.GetBytesParam(typesUtil.AppBaselineStakeRateOwner, height)
	case typesUtil.AppStakingAdjustmentParamName:
		return store.GetBytesParam(typesUtil.AppStakingAdjustmentOwner, height)
	case typesUtil.AppUnstakingBlocksParamName:
		return store.GetBytesParam(typesUtil.AppUnstakingBlocksOwner, height)
	case typesUtil.AppMinimumPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.AppMinimumPauseBlocksOwner, height)
	case typesUtil.AppMaxPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.AppMaxPausedBlocksOwner, height)
	case typesUtil.ServiceNodesPerSessionParamName:
		return store.GetBytesParam(typesUtil.ServiceNodesPerSessionOwner, height)
	case typesUtil.ServiceNodeMinimumStakeParamName:
		return store.GetBytesParam(typesUtil.ServiceNodeMinimumStakeOwner, height)
	case typesUtil.ServiceNodeMaxChainsParamName:
		return store.GetBytesParam(typesUtil.ServiceNodeMaxChainsOwner, height)
	case typesUtil.ServiceNodeUnstakingBlocksParamName:
		return store.GetBytesParam(typesUtil.ServiceNodeUnstakingBlocksOwner, height)
	case typesUtil.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.ServiceNodeMinimumPauseBlocksOwner, height)
	case typesUtil.ServiceNodeMaxPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.ServiceNodeMaxPausedBlocksOwner, height)
	case typesUtil.FishermanMinimumStakeParamName:
		return store.GetBytesParam(typesUtil.FishermanMinimumStakeOwner, height)
	case typesUtil.FishermanMaxChainsParamName:
		return store.GetBytesParam(typesUtil.FishermanMaxChainsOwner, height)
	case typesUtil.FishermanUnstakingBlocksParamName:
		return store.GetBytesParam(typesUtil.FishermanUnstakingBlocksOwner, height)
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.FishermanMinimumPauseBlocksOwner, height)
	case typesUtil.FishermanMaxPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.FishermanMaxPausedBlocksOwner, height)
	case typesUtil.ValidatorMinimumStakeParamName:
		return store.GetBytesParam(typesUtil.ValidatorMinimumStakeOwner, height)
	case typesUtil.ValidatorUnstakingBlocksParamName:
		return store.GetBytesParam(typesUtil.ValidatorUnstakingBlocksOwner, height)
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.ValidatorMinimumPauseBlocksOwner, height)
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		return store.GetBytesParam(typesUtil.ValidatorMaxPausedBlocksOwner, height)
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		return store.GetBytesParam(typesUtil.ValidatorMaximumMissedBlocksOwner, height)
	case typesUtil.ProposerPercentageOfFeesParamName:
		return store.GetBytesParam(typesUtil.ProposerPercentageOfFeesOwner, height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetBytesParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner, height)
	case typesUtil.MissedBlocksBurnPercentageParamName:
		return store.GetBytesParam(typesUtil.MissedBlocksBurnPercentageOwner, height)
	case typesUtil.DoubleSignBurnPercentageParamName:
		return store.GetBytesParam(typesUtil.DoubleSignBurnPercentageOwner, height)
	case typesUtil.MessageDoubleSignFee:
		return store.GetBytesParam(typesUtil.MessageDoubleSignFeeOwner, height)
	case typesUtil.MessageSendFee:
		return store.GetBytesParam(typesUtil.MessageSendFeeOwner, height)
	case typesUtil.MessageStakeFishermanFee:
		return store.GetBytesParam(typesUtil.MessageStakeFishermanFeeOwner, height)
	case typesUtil.MessageEditStakeFishermanFee:
		return store.GetBytesParam(typesUtil.MessageEditStakeFishermanFeeOwner, height)
	case typesUtil.MessageUnstakeFishermanFee:
		return store.GetBytesParam(typesUtil.MessageUnstakeFishermanFeeOwner, height)
	case typesUtil.MessagePauseFishermanFee:
		return store.GetBytesParam(typesUtil.MessagePauseFishermanFeeOwner, height)
	case typesUtil.MessageUnpauseFishermanFee:
		return store.GetBytesParam(typesUtil.MessageUnpauseFishermanFeeOwner, height)
	case typesUtil.MessageFishermanPauseServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessageFishermanPauseServiceNodeFeeOwner, height)
	case typesUtil.MessageTestScoreFee:
		return store.GetBytesParam(typesUtil.MessageTestScoreFeeOwner, height)
	case typesUtil.MessageProveTestScoreFee:
		return store.GetBytesParam(typesUtil.MessageProveTestScoreFeeOwner, height)
	case typesUtil.MessageStakeAppFee:
		return store.GetBytesParam(typesUtil.MessageStakeAppFeeOwner, height)
	case typesUtil.MessageEditStakeAppFee:
		return store.GetBytesParam(typesUtil.MessageEditStakeAppFeeOwner, height)
	case typesUtil.MessageUnstakeAppFee:
		return store.GetBytesParam(typesUtil.MessageUnstakeAppFeeOwner, height)
	case typesUtil.MessagePauseAppFee:
		return store.GetBytesParam(typesUtil.MessagePauseAppFeeOwner, height)
	case typesUtil.MessageUnpauseAppFee:
		return store.GetBytesParam(typesUtil.MessageUnpauseAppFeeOwner, height)
	case typesUtil.MessageStakeValidatorFee:
		return store.GetBytesParam(typesUtil.MessageStakeValidatorFeeOwner, height)
	case typesUtil.MessageEditStakeValidatorFee:
		return store.GetBytesParam(typesUtil.MessageEditStakeValidatorFeeOwner, height)
	case typesUtil.MessageUnstakeValidatorFee:
		return store.GetBytesParam(typesUtil.MessageUnstakeValidatorFeeOwner, height)
	case typesUtil.MessagePauseValidatorFee:
		return store.GetBytesParam(typesUtil.MessagePauseValidatorFeeOwner, height)
	case typesUtil.MessageUnpauseValidatorFee:
		return store.GetBytesParam(typesUtil.MessageUnpauseValidatorFeeOwner, height)
	case typesUtil.MessageStakeServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessageStakeServiceNodeFeeOwner, height)
	case typesUtil.MessageEditStakeServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessageEditStakeServiceNodeFeeOwner, height)
	case typesUtil.MessageUnstakeServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessageUnstakeServiceNodeFeeOwner, height)
	case typesUtil.MessagePauseServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessagePauseServiceNodeFeeOwner, height)
	case typesUtil.MessageUnpauseServiceNodeFee:
		return store.GetBytesParam(typesUtil.MessageUnpauseServiceNodeFeeOwner, height)
	case typesUtil.MessageChangeParameterFee:
		return store.GetBytesParam(typesUtil.MessageChangeParameterFeeOwner, height)
	case typesUtil.BlocksPerSessionOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppMaxChainsOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppMinimumStakeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppBaselineStakeRateOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppStakingAdjustmentOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppUnstakingBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppMinimumPauseBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.AppMaxPausedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodeMinimumStakeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodeMaxChainsOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodeUnstakingBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodeMaxPausedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServiceNodesPerSessionOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.FishermanMinimumStakeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.FishermanMaxChainsOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.FishermanUnstakingBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.FishermanMaxPausedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorMinimumStakeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorUnstakingBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ProposerPercentageOfFeesOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MissedBlocksBurnPercentageOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.DoubleSignBurnPercentageOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageSendFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageStakeFishermanFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessagePauseFishermanFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageTestScoreFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageProveTestScoreFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageStakeAppFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageEditStakeAppFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnstakeAppFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessagePauseAppFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnpauseAppFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageStakeValidatorFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessagePauseValidatorFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageStakeServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageEditStakeServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnstakeServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessagePauseServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnpauseServiceNodeFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageChangeParameterFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	default:
		return nil, typesUtil.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg typesUtil.Message, actorType typesUtil.ActorType) (amount *big.Int, err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *typesUtil.MessageSend:
		return u.GetMessageSendFee()
	case *typesUtil.MessageStake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageStakeAppFee()
		case typesUtil.ActorType_Fisherman:
			return u.GetMessageStakeFishermanFee()
		case typesUtil.ActorType_ServiceNode:
			return u.GetMessageStakeServiceNodeFee()
		case typesUtil.ActorType_Validator:
			return u.GetMessageStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageEditStakeAppFee()
		case typesUtil.ActorType_Fisherman:
			return u.GetMessageEditStakeFishermanFee()
		case typesUtil.ActorType_ServiceNode:
			return u.GetMessageEditStakeServiceNodeFee()
		case typesUtil.ActorType_Validator:
			return u.GetMessageEditStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageUnstakeAppFee()
		case typesUtil.ActorType_Fisherman:
			return u.GetMessageUnstakeFishermanFee()
		case typesUtil.ActorType_ServiceNode:
			return u.GetMessageUnstakeServiceNodeFee()
		case typesUtil.ActorType_Validator:
			return u.GetMessageUnstakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageUnpauseAppFee()
		case typesUtil.ActorType_Fisherman:
			return u.GetMessageUnpauseFishermanFee()
		case typesUtil.ActorType_ServiceNode:
			return u.GetMessageUnpauseServiceNodeFee()
		case typesUtil.ActorType_Validator:
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
	value, err := store.GetStringParam(paramName, height)
	if err != nil {
		log.Printf("err: %v\n", err)
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return typesUtil.StringToBigInt(value)
}

func (u *UtilityContext) getIntParam(paramName string) (int, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return 0, er
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *UtilityContext) getInt64Param(paramName string) (int64, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return 0, er
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *UtilityContext) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	value, err := store.GetBytesParam(paramName, height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}
