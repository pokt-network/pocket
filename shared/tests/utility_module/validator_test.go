package utility_module

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func TestUtilityContext_BurnValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount(genesis.ValidatorStakePoolName, big.NewInt(100000000000000))
	actor := GetAllTestingValidators(t, ctx)[0]
	burnPercentage := big.NewFloat(10)
	tokens, err := types.StringToBigInt(actor.StakedTokens)
	require.NoError(t, err)
	tokensFloat := big.NewFloat(0).SetInt(tokens)
	tokensFloat.Mul(tokensFloat, burnPercentage)
	tokensFloat.Quo(tokensFloat, big.NewFloat(100))
	tokensTrunc, _ := tokensFloat.Int(nil)
	afterTokensBig := big.NewInt(0).Sub(tokens, tokensTrunc)
	afterTokens := types.BigIntToString(afterTokensBig)
	if err := ctx.BurnActor(actor.Address, 10, typesUtil.ActorType_Val); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.StakedTokens != afterTokens {
		t.Fatalf("unexpected staked tokens after burn; expected %v got %v", afterTokens, actor.StakedTokens)
	}
}

func TestUtilityContext_GetMessageDoubleSignSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	msg := &typesUtil.MessageDoubleSign{
		ReporterAddress: actor.Address,
	}
	candidates, err := ctx.GetMessageDoubleSignSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actor.Address) {
		t.Fatalf("unexpected signer candidate: expected %v got %v", actor.Address, candidates[1])
	}
}

func TestUtilityContext_HandleMessageDoubleSign(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount(genesis.ValidatorStakePoolName, big.NewInt(100000000000000))
	actors := GetAllTestingValidators(t, ctx)
	reporter := actors[0]
	byzVal := actors[1]
	voteA := typesUtil.Vote{
		PublicKey: byzVal.PublicKey,
		Height:    0,
		Round:     0,
		Type:      0,
		BlockHash: crypto.SHA3Hash([]byte("voteA")),
	}
	voteB := voteA
	voteB.BlockHash = crypto.SHA3Hash([]byte("voteB"))
	msg := &typesUtil.MessageDoubleSign{
		VoteA:           &voteA,
		VoteB:           &voteB,
		ReporterAddress: reporter.Address,
	}
	if err := ctx.HandleMessageDoubleSign(msg); err != nil {
		t.Fatal(err)
	}
	stakedTokensAfterBig, err := ctx.GetStakeAmount(byzVal.Address, typesUtil.ActorType_Val)
	require.NoError(t, err)
	stakedTokensAfter := types.BigIntToString(stakedTokensAfterBig)
	burnPercentage, err := ctx.GetDoubleSignBurnPercentage()
	require.NoError(t, err)
	stakedTokensBeforeBig, err := types.StringToBigInt(byzVal.StakedTokens)
	require.NoError(t, err)
	stakedTokensBeforeFloat := big.NewFloat(0).SetInt(stakedTokensBeforeBig)
	stakedTokensBeforeFloat.Mul(stakedTokensBeforeFloat, big.NewFloat(float64(burnPercentage)))
	stakedTokensBeforeFloat.Quo(stakedTokensBeforeFloat, big.NewFloat(100))
	trunactedDiffTokens, _ := stakedTokensBeforeFloat.Int(nil)
	stakedTokensExpectedAfterBig := big.NewInt(0).Sub(stakedTokensBeforeBig, trunactedDiffTokens)
	stakedTokensExpectedAfter := types.BigIntToString(stakedTokensExpectedAfterBig)
	if stakedTokensAfter != stakedTokensExpectedAfter {
		t.Fatalf("unexpected token amount after double sign handling: expected %v got %v", stakedTokensExpectedAfter, stakedTokensAfter)
	}
}

