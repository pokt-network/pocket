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

func TestUtilityContext_HandleMessageStakeValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	out, _ := crypto.GenerateAddress()
	if err := ctx.SetAccountAmount(out, defaultAmount); err != nil {
		t.Fatal(err)
	}
	msg := &typesUtil.MessageStakeValidator{
		PublicKey:     pubKey.Bytes(),
		Amount:        defaultAmountString,
		ServiceUrl:    defaultServiceUrl,
		OutputAddress: out,
		Signer:        out,
	}
	if err := ctx.HandleMessageStakeValidator(msg); err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingValidators(t, ctx)
	var actor *genesis.Validator
	for _, a := range actors {
		if bytes.Equal(a.PublicKey, msg.PublicKey) {
			actor = a
			break
		}
	}
	if !bytes.Equal(actor.Address, pubKey.Address()) {
		t.Fatalf("incorrect address, expected %v, got %v", pubKey.Address(), actor.Address)
	}
	if actor.Status != typesUtil.StakedStatus {
		t.Fatalf("incorrect status, expected %v, got %v", typesUtil.StakedStatus, actor.Status)
	}
	if actor.ServiceUrl != defaultServiceUrl {
		t.Fatalf("incorrect chains, expected %v, got %v", actor.ServiceUrl, defaultServiceUrl)
	}
	if actor.Paused != false {
		t.Fatalf("incorrect paused status, expected %v, got %v", false, actor.Paused)
	}
	if actor.PausedHeight != types.HeightNotUsed {
		t.Fatalf("incorrect paused height, expected %v, got %v", types.HeightNotUsed, actor.PausedHeight)
	}
	if actor.StakedTokens != defaultAmountString {
		t.Fatalf("incorrect staked amount, expected %v, got %v", actor.StakedTokens, defaultAmountString)
	}
	if actor.UnstakingHeight != types.HeightNotUsed {
		t.Fatalf("incorrect unstaking height, expected %v, got %v", types.HeightNotUsed, actor.UnstakingHeight)
	}
	if !bytes.Equal(actor.Output, out) {
		t.Fatalf("incorrect output address, expected %v, got %v", actor.Output, out)
	}
}

func TestUtilityContext_HandleMessageEditStakeValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	msg := &typesUtil.MessageEditStakeValidator{
		Address:     actor.Address,
		ServiceUrl:  defaultServiceUrlEdited,
		AmountToAdd: zeroAmountString,
		Signer:      actor.Address,
	}
	msgServiceUrlEdited := msg
	msgServiceUrlEdited.ServiceUrl = defaultServiceUrlEdited
	if err := ctx.HandleMessageEditStakeValidator(msgServiceUrlEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.Paused != false {
		t.Fatalf("incorrect paused status, expected %v, got %v", false, actor.Paused)
	}
	if actor.PausedHeight != types.HeightNotUsed {
		t.Fatalf("incorrect paused height, expected %v, got %v", types.HeightNotUsed, actor.PausedHeight)
	}
	if actor.ServiceUrl != defaultServiceUrlEdited {
		t.Fatalf("incorrect serviceurl, expected %v, got %v", defaultServiceUrlEdited, actor.ServiceUrl)
	}
	if actor.StakedTokens != defaultAmountString {
		t.Fatalf("incorrect staked tokens, expected %v, got %v", defaultAmountString, actor.StakedTokens)
	}
	if actor.UnstakingHeight != types.HeightNotUsed {
		t.Fatalf("incorrect unstaking height, expected %v, got %v", types.HeightNotUsed, actor.UnstakingHeight)
	}
	if !bytes.Equal(actor.Output, actor.Output) {
		t.Fatalf("incorrect output address, expected %v, got %v", actor.Output, actor.Output)
	}
	amountEdited := big.NewInt(1)
	expectedAmount := types.BigIntToString(big.NewInt(0).Add(defaultAmount, amountEdited))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := msg
	msgAmountEdited.AmountToAdd = amountEditedString
	if err := ctx.HandleMessageEditStakeValidator(msgAmountEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("incorrect amount status, expected %v, got %v", expectedAmount, actor.StakedTokens)
	}
}

