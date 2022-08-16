package utility

import (
	"fmt"
	"log"
	"math/big"

	"github.com/pokt-network/pocket/shared/types"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (u *UtilityContext) UpdateParam(paramName string, value interface{}) types.Error {
	store := u.Store()
	switch t := value.(type) {
	case *wrapperspb.Int32Value:
		if err := store.SetParam(types.BlocksPerSessionParamName, (int(t.Value))); err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.StringValue:
		if err := store.SetParam(types.BlocksPerSessionParamName, t.Value); err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	case *wrapperspb.BytesValue:
		if err := store.SetParam(types.BlocksPerSessionParamName, t.Value); err != nil {
			return types.ErrUpdateParam(err)
		}
		return nil
	default:
		break
	}
	log.Fatalf("unhandled value type %T for %v", value, value)
	return types.ErrUnknownParam(paramName)
}

func (u *UtilityContext) GetBlocksPerSession() (int, types.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(types.BlocksPerSessionParamName, err)
	}
	blocksPerSession, err := store.GetBlocksPerSession(height)
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(types.BlocksPerSessionParamName, err)
	}
	return blocksPerSession, nil
}

func (u *UtilityContext) GetAppMinimumStake() (*big.Int, types.Error) {
	return u.getBigIntParam(types.AppMinimumStakeParamName)
}

func (u *UtilityContext) GetAppMaxChains() (int, types.Error) {
	return u.getIntParam(types.AppMaxChainsParamName)
}

func (u *UtilityContext) GetBaselineAppStakeRate() (int, types.Error) {
	return u.getIntParam(types.AppBaselineStakeRateParamName)
}

func (u *UtilityContext) GetStabilityAdjustment() (int, types.Error) {
	return u.getIntParam(types.AppMaxChainsParamName)
}

func (u *UtilityContext) GetAppUnstakingBlocks() (int64, types.Error) {
	return u.getInt64Param(types.AppUnstakingBlocksParamName)
}

func (u *UtilityContext) GetAppMinimumPauseBlocks() (int, types.Error) {
	return u.getIntParam(types.AppMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetAppMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	return u.getIntParam(types.AppMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumStake() (*big.Int, types.Error) {
	return u.getBigIntParam(types.ServiceNodeMinimumStakeParamName)
}

func (u *UtilityContext) GetServiceNodeMaxChains() (int, types.Error) {
	return u.getIntParam(types.ServiceNodeMaxChainsParamName)
}

func (u *UtilityContext) GetServiceNodeUnstakingBlocks() (int64, types.Error) {
	return u.getInt64Param(types.ServiceNodeUnstakingBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMinimumPauseBlocks() (int, types.Error) {
	return u.getIntParam(types.ServiceNodeMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetServiceNodeMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	return u.getIntParam(types.ServiceNodeMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumStake() (*big.Int, types.Error) {
	return u.getBigIntParam(types.ValidatorMinimumStakeParamName)
}

func (u *UtilityContext) GetValidatorUnstakingBlocks() (int64, types.Error) {
	return u.getInt64Param(types.ValidatorUnstakingBlocksParamName)
}

func (u *UtilityContext) GetValidatorMinimumPauseBlocks() (int, types.Error) {
	return u.getIntParam(types.ValidatorMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetValidatorMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	return u.getIntParam(types.ValidatorMaxPausedBlocksParamName)
}

func (u *UtilityContext) GetProposerPercentageOfFees() (proposerPercentage int, err types.Error) {
	return u.getIntParam(types.ProposerPercentageOfFeesParamName)
}

func (u *UtilityContext) GetValidatorMaxMissedBlocks() (maxMissedBlocks int, err types.Error) {
	return u.getIntParam(types.ValidatorMaximumMissedBlocksParamName)
}

func (u *UtilityContext) GetMaxEvidenceAgeInBlocks() (maxMissedBlocks int, err types.Error) {
	return u.getIntParam(types.ValidatorMaxEvidenceAgeInBlocksParamName)
}

func (u *UtilityContext) GetDoubleSignBurnPercentage() (burnPercentage int, err types.Error) {
	return u.getIntParam(types.DoubleSignBurnPercentageParamName)
}

func (u *UtilityContext) GetMissedBlocksBurnPercentage() (burnPercentage int, err types.Error) {
	return u.getIntParam(types.MissedBlocksBurnPercentageParamName)
}

func (u *UtilityContext) GetFishermanMinimumStake() (*big.Int, types.Error) {
	return u.getBigIntParam(types.FishermanMinimumStakeParamName)
}

func (u *UtilityContext) GetFishermanMaxChains() (int, types.Error) {
	return u.getIntParam(types.FishermanMaxChainsParamName)
}

func (u *UtilityContext) GetFishermanUnstakingBlocks() (int64, types.Error) {
	return u.getInt64Param(types.FishermanUnstakingBlocksParamName)
}

func (u *UtilityContext) GetFishermanMinimumPauseBlocks() (int, types.Error) {
	return u.getIntParam(types.FishermanMinimumPauseBlocksParamName)
}

func (u *UtilityContext) GetFishermanMaxPausedBlocks() (maxPausedBlocks int, err types.Error) {
	return u.getIntParam(types.FishermanMaxPauseBlocksParamName)
}

func (u *UtilityContext) GetMessageDoubleSignFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageDoubleSignFee)
}

func (u *UtilityContext) GetMessageSendFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageSendFee)
}

func (u *UtilityContext) GetMessageStakeFishermanFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageStakeFishermanFee)
}

func (u *UtilityContext) GetMessageEditStakeFishermanFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageEditStakeFishermanFee)
}

func (u *UtilityContext) GetMessageUnstakeFishermanFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnstakeFishermanFee)
}

