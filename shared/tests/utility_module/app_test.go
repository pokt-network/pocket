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
	msg := &typesUtil.MessageStakeApp{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
		Signer:        out,
	}
	if err := ctx.HandleMessageStakeApp(msg); err != nil {
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
	if actor.PausedHeight != 0 {
		t.Fatalf("incorrect paused status, expected %v, got %v", actor.PausedHeight, 0)
	}
	if actor.StakedTokens != defaultAmountString {
		t.Fatalf("incorrect paused status, expected %v, got %v", actor.StakedTokens, defaultAmountString)
	}
	if actor.UnstakingHeight != 0 {
		t.Fatalf("incorrect unstaking height, expected %v, got %v", 0, actor.UnstakingHeight)
	}
	if !bytes.Equal(actor.Output, out) {
		t.Fatalf("incorrect output address, expected %v, got %v", actor.Output, out)
	}
}

func TestUtilityContext_HandleMessageEditStakeApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessageEditStakeApp{
		Address:     actor.Address,
		Chains:      defaultTestingChains,
		AmountToAdd: zeroAmountString,
		Signer:      actor.Address,
	}
	msgChainsEdited := msg
	msgChainsEdited.Chains = defaultTestingChainsEdited
	if err := ctx.HandleMessageEditStakeApp(msgChainsEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if !reflect.DeepEqual(actor.Chains, msg.Chains) {
		t.Fatalf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains)
	}
	if actor.Paused != false {
		t.Fatalf("incorrect paused status, expected %v, got %v", false, actor.Paused)
	}
	if actor.PausedHeight != 0 {
		t.Fatalf("incorrect paused status, expected %v, got %v", actor.PausedHeight, 0)
	}
	if !reflect.DeepEqual(actor.Chains, msgChainsEdited.Chains) {
		t.Fatalf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains)
	}
	if actor.StakedTokens != defaultAmountString {
		t.Fatalf("incorrect staked tokens, expected %v, got %v", defaultAmountString, actor.StakedTokens)
	}
	if actor.UnstakingHeight != 0 {
		t.Fatalf("incorrect unstaking height, expected %v, got %v", 0, actor.UnstakingHeight)
	}
	if !bytes.Equal(actor.Output, actor.Output) {
		t.Fatalf("incorrect output address, expected %v, got %v", actor.Output, actor.Output)
	}
	amountEdited := big.NewInt(1)
	expectedAmount := types.BigIntToString(big.NewInt(0).Add(defaultAmount, amountEdited))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := msg
	msgAmountEdited.AmountToAdd = amountEditedString
	if err := ctx.HandleMessageEditStakeApp(msgAmountEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("incorrect amount status, expected %v, got %v", expectedAmount, actor.StakedTokens)
	}
}

func TestUtilityContext_HandleMessagePauseApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessagePauseApp{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseApp(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
}

func TestUtilityContext_HandleMessageUnpauseApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetAppMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessagePauseApp{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseApp(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
	msgU := &typesUtil.MessageUnpauseApp{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnpauseApp(msgU); err != nil {
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
	msg := &typesUtil.MessageUnstakeApp{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnstakeApp(msg); err != nil {
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
	if err := ctx.SetAppPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPausedApps(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetAppStatus(actor.Address)
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
	unstakingHeight, err := ctx.CalculateAppUnstakingHeight()
	require.NoError(t, err)
	if unstakingBlocks != unstakingHeight {
		t.Fatalf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight)
	}
}

func TestUtilityContext_DeleteApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	if err := ctx.DeleteApp(actor.Address); err != nil {
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
	exists, err := ctx.GetAppExists(actor.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = ctx.GetAppExists(randAddr)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that shouldn't exist does")
	}
}

func TestUtilityContext_GetAppOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	outputAddress, err := ctx.GetAppOutputAddress(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(outputAddress, actor.Output) {
		t.Fatalf("unexpected output address, expected %v got %v", actor.Output, outputAddress)
	}
}

func TestUtilityContext_GetAppPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	pauseHeight := int64(100)
	if err := ctx.SetAppPauseHeight(actor.Address, pauseHeight); err != nil {
		t.Fatal(err)
	}
	gotPauseHeight, err := ctx.GetAppPauseHeightIfExists(actor.Address)
	require.NoError(t, err)
	if pauseHeight != gotPauseHeight {
		t.Fatal("unable to get pause height from the actor")
	}
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetAppPauseHeightIfExists(addr)
	if err == nil {
		t.Fatal("no error on non-existent actor pause height")
	}
}

func TestUtilityContext_GetAppsReadyToUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	if err := ctx.SetAppUnstakingHeightAndStatus(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	actors, err := ctx.GetAppsReadyToUnstake()
	require.NoError(t, err)
	if !bytes.Equal(actors[0].Address, actor.Address) {
		t.Fatalf("unexpected actor ready to unstake: expected %s, got %s", actor.Address, actors[0].Address)
	}
}

func TestUtilityContext_GetMessageEditStakeAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msgEditStake := &typesUtil.MessageEditStakeApp{
		Address:     actors[0].Address,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeAppSignerCandidates(msgEditStake)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessagePauseAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessagePauseApp{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessagePauseAppSignerCandidates(msg)
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
	msg := &typesUtil.MessageStakeApp{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
	}
	candidates, err := ctx.GetMessageStakeAppSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnpauseAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnpauseApp{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseAppSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnstakeAppSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnstakeApp{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeAppSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_InsertApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	if err := ctx.InsertApp(addr, pubKey.Bytes(), addr, defaultAmountString, defaultAmountString, defaultTestingChains); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetAppExists(addr)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor does not exist after insert")
	}
	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		if bytes.Equal(actor.Address, addr) {
			if actor.Chains[0] != defaultTestingChains[0] {
				t.Fatal("wrong chains")
			}
			if actor.StakedTokens != defaultAmountString {
				t.Fatal("wrong staked tokens")
			}
			if actor.MaxRelays != defaultAmountString {
				t.Fatal("wrong max relays")
			}
			if !bytes.Equal(actor.Output, addr) {
				t.Fatal("wrong output addr")
			}
			return
		}
	}
	t.Fatal("actor not found after insert in GetAll() call")
}

func TestUtilityContext_UnstakeAppsPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetAppPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeAppsPausedBefore(1); err != nil {
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
	ctx.SetPoolAmount(typesUtil.AppStakePoolName, big.NewInt(math.MaxInt64))
	if err := ctx.Context.SetAppUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingApps(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetAppPauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeAppsPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeAppsThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingApps(t, ctx)) != 0 {
		t.Fatal("actor still exists after unstake that are ready() call")
	}
}

func TestUtilityContext_UpdateApp(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newAmountBig := big.NewInt(9999999999999999)
	newAmount := types.BigIntToString(newAmountBig)
	oldAmount := actor.StakedTokens
	oldAmountBig, err := types.StringToBigInt(oldAmount)
	require.NoError(t, err)
	expectedAmountBig := newAmountBig.Add(newAmountBig, oldAmountBig)
	expectedAmount := types.BigIntToString(expectedAmountBig)
	if err := ctx.UpdateApp(actor.Address, actor.MaxRelays, newAmount, actor.Chains); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingApps(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("updated amount is incorrect; expected %s got %s", expectedAmount, actor.StakedTokens)
	}
}

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []*genesis.App {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