func TestUtilityContext_HandleMessagePauseValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingValidators(t, ctx)[0]
	msg := &typesUtil.MessagePauseValidator{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseValidator(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
}

func TestUtilityContext_HandleMessageUnpauseValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetValidatorMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingValidators(t, ctx)[0]
	msg := &typesUtil.MessagePauseValidator{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseValidator(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
	msgU := &typesUtil.MessageUnpauseValidator{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnpauseValidator(msgU); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.Paused {
		t.Fatal("actor is paused after")
	}
}

func TestUtilityContext_HandleMessageUnstakeValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetValidatorMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingValidators(t, ctx)[0]
	msg := &typesUtil.MessageUnstakeValidator{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnstakeValidator(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("actor isn't unstaking")
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedValidators(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingValidators(t, ctx)[0]
	err := ctx.Context.SetValidatorMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetValidatorPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPausedValidators(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetValidatorStatus(actor.Address)
	if status != 1 {
		t.Fatalf("incorrect status; expected %d got %d", 1, actor.Status)
	}
}

func TestUtilityContext_CalculateValidatorUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetValidatorUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.CalculateValidatorUnstakingHeight()
	require.NoError(t, err)
	if unstakingBlocks != unstakingHeight {
		t.Fatalf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight)
	}
}

func TestUtilityContext_DeleteValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingValidators(t, ctx)
	actor := actors[0]
	if err := ctx.DeleteValidator(actor.Address); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingValidators(t, ctx)) != len(actors)-1 {
		t.Fatal("deletion unsuccessful")
	}
}

func TestUtilityContext_GetValidatorExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	randAddr, _ := crypto.GenerateAddress()
	actor := GetAllTestingValidators(t, ctx)[0]
	exists, err := ctx.GetValidatorExists(actor.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = ctx.GetValidatorExists(randAddr)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that shouldn't exist does")
	}
}

func TestUtilityContext_GetValidatorOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	outputAddress, err := ctx.GetValidatorOutputAddress(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(outputAddress, actor.Output) {
		t.Fatalf("unexpected output address, expected %v got %v", actor.Output, outputAddress)
	}
}

func TestUtilityContext_GetValidatorPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	pauseHeight := int64(100)
	if err := ctx.SetValidatorPauseHeight(actor.Address, pauseHeight); err != nil {
		t.Fatal(err)
	}
	gotPauseHeight, err := ctx.GetValidatorPauseHeightIfExists(actor.Address)
	require.NoError(t, err)
	if pauseHeight != gotPauseHeight {
		t.Fatal("unable to get pause height from the actor")
	}
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetValidatorPauseHeightIfExists(addr)
	if err == nil {
		t.Fatal("no error on non-existent actor pause height")
	}
}

func TestUtilityContext_GetValidatorsReadyToUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	if err := ctx.SetValidatorUnstakingHeightAndStatus(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	actors, err := ctx.GetValidatorsReadyToUnstake()
	require.NoError(t, err)
	if !bytes.Equal(actors[0].Address, actor.Address) {
		t.Fatalf("unexpected actor ready to unstake: expected %s, got %s", actor.Address, actors[0].Address)
	}
}