func (u *UtilityContext) GetMessagePauseFishermanFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessagePauseFishermanFee)
}

func (u *UtilityContext) GetMessageUnpauseFishermanFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnpauseFishermanFee)
}

func (u *UtilityContext) GetMessageFishermanPauseServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageFishermanPauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageTestScoreFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageTestScoreFee)
}

func (u *UtilityContext) GetMessageProveTestScoreFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageProveTestScoreFee)
}

func (u *UtilityContext) GetMessageStakeAppFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageStakeAppFee)
}

func (u *UtilityContext) GetMessageEditStakeAppFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageEditStakeAppFee)
}

func (u *UtilityContext) GetMessageUnstakeAppFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnstakeAppFee)
}

func (u *UtilityContext) GetMessagePauseAppFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessagePauseAppFee)
}

func (u *UtilityContext) GetMessageUnpauseAppFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnpauseAppFee)
}

func (u *UtilityContext) GetMessageStakeValidatorFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageStakeValidatorFee)
}

func (u *UtilityContext) GetMessageEditStakeValidatorFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageEditStakeValidatorFee)
}

func (u *UtilityContext) GetMessageUnstakeValidatorFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnstakeValidatorFee)
}

func (u *UtilityContext) GetMessagePauseValidatorFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessagePauseValidatorFee)
}

func (u *UtilityContext) GetMessageUnpauseValidatorFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnpauseValidatorFee)
}

func (u *UtilityContext) GetMessageStakeServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageEditStakeServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageEditStakeServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnstakeServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnstakeServiceNodeFee)
}

func (u *UtilityContext) GetMessagePauseServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessagePauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageUnpauseServiceNodeFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageUnpauseServiceNodeFee)
}

func (u *UtilityContext) GetMessageChangeParameterFee() (*big.Int, types.Error) {
	return u.getBigIntParam(types.MessageChangeParameterFee)
}

func (u *UtilityContext) GetDoubleSignFeeOwner() (owner []byte, err types.Error) {
	return u.getByteArrayParam(types.MessageDoubleSignFeeOwner)
}

