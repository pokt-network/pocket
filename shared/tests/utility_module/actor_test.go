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

func TestUtilityContext_HandleMessageStakeApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	out, _ := crypto.GenerateAddress()
	if err := ctx.SetAccountAmount(out, defaultAmount); err != nil {
		t.Fatal(err)
	}
	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
		Signer:        out,
	}
	if err := ctx.HandleStakeMessage(msg); err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingApps(t, ctx)
	var actor *genesis.App
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
		t.Fatalf("incorrect stake amount, expected %v, got %v", defaultAmountString, actor.StakedTokens)
	}
	if actor.UnstakingHeight != types.HeightNotUsed {
		t.Fatalf("incorrect unstaking height, expected %v, got %v", types.HeightNotUsed, actor.UnstakingHeight)
	}
	if !bytes.Equal(actor.Output, out) {
		t.Fatalf("incorrect output address, expected %v, got %v", actor.Output, out)
	}
}

func TestUtilityContext_HandleMessageEditStakeApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessageEditStake{
		Address:   actor.Address,
		Chains:    defaultTestingChains,
		Amount:    defaultAmountString,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	msgChainsEdited := msg
	msgChainsEdited.Chains = defaultTestingChainsEdited
	if err := ctx.HandleEditStakeMessage(msgChainsEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
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
	amountEdited := defaultAmount.Add(defaultAmount, big.NewInt(1))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := msg
	msgAmountEdited.Amount = amountEditedString
	if err := ctx.HandleEditStakeMessage(msgAmountEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.StakedTokens != types.BigIntToString(amountEdited) {
		t.Fatalf("incorrect amount status, expected %v, got %v", amountEdited, actor.StakedTokens)
	}
}

func TestUtilityContext_HandleMessageUnpauseApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetAppMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingApps(t, ctx)[0]
	if err := ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, 1); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
	msgU := &typesUtil.MessageUnpause{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	if err := ctx.HandleUnpauseMessage(msgU); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.Paused {
		t.Fatal("actor is paused after")
	}
}

func TestUtilityContext_HandleMessageUnstakeApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetAppMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessageUnstake{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	if err := ctx.HandleUnstakeMessage(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("actor isn't unstaking")
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedApps(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	if err = ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, 0); err != nil {
		t.Fatal(err)
	}
	if err = ctx.BeginUnstakingMaxPaused(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetActorStatus(actor.Address, typesUtil.ActorType_App)
	if status != 1 {
		t.Fatalf("incorrect status; expected %d got %d", 1, actor.Status)
	}
}

func TestUtilityContext_CalculateAppRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newMaxRelays, err := ctx.CalculateAppRelays(actor.StakedTokens)
	require.NoError(t, err)
	if actor.MaxRelays != newMaxRelays {
		t.Fatalf("unexpected max relay calculation; got %v wanted %v", actor.MaxRelays, newMaxRelays)
	}
}

func TestUtilityContext_CalculateAppUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.GetUnstakingHeight(typesUtil.ActorType_App)
	require.NoError(t, err)
	if unstakingBlocks != unstakingHeight {
		t.Fatalf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight)
	}
}

func TestUtilityContext_DeleteApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	if err := ctx.DeleteActor(actor.Address, typesUtil.ActorType_App); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingApps(t, ctx)) > 0 {
		t.Fatal("deletion unsuccessful")
	}
}

func TestUtilityContext_GetAppExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	randAddr, _ := crypto.GenerateAddress()
	actor := GetAllTestingApps(t, ctx)[0]
	exists, err := ctx.GetActorExists(actor.Address, typesUtil.ActorType_App)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = ctx.GetActorExists(randAddr, typesUtil.ActorType_App)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that shouldn't exist does")
	}
}

func TestUtilityContext_GetAppOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	outputAddress, err := ctx.GetActorOutputAddress(actor.Address, typesUtil.ActorType_App)
	require.NoError(t, err)
	if !bytes.Equal(outputAddress, actor.Output) {
		t.Fatalf("unexpected output address, expected %v got %v", actor.Output, outputAddress)
	}
}

func TestUtilityContext_GetAppPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	pauseHeight := int64(100)
	if err := ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, pauseHeight); err != nil {
		t.Fatal(err)
	}
	gotPauseHeight, err := ctx.GetPauseHeight(actor.Address, typesUtil.ActorType_App)
	require.NoError(t, err)
	if pauseHeight != gotPauseHeight {
		t.Fatal("unable to get pause height from the actor")
	}
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetPauseHeight(addr, typesUtil.ActorType_App)
	if err == nil {
		t.Fatal("no error on non-existent actor pause height")
	}
}

func TestUtilityContext_GetMessageEditStakeAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msgEditStake := &typesUtil.MessageEditStake{
		Address: actors[0].Address,
		Chains:  defaultTestingChains,
		Amount:  defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeSignerCandidates(msgEditStake)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageStakeAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	out, _ := crypto.GenerateAddress()
	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
	}
	candidates, err := ctx.GetMessageStakeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnpauseAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnpause{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseSignercandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnstakeAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnstake{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_UnstakeAppsPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	if err := ctx.SetActorPauseHeight(actor.Address, typesUtil.ActorType_App, 0); err != nil {
		t.Fatal(err)
	}
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.UnstakeActorPausedBefore(0, typesUtil.ActorType_App); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeActorPausedBefore(1, typesUtil.ActorType_App); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("status does not equal unstaking")
	}
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	if actor.UnstakingHeight != unstakingBlocks+1 {
		t.Fatal("incorrect unstaking height")
	}
}

func TestUtilityContext_UnstakeAppsThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(genesis.AppStakePoolName, big.NewInt(math.MaxInt64))
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
		t.Fatal("apps still exists after unstake that are ready() call")
	}
}

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []*genesis.App {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Validator {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
