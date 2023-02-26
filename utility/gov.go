package utility

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *utilityContext) updateParam(paramName string, value any) typesUtil.Error {
	switch t := value.(type) {
	case *wrapperspb.Int32Value:
		if err := u.store.SetParam(paramName, (int(t.Value))); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.StringValue:
		if err := u.store.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.BytesValue:
		if err := u.store.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	default:
		break
	}
	u.logger.Fatal().Msgf("unhandled value type %T for %v", value, value)
	return typesUtil.ErrUnknownParam(paramName)
}

func (u *utilityContext) getParameter(paramName string) (any, error) {
	return u.store.GetParameter(paramName, u.height)
}

func (u *utilityContext) getAppMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.AppMinimumStakeParamName)
}

func (u *utilityContext) getAppMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxChainsParamName)
}

func (u *utilityContext) getAppSessionTokensMultiplier() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppSessionTokensMultiplierParamName)
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
	switch paramName {
	case typesUtil.AclOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.BlocksPerSessionParamName:
		return u.store.GetBytesParam(typesUtil.BlocksPerSessionOwner, u.height)
	case typesUtil.AppMaxChainsParamName:
		return u.store.GetBytesParam(typesUtil.AppMaxChainsOwner, u.height)
	case typesUtil.AppMinimumStakeParamName:
		return u.store.GetBytesParam(typesUtil.AppMinimumStakeOwner, u.height)
	case typesUtil.AppSessionTokensMultiplierParamName:
		return u.store.GetBytesParam(typesUtil.AppSessionTokensMultiplierOwner, u.height)
	case typesUtil.AppUnstakingBlocksParamName:
		return u.store.GetBytesParam(typesUtil.AppUnstakingBlocksOwner, u.height)
	case typesUtil.AppMinimumPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.AppMinimumPauseBlocksOwner, u.height)
	case typesUtil.AppMaxPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.AppMaxPausedBlocksOwner, u.height)
	case typesUtil.ServicersPerSessionParamName:
		return u.store.GetBytesParam(typesUtil.ServicersPerSessionOwner, u.height)
	case typesUtil.ServicerMinimumStakeParamName:
		return u.store.GetBytesParam(typesUtil.ServicerMinimumStakeOwner, u.height)
	case typesUtil.ServicerMaxChainsParamName:
		return u.store.GetBytesParam(typesUtil.ServicerMaxChainsOwner, u.height)
	case typesUtil.ServicerUnstakingBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ServicerUnstakingBlocksOwner, u.height)
	case typesUtil.ServicerMinimumPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ServicerMinimumPauseBlocksOwner, u.height)
	case typesUtil.ServicerMaxPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ServicerMaxPausedBlocksOwner, u.height)
	case typesUtil.FishermanMinimumStakeParamName:
		return u.store.GetBytesParam(typesUtil.FishermanMinimumStakeOwner, u.height)
	case typesUtil.FishermanMaxChainsParamName:
		return u.store.GetBytesParam(typesUtil.FishermanMaxChainsOwner, u.height)
	case typesUtil.FishermanUnstakingBlocksParamName:
		return u.store.GetBytesParam(typesUtil.FishermanUnstakingBlocksOwner, u.height)
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.FishermanMinimumPauseBlocksOwner, u.height)
	case typesUtil.FishermanMaxPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.FishermanMaxPausedBlocksOwner, u.height)
	case typesUtil.ValidatorMinimumStakeParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorMinimumStakeOwner, u.height)
	case typesUtil.ValidatorUnstakingBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorUnstakingBlocksOwner, u.height)
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorMinimumPauseBlocksOwner, u.height)
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorMaxPausedBlocksOwner, u.height)
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorMaximumMissedBlocksOwner, u.height)
	case typesUtil.ProposerPercentageOfFeesParamName:
		return u.store.GetBytesParam(typesUtil.ProposerPercentageOfFeesOwner, u.height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		return u.store.GetBytesParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner, u.height)
	case typesUtil.MissedBlocksBurnPercentageParamName:
		return u.store.GetBytesParam(typesUtil.MissedBlocksBurnPercentageOwner, u.height)
	case typesUtil.DoubleSignBurnPercentageParamName:
		return u.store.GetBytesParam(typesUtil.DoubleSignBurnPercentageOwner, u.height)
	case typesUtil.MessageDoubleSignFee:
		return u.store.GetBytesParam(typesUtil.MessageDoubleSignFeeOwner, u.height)
	case typesUtil.MessageSendFee:
		return u.store.GetBytesParam(typesUtil.MessageSendFeeOwner, u.height)
	case typesUtil.MessageStakeFishermanFee:
		return u.store.GetBytesParam(typesUtil.MessageStakeFishermanFeeOwner, u.height)
	case typesUtil.MessageEditStakeFishermanFee:
		return u.store.GetBytesParam(typesUtil.MessageEditStakeFishermanFeeOwner, u.height)
	case typesUtil.MessageUnstakeFishermanFee:
		return u.store.GetBytesParam(typesUtil.MessageUnstakeFishermanFeeOwner, u.height)
	case typesUtil.MessagePauseFishermanFee:
		return u.store.GetBytesParam(typesUtil.MessagePauseFishermanFeeOwner, u.height)
	case typesUtil.MessageUnpauseFishermanFee:
		return u.store.GetBytesParam(typesUtil.MessageUnpauseFishermanFeeOwner, u.height)
	case typesUtil.MessageFishermanPauseServicerFee:
		return u.store.GetBytesParam(typesUtil.MessageFishermanPauseServicerFeeOwner, u.height)
	case typesUtil.MessageTestScoreFee:
		return u.store.GetBytesParam(typesUtil.MessageTestScoreFeeOwner, u.height)
	case typesUtil.MessageProveTestScoreFee:
		return u.store.GetBytesParam(typesUtil.MessageProveTestScoreFeeOwner, u.height)
	case typesUtil.MessageStakeAppFee:
		return u.store.GetBytesParam(typesUtil.MessageStakeAppFeeOwner, u.height)
	case typesUtil.MessageEditStakeAppFee:
		return u.store.GetBytesParam(typesUtil.MessageEditStakeAppFeeOwner, u.height)
	case typesUtil.MessageUnstakeAppFee:
		return u.store.GetBytesParam(typesUtil.MessageUnstakeAppFeeOwner, u.height)
	case typesUtil.MessagePauseAppFee:
		return u.store.GetBytesParam(typesUtil.MessagePauseAppFeeOwner, u.height)
	case typesUtil.MessageUnpauseAppFee:
		return u.store.GetBytesParam(typesUtil.MessageUnpauseAppFeeOwner, u.height)
	case typesUtil.MessageStakeValidatorFee:
		return u.store.GetBytesParam(typesUtil.MessageStakeValidatorFeeOwner, u.height)
	case typesUtil.MessageEditStakeValidatorFee:
		return u.store.GetBytesParam(typesUtil.MessageEditStakeValidatorFeeOwner, u.height)
	case typesUtil.MessageUnstakeValidatorFee:
		return u.store.GetBytesParam(typesUtil.MessageUnstakeValidatorFeeOwner, u.height)
	case typesUtil.MessagePauseValidatorFee:
		return u.store.GetBytesParam(typesUtil.MessagePauseValidatorFeeOwner, u.height)
	case typesUtil.MessageUnpauseValidatorFee:
		return u.store.GetBytesParam(typesUtil.MessageUnpauseValidatorFeeOwner, u.height)
	case typesUtil.MessageStakeServicerFee:
		return u.store.GetBytesParam(typesUtil.MessageStakeServicerFeeOwner, u.height)
	case typesUtil.MessageEditStakeServicerFee:
		return u.store.GetBytesParam(typesUtil.MessageEditStakeServicerFeeOwner, u.height)
	case typesUtil.MessageUnstakeServicerFee:
		return u.store.GetBytesParam(typesUtil.MessageUnstakeServicerFeeOwner, u.height)
	case typesUtil.MessagePauseServicerFee:
		return u.store.GetBytesParam(typesUtil.MessagePauseServicerFeeOwner, u.height)
	case typesUtil.MessageUnpauseServicerFee:
		return u.store.GetBytesParam(typesUtil.MessageUnpauseServicerFeeOwner, u.height)
	case typesUtil.MessageChangeParameterFee:
		return u.store.GetBytesParam(typesUtil.MessageChangeParameterFeeOwner, u.height)
	case typesUtil.BlocksPerSessionOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMaxChainsOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMinimumStakeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppSessionTokensMultiplierOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppUnstakingBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMinimumPauseBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMaxPausedBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMinimumStakeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMaxChainsOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerUnstakingBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMinimumPauseBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMaxPausedBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicersPerSessionOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMinimumStakeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMaxChainsOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanUnstakingBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMaxPausedBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMinimumStakeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorUnstakingBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ProposerPercentageOfFeesOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MissedBlocksBurnPercentageOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.DoubleSignBurnPercentageOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageSendFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeFishermanFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseFishermanFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageFishermanPauseServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageTestScoreFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageProveTestScoreFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeAppFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeAppFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeAppFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseAppFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseAppFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeValidatorFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseValidatorFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseServicerFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageChangeParameterFeeOwner:
		return u.store.GetBytesParam(typesUtil.AclOwner, u.height)
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
	value, err := u.store.GetStringParam(paramName, u.height)
	if err != nil {
		u.logger.Err(err)
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	amount, err := utils.StringToBigInt(value)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *utilityContext) getIntParam(paramName string) (int, typesUtil.Error) {
	value, err := u.store.GetIntParam(paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *utilityContext) getInt64Param(paramName string) (int64, typesUtil.Error) {
	value, err := u.store.GetIntParam(paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *utilityContext) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	value, err := u.store.GetBytesParam(paramName, u.height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}
