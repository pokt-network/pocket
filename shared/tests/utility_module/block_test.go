package utility_module

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/tests"

	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_ApplyBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, startingBalance, amount, signer := newTestingTransaction(t, ctx)

	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)

	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		require.NoError(t, err, "apply block")
	}

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.True(t, missed == 1, fmt.Sprintf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks))

	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAmountSubtracted := big.NewInt(0).Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amountAfter, "unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter)
	// end-block logic verify

	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)

	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.Equal(t, expectedProposerBalanceDifference, proposerBalanceDifference, "unexpected before / after balance difference")

	tests.CleanupTest(ctx)
}

func TestUtilityContext_BeginBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		require.NoError(t, err)
	}

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.False(t, missed != 1, fmt.Sprintf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)
		actor := GetFirstActor(t, ctx, actorType)

		var err error
		switch actorType {
		case typesUtil.ActorType_App:
			err = ctx.Context.SetAppMaxPausedBlocks(0)
		case typesUtil.ActorType_Val:
			err = ctx.Context.SetValidatorMaxPausedBlocks(0)
		case typesUtil.ActorType_Fish:
			err = ctx.Context.SetFishermanMaxPausedBlocks(0)
		case typesUtil.ActorType_Node:
			err = ctx.Context.SetServiceNodeMaxPausedBlocks(0)
		default:
			t.Fatalf("unexpected actor type %s", actorType.GetActorName())
		}
		require.NoError(t, err)

		err = ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 0)
		require.NoError(t, err)

		err = ctx.BeginUnstakingMaxPaused()
		require.NoError(t, err)

		status, err := ctx.GetActorStatus(actorType, actor.GetAddress())
		require.Equal(t, typesUtil.UnstakingStatus, status, "incorrect status")

		ctx.Context.Release()
		tests.CleanupTest(ctx)
	}
}

func TestUtilityContext_EndBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)

	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		require.NoError(t, err)
	}
	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	// end-block logic verify
	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)

	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.False(t, proposerBalanceDifference.Cmp(expectedProposerBalanceDifference) != 0, fmt.Sprintf("unexpected before / after balance difference: expected %v got %v", expectedProposerBalanceDifference, proposerBalanceDifference))

	tests.CleanupTest(ctx)
}

func TestUtilityContext_GetAppHash(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	appHashTest, err := ctx.GetAppHash()
	require.NoError(t, err)

	appHashSource, er := ctx.Context.AppHash()
	require.NoError(t, er)
	require.Equal(t, appHashSource, appHashTest, "unexpected appHash")

	tests.CleanupTest(ctx)
}

func TestUtilityContext_UnstakeValidatorsActorsThatAreReady(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)
		poolName := actorType.GetActorPoolName()

		ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		err := ctx.Context.SetAppUnstakingBlocks(0)
		require.NoError(t, err)

		err = ctx.Context.SetAppMaxPausedBlocks(0)
		require.NoError(t, err)

		actors := GetAllTestingActors(t, ctx, actorType)
		for _, actor := range actors {
			require.False(t, actor.GetStatus() != typesUtil.StakedStatus, "wrong starting status")
			er := ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 1)
			require.NoError(t, er)
		}

		err = ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err)

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err)

		actors = GetAllTestingActors(t, ctx, actorType)
		require.False(t, actors[0].GetUnstakingHeight() == -1, fmt.Sprintf("validators still exists after unstake that are ready() call"))

		// TODO: We need to better define what 'deleted' really is in the postgres world.
		// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight

		ctx.Context.Release()
		tests.CleanupTest(ctx)
	}
}
