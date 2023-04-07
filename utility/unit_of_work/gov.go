package unit_of_work

import (
	"math/big"
	"strings"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func init() {
	govParamTypes = prepareGovParamParamTypesMap()
	for _, key := range utils.GovParamMetadataKeys {
		if isOwner := strings.Contains(key, "_owner"); isOwner {
			continue
		}
		if _, ok := govParamTypes[key]; !ok {
			logger.Global.Fatal().Msgf("govParamTypes map does not contain: %s", key)
		}
	}
}

var (
	govParamTypes map[string]int
)

const (
	BIGINT int = iota
	INT
	INT64
	BYTES
	STRING
)

func prepareGovParamParamTypesMap() map[string]int {
	return map[string]int{
		typesUtil.AppMinimumStakeParamName:                 BIGINT,
		typesUtil.AppMaxChainsParamName:                    INT,
		typesUtil.AppSessionTokensMultiplierParamName:      INT,
		typesUtil.AppUnstakingBlocksParamName:              INT64,
		typesUtil.AppMinimumPauseBlocksParamName:           INT,
		typesUtil.AppMaxPauseBlocksParamName:               INT,
		typesUtil.BlocksPerSessionParamName:                INT,
		typesUtil.ServicerMinimumStakeParamName:            BIGINT,
		typesUtil.ServicerMaxChainsParamName:               INT,
		typesUtil.ServicerUnstakingBlocksParamName:         INT64,
		typesUtil.ServicerMinimumPauseBlocksParamName:      INT,
		typesUtil.ServicerMaxPauseBlocksParamName:          INT,
		typesUtil.ServicersPerSessionParamName:             INT,
		typesUtil.ValidatorMinimumStakeParamName:           BIGINT,
		typesUtil.ValidatorUnstakingBlocksParamName:        INT64,
		typesUtil.ValidatorMinimumPauseBlocksParamName:     INT,
		typesUtil.ValidatorMaxPausedBlocksParamName:        INT,
		typesUtil.ProposerPercentageOfFeesParamName:        INT,
		typesUtil.ValidatorMaximumMissedBlocksParamName:    INT,
		typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName: INT,
		typesUtil.DoubleSignBurnPercentageParamName:        INT,
		typesUtil.MissedBlocksBurnPercentageParamName:      INT,
		typesUtil.FishermanMinimumStakeParamName:           BIGINT,
		typesUtil.FishermanMaxChainsParamName:              INT,
		typesUtil.FishermanUnstakingBlocksParamName:        INT64,
		typesUtil.FishermanMinimumPauseBlocksParamName:     INT,
		typesUtil.FishermanMaxPauseBlocksParamName:         INT,
		typesUtil.MessageDoubleSignFee:                     BIGINT,
		typesUtil.MessageSendFee:                           BIGINT,
		typesUtil.MessageStakeFishermanFee:                 BIGINT,
		typesUtil.MessageEditStakeFishermanFee:             BIGINT,
		typesUtil.MessageUnstakeFishermanFee:               BIGINT,
		typesUtil.MessagePauseFishermanFee:                 BIGINT,
		typesUtil.MessageUnpauseFishermanFee:               BIGINT,
		typesUtil.MessageFishermanPauseServicerFee:         BIGINT,
		typesUtil.MessageTestScoreFee:                      BIGINT,
		typesUtil.MessageProveTestScoreFee:                 BIGINT,
		typesUtil.MessageStakeAppFee:                       BIGINT,
		typesUtil.MessageEditStakeAppFee:                   BIGINT,
		typesUtil.MessageUnstakeAppFee:                     BIGINT,
		typesUtil.MessagePauseAppFee:                       BIGINT,
		typesUtil.MessageUnpauseAppFee:                     BIGINT,
		typesUtil.MessageStakeValidatorFee:                 BIGINT,
		typesUtil.MessageEditStakeValidatorFee:             BIGINT,
		typesUtil.MessageUnstakeValidatorFee:               BIGINT,
		typesUtil.MessagePauseValidatorFee:                 BIGINT,
		typesUtil.MessageUnpauseValidatorFee:               BIGINT,
		typesUtil.MessageStakeServicerFee:                  BIGINT,
		typesUtil.MessageEditStakeServicerFee:              BIGINT,
		typesUtil.MessageUnstakeServicerFee:                BIGINT,
		typesUtil.MessagePauseServicerFee:                  BIGINT,
		typesUtil.MessageUnpauseServicerFee:                BIGINT,
		typesUtil.MessageChangeParameterFee:                BIGINT,
	}
}

func getGovParam[T *big.Int | int | int64 | []byte | string](uow *baseUtilityUnitOfWork, paramName string) (i T, err typesUtil.Error) {
	switch tp := any(i).(type) {
	case *big.Int:
		v, er := uow.getBigIntParam(paramName)
		return any(v).(T), er
	case int:
		v, er := uow.getIntParam(paramName)
		return any(v).(T), er
	case int64:
		v, er := uow.getInt64Param(paramName)
		return any(v).(T), er
	case []byte:
		v, er := uow.getByteArrayParam(paramName)
		return any(v).(T), er
	case string:
		v, er := uow.getStringParam(paramName)
		return any(v).(T), er
	default:
		uow.logger.Fatal().Msgf("unhandled parameter type: %T", tp)
	}
	return
}

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

func (u *baseUtilityUnitOfWork) getParamOwner(paramName string) ([]byte, typesUtil.Error) {
	if paramOwner := utils.GovParamMetadataMap[paramName].ParamOwner; paramOwner != "" {
		return u.getByteArrayParam(paramOwner)
	}
	return nil, typesUtil.ErrUnknownParam(paramName)
}

func (u *baseUtilityUnitOfWork) getFee(msg typesUtil.Message, actorType coreTypes.ActorType) (amount *big.Int, err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageSend:
		return getGovParam[*big.Int](u, typesUtil.MessageSendFee)
	case *typesUtil.MessageStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return getGovParam[*big.Int](u, typesUtil.MessageStakeAppFee)
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return getGovParam[*big.Int](u, typesUtil.MessageStakeFishermanFee)
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return getGovParam[*big.Int](u, typesUtil.MessageStakeServicerFee)
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return getGovParam[*big.Int](u, typesUtil.MessageStakeValidatorFee)
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return getGovParam[*big.Int](u, typesUtil.MessageEditStakeAppFee)
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return getGovParam[*big.Int](u, typesUtil.MessageEditStakeFishermanFee)
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return getGovParam[*big.Int](u, typesUtil.MessageEditStakeServicerFee)
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return getGovParam[*big.Int](u, typesUtil.MessageEditStakeValidatorFee)
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return getGovParam[*big.Int](u, typesUtil.MessageUnstakeAppFee)
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return getGovParam[*big.Int](u, typesUtil.MessageUnstakeFishermanFee)
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return getGovParam[*big.Int](u, typesUtil.MessageUnstakeServicerFee)
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return getGovParam[*big.Int](u, typesUtil.MessageUnstakeValidatorFee)
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			return getGovParam[*big.Int](u, typesUtil.MessageUnpauseAppFee)
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			return getGovParam[*big.Int](u, typesUtil.MessageUnpauseFishermanFee)
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			return getGovParam[*big.Int](u, typesUtil.MessageUnpauseServicerFee)
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			return getGovParam[*big.Int](u, typesUtil.MessageUnpauseValidatorFee)
		default:
			return nil, typesUtil.ErrUnknownActorType(actorType.String())
		}
	case *typesUtil.MessageChangeParameter:
		return getGovParam[*big.Int](u, typesUtil.MessageChangeParameterFee)
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
	value, err := persistence.GetParameter[string](u.persistenceReadContext, paramName, u.height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	amount, err := utils.StringToBigInt(value)
	if err != nil {
		return nil, typesUtil.ErrStringToBigInt(err)
	}
	return amount, nil
}

func (u *baseUtilityUnitOfWork) getIntParam(paramName string) (int, typesUtil.Error) {
	value, err := persistence.GetParameter[int](u.persistenceReadContext, paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *baseUtilityUnitOfWork) getInt64Param(paramName string) (int64, typesUtil.Error) {
	value, err := persistence.GetParameter[int](u.persistenceReadContext, paramName, u.height)
	if err != nil {
		return 0, typesUtil.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *baseUtilityUnitOfWork) getByteArrayParam(paramName string) ([]byte, typesUtil.Error) {
	value, err := persistence.GetParameter[[]byte](u.persistenceReadContext, paramName, u.height)
	if err != nil {
		return nil, typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *baseUtilityUnitOfWork) getStringParam(paramName string) (string, typesUtil.Error) {
	value, err := persistence.GetParameter[string](u.persistenceReadContext, paramName, u.height)
	if err != nil {
		return "", typesUtil.ErrGetParam(paramName, err)
	}
	return value, nil
}
