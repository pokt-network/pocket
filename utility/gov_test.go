package utility

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TODO : After we change the interface to pass param name, simply use reflection to
//  iterate over all the params and test them. Suggestion: [Google's go-cmp] (https://github.com/google/go-cmp)

func DefaultTestingParams(_ *testing.T) *genesis.Params {
	return test_artifacts.DefaultParams()
}

func TestUtilityContext_GetAppMaxChains(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	maxChains, err := ctx.getAppMaxChains()
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxChains()), maxChains)
}

func TestUtilityContext_GetAppMaxPausedBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	gotParam, err := ctx.getAppMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxPauseBlocks()), gotParam)
}

func TestUtilityContext_GetAppMinimumPauseBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppMinimumPauseBlocks())
	gotParam, err := ctx.getAppMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetAppMinimumStake(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetAppMinimumStake()
	gotParam, err := ctx.getAppMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetAppUnstakingBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetAppUnstakingBlocks())
	gotParam, err := ctx.getAppUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetBaselineAppStakeRate(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppBaselineStakeRate())
	gotParam, err := ctx.getBaselineAppStakeRate()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetBlocksPerSession(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetBlocksPerSession())
	gotParam, err := ctx.getParameter(typesUtil.BlocksPerSessionParamName, 0)
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetDoubleSignBurnPercentage(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetDoubleSignBurnPercentage())
	gotParam, err := ctx.getDoubleSignBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetDoubleSignFeeOwner(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFeeOwner()
	gotParam, err := ctx.getDoubleSignFeeOwner()
	require.NoError(t, err)

	defaultParamTx, er := hex.DecodeString(defaultParam)
	require.NoError(t, er)

	require.Equal(t, defaultParamTx, gotParam)

}

func TestUtilityContext_GetFishermanMaxChains(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxChains())
	gotParam, err := ctx.getFishermanMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetFishermanMaxPausedBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxPauseBlocks())
	gotParam, err := ctx.getFishermanMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetFishermanMinimumPauseBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMinimumPauseBlocks())
	gotParam, err := ctx.getFishermanMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetFishermanMinimumStake(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetFishermanMinimumStake()
	gotParam, err := ctx.getFishermanMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetFishermanUnstakingBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetFishermanUnstakingBlocks())
	gotParam, err := ctx.getFishermanUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetMaxEvidenceAgeInBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxEvidenceAgeInBlocks())
	gotParam, err := ctx.getMaxEvidenceAgeInBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

}

func TestUtilityContext_GetMessageChangeParameterFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageChangeParameterFee()
	gotParam, err := ctx.getMessageChangeParameterFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageDoubleSignFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFee()
	gotParam, err := ctx.getMessageDoubleSignFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageEditStakeAppFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeAppFee()
	gotParam, err := ctx.getMessageEditStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageEditStakeFishermanFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeFishermanFee()
	gotParam, err := ctx.getMessageEditStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageEditStakeServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeServicerFee()
	gotParam, err := ctx.getMessageEditStakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageEditStakeValidatorFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeValidatorFee()
	gotParam, err := ctx.getMessageEditStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageFishermanPauseServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageFishermanPauseServicerFee()
	gotParam, err := ctx.getMessageFishermanPauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessagePauseAppFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseAppFee()
	gotParam, err := ctx.getMessagePauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessagePauseFishermanFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseFishermanFee()
	gotParam, err := ctx.getMessagePauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessagePauseServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseServicerFee()
	gotParam, err := ctx.getMessagePauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessagePauseValidatorFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseValidatorFee()
	gotParam, err := ctx.getMessagePauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageProveTestScoreFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageProveTestScoreFee()
	gotParam, err := ctx.getMessageProveTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageSendFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageSendFee()
	gotParam, err := ctx.getMessageSendFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageStakeAppFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeAppFee()
	gotParam, err := ctx.getMessageStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageStakeFishermanFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeFishermanFee()
	gotParam, err := ctx.getMessageStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageStakeServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeServicerFee()
	gotParam, err := ctx.getMessageStakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageStakeValidatorFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeValidatorFee()
	gotParam, err := ctx.getMessageStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageTestScoreFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageTestScoreFee()
	gotParam, err := ctx.getMessageTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnpauseAppFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseAppFee()
	gotParam, err := ctx.getMessageUnpauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnpauseFishermanFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseFishermanFee()
	gotParam, err := ctx.getMessageUnpauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnpauseServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseServicerFee()
	gotParam, err := ctx.getMessageUnpauseServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnpauseValidatorFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseValidatorFee()
	gotParam, err := ctx.getMessageUnpauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageUnstakeAppFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeAppFee()
	gotParam, err := ctx.getMessageUnstakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))

}

