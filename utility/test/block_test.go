package test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_ApplyBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, startingBalance, amountSent, signer := newTestingTransaction(t, ctx)

	txBz, er := tx.Bytes()
	require.NoError(t, er)

	proposer := getFirstActor(t, &ctx, coreTypes.ActorType_ACTOR_TYPE_VAL)

	addrBz, err := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, err)

	proposerBeforeBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	err = ctx.SetProposalBlock("", addrBz, [][]byte{txBz})
	require.NoError(t, err)

	appHash, err := ctx.ApplyBlock()
	require.NoError(t, err)
	require.NotNil(t, appHash)

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)

	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

	expectedAmountSubtracted := big.NewInt(0).Add(amountSent, feeBig)
	expectedAfterBalance := big.NewInt(0).Sub(startingBalance, expectedAmountSubtracted)
	amountAfter, err := ctx.GetAccountAmount(signer.Address())
	require.NoError(t, err)
	require.Equal(t, expectedAfterBalance, amountAfter, "unexpected after balance; expected %v got %v", expectedAfterBalance, amountAfter)

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

	test_artifacts.CleanupTest(&ctx)
}

func TestUtilityContext_BeginBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)

	proposer := getFirstActor(t, &ctx, coreTypes.ActorType_ACTOR_TYPE_VAL)

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)

	er = ctx.SetProposalBlock("", addrBz, [][]byte{txBz})
	require.NoError(t, er)

	_, er = ctx.ApplyBlock()
	require.NoError(t, er)

	// // TODO: Uncomment this once `GetValidatorMissedBlocks` is implemented.
	// beginBlock logic verify
	// missed, err := ctx.GetValidatorMissedBlocks(byzantine.Address)
	// require.NoError(t, err)
	// require.Equal(t, missed, 1)

	test_artifacts.CleanupTest(&ctx)
}

func TestUtilityContext_EndBlock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	tx, _, _, _ := newTestingTransaction(t, ctx)

	proposer := getFirstActor(t, &ctx, coreTypes.ActorType_ACTOR_TYPE_VAL)

	txBz, err := tx.Bytes()
	require.NoError(t, err)

	addrBz, er := hex.DecodeString(proposer.GetAddress())
	require.NoError(t, er)

	proposerBeforeBalance, err := ctx.GetAccountAmount(addrBz)
	require.NoError(t, err)

	er = ctx.SetProposalBlock("", addrBz, [][]byte{txBz})
	require.NoError(t, er)

	_, er = ctx.ApplyBlock()
	require.NoError(t, er)

	feeBig, err := ctx.GetMessageSendFee()
	require.NoError(t, err)

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

	test_artifacts.CleanupTest(&ctx)
}