func (u *UtilityContext) GetParamOwner(paramName string) ([]byte, error) {
	// DISCUSS (@deblasis): here we could potentially leverage the struct tags in gov.proto by specifying an `owner` key
	// eg: `app_minimum_stake` could have `pokt:"owner=app_minimum_stake_owner"`
	// in here we would use that map to point to the owner, removing this switch, centralizing the logic and making it declarative
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, err
	}
	switch paramName {
	case types.AclOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.BlocksPerSessionParamName:
		return store.GetBytesParam(types.BlocksPerSessionOwner, height)
	case types.AppMaxChainsParamName:
		return store.GetBytesParam(types.AppMaxChainsOwner, height)
	case types.AppMinimumStakeParamName:
		return store.GetBytesParam(types.AppMinimumStakeOwner, height)
	case types.AppBaselineStakeRateParamName:
		return store.GetBytesParam(types.AppBaselineStakeRateOwner, height)
	case types.AppStakingAdjustmentParamName:
		return store.GetBytesParam(types.AppStakingAdjustmentOwner, height)
	case types.AppUnstakingBlocksParamName:
		return store.GetBytesParam(types.AppUnstakingBlocksOwner, height)
	case types.AppMinimumPauseBlocksParamName:
		return store.GetBytesParam(types.AppMinimumPauseBlocksOwner, height)
	case types.AppMaxPauseBlocksParamName:
		return store.GetBytesParam(types.AppMaxPausedBlocksOwner, height)
	case types.ServiceNodesPerSessionParamName:
		return store.GetBytesParam(types.ServiceNodesPerSessionOwner, height)
	case types.ServiceNodeMinimumStakeParamName:
		return store.GetBytesParam(types.ServiceNodeMinimumStakeOwner, height)
	case types.ServiceNodeMaxChainsParamName:
		return store.GetBytesParam(types.ServiceNodeMaxChainsOwner, height)
	case types.ServiceNodeUnstakingBlocksParamName:
		return store.GetBytesParam(types.ServiceNodeUnstakingBlocksOwner, height)
	case types.ServiceNodeMinimumPauseBlocksParamName:
		return store.GetBytesParam(types.ServiceNodeMinimumPauseBlocksOwner, height)
	case types.ServiceNodeMaxPauseBlocksParamName:
		return store.GetBytesParam(types.ServiceNodeMaxPausedBlocksOwner, height)
	case types.FishermanMinimumStakeParamName:
		return store.GetBytesParam(types.FishermanMinimumStakeOwner, height)
	case types.FishermanMaxChainsParamName:
		return store.GetBytesParam(types.FishermanMaxChainsOwner, height)
	case types.FishermanUnstakingBlocksParamName:
		return store.GetBytesParam(types.FishermanUnstakingBlocksOwner, height)
	case types.FishermanMinimumPauseBlocksParamName:
		return store.GetBytesParam(types.FishermanMinimumPauseBlocksOwner, height)
	case types.FishermanMaxPauseBlocksParamName:
		return store.GetBytesParam(types.FishermanMaxPausedBlocksOwner, height)
	case types.ValidatorMinimumStakeParamName:
		return store.GetBytesParam(types.ValidatorMinimumStakeOwner, height)
	case types.ValidatorUnstakingBlocksParamName:
		return store.GetBytesParam(types.ValidatorUnstakingBlocksOwner, height)
	case types.ValidatorMinimumPauseBlocksParamName:
		return store.GetBytesParam(types.ValidatorMinimumPauseBlocksOwner, height)
	case types.ValidatorMaxPausedBlocksParamName:
		return store.GetBytesParam(types.ValidatorMaxPausedBlocksOwner, height)
	case types.ValidatorMaximumMissedBlocksParamName:
		return store.GetBytesParam(types.ValidatorMaximumMissedBlocksOwner, height)
	case types.ProposerPercentageOfFeesParamName:
		return store.GetBytesParam(types.ProposerPercentageOfFeesOwner, height)
	case types.ValidatorMaxEvidenceAgeInBlocksParamName:
		return store.GetBytesParam(types.ValidatorMaxEvidenceAgeInBlocksOwner, height)
	case types.MissedBlocksBurnPercentageParamName:
		return store.GetBytesParam(types.MissedBlocksBurnPercentageOwner, height)
	case types.DoubleSignBurnPercentageParamName:
		return store.GetBytesParam(types.DoubleSignBurnPercentageOwner, height)
	case types.MessageDoubleSignFee:
		return store.GetBytesParam(types.MessageDoubleSignFeeOwner, height)
	case types.MessageSendFee:
		return store.GetBytesParam(types.MessageSendFeeOwner, height)
	case types.MessageStakeFishermanFee:
		return store.GetBytesParam(types.MessageStakeFishermanFeeOwner, height)
	case types.MessageEditStakeFishermanFee:
		return store.GetBytesParam(types.MessageEditStakeFishermanFeeOwner, height)
	case types.MessageUnstakeFishermanFee:
		return store.GetBytesParam(types.MessageUnstakeFishermanFeeOwner, height)
	case types.MessagePauseFishermanFee:
		return store.GetBytesParam(types.MessagePauseFishermanFeeOwner, height)
	case types.MessageUnpauseFishermanFee:
		return store.GetBytesParam(types.MessageUnpauseFishermanFeeOwner, height)
	case types.MessageFishermanPauseServiceNodeFee:
		return store.GetBytesParam(types.MessageFishermanPauseServiceNodeFeeOwner, height)
	case types.MessageTestScoreFee:
		return store.GetBytesParam(types.MessageTestScoreFeeOwner, height)
	case types.MessageProveTestScoreFee:
		return store.GetBytesParam(types.MessageProveTestScoreFeeOwner, height)
	case types.MessageStakeAppFee:
		return store.GetBytesParam(types.MessageStakeAppFeeOwner, height)
	case types.MessageEditStakeAppFee:
		return store.GetBytesParam(types.MessageEditStakeAppFeeOwner, height)
	case types.MessageUnstakeAppFee:
		return store.GetBytesParam(types.MessageUnstakeAppFeeOwner, height)
	case types.MessagePauseAppFee:
		return store.GetBytesParam(types.MessagePauseAppFeeOwner, height)
	case types.MessageUnpauseAppFee:
		return store.GetBytesParam(types.MessageUnpauseAppFeeOwner, height)
	case types.MessageStakeValidatorFee:
		return store.GetBytesParam(types.MessageStakeValidatorFeeOwner, height)
	case types.MessageEditStakeValidatorFee:
		return store.GetBytesParam(types.MessageEditStakeValidatorFeeOwner, height)
	case types.MessageUnstakeValidatorFee:
		return store.GetBytesParam(types.MessageUnstakeValidatorFeeOwner, height)
	case types.MessagePauseValidatorFee:
		return store.GetBytesParam(types.MessagePauseValidatorFeeOwner, height)
	case types.MessageUnpauseValidatorFee:
		return store.GetBytesParam(types.MessageUnpauseValidatorFeeOwner, height)
	case types.MessageStakeServiceNodeFee:
		return store.GetBytesParam(types.MessageStakeServiceNodeFeeOwner, height)
	case types.MessageEditStakeServiceNodeFee:
		return store.GetBytesParam(types.MessageEditStakeServiceNodeFeeOwner, height)
	case types.MessageUnstakeServiceNodeFee:
		return store.GetBytesParam(types.MessageUnstakeServiceNodeFeeOwner, height)
	case types.MessagePauseServiceNodeFee:
		return store.GetBytesParam(types.MessagePauseServiceNodeFeeOwner, height)
	case types.MessageUnpauseServiceNodeFee:
		return store.GetBytesParam(types.MessageUnpauseServiceNodeFeeOwner, height)
	case types.MessageChangeParameterFee:
		return store.GetBytesParam(types.MessageChangeParameterFeeOwner, height)
	case types.BlocksPerSessionOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppMaxChainsOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppMinimumStakeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppBaselineStakeRateOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppStakingAdjustmentOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppUnstakingBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppMinimumPauseBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.AppMaxPausedBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodeMinimumStakeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodeMaxChainsOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodeUnstakingBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodeMinimumPauseBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodeMaxPausedBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ServiceNodesPerSessionOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.FishermanMinimumStakeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.FishermanMaxChainsOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.FishermanUnstakingBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.FishermanMinimumPauseBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.FishermanMaxPausedBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorMinimumStakeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorUnstakingBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorMinimumPauseBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorMaxPausedBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorMaximumMissedBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ProposerPercentageOfFeesOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.ValidatorMaxEvidenceAgeInBlocksOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MissedBlocksBurnPercentageOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.DoubleSignBurnPercentageOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageSendFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageStakeFishermanFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageEditStakeFishermanFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnstakeFishermanFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessagePauseFishermanFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnpauseFishermanFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageFishermanPauseServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageTestScoreFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageProveTestScoreFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageStakeAppFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageEditStakeAppFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnstakeAppFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessagePauseAppFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnpauseAppFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageStakeValidatorFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageEditStakeValidatorFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnstakeValidatorFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessagePauseValidatorFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnpauseValidatorFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageStakeServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageEditStakeServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnstakeServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessagePauseServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageUnpauseServiceNodeFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	case types.MessageChangeParameterFeeOwner:
		return store.GetBytesParam(types.AclOwner, height)
	default:
		return nil, types.ErrUnknownParam(paramName)
	}
}

