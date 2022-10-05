package test

import (
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/test_artifacts"

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
	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)
	byzantineAddrBz, er := hex.DecodeString(byzantine.GetAddress())
	require.NoError(t, er)
	proposerBeforeBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)
	// apply block
	if _, err := ctx.ApplyBlock(0, addrBz, [][]byte{txBz}, [][]byte{byzantineAddrBz}); err != nil {
		require.NoError(t, err, "apply block")
	}

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)

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
	proposerAfterBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.Equal(t, expectedProposerBalanceDifference, proposerBalanceDifference, "unexpected before / after balance difference")

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_BeginBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)

	byzantineBz, er := hex.DecodeString(byzantine.GetAddress())
	require.NoError(t, er)

	// apply block
	if _, err := ctx.ApplyBlock(0, addrBz, [][]byte{txBz}, [][]byte{byzantineBz}); err != nil {
		require.NoError(t, err)
	}

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)
		actor := GetFirstActor(t, ctx, actorType)

		var err error
		switch actorType {
		case typesUtil.UtilActorType_App:
			err = ctx.Context.SetParam(modules.AppMaxPauseBlocksParamName, 0)
		case typesUtil.UtilActorType_Val:
			err = ctx.Context.SetParam(modules.ValidatorMaxPausedBlocksParamName, 0)
		case typesUtil.UtilActorType_Fish:
			err = ctx.Context.SetParam(modules.FishermanMaxPauseBlocksParamName, 0)
		case typesUtil.UtilActorType_Node:
			err = ctx.Context.SetParam(modules.ServiceNodeMaxPauseBlocksParamName, 0)
		default:
			t.Fatalf("unexpected actor type %s", actorType.GetActorName())
		}
		require.NoError(t, err)
		addrBz, er := hex.DecodeString(actor.GetAddress())
		require.NoError(t, er)
		err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
		require.NoError(t, err)

		err = ctx.BeginUnstakingMaxPaused()
		require.NoError(t, err)

		status, err := ctx.GetActorStatus(actorType, addrBz)
		require.Equal(t, typesUtil.UnstakingStatus, status, "incorrect status")

		test_artifacts.CleanupTest(ctx)
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
	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)
	byzantineAddrBz, er := hex.DecodeString(byzantine.GetAddress())
	require.NoError(t, er)
	proposerBeforeBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	// apply block
	if _, err := ctx.ApplyBlock(0, addrBz, [][]byte{txBz}, [][]byte{byzantineAddrBz}); err != nil {
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
	proposerAfterBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	require.Equal(t, expectedProposerBalanceDifference, proposerBalanceDifference)

	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_GetAppHash(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	appHashTest, err := ctx.GetAppHash()
	require.NoError(t, err)

	appHashSource, er := ctx.Context.AppHash()
	require.NoError(t, er)
	require.Equal(t, appHashTest, appHashSource, "unexpected appHash")

	test_artifacts.CleanupTest(ctx)
}
func TestUtilityContext_UnstakeValidatorsActorsThatAreReady(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)
		poolName := actorType.GetActorPoolName()

		ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		err := ctx.Context.SetParam(modules.AppUnstakingBlocksParamName, 0)
		require.NoError(t, err)

		err = ctx.Context.SetParam(modules.AppMaxPauseBlocksParamName, 0)
		require.NoError(t, err)

		actors := GetAllTestingActors(t, ctx, actorType)
		for _, actor := range actors {
			// require.Equal(t, int32(typesUtil.StakedStatus), actor.GetStatus(), "wrong starting status")
			addrBz, er := hex.DecodeString(actor.GetAddress())
			require.NoError(t, er)
			er = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, er)
		}

		err = ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err)

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err)

		actors = GetAllTestingActors(t, ctx, actorType)
		require.NotEqual(t, actors[0].GetUnstakingHeight(), -1, "validators still exists after unstake that are ready() call")

		// TODO: We need to better define what 'deleted' really is in the postgres world.
		// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight

		test_artifacts.CleanupTest(ctx)
	}
}
