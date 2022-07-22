package utility_module

import (
	"bytes"
	"math"
	"math/big"
	"testing"

	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
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
	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
	}
	// beginBlock logic verify
	missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	require.NoError(t, err)
	if missed != 1 {
		t.Fatalf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks)
	}
	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)
	expectedAmountSubtracted := big.NewInt(0).Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	if amountAfter.Cmp(expectedAfterBalance) != 0 {
		t.Fatalf("unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter)
	}
	// end-block logic verify
	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)
	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	// DISCUSS/HACK: Why did we need to add the line below?
	feesAndRewardsCollectedFloat.Add(feesAndRewardsCollectedFloat, big.NewFloat(float64(feeBig.Int64())))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)
	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	if proposerBalanceDifference.Cmp(expectedProposerBalanceDifference) != 0 {
		t.Fatalf("unexpected before / after balance difference: expected %v got %v", expectedProposerBalanceDifference, proposerBalanceDifference)
	}
}

func TestUtilityContext_BeginBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := NewTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]
	txBz, err := tx.Bytes()
	require.NoError(t, err)
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
	}
	// beginBlock logic verify
	missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	require.NoError(t, err)
	if missed != 1 {
		t.Fatalf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks)
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPaused(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetActorStatus(actor.Address, typesUtil.ActorType_App)
	if status != 1 {
		t.Fatalf("incorrect status; expected %d got %d", 1, actor.Status)
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
	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
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
	// DISCUSS/HACK: Why did we need to add the line below?
	feesAndRewardsCollectedFloat.Add(feesAndRewardsCollectedFloat, big.NewFloat(float64(feeBig.Int64())))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	require.NoError(t, err)
	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	if proposerBalanceDifference.Cmp(expectedProposerBalanceDifference) != 0 {
		t.Fatalf("unexpected before / after balance difference: expected %v got %v", expectedProposerBalanceDifference, proposerBalanceDifference)
	}
}

func TestUtilityContext_GetAppHash(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	appHashTest, err := ctx.GetAppHash()
	require.NoError(t, err)
	appHashSource, er := ctx.Context.AppHash()
	if er != nil {
		t.Fatal(er)
	}
	if !bytes.Equal(appHashSource, appHashTest) {
		t.Fatalf("unexpected appHash, expected %v got %v", appHashSource, appHashTest)
	}
}

func TestUtilityContext_UnstakeValidatorsActorsThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(typesGenesis.AppStakePoolName, big.NewInt(math.MaxInt64))
	if err := ctx.Context.SetAppUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	if err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		if actor.Status != typesUtil.StakedStatus {
			t.Fatal("wrong starting status")
		}
		if err := ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.UnstakeActorPausedBefore(2, typesUtil.ActorType_App); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeActorsThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingApps(t, ctx)) != 0 {
		t.Fatal("validators still exists after unstake that are ready() call")
	}
}
