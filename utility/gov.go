package utility

import (
	"fmt"
	"github.com/pokt-network/pocket/shared/modules"
	"log"
	"math/big"

	typesUtil "github.com/pokt-network/pocket/utility/types"
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
		return typesUtil.ZeroInt, typesUtil.ErrGetParam(modules.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.AppMinimumStakeParamName)
}

func (u *UtilityContext) GetAppMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(modules.AppMaxChainsParamName)
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, typesUtil.Error) {
	return u.getIntParam(modules.AppBaselineStakeRateParamName)
}

func (u *UtilityContext) GetStabilityAdjustment() (int, typesUtil.Error) {
	return u.getIntParam(modules.AppStakingAdjustmentParamName)
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(modules.AppUnstakingBlocksParamName)
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(modules.AppMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.AppMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.ServiceNodeMinimumStakeParamName)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(modules.ServiceNodeMaxChainsParamName)
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(modules.ServiceNodeUnstakingBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(modules.ServiceNodeMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.ServiceNodeMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.ValidatorMinimumStakeParamName)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(modules.ValidatorUnstakingBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(modules.ValidatorMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.ValidatorMaxPausedBlocksParamName)
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err typesUtil.Error) {
	return u.getIntParam(modules.ProposerPercentageOfFeesParamName)
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.ValidatorMaximumMissedBlocksParamName)
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(modules.DoubleSignBurnPercentageParamName)
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err typesUtil.Error) {
	return u.getIntParam(modules.MissedBlocksBurnPercentageParamName)
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.FishermanMinimumStakeParamName)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, typesUtil.Error) {
	return u.getIntParam(modules.FishermanMaxChainsParamName)
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, typesUtil.Error) {
	return u.getInt64Param(modules.FishermanUnstakingBlocksParamName)
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, typesUtil.Error) {
	return u.getIntParam(modules.FishermanMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err typesUtil.Error) {
	return u.getIntParam(modules.FishermanMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageDoubleSignFee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageSendFee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageStakeFishermanFee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageEditStakeFishermanFee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnstakeFishermanFee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessagePauseFishermanFee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnpauseFishermanFee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageFishermanPauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageTestScoreFee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageProveTestScoreFee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageStakeAppFee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageEditStakeAppFee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnstakeAppFee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessagePauseAppFee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnpauseAppFee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageStakeValidatorFee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageEditStakeValidatorFee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnstakeValidatorFee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessagePauseValidatorFee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnpauseValidatorFee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageEditStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnstakeServiceNodeFee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessagePauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageUnpauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, typesUtil.Error) {
	return u.getBigIntParam(modules.MessageChangeParameterFee)
}

func (u *UtilityContext) GetDoubleSignFeeOwner() (owner []byte, err typesUtil.Error) {
	return u.getByteArrayParam(modules.MessageDoubleSignFeeOwner)
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
	case modules.AclOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.BlocksPerSessionParamName:
		return store.GetBytesParam(modules.BlocksPerSessionOwner, height)
	case modules.AppMaxChainsParamName:
		return store.GetBytesParam(modules.AppMaxChainsOwner, height)
	case modules.AppMinimumStakeParamName:
		return store.GetBytesParam(modules.AppMinimumStakeOwner, height)
	case modules.AppBaselineStakeRateParamName:
		return store.GetBytesParam(modules.AppBaselineStakeRateOwner, height)
	case modules.AppStakingAdjustmentParamName:
		return store.GetBytesParam(modules.AppStakingAdjustmentOwner, height)
	case modules.AppUnstakingBlocksParamName:
		return store.GetBytesParam(modules.AppUnstakingBlocksOwner, height)
	case modules.AppMinimumPauseBlocksParamName:
		return store.GetBytesParam(modules.AppMinimumPauseBlocksOwner, height)
	case modules.AppMaxPauseBlocksParamName:
		return store.GetBytesParam(modules.AppMaxPausedBlocksOwner, height)
	case modules.ServiceNodesPerSessionParamName:
		return store.GetBytesParam(modules.ServiceNodesPerSessionOwner, height)
	case modules.ServiceNodeMinimumStakeParamName:
		return store.GetBytesParam(modules.ServiceNodeMinimumStakeOwner, height)
	case modules.ServiceNodeMaxChainsParamName:
		return store.GetBytesParam(modules.ServiceNodeMaxChainsOwner, height)
	case modules.ServiceNodeUnstakingBlocksParamName:
		return store.GetBytesParam(modules.ServiceNodeUnstakingBlocksOwner, height)
	case modules.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetBytesParam(modules.ServiceNodeMinimumPauseBlocksOwner, height)
	case modules.ServiceNodeMaxPauseBlocksParamName:
		return store.GetBytesParam(modules.ServiceNodeMaxPausedBlocksOwner, height)
	case modules.FishermanMinimumStakeParamName:
		return store.GetBytesParam(modules.FishermanMinimumStakeOwner, height)
	case modules.FishermanMaxChainsParamName:
		return store.GetBytesParam(modules.FishermanMaxChainsOwner, height)
	case modules.FishermanUnstakingBlocksParamName:
		return store.GetBytesParam(modules.FishermanUnstakingBlocksOwner, height)
	case modules.FishermanMinimumPauseBlocksParamName:
		return store.GetBytesParam(modules.FishermanMinimumPauseBlocksOwner, height)
	case modules.FishermanMaxPauseBlocksParamName:
		return store.GetBytesParam(modules.FishermanMaxPausedBlocksOwner, height)
	case modules.ValidatorMinimumStakeParamName:
		return store.GetBytesParam(modules.ValidatorMinimumStakeOwner, height)
	case modules.ValidatorUnstakingBlocksParamName:
		return store.GetBytesParam(modules.ValidatorUnstakingBlocksOwner, height)
	case modules.ValidatorMinimumPauseBlocksParamName:
		return store.GetBytesParam(modules.ValidatorMinimumPauseBlocksOwner, height)
	case modules.ValidatorMaxPausedBlocksParamName:
		return store.GetBytesParam(modules.ValidatorMaxPausedBlocksOwner, height)
	case modules.ValidatorMaximumMissedBlocksParamName:
		return store.GetBytesParam(modules.ValidatorMaximumMissedBlocksOwner, height)
	case modules.ProposerPercentageOfFeesParamName:
		return store.GetBytesParam(modules.ProposerPercentageOfFeesOwner, height)
	case modules.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetBytesParam(modules.ValidatorMaxEvidenceAgeInBlocksOwner, height)
	case modules.MissedBlocksBurnPercentageParamName:
		return store.GetBytesParam(modules.MissedBlocksBurnPercentageOwner, height)
	case modules.DoubleSignBurnPercentageParamName:
		return store.GetBytesParam(modules.DoubleSignBurnPercentageOwner, height)
	case modules.MessageDoubleSignFee:
		return store.GetBytesParam(modules.MessageDoubleSignFeeOwner, height)
	case modules.MessageSendFee:
		return store.GetBytesParam(modules.MessageSendFeeOwner, height)
	case modules.MessageStakeFishermanFee:
		return store.GetBytesParam(modules.MessageStakeFishermanFeeOwner, height)
	case modules.MessageEditStakeFishermanFee:
		return store.GetBytesParam(modules.MessageEditStakeFishermanFeeOwner, height)
	case modules.MessageUnstakeFishermanFee:
		return store.GetBytesParam(modules.MessageUnstakeFishermanFeeOwner, height)
	case modules.MessagePauseFishermanFee:
		return store.GetBytesParam(modules.MessagePauseFishermanFeeOwner, height)
	case modules.MessageUnpauseFishermanFee:
		return store.GetBytesParam(modules.MessageUnpauseFishermanFeeOwner, height)
	case modules.MessageFishermanPauseServiceNodeFee:
		return store.GetBytesParam(modules.MessageFishermanPauseServiceNodeFeeOwner, height)
	case modules.MessageTestScoreFee:
		return store.GetBytesParam(modules.MessageTestScoreFeeOwner, height)
	case modules.MessageProveTestScoreFee:
		return store.GetBytesParam(modules.MessageProveTestScoreFeeOwner, height)
	case modules.MessageStakeAppFee:
		return store.GetBytesParam(modules.MessageStakeAppFeeOwner, height)
	case modules.MessageEditStakeAppFee:
		return store.GetBytesParam(modules.MessageEditStakeAppFeeOwner, height)
	case modules.MessageUnstakeAppFee:
		return store.GetBytesParam(modules.MessageUnstakeAppFeeOwner, height)
	case modules.MessagePauseAppFee:
		return store.GetBytesParam(modules.MessagePauseAppFeeOwner, height)
	case modules.MessageUnpauseAppFee:
		return store.GetBytesParam(modules.MessageUnpauseAppFeeOwner, height)
	case modules.MessageStakeValidatorFee:
		return store.GetBytesParam(modules.MessageStakeValidatorFeeOwner, height)
	case modules.MessageEditStakeValidatorFee:
		return store.GetBytesParam(modules.MessageEditStakeValidatorFeeOwner, height)
	case modules.MessageUnstakeValidatorFee:
		return store.GetBytesParam(modules.MessageUnstakeValidatorFeeOwner, height)
	case modules.MessagePauseValidatorFee:
		return store.GetBytesParam(modules.MessagePauseValidatorFeeOwner, height)
	case modules.MessageUnpauseValidatorFee:
		return store.GetBytesParam(modules.MessageUnpauseValidatorFeeOwner, height)
	case modules.MessageStakeServiceNodeFee:
		return store.GetBytesParam(modules.MessageStakeServiceNodeFeeOwner, height)
	case modules.MessageEditStakeServiceNodeFee:
		return store.GetBytesParam(modules.MessageEditStakeServiceNodeFeeOwner, height)
	case modules.MessageUnstakeServiceNodeFee:
		return store.GetBytesParam(modules.MessageUnstakeServiceNodeFeeOwner, height)
	case modules.MessagePauseServiceNodeFee:
		return store.GetBytesParam(modules.MessagePauseServiceNodeFeeOwner, height)
	case modules.MessageUnpauseServiceNodeFee:
		return store.GetBytesParam(modules.MessageUnpauseServiceNodeFeeOwner, height)
	case modules.MessageChangeParameterFee:
		return store.GetBytesParam(modules.MessageChangeParameterFeeOwner, height)
	case modules.BlocksPerSessionOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppMaxChainsOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppMinimumStakeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppBaselineStakeRateOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppStakingAdjustmentOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppUnstakingBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppMinimumPauseBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.AppMaxPausedBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodeMinimumStakeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodeMaxChainsOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodeUnstakingBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodeMaxPausedBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ServiceNodesPerSessionOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.FishermanMinimumStakeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.FishermanMaxChainsOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.FishermanUnstakingBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.FishermanMinimumPauseBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.FishermanMaxPausedBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorMinimumStakeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorUnstakingBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorMinimumPauseBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorMaxPausedBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorMaximumMissedBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ProposerPercentageOfFeesOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MissedBlocksBurnPercentageOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.DoubleSignBurnPercentageOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageSendFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageStakeFishermanFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageEditStakeFishermanFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnstakeFishermanFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessagePauseFishermanFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnpauseFishermanFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageTestScoreFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageProveTestScoreFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageStakeAppFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageEditStakeAppFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnstakeAppFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessagePauseAppFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnpauseAppFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageStakeValidatorFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageEditStakeValidatorFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnstakeValidatorFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessagePauseValidatorFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnpauseValidatorFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageStakeServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageEditStakeServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnstakeServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessagePauseServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageUnpauseServiceNodeFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	case modules.MessageChangeParameterFeeOwner:
		return store.GetBytesParam(modules.AclOwner, height)
	default:
		return nil, typesUtil.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg typesUtil.Message, actorType typesUtil.UtilActorType) (amount *big.Int, err typesUtil.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *typesUtil.MessageSend:
		return u.GetMessageSendFee()
	case *typesUtil.MessageStake:
		switch actorType {
		case typesUtil.UtilActorType_App:
			return u.GetMessageStakeAppFee()
		case typesUtil.UtilActorType_Fish:
			return u.GetMessageStakeFishermanFee()
		case typesUtil.UtilActorType_Node:
			return u.GetMessageStakeServiceNodeFee()
		case typesUtil.UtilActorType_Val:
			return u.GetMessageStakeValidatorFee()
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case typesUtil.UtilActorType_App:
			return u.GetMessageEditStakeAppFee()
		case typesUtil.UtilActorType_Fish:
			return u.GetMessageEditStakeFishermanFee()
		case typesUtil.UtilActorType_Node:
			return u.GetMessageEditStakeServiceNodeFee()
		case typesUtil.UtilActorType_Val:
			return u.GetMessageEditStakeValidatorFee()
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case typesUtil.UtilActorType_App:
			return u.GetMessageUnstakeAppFee()
		case typesUtil.UtilActorType_Fish:
			return u.GetMessageUnstakeFishermanFee()
		case typesUtil.UtilActorType_Node:
			return u.GetMessageUnstakeServiceNodeFee()
		case typesUtil.UtilActorType_Val:
			return u.GetMessageUnstakeValidatorFee()
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case typesUtil.UtilActorType_App:
			return u.GetMessageUnpauseAppFee()
		case typesUtil.UtilActorType_Fish:
			return u.GetMessageUnpauseFishermanFee()
		case typesUtil.UtilActorType_Node:
			return u.GetMessageUnpauseServiceNodeFee()
		case typesUtil.UtilActorType_Val:
			return u.GetMessageUnpauseValidatorFee()
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
	var err error
	store, height, er := u.GetStoreAndHeight()
	if er != nil {
		return nil, er
	}
	value, err := store.GetStringParam(paramName, height)
	if err != nil {
		fmt.Printf("err: %v\n", err)
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
