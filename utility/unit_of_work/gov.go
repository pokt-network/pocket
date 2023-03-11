package unit_of_work

import (
	"math/big"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *baseUtilityUnitOfWork) updateParam(paramName string, value any) typesUtil.Error {
	switch t := value.(type) {
	case *wrapperspb.Int32Value:
		if err := u.persistenceRWContext.SetParam(paramName, (int(t.Value))); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.StringValue:
		if err := u.persistenceRWContext.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.BytesValue:
		if err := u.persistenceRWContext.SetParam(paramName, t.Value); err != nil {
			return typesUtil.ErrUpdateParam(err)
		}
		return nil
	default:
		break
	}
	u.logger.Fatal().Msgf("unhandled value type %T for %v", value, value)
	return typesUtil.ErrUnknownParam(paramName)
}

func (u *baseUtilityUnitOfWork) getParameter(paramName string) (any, error) {
	return u.persistenceReadContext.GetParameter(paramName, u.height)
}

func (u *baseUtilityUnitOfWork) getAppMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.AppMinimumStakeParamName)
}

func (u *baseUtilityUnitOfWork) getAppMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxChainsParamName)
}

func (u *baseUtilityUnitOfWork) getAppSessionTokensMultiplier() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppSessionTokensMultiplierParamName)
}

func (u *baseUtilityUnitOfWork) getAppUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.AppUnstakingBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getAppMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMinimumPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getAppMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.AppMaxPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getServicerMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ServicerMinimumStakeParamName)
}

func (u *baseUtilityUnitOfWork) getServicerMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMaxChainsParamName)
}

func (u *baseUtilityUnitOfWork) getServicerUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ServicerUnstakingBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getServicerMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMinimumPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getServicerMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ServicerMaxPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getValidatorMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.ValidatorMinimumStakeParamName)
}

func (u *baseUtilityUnitOfWork) getValidatorUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.ValidatorUnstakingBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getValidatorMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMinimumPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getValidatorMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxPausedBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getProposerPercentageOfFees() (proposerPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ProposerPercentageOfFeesParamName)
}

func (u *baseUtilityUnitOfWork) getValidatorMaxMissedBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaximumMissedBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getDoubleSignBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.DoubleSignBurnPercentageParamName)
}

func (u *baseUtilityUnitOfWork) getMissedBlocksBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.MissedBlocksBurnPercentageParamName)
}

func (u *baseUtilityUnitOfWork) getFishermanMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.FishermanMinimumStakeParamName)
}

func (u *baseUtilityUnitOfWork) getFishermanMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxChainsParamName)
}

func (u *baseUtilityUnitOfWork) getFishermanUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(typesUtil.FishermanUnstakingBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getFishermanMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMinimumPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getFishermanMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(typesUtil.FishermanMaxPauseBlocksParamName)
}

func (u *baseUtilityUnitOfWork) getMessageDoubleSignFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageDoubleSignFee)
}

func (u *baseUtilityUnitOfWork) getMessageSendFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageSendFee)
}

func (u *baseUtilityUnitOfWork) getMessageStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeFishermanFee)
}

func (u *baseUtilityUnitOfWork) getMessageEditStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeFishermanFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnstakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeFishermanFee)
}

func (u *baseUtilityUnitOfWork) getMessagePauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseFishermanFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnpauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseFishermanFee)
}

func (u *baseUtilityUnitOfWork) getMessageFishermanPauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageFishermanPauseServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessageTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageTestScoreFee)
}

func (u *baseUtilityUnitOfWork) getMessageProveTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageProveTestScoreFee)
}

func (u *baseUtilityUnitOfWork) getMessageStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeAppFee)
}

func (u *baseUtilityUnitOfWork) getMessageEditStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeAppFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnstakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeAppFee)
}

func (u *baseUtilityUnitOfWork) getMessagePauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseAppFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnpauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseAppFee)
}

func (u *baseUtilityUnitOfWork) getMessageStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeValidatorFee)
}

func (u *baseUtilityUnitOfWork) getMessageEditStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeValidatorFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnstakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeValidatorFee)
}

func (u *baseUtilityUnitOfWork) getMessagePauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseValidatorFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnpauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseValidatorFee)
}

func (u *baseUtilityUnitOfWork) getMessageStakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageStakeServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessageEditStakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageEditStakeServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnstakeServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnstakeServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessagePauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessagePauseServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessageUnpauseServicerFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageUnpauseServicerFee)
}