func TestUtilityContext_GetValidatorMissedBlocks(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	missedBlocks := 3
	if int(actor.MissedBlocks) == missedBlocks {
		t.Fatal("wrong missed blocks starting amount")
	}
	actor.MissedBlocks = uint32(missedBlocks)
	if err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).SetValidatorMissedBlocks(actor.Address, int(actor.MissedBlocks)); err != nil {
		t.Fatal(err)
	}
	gotMissedBlocks, err := ctx.GetValidatorMissedBlocks(actor.Address)
	require.NoError(t, err)
	if gotMissedBlocks != missedBlocks {
		t.Fatalf("unexpected missed blocks: expected %v got %v", missedBlocks, gotMissedBlocks)
	}
}

func TestUtilityContext_HandleByzantineValidators(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount(genesis.ValidatorStakePoolName, big.NewInt(100000000000000))
	actor := GetAllTestingValidators(t, ctx)[0]
	stakedTokensBeforeBig, err := types.StringToBigInt(actor.StakedTokens)
	require.NoError(t, err)
	maxMissed, err := ctx.GetValidatorMaxMissedBlocks()
	require.NoError(t, err)
	if err := ctx.SetValidatorMissedBlocks(actor.Address, maxMissed); err != nil {
		t.Fatal(err)
	}
	// Pause scenario only
	// TODO add more situations / paths to test
	if err := ctx.HandleByzantineValidators([][]byte{actor.Address}); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor should be paused after byzantine handling")
	}
	stakedTokensAfterBig, err := types.StringToBigInt(actor.StakedTokens)
	require.NoError(t, err)
	stakedTokensAfter := types.BigIntToString(stakedTokensAfterBig)
	burnPercentage, err := ctx.GetMissedBlocksBurnPercentage()
	require.NoError(t, err)
	stakedTokensBeforeFloat := big.NewFloat(0).SetInt(stakedTokensBeforeBig)
	stakedTokensBeforeFloat.Mul(stakedTokensBeforeFloat, big.NewFloat(float64(burnPercentage)))
	stakedTokensBeforeFloat.Quo(stakedTokensBeforeFloat, big.NewFloat(100))
	trunactedDiffTokens, _ := stakedTokensBeforeFloat.Int(nil)
	stakedTokensExpectedAfterBig := big.NewInt(0).Sub(stakedTokensBeforeBig, trunactedDiffTokens)
	stakedTokensExpectedAfter := types.BigIntToString(stakedTokensExpectedAfterBig)
	if stakedTokensAfter != stakedTokensExpectedAfter {
		t.Fatalf("tokens are not as expected after handling: expected %v got %v", stakedTokensExpectedAfter, stakedTokensAfter)
	}
}

func TestUtilityContext_HandleProposalRewards(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	actorTokensBeforeBig, err := ctx.GetAccountAmount(actor.Address)
	require.NoError(t, err)
	require.NoError(t, err)
	feeAndRewardsCollected := big.NewInt(100)
	err = ctx.SetPoolAmount(genesis.FeePoolName, feeAndRewardsCollected)
	require.NoError(t, err)
	proposerCutPercentage, err := ctx.GetProposerPercentageOfFees()
	require.NoError(t, err)
	daoCutPercentage := 100 - proposerCutPercentage
	if daoCutPercentage < 0 {
		t.Fatal("dao cut percentage negative")
	}
	feesAndRewardsCollectedFloat := new(big.Float).SetInt(feeAndRewardsCollected)
	feesAndRewardsCollectedFloat.Mul(feesAndRewardsCollectedFloat, big.NewFloat(float64(proposerCutPercentage)))
	feesAndRewardsCollectedFloat.Quo(feesAndRewardsCollectedFloat, big.NewFloat(100))
	amountToProposer, _ := feesAndRewardsCollectedFloat.Int(nil)
	expectedResultBig := actorTokensBeforeBig.Add(actorTokensBeforeBig, amountToProposer)
	expectedResult := types.BigIntToString(expectedResultBig)
	if err := ctx.HandleProposalRewards(actor.Address); err != nil {
		t.Fatal(err)
	}
	actorTokensAfterBig, err := ctx.GetAccountAmount(actor.Address)
	require.NoError(t, err)
	actorTokensAfter := types.BigIntToString(actorTokensAfterBig)
	if actorTokensAfter != expectedResult {
		t.Fatalf("unexpected token amount after; expected %v got %v", expectedResult, actorTokensAfter)
	}
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Validator {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
