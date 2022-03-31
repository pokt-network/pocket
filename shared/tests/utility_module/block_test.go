package utility_module

import (
	"bytes"
	"github.com/pokt-network/pocket/utility/types"
	"math"
	"math/big"
	"testing"
)

func TestUtilityContext_ApplyBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, startingBalance, amount, signer := NewTestingTransaction(t, ctx)
	vals := GetAllTestingValidators(t, ctx)
	proposer := vals[0]
	byzantine := vals[1]
	txBz, err := tx.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	if err != nil {
		t.Fatal(err)
	}
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
	}
	// beginBlock logic verify
	missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	if err != nil {
		t.Fatal(err)
	}
	if missed != 1 {
		t.Fatalf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks)
	}
	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	if err != nil {
		t.Fatal(err)
	}
	expectedAmountSubtracted := big.NewInt(0).Add(amount, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := ctx.GetAccountAmount(signer.Address())
	if err != nil {
		t.Fatal(err)
	}
	if amountAfter.Cmp(expectedAfterBalance) != 0 {
		t.Fatalf("unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter)
	}
	// end-block logic verify
	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	if err != nil {
		t.Fatal(err)
	}
	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
	}
	// beginBlock logic verify
	missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	if err != nil {
		t.Fatal(err)
	}
	if missed != 1 {
		t.Fatalf("wrong missed blocks amount; expected %v got %v", 1, byzantine.MissedBlocks)
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingValidators(t, ctx)[0]
	err := ctx.Context.SetValidatorMaxPausedBlocks(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetValidatorPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPausedActors(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetValidatorStatus(actor.Address)
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
	if err != nil {
		t.Fatal(err)
	}
	proposerBeforeBalance, err := ctx.GetAccountAmount(proposer.Address)
	if err != nil {
		t.Fatal(err)
	}
	// apply block
	if _, err := ctx.ApplyBlock(0, proposer.Address, [][]byte{txBz}, [][]byte{byzantine.Address}); err != nil {
		t.Fatal(err)
	}
	// deliverTx logic verify
	feeBig, err := ctx.GetMessageSendFee()
	if err != nil {
		t.Fatal(err)
	}
	// end-block logic verify
	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	if err != nil {
		t.Fatal(err)
	}
	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeBig)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	expectedProposerBalanceDifference, _ := feesAndRewardsCollectedFloat.Int(nil)
	proposerAfterBalance, err := ctx.GetAccountAmount(proposer.Address)
	if err != nil {
		t.Fatal(err)
	}
	proposerBalanceDifference := big.NewInt(0).Sub(proposerAfterBalance, proposerBeforeBalance)
	if proposerBalanceDifference.Cmp(expectedProposerBalanceDifference) != 0 {
		t.Fatalf("unexpected before / after balance difference: expected %v got %v", expectedProposerBalanceDifference, proposerBalanceDifference)
	}
}

func TestUtilityContext_GetAppHash(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	appHashTest, err := ctx.GetAppHash()
	if err != nil {
		t.Fatal(err)
	}
	appHashSource, er := ctx.Context.AppHash()
	if er != nil {
		t.Fatal(er)
	}
	if !bytes.Equal(appHashSource, appHashTest) {
		t.Fatalf("unexpected appHash, expected %v got %v", appHashSource, appHashTest)
	}
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(types.ValidatorStakePoolName, big.NewInt(math.MaxInt64))
	if err := ctx.Context.SetValidatorUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingValidators(t, ctx)[0]
	if actor.Status != types.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetValidatorMaxPausedBlocks(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ctx.SetValidatorPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeValidatorsPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeActorsThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingValidators(t, ctx)) != 0 {
		t.Fatal("actor still exists after unstake that are ready() call")
	}
}