func TestUtilityContext_GetMessageEditStakeValidatorSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingValidators(t, ctx)
	msgEditStake := &typesUtil.MessageEditStakeValidator{
		Address:     actors[0].Address,
		ServiceUrl:  defaultServiceUrl,
		AmountToAdd: defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeValidatorSignerCandidates(msgEditStake)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessagePauseValidatorSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingValidators(t, ctx)
	msg := &typesUtil.MessagePauseValidator{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessagePauseValidatorSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageStakeValidatorSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	out, _ := crypto.GenerateAddress()
	msg := &typesUtil.MessageStakeValidator{
		PublicKey:     pubKey.Bytes(),
		Amount:        defaultAmountString,
		ServiceUrl:    defaultServiceUrl,
		OutputAddress: out,
		Signer:        nil,
	}
	candidates, err := ctx.GetMessageStakeValidatorSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnpauseValidatorSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingValidators(t, ctx)
	msg := &typesUtil.MessageUnpauseValidator{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseValidatorSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnstakeValidatorSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingValidators(t, ctx)
	msg := &typesUtil.MessageUnstakeValidator{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeValidatorSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_InsertValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	if err := ctx.InsertValidator(addr, pubKey.Bytes(), addr, defaultServiceUrl, defaultAmountString); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetValidatorExists(addr)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor does not exist after insert")
	}
	actors := GetAllTestingValidators(t, ctx)
	for _, actor := range actors {
		if bytes.Equal(actor.Address, addr) {
			if actor.StakedTokens != defaultAmountString {
				t.Fatal("wrong staked tokens")
			}
			if actor.ServiceUrl != defaultServiceUrl {
				t.Fatal("wrong serviceURL")
			}
			if !bytes.Equal(actor.Output, addr) {
				t.Fatal("wrong output addr")
			}
			return
		}
	}
	t.Fatal("actor not found after insert in GetAll() call")
}

func TestUtilityContext_UnstakeValidatorsPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingValidators(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetValidatorMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetValidatorPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeValidatorsPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("status does not equal unstaking")
	}
	unstakingBlocks, err := ctx.GetValidatorUnstakingBlocks()
	require.NoError(t, err)
	if actor.UnstakingHeight != unstakingBlocks+1 {
		t.Fatal("incorrect unstaking height")
	}
}

func TestUtilityContext_UnstakeValidatorsThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(genesis.ValidatorStakePoolName, big.NewInt(100000000000000000))
	if err := ctx.Context.SetValidatorUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	err := ctx.Context.SetValidatorMaxPausedBlocks(0)
	if err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingValidators(t, ctx)
	for _, actor := range actors {
		if actor.Status != typesUtil.StakedStatus {
			t.Fatal("wrong starting status")
		}
		if err := ctx.SetValidatorPauseHeight(actor.Address, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.UnstakeValidatorsPausedBefore(2); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeValidatorsThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingValidators(t, ctx)) != 0 {
		t.Fatal("validators still exists after unstake that are ready() call")
	}
}

func TestUtilityContext_UpdateValidator(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingValidators(t, ctx)[0]
	newAmountBig := big.NewInt(9999999999999999)
	newAmount := types.BigIntToString(newAmountBig)
	oldAmount := actor.StakedTokens
	oldAmountBig, err := types.StringToBigInt(oldAmount)
	require.NoError(t, err)
	expectedAmountBig := newAmountBig.Add(newAmountBig, oldAmountBig)
	expectedAmount := types.BigIntToString(expectedAmountBig)
	if err := ctx.UpdateValidator(actor.Address, actor.ServiceUrl, newAmount); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("updated amount is incorrect; expected %s got %s", expectedAmount, actor.StakedTokens)
	}
}

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
	if err := ctx.BurnValidator(actor.Address, 10); err != nil {
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
	stakedTokensAfterBig, err := ctx.GetValidatorStakedTokens(byzVal.Address)
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

func TestUtilityContext_GetValidatorStakedTokens(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	tokensBig, err := ctx.GetValidatorStakedTokens(actor.Address)
	require.NoError(t, err)
	tokens := types.BigIntToString(tokensBig)
	if actor.StakedTokens != tokens {
		t.Fatalf("unexpected staked tokens: expected %v got %v ", actor.StakedTokens, tokens)
	}
}

func TestUtilityContext_GetValidatorStatus(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingValidators(t, ctx)[0]
	status, err := ctx.GetValidatorStatus(actor.Address)
	require.NoError(t, err)
	if int(actor.Status) != status {
		t.Fatalf("unexpected staked tokens: expected %v got %v ", int(actor.Status), status)
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

func TestUtilityContext_SetValidatorStakedTokens(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	afterTokensExpectedBig := big.NewInt(100)
	afterTokensExpected := types.BigIntToString(afterTokensExpectedBig)
	actor := GetAllTestingValidators(t, ctx)[0]
	if actor.StakedTokens == afterTokensExpected {
		t.Fatal("bad starting amount for staked tokens")
	}
	if err := ctx.SetValidatorStakedTokens(actor.Address, afterTokensExpectedBig); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingValidators(t, ctx)[0]
	if actor.StakedTokens != afterTokensExpected {
		t.Fatalf("unexpected after tokens: expected %v got %v", afterTokensExpected, actor.StakedTokens)
	}
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Validator {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