func (u *UtilityContext) GetFee(msg typesUtil.Message, actorType typesUtil.ActorType) (amount *big.Int, err types.Error) {
	switch x := msg.(type) {
	case *typesUtil.MessageDoubleSign:
		return u.GetMessageDoubleSignFee()
	case *typesUtil.MessageSend:
		return u.GetMessageSendFee()
	case *typesUtil.MessageStake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageStakeAppFee()
		case typesUtil.ActorType_Fish:
			return u.GetMessageStakeFishermanFee()
		case typesUtil.ActorType_Node:
			return u.GetMessageStakeServiceNodeFee()
		case typesUtil.ActorType_Val:
			return u.GetMessageStakeValidatorFee()
		}
	case *typesUtil.MessageEditStake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageEditStakeAppFee()
		case typesUtil.ActorType_Fish:
			return u.GetMessageEditStakeFishermanFee()
		case typesUtil.ActorType_Node:
			return u.GetMessageEditStakeServiceNodeFee()
		case typesUtil.ActorType_Val:
			return u.GetMessageEditStakeValidatorFee()
		}
	case *typesUtil.MessageUnstake:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageUnstakeAppFee()
		case typesUtil.ActorType_Fish:
			return u.GetMessageUnstakeFishermanFee()
		case typesUtil.ActorType_Node:
			return u.GetMessageUnstakeServiceNodeFee()
		case typesUtil.ActorType_Val:
			return u.GetMessageUnstakeValidatorFee()
		}
	case *typesUtil.MessageUnpause:
		switch actorType {
		case typesUtil.ActorType_App:
			return u.GetMessageUnpauseAppFee()
		case typesUtil.ActorType_Fish:
			return u.GetMessageUnpauseFishermanFee()
		case typesUtil.ActorType_Node:
			return u.GetMessageUnpauseServiceNodeFee()
		case typesUtil.ActorType_Val:
			return u.GetMessageUnpauseValidatorFee()
		}
	case *typesUtil.MessageChangeParameter:
		return u.GetMessageChangeParameterFee()
	default:
		return nil, types.ErrUnknownMessage(x)
	}
	return nil, nil
}

