package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
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
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	maxChains, err := ctx.GetParameter(typesUtil.AppMaxChainsParamName, 0)
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxChains()), maxChains)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	gotParam, err := ctx.GetParameter(typesUtil.AppMaxPauseBlocksParamName, 0)
	require.NoError(t, err)
	require.Equal(t, int(defaultParams.GetAppMaxPauseBlocks()), gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppMinimumPauseBlocks())
	gotParam, err := ctx.GetParameter(typesUtil.AppMinimumPauseBlocksParamName, 0)
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetAppMinimumStake()
	gotParam, err := ctx.GetAppMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetAppUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetAppUnstakingBlocks())
	gotParam, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetBaselineAppStakeRate(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppBaselineStakeRate())
	gotParam, err := ctx.GetBaselineAppStakeRate()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetBlocksPerSession(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetBlocksPerSession())
	gotParam, err := ctx.GetParameter(typesUtil.BlocksPerSessionParamName, 0)
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetDoubleSignBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetDoubleSignBurnPercentage())
	gotParam, err := ctx.GetDoubleSignBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetDoubleSignFeeOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFeeOwner()
	gotParam, err := ctx.GetDoubleSignFeeOwner()
	require.NoError(t, err)

	defaultParamTx, er := hex.DecodeString(defaultParam)
	require.NoError(t, er)

	require.Equal(t, defaultParamTx, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxChains())
	gotParam, err := ctx.GetFishermanMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMaxPauseBlocks())
	gotParam, err := ctx.GetFishermanMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetFishermanMinimumPauseBlocks())
	gotParam, err := ctx.GetFishermanMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetFishermanMinimumStake()
	gotParam, err := ctx.GetFishermanMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetFishermanUnstakingBlocks())
	gotParam, err := ctx.GetFishermanUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMaxEvidenceAgeInBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxEvidenceAgeInBlocks())
	gotParam, err := ctx.GetMaxEvidenceAgeInBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageChangeParameterFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageChangeParameterFee()
	gotParam, err := ctx.GetMessageChangeParameterFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageDoubleSignFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFee()
	gotParam, err := ctx.GetMessageDoubleSignFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeAppFee()
	gotParam, err := ctx.GetMessageEditStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeFishermanFee()
	gotParam, err := ctx.GetMessageEditStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeServiceNodeFee()
	gotParam, err := ctx.GetMessageEditStakeServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageEditStakeValidatorFee()
	gotParam, err := ctx.GetMessageEditStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageFishermanPauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageFishermanPauseServiceNodeFee()
	gotParam, err := ctx.GetMessageFishermanPauseServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseAppFee()
	gotParam, err := ctx.GetMessagePauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseFishermanFee()
	gotParam, err := ctx.GetMessagePauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseServiceNodeFee()
	gotParam, err := ctx.GetMessagePauseServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessagePauseValidatorFee()
	gotParam, err := ctx.GetMessagePauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageProveTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageProveTestScoreFee()
	gotParam, err := ctx.GetMessageProveTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageSendFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageSendFee()
	gotParam, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeAppFee()
	gotParam, err := ctx.GetMessageStakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeFishermanFee()
	gotParam, err := ctx.GetMessageStakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeServiceNodeFee()
	gotParam, err := ctx.GetMessageStakeServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageStakeValidatorFee()
	gotParam, err := ctx.GetMessageStakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageTestScoreFee()
	gotParam, err := ctx.GetMessageTestScoreFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseAppFee()
	gotParam, err := ctx.GetMessageUnpauseAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseFishermanFee()
	gotParam, err := ctx.GetMessageUnpauseFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseServiceNodeFee()
	gotParam, err := ctx.GetMessageUnpauseServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnpauseValidatorFee()
	gotParam, err := ctx.GetMessageUnpauseValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeAppFee()
	gotParam, err := ctx.GetMessageUnstakeAppFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeFishermanFee()
	gotParam, err := ctx.GetMessageUnstakeFishermanFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeServiceNodeFee()
	gotParam, err := ctx.GetMessageUnstakeServiceNodeFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageUnstakeValidatorFee()
	gotParam, err := ctx.GetMessageUnstakeValidatorFee()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetMissedBlocksBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetProposerPercentageOfFees(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetProposerPercentageOfFees())
	gotParam, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServiceNodeMaxChains())
	gotParam, err := ctx.GetServiceNodeMaxChains()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServiceNodeMaxPauseBlocks())
	gotParam, err := ctx.GetServiceNodeMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetServiceNodeMinimumPauseBlocks())
	gotParam, err := ctx.GetServiceNodeMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetServiceNodeMinimumStake()
	gotParam, err := ctx.GetServiceNodeMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetServiceNodeUnstakingBlocks())
	gotParam, err := ctx.GetServiceNodeUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetStakingAdjustment(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetAppStakingAdjustment())
	gotParam, err := ctx.GetStabilityAdjustment()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMaxMissedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaximumMissedBlocks())
	gotParam, err := ctx.GetValidatorMaxMissedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMaxPauseBlocks())
	gotParam, err := ctx.GetValidatorMaxPausedBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetValidatorMinimumPauseBlocks())
	gotParam, err := ctx.GetValidatorMinimumPauseBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetValidatorMinimumStake()
	gotParam, err := ctx.GetValidatorMinimumStake()
	require.NoError(t, err)
	require.Equal(t, defaultParam, typesUtil.BigIntToString(gotParam))

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.GetValidatorUnstakingBlocks())
	gotParam, err := ctx.GetValidatorUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, defaultParam, gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_HandleMessageChangeParameter(t *testing.T) {
	cdc := codec.GetCodec()
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.GetMissedBlocksBurnPercentage())
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
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
	require.NoError(t, ctx.HandleMessageChangeParameter(msg), "handle message change param")
	gotParam, err = ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.Equal(t, int(newParamValue), gotParam)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetParamOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetAclOwner()
	gotParam, err := ctx.GetParamOwner(typesUtil.AclOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetBlocksPerSessionOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.BlocksPerSessionParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMaxChainsOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMaxChainsParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMinimumStakeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppBaselineStakeRateOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppBaselineStakeRateParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppStakingAdjustmentOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppUnstakingBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMinimumPauseBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAppMaxPausedBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodesPerSessionOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodesPerSessionParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeMinimumStakeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeMaxChainsOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMaxChainsParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeUnstakingBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeMinimumPauseBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeMaxPausedBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMinimumStakeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetServiceNodeMaxChainsOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanUnstakingBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMinimumPauseBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetFishermanMaxPausedBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMinimumStakeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMinimumStakeParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorUnstakingBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorUnstakingBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMinimumPauseBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaxPausedBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaxPausedBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaximumMissedBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaximumMissedBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetProposerPercentageOfFeesOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ProposerPercentageOfFeesParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetValidatorMaxEvidenceAgeInBlocksOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaxEvidenceAgeInBlocksParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMissedBlocksBurnPercentageOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MissedBlocksBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetDoubleSignBurnPercentageOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.DoubleSignBurnPercentageParamName)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageDoubleSignFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageDoubleSignFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageSendFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageSendFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeFishermanFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeFishermanFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeFishermanFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseFishermanFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseFishermanFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseFishermanFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageTestScoreFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageTestScoreFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageFishermanPauseServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageFishermanPauseServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageProveTestScoreFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageProveTestScoreFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeAppFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeAppFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeAppFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseAppFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseAppFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseAppFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeValidatorFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeValidatorFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeValidatorFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseValidatorFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseValidatorFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseValidatorFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageStakeServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageEditStakeServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnstakeServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessagePauseServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageUnpauseServiceNodeFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseServiceNodeFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetMessageChangeParameterFeeOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageChangeParameterFee)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	// owners
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.BlocksPerSessionOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppBaselineStakeRateOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.AppMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodeMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ServiceNodesPerSessionOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMaxChainsOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.FishermanMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMinimumStakeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorUnstakingBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ProposerPercentageOfFeesOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.ValidatorMaxEvidenceAgeInBlocksOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MissedBlocksBurnPercentageOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.DoubleSignBurnPercentageOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageSendFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseFishermanFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageFishermanPauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageTestScoreFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageProveTestScoreFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseAppFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseValidatorFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageStakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageEditStakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnstakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessagePauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageUnpauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.Equal(t, defaultParam, hex.EncodeToString(gotParam))
	defaultParam = defaultParams.GetAclOwner()
	gotParam, err = ctx.GetParamOwner(typesUtil.MessageChangeParameterFeeOwner)
	require.NoError(t, err)
	defaultParamBz, err := hex.DecodeString(defaultParam)
	require.NoError(t, err)
	require.Equal(t, defaultParamBz, gotParam)

	test_artifacts.CleanupTest(ctx)
}
