package utility_module

import (
	"bytes"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func TestUtilityContext_HandleMessageStakeFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	out, _ := crypto.GenerateAddress()
	if err := ctx.SetAccountAmount(out, defaultAmount); err != nil {
		t.Fatal(err)
	}
	msg := &typesUtil.MessageStakeFisherman{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
		Signer:        out,
	}
	if err := ctx.HandleMessageStakeFisherman(msg); err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingFishermen(t, ctx)
	var actor *genesis.Fisherman
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
	if !reflect.DeepEqual(actor.Chains, msg.Chains) {
		t.Fatalf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains)
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

func TestUtilityContext_HandleMessageEditStakeFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingFishermen(t, ctx)[0]
	msg := &typesUtil.MessageEditStakeFisherman{
		Address:     actor.Address,
		Chains:      defaultTestingChains,
		AmountToAdd: zeroAmountString,
		Signer:      actor.Address,
	}
	msgChainsEdited := msg
	msgChainsEdited.Chains = defaultTestingChainsEdited
	if err := ctx.HandleMessageEditStakeFisherman(msgChainsEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if !reflect.DeepEqual(actor.Chains, msg.Chains) {
		t.Fatalf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains)
	}
	if actor.Paused != false {
		t.Fatalf("incorrect paused status, expected %v, got %v", false, actor.Paused)
	}
	if actor.PausedHeight != types.HeightNotUsed {
		t.Fatalf("incorrect paused height, expected %v, got %v", types.HeightNotUsed, actor.PausedHeight)
	}
	if !reflect.DeepEqual(actor.Chains, msgChainsEdited.Chains) {
		t.Fatalf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains)
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
	if err := ctx.HandleMessageEditStakeFisherman(msgAmountEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("incorrect amount status, expected %v, got %v", expectedAmount, actor.StakedTokens)
	}
}

func TestUtilityContext_HandleMessagePauseFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	msg := &typesUtil.MessagePauseFisherman{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseFisherman(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
}

func TestUtilityContext_HandleMessageUnpauseFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetFishermanMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingFishermen(t, ctx)[0]
	msg := &typesUtil.MessagePauseFisherman{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseFisherman(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
	msgU := &typesUtil.MessageUnpauseFisherman{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnpauseFisherman(msgU); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if actor.Paused {
		t.Fatal("actor is paused after")
	}
}

func TestUtilityContext_HandleMessageUnstakeFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetFishermanMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingFishermen(t, ctx)[0]
	msg := &typesUtil.MessageUnstakeFisherman{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnstakeFisherman(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("actor isn't unstaking")
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedFishermen(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	err := ctx.Context.SetFishermanMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetFishermanPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPausedFishermen(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetFishermanStatus(actor.Address)
	if status != 1 {
		t.Fatalf("incorrect status; expected %d got %d", 1, actor.Status)
	}
}

func TestUtilityContext_CalculateFishermanUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetFishermanUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.CalculateFishermanUnstakingHeight()
	require.NoError(t, err)
	if unstakingBlocks != unstakingHeight {
		t.Fatalf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight)
	}
}

func TestUtilityContext_DeleteFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingFishermen(t, ctx)[0]
	if err := ctx.DeleteFisherman(actor.Address); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingFishermen(t, ctx)) > 0 {
		t.Fatal("deletion unsuccessful")
	}
}

func TestUtilityContext_GetFishermanExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	randAddr, _ := crypto.GenerateAddress()
	actor := GetAllTestingFishermen(t, ctx)[0]
	exists, err := ctx.GetFishermanExists(actor.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = ctx.GetFishermanExists(randAddr)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that shouldn't exist does")
	}
}

func TestUtilityContext_GetFishermanOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingFishermen(t, ctx)[0]
	outputAddress, err := ctx.GetFishermanOutputAddress(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(outputAddress, actor.Output) {
		t.Fatalf("unexpected output address, expected %v got %v", actor.Output, outputAddress)
	}
}

func TestUtilityContext_GetFishermanPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingFishermen(t, ctx)[0]
	pauseHeight := int64(100)
	if err := ctx.SetFishermanPauseHeight(actor.Address, pauseHeight); err != nil {
		t.Fatal(err)
	}
	gotPauseHeight, err := ctx.GetFishermanPauseHeightIfExists(actor.Address)
	require.NoError(t, err)
	if pauseHeight != gotPauseHeight {
		t.Fatal("unable to get pause height from the actor")
	}
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetFishermanPauseHeightIfExists(addr)
	if err == nil {
		t.Fatal("no error on non-existent actor pause height")
	}
}

func TestUtilityContext_GetFishermenReadyToUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingFishermen(t, ctx)[0]
	if err := ctx.SetFishermanUnstakingHeightAndStatus(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	actors, err := ctx.GetFishermenReadyToUnstake()
	require.NoError(t, err)
	if !bytes.Equal(actors[0].Address, actor.Address) {
		t.Fatalf("unexpected actor ready to unstake: expected %s, got %s", actor.Address, actors[0].Address)
	}
}

