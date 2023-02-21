package utility

import (
	"math/big"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *utilityContext) updateParam(paramName string, value any) typesUtil.Error {
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
	u.logger.Fatal().Msgf("unhandled value type %T for %v", value, value)
	return typesUtil.ErrUnknownParam(paramName)
}

func (u *utilityContext) getParameter(paramName string, height int64) (any, error) {
	return u.Store().GetParameter(paramName, height)
}

func (u *utilityContext) getAppMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.AppMinimumStakeParamName)
}

func (u *utilityContext) getAppMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxChainsParamName)
}

func (u *utilityContext) getBaselineAppStakeRate() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppBaselineStakeRateParamName)
}

func (u *utilityContext) getStabilityAdjustment() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppStakingAdjustmentParamName)
}

func (u *utilityContext) getAppUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.AppUnstakingBlocksParamName)
}

func (u *utilityContext) getAppMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMinimumPauseBlocksParamName)
}

func (u *utilityContext) getAppMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxPauseBlocksParamName)
}

func (u *utilityContext) getServicerMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ServicerMinimumStakeParamName)
}

func (u *utilityContext) getServicerMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMaxChainsParamName)
}

func (u *utilityContext) getServicerUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ServicerUnstakingBlocksParamName)
}

func (u *utilityContext) getServicerMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMinimumPauseBlocksParamName)
}

func (u *utilityContext) getServicerMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMaxPauseBlocksParamName)
}

func (u *utilityContext) getValidatorMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ValidatorMinimumStakeParamName)
}

func (u *utilityContext) getValidatorUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ValidatorUnstakingBlocksParamName)
}

func (u *utilityContext) getValidatorMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMinimumPauseBlocksParamName)
}

func (u *utilityContext) getValidatorMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxPausedBlocksParamName)
}

func (u *utilityContext) getProposerPercentageOfFees() (proposerPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ProposerPercentageOfFeesParamName)
}

func (u *utilityContext) getValidatorMaxMissedBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaximumMissedBlocksParamName)
}

func (u *utilityContext) getMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (u *utilityContext) getDoubleSignBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.DoubleSignBurnPercentageParamName)
}

func (u *utilityContext) getMissedBlocksBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.MissedBlocksBurnPercentageParamName)
}

func (u *utilityContext) getFishermanMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.FishermanMinimumStakeParamName)
}

func (u *utilityContext) getFishermanMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxChainsParamName)
}

func (u *utilityContext) getFishermanUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.FishermanUnstakingBlocksParamName)
}

func (u *utilityContext) getFishermanMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMinimumPauseBlocksParamName)
}

func (u *utilityContext) getFishermanMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxPauseBlocksParamName)
}

func (u *utilityContext) getMessageDoubleSignFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageDoubleSignFee)
}

func (u *utilityContext) getMessageSendFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageSendFee)
}

func (u *utilityContext) getMessageStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeFishermanFee)
}

func (u *utilityContext) getMessageEditStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeFishermanFee)
}

func (u *utilityContext) getMessageUnstakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeFishermanFee)
}

func (u *utilityContext) getMessagePauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseFishermanFee)
}

func (u *utilityContext) getMessageUnpauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseFishermanFee)
}

func (u *utilityContext) getMessageFishermanPauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageFishermanPauseServicerFee)
}

func (u *utilityContext) getMessageTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageTestScoreFee)
}

func (u *utilityContext) getMessageProveTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageProveTestScoreFee)
}

func (u *utilityContext) getMessageStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeAppFee)
}

func (u *utilityContext) getMessageEditStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeAppFee)
}

func (u *utilityContext) getMessageUnstakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeAppFee)
}

func (u *utilityContext) getMessagePauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseAppFee)
}

func (u *utilityContext) getMessageUnpauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseAppFee)
}

func (u *utilityContext) getMessageStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeValidatorFee)
}

func (u *utilityContext) getMessageEditStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeValidatorFee)
}

func (u *utilityContext) getMessageUnstakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeValidatorFee)
}

func (u *utilityContext) getMessagePauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseValidatorFee)
}

func (u *utilityContext) getMessageUnpauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseValidatorFee)
}

func (u *utilityContext) getMessageStakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeServicerFee)
}

func (u *utilityContext) getMessageEditStakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeServicerFee)
}

func (u *utilityContext) getMessageUnstakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeServicerFee)
}

func (u *utilityContext) getMessagePauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseServicerFee)
}

func (u *utilityContext) getMessageUnpauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseServicerFee)
}

func (u *utilityContext) getMessageChangeParameterFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageChangeParameterFee)
}

func (u *utilityContext) getDoubleSignFeeOwner() (owner []byte, err typesUtil.Error) {
	return u.getByteArrayParam(typesUtil.MessageDoubleSignFeeOwner)
}

