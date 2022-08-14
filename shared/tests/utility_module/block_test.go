package utility_module

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/shared/tests"
	"math"
	"math/big"
	"testing"

	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_ApplyBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, startingBalance, amount, signer := NewTestingTransaction(t, ctx)

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
	// beginBlock logic verify
	//missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address) TODO not implemented in persistence context yet
	//require.NoError(t, err)
	//require.True(t, missed == 1, fmt.Sprintf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks))
	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	expectedAmountSubtracted := big.NewInt(0).Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.True(t, amountAfter.Cmp(expectedAfterBalance) == 0, fmt.Sprintf("unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter))
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
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func TestUtilityContext_BeginBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]

	txBz, err := tx.Bytes()
	require.NoError(t, err)
	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)
	byzantineAddrBz, er := hex.DecodeString(byzantine.GetAddress())
	require.NoError(t, er)
	// apply block
	if _, err := ctx.ApplyBlock(0, addrBz, [][]byte{txBz}, [][]byte{byzantineAddrBz}); err != nil {
		require.NoError(t, err)
	}
	// beginBlock logic verify
	//missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address) TODO not yet implemented
	//require.NoError(t, err)
	//require.False(t, missed != 1, fmt.Sprintf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
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
		addrBz, er := hex.DecodeString(actor.GetAddress())
		require.NoError(t, er)
		err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
		require.NoError(t, err)

		err = ctx.BeginUnstakingMaxPaused()
		require.NoError(t, err)

		status, err := ctx.GetActorStatus(actorType, addrBz)
		require.False(t, status != 1, fmt.Sprintf("incorrect status; expected %d got %d", 1, actor.UnstakingHeight))
		ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
		tests.CleanupTest()
	}
}

func TestUtilityContext_EndBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
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
	require.Equal(t, proposerBalanceDifference, expectedProposerBalanceDifference)
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func TestUtilityContext_GetAppHash(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	appHashTest, err := ctx.GetAppHash()
	require.NoError(t, err)

	appHashSource, er := ctx.Context.AppHash()
	require.NoError(t, er)
	require.False(t, !bytes.Equal(appHashSource, appHashTest), fmt.Sprintf("unexpected appHash, expected %v got %v", appHashSource, appHashTest))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
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
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			require.False(t, actor.UnstakingHeight != -1, "wrong starting status")
			er := ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, er)
		}

		err = ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err)

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err)
		actors = GetAllTestingActors(t, ctx, actorType)
		require.False(t, actors[0].GetUnstakingHeight() == -1, fmt.Sprintf("validators still exists after unstake that are ready() call"))
		// TODO (Team) we need to better define what 'deleted' really is in the postgres world.
		// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight
		ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
		tests.CleanupTest()
	}
}