func (u *UtilityContext) GetMessageChangeParameterSignerCandidates(msg *typesUtil.MessageChangeParameter) ([][]byte, types.Error) {
	owner, err := u.GetParamOwner(msg.ParameterKey)
	if err != nil {
		return nil, types.ErrGetParam(msg.ParameterKey, err)
	}
	return [][]byte{owner}, nil
}

func (u *UtilityContext) getBigIntParam(paramName string) (*big.Int, types.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, types.ErrGetParam(paramName, err)
	}
	value, err := store.GetStringParam(paramName, height)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, types.ErrGetParam(paramName, err)
	}
	return types.StringToBigInt(value)
}

func (u *UtilityContext) getIntParam(paramName string) (int, types.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, err)
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, err)
	}
	return value, nil
}

func (u *UtilityContext) getInt64Param(paramName string) (int64, types.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, err)
	}
	value, err := store.GetIntParam(paramName, height)
	if err != nil {
		return typesUtil.ZeroInt, types.ErrGetParam(paramName, err)
	}
	return int64(value), nil
}

func (u *UtilityContext) getByteArrayParam(paramName string) ([]byte, types.Error) {
	store := u.Store()
	height, err := store.GetHeight()
	if err != nil {
		return nil, types.ErrGetParam(paramName, err)
	}
	value, er := store.GetBytesParam(paramName, height)
	if er != nil {
		return nil, types.ErrGetParam(paramName, er)
	}
	return value, nil
}