func (u *utilityContext) getParamOwner(paramName string) ([]byte, error) {
	// DISCUSS (@deblasis): here we could potentially leverage the struct tags in gov.proto by specifying an `owner` key
	// eg: `app_minimum_stake` could have `pokt:"owner=app_minimum_stake_owner"`
	// in here we would use that map to point to the owner, removing this switch, centralizing the logic and making it declarative
	store, height, er := u.getStoreAndHeight()
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
	case typesUtil.ServicersPerSessionParamName:
		return store.GetBytesParam(typesUtil.ServicersPerSessionOwner, height)
	case typesUtil.ServicerMinimumStakeParamName:
		return store.GetBytesParam(typesUtil.ServicerMinimumStakeOwner, height)
	case typesUtil.ServicerMaxChainsParamName:
		return store.GetBytesParam(typesUtil.ServicerMaxChainsOwner, height)
	case typesUtil.ServicerUnstakingBlocksParamName:
		return store.GetBytesParam(typesUtil.ServicerUnstakingBlocksOwner, height)
	case typesUtil.ServicerMinimumPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.ServicerMinimumPauseBlocksOwner, height)
	case typesUtil.ServicerMaxPauseBlocksParamName:
		return store.GetBytesParam(typesUtil.ServicerMaxPausedBlocksOwner, height)
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
	case typesUtil.MessageFishermanPauseServicerFee:
		return store.GetBytesParam(typesUtil.MessageFishermanPauseServicerFeeOwner, height)
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
	case typesUtil.MessageStakeServicerFee:
		return store.GetBytesParam(typesUtil.MessageStakeServicerFeeOwner, height)
	case typesUtil.MessageEditStakeServicerFee:
		return store.GetBytesParam(typesUtil.MessageEditStakeServicerFeeOwner, height)
	case typesUtil.MessageUnstakeServicerFee:
		return store.GetBytesParam(typesUtil.MessageUnstakeServicerFeeOwner, height)
	case typesUtil.MessagePauseServicerFee:
		return store.GetBytesParam(typesUtil.MessagePauseServicerFeeOwner, height)
	case typesUtil.MessageUnpauseServicerFee:
		return store.GetBytesParam(typesUtil.MessageUnpauseServicerFeeOwner, height)
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
	case typesUtil.ServicerMinimumStakeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServicerMaxChainsOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServicerUnstakingBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServicerMinimumPauseBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServicerMaxPausedBlocksOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.ServicersPerSessionOwner:
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
	case typesUtil.MessageFishermanPauseServicerFeeOwner:
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
	case typesUtil.MessageStakeServicerFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageEditStakeServicerFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnstakeServicerFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessagePauseServicerFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageUnpauseServicerFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	case typesUtil.MessageChangeParameterFeeOwner:
		return store.GetBytesParam(typesUtil.AclOwner, height)
	default:
		return nil, typesUtil.ErrUnknownParam(paramName)
	}
}

func (u *utilityContext) getFee(msg typesUtil.Message, actorType coreTypes.ActorType) (amount *big.Int, err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageSend:
		return u.getMessageSendFee()
	case *typesUtil.MessageStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.getMessageStakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.getMessageStakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return u.getMessageStakeServicerFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.getMessageStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.getMessageEditStakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.getMessageEditStakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return u.getMessageEditStakeServicerFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.getMessageEditStakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.getMessageUnstakeAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.getMessageUnstakeFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return u.getMessageUnstakeServicerFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.getMessageUnstakeValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return u.getMessageUnpauseAppFee()
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return u.getMessageUnpauseFishermanFee()
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return u.getMessageUnpauseServicerFee()
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return u.getMessageUnpauseValidatorFee()
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageChangeParameter:
		return u.getMessageChangeParameterFee()
	default:
		return nil, typesUtil.ErrUnknownMessage(x)
	}
}

func (u *utilityContext) getMessageChangeParameterSignerCandidates(msg *typesUtil.MessageChangeParameter) ([][]byte, typesUtil.Error) {
	owner, err := u.getParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, typesUtil.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}

func (u *utilityContext) getBigIntParam(paramName string) (*big.Int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	value, err := store.GetStringParam(paramName, height)
	if err != nil {
		u.logger.Err(err)
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	amount, err := converters.StringToBigInt(value)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) getIntParam(paramName string) (int, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *utilityContext) getInt64Param(paramName string) (int64, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return 0, typesUtil.ErrGetHeight(err)
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *utilityContext) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	store, height, err := u.getStoreAndHeight()
	if err != nil {
		return nil, typesUtil.ErrGetHeight(err)
	}
	value, err := store.GetBytesParam(paramName, height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}
