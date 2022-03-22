package utility_module

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/pokt-network/pocket/shared/types"
	utilTypes "github.com/pokt-network/pocket/utility/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func DefaultTestingParams(_ *testing.T) *pre_persistence.Params {
	return pre_persistence.DefaultParams()
}

func TestUtilityContext_GetAppMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	maxChains, err := ctx.GetAppMaxChains()
	if err != nil {
		t.Fatal(err)
	}
	if int(defaultParams.AppMaxChains) != maxChains {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParams.AppMaxChains, maxChains)
	}
}

func TestUtilityContext_GetAppMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	gotParam, err := ctx.GetAppMaxPausedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if int(defaultParams.AppMaxPauseBlocks) != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParams.AppMaxPausedBlocksOwner, gotParam)
	}
}

func TestUtilityContext_GetAppMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppMinimumPauseBlocks)
	gotParam, err := ctx.GetAppMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetAppMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.AppMinimumStake
	gotParam, err := ctx.GetAppMinimumStake()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetAppUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.AppUnstakingBlocks)
	gotParam, err := ctx.GetAppUnstakingBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetBaselineAppStakeRate(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppBaselineStakeRate)
	gotParam, err := ctx.GetBaselineAppStakeRate()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetBlocksPerSession(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.BlocksPerSession)
	gotParam, err := ctx.GetBlocksPerSession()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetDoubleSignBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.DoubleSignBurnPercentage)
	gotParam, err := ctx.GetDoubleSignBurnPercentage()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetDoubleSignFeeOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageDoubleSignFeeOwner
	gotParam, err := ctx.GetDoubleSignFeeOwner()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(defaultParam, gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetFishermanMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMaxChains)
	gotParam, err := ctx.GetFishermanMaxChains()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetFishermanMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMaxPauseBlocks)
	gotParam, err := ctx.GetFishermanMaxPausedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetFishermanMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.FishermanMinimumPauseBlocks)
	gotParam, err := ctx.GetFishermanMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetFishermanMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.FishermanMinimumStake
	gotParam, err := ctx.GetFishermanMinimumStake()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetFishermanUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.FishermanUnstakingBlocks)
	gotParam, err := ctx.GetFishermanUnstakingBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMaxEvidenceAgeInBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaxEvidenceAgeInBlocks)
	gotParam, err := ctx.GetMaxEvidenceAgeInBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageChangeParameterFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageChangeParameterFee
	gotParam, err := ctx.GetMessageChangeParameterFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageDoubleSignFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.GetMessageDoubleSignFee()
	gotParam, err := ctx.GetMessageDoubleSignFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageEditStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeAppFee
	gotParam, err := ctx.GetMessageEditStakeAppFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageEditStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeFishermanFee
	gotParam, err := ctx.GetMessageEditStakeFishermanFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageEditStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeServiceNodeFee
	gotParam, err := ctx.GetMessageEditStakeServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageEditStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageEditStakeValidatorFee
	gotParam, err := ctx.GetMessageEditStakeValidatorFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageFishermanPauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageFishermanPauseServiceNodeFee
	gotParam, err := ctx.GetMessageFishermanPauseServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessagePauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseAppFee
	gotParam, err := ctx.GetMessagePauseAppFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessagePauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseFishermanFee
	gotParam, err := ctx.GetMessagePauseFishermanFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessagePauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseServiceNodeFee
	gotParam, err := ctx.GetMessagePauseServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessagePauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessagePauseValidatorFee
	gotParam, err := ctx.GetMessagePauseValidatorFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageProveTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageProveTestScoreFee
	gotParam, err := ctx.GetMessageProveTestScoreFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageSendFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageSendFee
	gotParam, err := ctx.GetMessageSendFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageStakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeAppFee
	gotParam, err := ctx.GetMessageStakeAppFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageStakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeFishermanFee
	gotParam, err := ctx.GetMessageStakeFishermanFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageStakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeServiceNodeFee
	gotParam, err := ctx.GetMessageStakeServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageStakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageStakeValidatorFee
	gotParam, err := ctx.GetMessageStakeValidatorFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageTestScoreFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageTestScoreFee
	gotParam, err := ctx.GetMessageTestScoreFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnpauseAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseAppFee
	gotParam, err := ctx.GetMessageUnpauseAppFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnpauseFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseFishermanFee
	gotParam, err := ctx.GetMessageUnpauseFishermanFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnpauseServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseServiceNodeFee
	gotParam, err := ctx.GetMessageUnpauseServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnpauseValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnpauseValidatorFee
	gotParam, err := ctx.GetMessageUnpauseValidatorFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnstakeAppFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeAppFee
	gotParam, err := ctx.GetMessageUnstakeAppFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnstakeFishermanFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeFishermanFee
	gotParam, err := ctx.GetMessageUnstakeFishermanFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnstakeServiceNodeFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeServiceNodeFee
	gotParam, err := ctx.GetMessageUnstakeServiceNodeFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMessageUnstakeValidatorFee(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.MessageUnstakeValidatorFee
	gotParam, err := ctx.GetMessageUnstakeValidatorFee()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetMissedBlocksBurnPercentage(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.MissedBlocksBurnPercentage)
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetProposerPercentageOfFees(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ProposerPercentageOfFees)
	gotParam, err := ctx.GetProposerPercentageOfFees()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetServiceNodeMaxChains(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMaxChains)
	gotParam, err := ctx.GetServiceNodeMaxChains()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetServiceNodeMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMaxPauseBlocks)
	gotParam, err := ctx.GetServiceNodeMaxPausedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetServiceNodeMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ServiceNodeMinimumPauseBlocks)
	gotParam, err := ctx.GetServiceNodeMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetServiceNodeMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.ServiceNodeMinimumStake
	gotParam, err := ctx.GetServiceNodeMinimumStake()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetServiceNodeUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.ServiceNodeUnstakingBlocks)
	gotParam, err := ctx.GetServiceNodeUnstakingBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetStakingAdjustment(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.AppStakingAdjustment)
	gotParam, err := ctx.GetStakingAdjustment()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetValidatorMaxMissedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaximumMissedBlocks)
	gotParam, err := ctx.GetValidatorMaxMissedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetValidatorMaxPausedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMaxPauseBlocks)
	gotParam, err := ctx.GetValidatorMaxPausedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetValidatorMinimumPauseBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.ValidatorMinimumPauseBlocks)
	gotParam, err := ctx.GetValidatorMinimumPauseBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetValidatorMinimumStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.ValidatorMinimumStake
	gotParam, err := ctx.GetValidatorMinimumStake()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != types.BigIntToString(gotParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_GetValidatorUnstakingBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int64(defaultParams.ValidatorUnstakingBlocks)
	gotParam, err := ctx.GetValidatorUnstakingBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}

