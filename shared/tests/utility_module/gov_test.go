package utility_module

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CLEANUP: cleanup this file as part of https://github.com/pokt-network/pocket/issues/76

func DefaultTestingParams(_ *testing.T) *genesis.Params {
	return test_artifacts.DefaultParams()
}

func TestUtilityContext_GetAppMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	maxChains, err := ctx.GetAppMaxChains()
	require.NoError(t, err)
	require.False(t, int(defaultParams.AppMaxChains) != maxChains, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParams.AppMaxChains, maxChains))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	gotParam, err := ctx.GetAppMaxPausedBlocks()
	require.NoError(t, err)
	require.False(t, int(defaultParams.AppMaxPauseBlocks) != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParams.AppMaxPausedBlocksOwner, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppMinimumPauseBlocks)
	gotParam, err := ctx.GetAppMinimumPauseBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetAppMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.AppMinimumStake
	gotParam, err := ctx.GetAppMinimumStake()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetAppUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.AppUnstakingBlocks)
	gotParam, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetBaselineAppStakeRate(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppBaselineStakeRate)
	gotParam, err := ctx.GetBaselineAppStakeRate()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetBlocksPerSession(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.BlocksPerSession)
	gotParam, err := ctx.GetBlocksPerSession()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetDoubleSignBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.DoubleSignBurnPercentage)
	gotParam, err := ctx.GetDoubleSignBurnPercentage()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetDoubleSignFeeOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageDoubleSignFeeOwner
	gotParam, err := ctx.GetDoubleSignFeeOwner()
	require.NoError(t, err)

	defaultParamTx, er := hex.DecodeString(defaultParam)
	require.NoError(t, er)

	require.Equal(t, defaultParamTx, gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMaxChains)
	gotParam, err := ctx.GetFishermanMaxChains()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMaxPauseBlocks)
	gotParam, err := ctx.GetFishermanMaxPausedBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMinimumPauseBlocks)
	gotParam, err := ctx.GetFishermanMinimumPauseBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.FishermanMinimumStake
	gotParam, err := ctx.GetFishermanMinimumStake()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetFishermanUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.FishermanUnstakingBlocks)
	gotParam, err := ctx.GetFishermanUnstakingBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMaxEvidenceAgeInBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaxEvidenceAgeInBlocks)
	gotParam, err := ctx.GetMaxEvidenceAgeInBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageChangeParameterFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageChangeParameterFee
	gotParam, err := ctx.GetMessageChangeParameterFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageDoubleSignFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFee()
	gotParam, err := ctx.GetMessageDoubleSignFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeAppFee
	gotParam, err := ctx.GetMessageEditStakeAppFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeFishermanFee
	gotParam, err := ctx.GetMessageEditStakeFishermanFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeServiceNodeFee
	gotParam, err := ctx.GetMessageEditStakeServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageEditStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeValidatorFee
	gotParam, err := ctx.GetMessageEditStakeValidatorFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageFishermanPauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageFishermanPauseServiceNodeFee
	gotParam, err := ctx.GetMessageFishermanPauseServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseAppFee
	gotParam, err := ctx.GetMessagePauseAppFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseFishermanFee
	gotParam, err := ctx.GetMessagePauseFishermanFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseServiceNodeFee
	gotParam, err := ctx.GetMessagePauseServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessagePauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseValidatorFee
	gotParam, err := ctx.GetMessagePauseValidatorFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageProveTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageProveTestScoreFee
	gotParam, err := ctx.GetMessageProveTestScoreFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageSendFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageSendFee
	gotParam, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeAppFee
	gotParam, err := ctx.GetMessageStakeAppFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeFishermanFee
	gotParam, err := ctx.GetMessageStakeFishermanFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeServiceNodeFee
	gotParam, err := ctx.GetMessageStakeServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeValidatorFee
	gotParam, err := ctx.GetMessageStakeValidatorFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageTestScoreFee
	gotParam, err := ctx.GetMessageTestScoreFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseAppFee
	gotParam, err := ctx.GetMessageUnpauseAppFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseFishermanFee
	gotParam, err := ctx.GetMessageUnpauseFishermanFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseServiceNodeFee
	gotParam, err := ctx.GetMessageUnpauseServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnpauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseValidatorFee
	gotParam, err := ctx.GetMessageUnpauseValidatorFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeAppFee
	gotParam, err := ctx.GetMessageUnstakeAppFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeFishermanFee
	gotParam, err := ctx.GetMessageUnstakeFishermanFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeServiceNodeFee
	gotParam, err := ctx.GetMessageUnstakeServiceNodeFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMessageUnstakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeValidatorFee
	gotParam, err := ctx.GetMessageUnstakeValidatorFee()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetMissedBlocksBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.MissedBlocksBurnPercentage)
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetProposerPercentageOfFees(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ProposerPercentageOfFees)
	gotParam, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMaxChains)
	gotParam, err := ctx.GetServiceNodeMaxChains()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMaxPauseBlocks)
	gotParam, err := ctx.GetServiceNodeMaxPausedBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMinimumPauseBlocks)
	gotParam, err := ctx.GetServiceNodeMinimumPauseBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.ServiceNodeMinimumStake
	gotParam, err := ctx.GetServiceNodeMinimumStake()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetServiceNodeUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.ServiceNodeUnstakingBlocks)
	gotParam, err := ctx.GetServiceNodeUnstakingBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetStakingAdjustment(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppStakingAdjustment)
	gotParam, err := ctx.GetStabilityAdjustment()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMaxMissedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaximumMissedBlocks)
	gotParam, err := ctx.GetValidatorMaxMissedBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaxPauseBlocks)
	gotParam, err := ctx.GetValidatorMaxPausedBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMinimumPauseBlocks)
	gotParam, err := ctx.GetValidatorMinimumPauseBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.ValidatorMinimumStake
	gotParam, err := ctx.GetValidatorMinimumStake()
	require.NoError(t, err)
	require.False(t, defaultParam != types.BigIntToString(gotParam), fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetValidatorUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.ValidatorUnstakingBlocks)
	gotParam, err := ctx.GetValidatorUnstakingBlocks()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_HandleMessageChangeParameter(t *testing.T) {
	cdc := types.GetCodec()
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.MissedBlocksBurnPercentage)
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.False(t, defaultParam != gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	newParamValue := int32(2)
	paramOwnerPK := test_artifacts.DefaultParamsOwner
	any, err := cdc.ToAny(&wrapperspb.Int32Value{
		Value: newParamValue,
	})
	require.NoError(t, err)
	msg := &typesUtil.MessageChangeParameter{
		Owner:          paramOwnerPK.Address(),
		ParameterKey:   types.MissedBlocksBurnPercentageParamName,
		ParameterValue: any,
	}
	require.NoError(t, ctx.HandleMessageChangeParameter(msg), "handle message change param")
	gotParam, err = ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	require.False(t, int(newParamValue) != gotParam, fmt.Sprintf("wrong param value after handling, expected %v got %v", newParamValue, gotParam))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetParamOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.AclOwner
	gotParam, err := ctx.GetParamOwner(types.AclOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.BlocksPerSessionOwner
	gotParam, err = ctx.GetParamOwner(types.BlocksPerSessionParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppMaxChainsOwner
	gotParam, err = ctx.GetParamOwner(types.AppMaxChainsParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(types.AppMinimumStakeParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppBaselineStakeRateOwner
	gotParam, err = ctx.GetParamOwner(types.AppBaselineStakeRateParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppStakingAdjustmentOwner
	gotParam, err = ctx.GetParamOwner(types.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.AppUnstakingBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.AppMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AppMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.AppMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodesPerSessionOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodesPerSessionParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodeMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMinimumStakeParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodeMaxChainsOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMaxChainsParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodeUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeUnstakingBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodeMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ServiceNodeMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.FishermanMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMinimumStakeParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.GetServiceNodeMaxChainsOwner()
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.FishermanUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanUnstakingBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.FishermanMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.FishermanMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMaxPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMinimumStakeParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorUnstakingBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMinimumPauseBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaxPausedBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorMaximumMissedBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaximumMissedBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ProposerPercentageOfFeesOwner
	gotParam, err = ctx.GetParamOwner(types.ProposerPercentageOfFeesParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.ValidatorMaxEvidenceAgeInBlocksOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaxEvidenceAgeInBlocksParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MissedBlocksBurnPercentageOwner
	gotParam, err = ctx.GetParamOwner(types.MissedBlocksBurnPercentageParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.DoubleSignBurnPercentageOwner
	gotParam, err = ctx.GetParamOwner(types.DoubleSignBurnPercentageParamName)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageDoubleSignFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageDoubleSignFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageSendFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageSendFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageStakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeFishermanFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageEditStakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeFishermanFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnstakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeFishermanFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessagePauseFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseFishermanFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnpauseFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseFishermanFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageTestScoreFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageTestScoreFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageFishermanPauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageFishermanPauseServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageProveTestScoreFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageProveTestScoreFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageStakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeAppFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageEditStakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeAppFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnstakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeAppFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessagePauseAppFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseAppFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnpauseAppFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseAppFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageStakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeValidatorFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageEditStakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeValidatorFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnstakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeValidatorFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessagePauseValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseValidatorFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnpauseValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseValidatorFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageStakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageEditStakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnstakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessagePauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageUnpauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseServiceNodeFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.MessageChangeParameterFeeOwner
	gotParam, err = ctx.GetParamOwner(types.MessageChangeParameterFee)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	// owners
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.BlocksPerSessionOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppMaxChainsOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppMinimumStakeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppBaselineStakeRateOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppStakingAdjustmentOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppUnstakingBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.AppMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMaxChainsOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeUnstakingBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMinimumStakeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodeMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ServiceNodesPerSessionOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMinimumStakeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMaxChainsOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanUnstakingBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.FishermanMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMinimumStakeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorUnstakingBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMinimumPauseBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaxPausedBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ProposerPercentageOfFeesOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.ValidatorMaxEvidenceAgeInBlocksOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MissedBlocksBurnPercentageOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.DoubleSignBurnPercentageOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageSendFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeFishermanFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeFishermanFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseFishermanFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseFishermanFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageFishermanPauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageTestScoreFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageProveTestScoreFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeAppFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeAppFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeAppFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseAppFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseAppFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeValidatorFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeValidatorFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseValidatorFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseValidatorFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageStakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageEditStakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnstakeServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessagePauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageUnpauseServiceNodeFeeOwner)
	require.NoError(t, err)
	require.False(t, hex.EncodeToString(gotParam) != defaultParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(types.MessageChangeParameterFeeOwner)
	require.NoError(t, err)
	defaultParamBz, err := hex.DecodeString(defaultParam)
	require.NoError(t, err)
	require.Equal(t, defaultParamBz, gotParam, fmt.Sprintf("unexpected param value: expected %v got %v", defaultParam, gotParam))

	tests.CleanupTest(ctx)
}