func TestUtilityContext_GetMessageEditStakeFishermanSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingFishermen(t, ctx)
	msgEditStake := &typesUtil.MessageEditStakeFisherman{
		Address:     actors[0].Address,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeFishermanSignerCandidates(msgEditStake)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessagePauseFishermanSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingFishermen(t, ctx)
	msg := &typesUtil.MessagePauseFisherman{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessagePauseFishermanSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageStakeFishermanSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	out, _ := crypto.GenerateAddress()
	msg := &typesUtil.MessageStakeFisherman{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
	}
	candidates, err := ctx.GetMessageStakeFishermanSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnpauseFishermanSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingFishermen(t, ctx)
	msg := &typesUtil.MessageUnpauseFisherman{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseFishermanSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnstakeFishermanSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingFishermen(t, ctx)
	msg := &typesUtil.MessageUnstakeFisherman{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeFishermanSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_InsertFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	if err := ctx.InsertFisherman(addr, pubKey.Bytes(), addr, defaultServiceUrl, defaultAmountString, defaultTestingChains); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetFishermanExists(addr)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor does not exist after insert")
	}
	actors := GetAllTestingFishermen(t, ctx)
	for _, actor := range actors {
		if bytes.Equal(actor.Address, addr) {
			if actor.Chains[0] != defaultTestingChains[0] {
				t.Fatal("wrong chains")
			}
			if actor.ServiceUrl != defaultServiceUrl {
				t.Fatal("wrong serviceURL")
			}
			if actor.StakedTokens != defaultAmountString {
				t.Fatal("wrong staked tokens")
			}
			if !bytes.Equal(actor.Output, addr) {
				t.Fatal("wrong output addr")
			}
			return
		}
	}
	t.Fatal("actor not found after insert in GetAll() call")
}

func TestUtilityContext_UnstakeFishermenPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetFishermanMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetFishermanPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeFishermenPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("status does not equal unstaking")
	}
	unstakingBlocks, err := ctx.GetFishermanUnstakingBlocks()
	require.NoError(t, err)
	if actor.UnstakingHeight != unstakingBlocks+1 {
		t.Fatal("incorrect unstaking height")
	}
}

func TestUtilityContext_UnstakeFishermenThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(genesis.FishermanStakePoolName, big.NewInt(math.MaxInt64))
	if err := ctx.Context.SetFishermanUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	err := ctx.Context.SetFishermanMaxPausedBlocks(0)
	if err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingFishermen(t, ctx)
	for _, actor := range actors {
		if actor.Status != typesUtil.StakedStatus {
			t.Fatal("wrong starting status")
		}
		if err := ctx.SetFishermanPauseHeight(actor.Address, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := ctx.UnstakeFishermenPausedBefore(2); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeFishermenThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingFishermen(t, ctx)) != 0 {
		t.Fatal("fishermen still exists after unstake that are ready() call")
	}
}

func TestUtilityContext_UpdateFisherman(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	newAmountBig := big.NewInt(9999999999999999)
	newAmount := types.BigIntToString(newAmountBig)
	oldAmount := actor.StakedTokens
	oldAmountBig, err := types.StringToBigInt(oldAmount)
	require.NoError(t, err)
	expectedAmountBig := newAmountBig.Add(newAmountBig, oldAmountBig)
	expectedAmount := types.BigIntToString(expectedAmountBig)
	if err := ctx.UpdateFisherman(actor.Address, actor.ServiceUrl, newAmount, actor.Chains); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingFishermen(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("updated amount is incorrect; expected %s got %s", expectedAmount, actor.StakedTokens)
	}
}

func GetAllTestingFishermen(t *testing.T, ctx utility.UtilityContext) []*genesis.Fisherman {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllFishermen(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func TestUtilityContext_GetMessageFishermanPauseServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	candidates, err := ctx.GetMessageFishermanPauseServiceNodeSignerCandidates(&typesUtil.MessageFishermanPauseServiceNode{
		Reporter: actor.Address,
	})
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actor.Output) {
		t.Fatal("output address is not a signer candidate")
	}
	if !bytes.Equal(candidates[1], actor.Address) {
		t.Fatal("operator address is not a signer candidate")
	}
}

func TestUtilityContext_HandleMessageFishermanPauseServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingFishermen(t, ctx)[0]
	sn := GetAllTestingServiceNodes(t, ctx)[0]
	if sn.Paused == true {
		t.Fatal("incorrect starting pause status")
	}
	if err := ctx.HandleMessageFishermanPauseServiceNode(&typesUtil.MessageFishermanPauseServiceNode{
		Address:  sn.Address,
		Reporter: actor.Address,
	}); err != nil {
		t.Fatal(err)
	}
	sn = GetAllTestingServiceNodes(t, ctx)[0]
	if !sn.Paused || sn.PausedHeight != 1 {
		t.Fatal("service node is not correctly paused after message")
	}
}

func TestUtilityContext_HandleMessageProveTestScore(t *testing.T) {
	// Not Implemented Yet TODO
}

func TestUtilityContext_HandleMessageTestScore(t *testing.T) {
	// Not Implemented Yet TODO
}