func TestUtilityContext_HandleMessageChangeParameter(t *testing.T) {
	cdc := utilTypes.UtilityCodec()
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := int(defaultParams.MissedBlocksBurnPercentage)
	gotParam, err := ctx.GetMissedBlocksBurnPercentage()
	if err != nil {
		t.Fatal(err)
	}
	if defaultParam != gotParam {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	newParamValue := int32(2)
	paramOwnerPK := pre_persistence.DefaultParamsOwner
	any, err := cdc.ToAny(&wrapperspb.Int32Value{
		Value: newParamValue,
	})
	if err != nil {
		t.Fatal(err)
	}
	msg := &utilTypes.MessageChangeParameter{
		Owner:          paramOwnerPK.Address(),
		ParameterKey:   utilTypes.MissedBlocksBurnPercentageParamName,
		ParameterValue: any,
	}
	if err := ctx.HandleMessageChangeParameter(msg); err != nil {
		t.Fatal(err)
	}
	gotParam, err = ctx.GetMissedBlocksBurnPercentage()
	if err != nil {
		t.Fatal(err)
	}
	if int(newParamValue) != gotParam {
		t.Fatalf("wrong param value after handling, expected %v got %v", newParamValue, gotParam)
	}
}

func TestUtilityContext_GetParamOwner(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	defaultParams := DefaultTestingParams(t)
	defaultParam := defaultParams.AclOwner
	gotParam, err := ctx.GetParamOwner(utilTypes.AclOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.BlocksPerSessionOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.BlocksPerSessionParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppMaxChainsOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMaxChainsParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMinimumStakeParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppBaselineStakeRateOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppBaselineStakeRateParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppStakingAdjustmentOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppStakingAdjustmentOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppUnstakingBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMinimumPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AppMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMaxPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodesPerSessionOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodesPerSessionParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodeMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMinimumStakeParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodeMaxChainsOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMaxChainsParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodeUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeUnstakingBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodeMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMinimumPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ServiceNodeMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMaxPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.FishermanMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMinimumStakeParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.GetServiceNodeMaxChainsOwner()
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMaxPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.FishermanUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanUnstakingBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.FishermanMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMinimumPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.FishermanMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMaxPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorMinimumStakeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMinimumStakeParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorUnstakingBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorUnstakingBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorMinimumPauseBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMinimumPauseBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorMaxPausedBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaxPausedBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorMaximumMissedBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaximumMissedBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ProposerPercentageOfFeesOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ProposerPercentageOfFeesParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.ValidatorMaxEvidenceAgeInBlocksOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaxEvidenceAgeInBlocksParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MissedBlocksBurnPercentageOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MissedBlocksBurnPercentageParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.DoubleSignBurnPercentageOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.DoubleSignBurnPercentageParamName)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageDoubleSignFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageDoubleSignFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageSendFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageSendFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageStakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeFishermanFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageEditStakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeFishermanFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnstakeFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeFishermanFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessagePauseFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseFishermanFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnpauseFishermanFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseFishermanFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageTestScoreFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageTestScoreFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageFishermanPauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageFishermanPauseServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageProveTestScoreFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageProveTestScoreFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageStakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeAppFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageEditStakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeAppFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnstakeAppFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeAppFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessagePauseAppFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseAppFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnpauseAppFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseAppFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageStakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeValidatorFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageEditStakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeValidatorFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnstakeValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeValidatorFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessagePauseValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseValidatorFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnpauseValidatorFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseValidatorFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageStakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageEditStakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnstakeServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessagePauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageUnpauseServiceNodeFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseServiceNodeFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.MessageChangeParameterFeeOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageChangeParameterFee)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	// owners
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.BlocksPerSessionOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMaxChainsOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMinimumStakeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppBaselineStakeRateOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppStakingAdjustmentOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppUnstakingBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMinimumPauseBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.AppMaxPausedBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMinimumPauseBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMaxChainsOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeUnstakingBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMinimumStakeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodeMaxPausedBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ServiceNodesPerSessionOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMinimumStakeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMaxChainsOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanUnstakingBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMinimumPauseBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.FishermanMaxPausedBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMinimumStakeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorUnstakingBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMinimumPauseBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaxPausedBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaxPausedBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ProposerPercentageOfFeesOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.ValidatorMaxEvidenceAgeInBlocksOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MissedBlocksBurnPercentageOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.DoubleSignBurnPercentageOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageSendFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeFishermanFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeFishermanFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeFishermanFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseFishermanFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseFishermanFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageFishermanPauseServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageTestScoreFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageProveTestScoreFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeAppFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeAppFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeAppFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseAppFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseAppFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeValidatorFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeValidatorFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeValidatorFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseValidatorFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseValidatorFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageStakeServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageEditStakeServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnstakeServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessagePauseServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageUnpauseServiceNodeFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
	defaultParam = defaultParams.AclOwner
	gotParam, err = ctx.GetParamOwner(utilTypes.MessageChangeParameterFeeOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotParam, defaultParam) {
		t.Fatalf("unexpected param value: expected %v got %v", defaultParam, gotParam)
	}
}