func (u *baseUtilityUnitOfWork) getMessageChangeParameterFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(typesUtil.MessageChangeParameterFee)
}

func (u *baseUtilityUnitOfWork) getDoubleSignFeeOwner() (owner []byte, err typesUtil.Error) {
	return u.getByteArrayParam(typesUtil.MessageDoubleSignFeeOwner)
}

func (u *baseUtilityUnitOfWork) getParamOwner(paramName string) ([]byte, error) {
	// DISCUSS (@deblasis): here we could potentially leverage the struct tags in gov.proto by specifying an `owner` key
	// eg: `app_minimum_stake` could have `pokt:"owner=app_minimum_stake_owner"`
	// in here we would use that map to point to the owner, removing this switch, centralizing the logic and making it declarative
	switch paramName {
	case typesUtil.AclOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.BlocksPerSessionParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.BlocksPerSessionOwner, u.height)
	case typesUtil.AppMaxChainsParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppMaxChainsOwner, u.height)
	case typesUtil.AppMinimumStakeParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppMinimumStakeOwner, u.height)
	case typesUtil.AppSessionTokensMultiplierParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppSessionTokensMultiplierOwner, u.height)
	case typesUtil.AppUnstakingBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppUnstakingBlocksOwner, u.height)
	case typesUtil.AppMinimumPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppMinimumPauseBlocksOwner, u.height)
	case typesUtil.AppMaxPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AppMaxPausedBlocksOwner, u.height)
	case typesUtil.ServicersPerSessionParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicersPerSessionOwner, u.height)
	case typesUtil.ServicerMinimumStakeParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicerMinimumStakeOwner, u.height)
	case typesUtil.ServicerMaxChainsParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicerMaxChainsOwner, u.height)
	case typesUtil.ServicerUnstakingBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicerUnstakingBlocksOwner, u.height)
	case typesUtil.ServicerMinimumPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicerMinimumPauseBlocksOwner, u.height)
	case typesUtil.ServicerMaxPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ServicerMaxPausedBlocksOwner, u.height)
	case typesUtil.FishermanMinimumStakeParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.FishermanMinimumStakeOwner, u.height)
	case typesUtil.FishermanMaxChainsParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.FishermanMaxChainsOwner, u.height)
	case typesUtil.FishermanUnstakingBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.FishermanUnstakingBlocksOwner, u.height)
	case typesUtil.FishermanMinimumPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.FishermanMinimumPauseBlocksOwner, u.height)
	case typesUtil.FishermanMaxPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.FishermanMaxPausedBlocksOwner, u.height)
	case typesUtil.ValidatorMinimumStakeParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorMinimumStakeOwner, u.height)
	case typesUtil.ValidatorUnstakingBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorUnstakingBlocksOwner, u.height)
	case typesUtil.ValidatorMinimumPauseBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorMinimumPauseBlocksOwner, u.height)
	case typesUtil.ValidatorMaxPausedBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorMaxPausedBlocksOwner, u.height)
	case typesUtil.ValidatorMaximumMissedBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorMaximumMissedBlocksOwner, u.height)
	case typesUtil.ProposerPercentageOfFeesParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ProposerPercentageOfFeesOwner, u.height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner, u.height)
	case typesUtil.MissedBlocksBurnPercentageParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MissedBlocksBurnPercentageOwner, u.height)
	case typesUtil.DoubleSignBurnPercentageParamName:
		return u.persistenceReadContext.GetBytesParam(typesUtil.DoubleSignBurnPercentageOwner, u.height)
	case typesUtil.MessageDoubleSignFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageDoubleSignFeeOwner, u.height)
	case typesUtil.MessageSendFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageSendFeeOwner, u.height)
	case typesUtil.MessageStakeFishermanFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageStakeFishermanFeeOwner, u.height)
	case typesUtil.MessageEditStakeFishermanFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageEditStakeFishermanFeeOwner, u.height)
	case typesUtil.MessageUnstakeFishermanFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnstakeFishermanFeeOwner, u.height)
	case typesUtil.MessagePauseFishermanFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessagePauseFishermanFeeOwner, u.height)
	case typesUtil.MessageUnpauseFishermanFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnpauseFishermanFeeOwner, u.height)
	case typesUtil.MessageFishermanPauseServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageFishermanPauseServicerFeeOwner, u.height)
	case typesUtil.MessageTestScoreFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageTestScoreFeeOwner, u.height)
	case typesUtil.MessageProveTestScoreFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageProveTestScoreFeeOwner, u.height)
	case typesUtil.MessageStakeAppFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageStakeAppFeeOwner, u.height)
	case typesUtil.MessageEditStakeAppFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageEditStakeAppFeeOwner, u.height)
	case typesUtil.MessageUnstakeAppFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnstakeAppFeeOwner, u.height)
	case typesUtil.MessagePauseAppFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessagePauseAppFeeOwner, u.height)
	case typesUtil.MessageUnpauseAppFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnpauseAppFeeOwner, u.height)
	case typesUtil.MessageStakeValidatorFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageStakeValidatorFeeOwner, u.height)
	case typesUtil.MessageEditStakeValidatorFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageEditStakeValidatorFeeOwner, u.height)
	case typesUtil.MessageUnstakeValidatorFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnstakeValidatorFeeOwner, u.height)
	case typesUtil.MessagePauseValidatorFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessagePauseValidatorFeeOwner, u.height)
	case typesUtil.MessageUnpauseValidatorFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnpauseValidatorFeeOwner, u.height)
	case typesUtil.MessageStakeServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageStakeServicerFeeOwner, u.height)
	case typesUtil.MessageEditStakeServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageEditStakeServicerFeeOwner, u.height)
	case typesUtil.MessageUnstakeServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnstakeServicerFeeOwner, u.height)
	case typesUtil.MessagePauseServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessagePauseServicerFeeOwner, u.height)
	case typesUtil.MessageUnpauseServicerFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageUnpauseServicerFeeOwner, u.height)
	case typesUtil.MessageChangeParameterFee:
		return u.persistenceReadContext.GetBytesParam(typesUtil.MessageChangeParameterFeeOwner, u.height)
	case typesUtil.BlocksPerSessionOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMaxChainsOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMinimumStakeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppSessionTokensMultiplierOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppUnstakingBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMinimumPauseBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.AppMaxPausedBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMinimumStakeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMaxChainsOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerUnstakingBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMinimumPauseBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicerMaxPausedBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ServicersPerSessionOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMinimumStakeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMaxChainsOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanUnstakingBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMinimumPauseBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.FishermanMaxPausedBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMinimumStakeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorUnstakingBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMinimumPauseBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaxPausedBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaximumMissedBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ProposerPercentageOfFeesOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MissedBlocksBurnPercentageOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.DoubleSignBurnPercentageOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageSendFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeFishermanFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeFishermanFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeFishermanFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseFishermanFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseFishermanFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageFishermanPauseServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageTestScoreFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageProveTestScoreFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeAppFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeAppFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeAppFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseAppFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseAppFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeValidatorFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeValidatorFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeValidatorFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseValidatorFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseValidatorFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageStakeServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageEditStakeServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnstakeServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessagePauseServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageUnpauseServicerFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	case typesUtil.MessageChangeParameterFeeOwner:
		return u.persistenceReadContext.GetBytesParam(typesUtil.AclOwner, u.height)
	default:
		return nil, typesUtil.ErrUnknownParam(paramName)
	}
}

func (u *baseUtilityUnitOfWork) getFee(msg typesUtil.Message, actorType coreTypes.ActorType) (amount *big.Int, err typesUtil.Error) {
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

func (u *baseUtilityUnitOfWork) getMessageChangeParameterSignerCandidates(msg *typesUtil.MessageChangeParameter) ([][]byte, typesUtil.Error) {
	owner, err := u.getParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, typesUtil.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}

func (u *baseUtilityUnitOfWork) getBigIntParam(paramName string) (*big.Int, typesUtil.Error) {
	value, err := u.persistenceReadContext.GetStringParam(paramName, u.height)
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

func (u *baseUtilityUnitOfWork) getIntParam(paramName string) (int, typesUtil.Error) {
	value, err := u.persistenceReadContext.GetIntParam(paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *baseUtilityUnitOfWork) getInt64Param(paramName string) (int64, typesUtil.Error) {
	value, err := u.persistenceReadContext.GetIntParam(paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *baseUtilityUnitOfWork) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	value, err := u.persistenceReadContext.GetBytesParam(paramName, u.height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}
