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
	if paramOwner := utils.GovParamMetadataMap[paramName].ParamOwner; paramOwner != "" {
		return u.persistenceReadContext.GetBytesParam(paramOwner, u.height)
	}
	return u.persistenceReadContext.GetBytesParam(paramName, u.height)
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