func TestUtilityContext_GetMessageUnstakeFishermanFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeFishermanFee()
	gotParam, err := ctx.getMessageUnstakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnstakeServicerFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeServicerFee()
	gotParam, err := ctx.getMessageUnstakeServicerFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMessageUnstakeValidatorFee(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeValidatorFee()
	gotParam, err := ctx.getMessageUnstakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetMissedBlocksBurnPercentage(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := ctx.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetProposerPercentageOfFees(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetProposerPercentageOfFees())
	gotParam, err := ctx.getProposerPercentageOfFees()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetServicerMaxChains(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMaxChains())
	gotParam, err := ctx.getServicerMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetServicerMaxPausedBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMaxPauseBlocks())
	gotParam, err := ctx.getServicerMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetServicerMinimumPauseBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServicerMinimumPauseBlocks())
	gotParam, err := ctx.getServicerMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetServicerMinimumStake(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetServicerMinimumStake()
	gotParam, err := ctx.getServicerMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetServicerUnstakingBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetServicerUnstakingBlocks())
	gotParam, err := ctx.getServicerUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetStakingAdjustment(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppStakingAdjustment())
	gotParam, err := ctx.getStabilityAdjustment()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetValidatorMaxMissedBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaximumMissedBlocks())
	gotParam, err := ctx.getValidatorMaxMissedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetValidatorMaxPausedBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxPauseBlocks())
	gotParam, err := ctx.getValidatorMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetValidatorMinimumPauseBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMinimumPauseBlocks())
	gotParam, err := ctx.getValidatorMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_GetValidatorMinimumStake(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetValidatorMinimumStake()
	gotParam, err := ctx.getValidatorMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, converters.BigIntToString(gotParam))
}

func TestUtilityContext_GetValidatorUnstakingBlocks(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetValidatorUnstakingBlocks())
	gotParam, err := ctx.getValidatorUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
}

func TestUtilityContext_HandleMessageChangeParameter(t *testing.T) {
	cdc := codec.GetCodec()
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := ctx.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)
	newParamValue := int32(2)
	paramOwnerPK := test_artifacts.DefaultParamsOwner
	any, er := cdc.ToAny(&wrapperspb.Int32Value{
		Value: newParamValue,
	})
	require.NoError(t, er)
	msg := &typesUtil.MessageChangeParameter{
		Owner:          paramOwnerPK.Address(),
		ParameterKey:   typesUtil.MissedBlocksBurnPercentageParamName,
		ParameterValue: any,
	}
	require.NoError(t, ctx.handleMessageChangeParameter(msg), "handle message change param")
	gotParam, err = ctx.getMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, int(newParamValue), gotParam)

}

func TestUtilityContext_GetParamOwner(t *testing.T) {
	ctx := newTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetAclOwner()
	gotParam, err := ctx.getParamOwner(typesUtil.AclOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetBlocksPerSessionOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.BlocksPerSessionParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMaxChainsOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMaxChainsParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMinimumStakeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppBaselineStakeRateOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppBaselineStakeRateParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppStakingAdjustmentOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppUnstakingBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMinimumPauseBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMaxPausedBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicersPerSessionOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicersPerSessionParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerMinimumStakeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerMaxChainsOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMaxChainsParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerUnstakingBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerMinimumPauseBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerMaxPausedBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMinimumStakeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServicerMaxChainsOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanUnstakingBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMinimumPauseBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMaxPausedBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMinimumStakeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorUnstakingBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMinimumPauseBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaxPausedBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaxPausedBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaximumMissedBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaximumMissedBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetProposerPercentageOfFeesOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ProposerPercentageOfFeesParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaxEvidenceAgeInBlocksOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMissedBlocksBurnPercentageOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MissedBlocksBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetDoubleSignBurnPercentageOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.DoubleSignBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageDoubleSignFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageDoubleSignFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageSendFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageSendFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeFishermanFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeFishermanFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeFishermanFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseFishermanFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseFishermanFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageTestScoreFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageTestScoreFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageFishermanPauseServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageFishermanPauseServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageProveTestScoreFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageProveTestScoreFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeAppFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeAppFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeAppFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseAppFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseAppFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeValidatorFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeValidatorFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeValidatorFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseValidatorFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseValidatorFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseServicerFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseServicerFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageChangeParameterFeeOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageChangeParameterFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	// owners
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.BlocksPerSessionOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppBaselineStakeRateOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.AppMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicerMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ServicersPerSessionOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.FishermanMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ProposerPercentageOfFeesOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MissedBlocksBurnPercentageOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.DoubleSignBurnPercentageOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageSendFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageFishermanPauseServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageTestScoreFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageProveTestScoreFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageStakeServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageEditStakeServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnstakeServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessagePauseServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageUnpauseServicerFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.getParamOwner(typesUtil.MessageChangeParameterFeeOwner)
	require.NoError(t, err)
	defaultParamBz, err := hex.DecodeString(defaultParam)
	require.NoError(t, err)
	require.Equal(t, defaultParamBz, gotParam)
}
