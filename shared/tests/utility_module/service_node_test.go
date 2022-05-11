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

func TestUtilityContext_HandleMessageStakeServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	out, _ := crypto.GenerateAddress()
	if err := ctx.SetAccountAmount(out, defaultAmount); err != nil {
		t.Fatal(err)
	}
	msg := &typesUtil.MessageStakeServiceNode{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
		Signer:        out,
	}
	if err := ctx.HandleMessageStakeServiceNode(msg); err != nil {
		t.Fatal(err)
	}
	actors := GetAllTestingServiceNodes(t, ctx)
	var actor *genesis.ServiceNode
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

func TestUtilityContext_GetServiceNodeCount(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	count, err := ctx.GetServiceNodeCount(defaultTestingChains[0], 0)
	require.NoError(t, err)
	if count != len(actors) {
		t.Fatalf("wrong chain count, expected %d, got %d", len(actors), count)
	}
}

func TestUtilityContext_GetServiceNodesPerSession(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	count, err := ctx.GetServiceNodesPerSession(0)
	require.NoError(t, err)
	if count != defaultServiceNodesPerSession {
		t.Fatalf("incorrect service node per session, expected %d got %d", defaultServiceNodesPerSession, count)
	}
}

func TestUtilityContext_HandleMessageEditStakeServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	msg := &typesUtil.MessageEditStakeServiceNode{
		Address:     actor.Address,
		Chains:      defaultTestingChains,
		AmountToAdd: zeroAmountString,
		Signer:      actor.Address,
	}
	msgChainsEdited := msg
	msgChainsEdited.Chains = defaultTestingChainsEdited
	if err := ctx.HandleMessageEditStakeServiceNode(msgChainsEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
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
	if err := ctx.HandleMessageEditStakeServiceNode(msgAmountEdited); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("incorrect amount status, expected %v, got %v", expectedAmount, actor.StakedTokens)
	}
}

func TestUtilityContext_HandleMessagePauseServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	msg := &typesUtil.MessagePauseServiceNode{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseServiceNode(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
}

func TestUtilityContext_HandleMessageUnpauseServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetServiceNodeMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	msg := &typesUtil.MessagePauseServiceNode{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessagePauseServiceNode(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if !actor.Paused {
		t.Fatal("actor isn't paused after")
	}
	msgU := &typesUtil.MessageUnpauseServiceNode{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnpauseServiceNode(msgU); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if actor.Paused {
		t.Fatal("actor is paused after")
	}
}

func TestUtilityContext_HandleMessageUnstakeServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	if err := ctx.Context.SetServiceNodeMinimumPauseBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	msg := &typesUtil.MessageUnstakeServiceNode{
		Address: actor.Address,
		Signer:  actor.Address,
	}
	if err := ctx.HandleMessageUnstakeServiceNode(msg); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("actor isn't unstaking")
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedServiceNodes(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	err := ctx.Context.SetServiceNodeMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetServiceNodePauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.BeginUnstakingMaxPausedServiceNodes(); err != nil {
		t.Fatal(err)
	}
	status, err := ctx.GetServiceNodeStatus(actor.Address)
	if status != 1 {
		t.Fatalf("incorrect status; expected %d got %d", 1, actor.Status)
	}
}

func TestUtilityContext_CalculateServiceNodeUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetServiceNodeUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.CalculateServiceNodeUnstakingHeight()
	require.NoError(t, err)
	if unstakingBlocks != unstakingHeight {
		t.Fatalf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight)
	}
}

func TestUtilityContext_DeleteServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	actor := actors[0]
	if err := ctx.DeleteServiceNode(actor.Address); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingServiceNodes(t, ctx)) != len(actors)-1 {
		t.Fatal("deletion unsuccessful")
	}
}

func TestUtilityContext_GetServiceNodeExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	randAddr, _ := crypto.GenerateAddress()
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	exists, err := ctx.GetServiceNodeExists(actor.Address)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor that should exist does not")
	}
	exists, err = ctx.GetServiceNodeExists(randAddr)
	require.NoError(t, err)
	if exists {
		t.Fatal("actor that shouldn't exist does")
	}
}

func TestUtilityContext_GetServiceNodeOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	outputAddress, err := ctx.GetServiceNodeOutputAddress(actor.Address)
	require.NoError(t, err)
	if !bytes.Equal(outputAddress, actor.Output) {
		t.Fatalf("unexpected output address, expected %v got %v", actor.Output, outputAddress)
	}
}

func TestUtilityContext_GetServiceNodePauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	pauseHeight := int64(100)
	if err := ctx.SetServiceNodePauseHeight(actor.Address, pauseHeight); err != nil {
		t.Fatal(err)
	}
	gotPauseHeight, err := ctx.GetServiceNodePauseHeightIfExists(actor.Address)
	require.NoError(t, err)
	if pauseHeight != gotPauseHeight {
		t.Fatal("unable to get pause height from the actor")
	}
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetServiceNodePauseHeightIfExists(addr)
	if err == nil {
		t.Fatal("no error on non-existent actor pause height")
	}
}

func TestUtilityContext_GetServiceNodesReadyToUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	if err := ctx.SetServiceNodeUnstakingHeightAndStatus(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	actors, err := ctx.GetServiceNodesReadyToUnstake()
	require.NoError(t, err)
	if !bytes.Equal(actors[0].Address, actor.Address) {
		t.Fatalf("unexpected actor ready to unstake: expected %s, got %s", actor.Address, actors[0].Address)
	}
}

func TestUtilityContext_GetMessageEditStakeServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	msgEditStake := &typesUtil.MessageEditStakeServiceNode{
		Address:     actors[0].Address,
		Chains:      defaultTestingChains,
		AmountToAdd: defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeServiceNodeSignerCandidates(msgEditStake)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessagePauseServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	msg := &typesUtil.MessagePauseServiceNode{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessagePauseServiceNodeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageStakeServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	out, _ := crypto.GenerateAddress()
	msg := &typesUtil.MessageStakeServiceNode{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
	}
	candidates, err := ctx.GetMessageStakeServiceNodeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnpauseServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	msg := &typesUtil.MessageUnpauseServiceNode{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseServiceNodeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_GetMessageUnstakeServiceNodeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingServiceNodes(t, ctx)
	msg := &typesUtil.MessageUnstakeServiceNode{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeServiceNodeSignerCandidates(msg)
	require.NoError(t, err)
	if !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address) {
		t.Fatal(err)
	}
}

func TestUtilityContext_InsertServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	addr := pubKey.Address()
	if err := ctx.InsertServiceNode(addr, pubKey.Bytes(), addr, defaultServiceUrl, defaultAmountString, defaultTestingChains); err != nil {
		t.Fatal(err)
	}
	exists, err := ctx.GetServiceNodeExists(addr)
	require.NoError(t, err)
	if !exists {
		t.Fatal("actor does not exist after insert")
	}
	actors := GetAllTestingServiceNodes(t, ctx)
	for _, actor := range actors {
		if bytes.Equal(actor.Address, addr) {
			if actor.Chains[0] != defaultTestingChains[0] {
				t.Fatal("wrong chains")
			}
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

func TestUtilityContext_UnstakeServiceNodesPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetServiceNodeMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetServiceNodePauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeServiceNodesPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if actor.Status != typesUtil.UnstakingStatus {
		t.Fatal("status does not equal unstaking")
	}
	unstakingBlocks, err := ctx.GetServiceNodeUnstakingBlocks()
	require.NoError(t, err)
	if actor.UnstakingHeight != unstakingBlocks+1 {
		t.Fatal("incorrect unstaking height")
	}
}

func TestUtilityContext_UnstakeServiceNodesThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(typesUtil.ServiceNodeStakePoolName, big.NewInt(math.MaxInt64))
	if err := ctx.Context.SetServiceNodeUnstakingBlocks(0); err != nil {
		t.Fatal(err)
	}
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	if actor.Status != typesUtil.StakedStatus {
		t.Fatal("wrong starting status")
	}
	err := ctx.Context.SetServiceNodeMaxPausedBlocks(0)
	require.NoError(t, err)
	if err := ctx.SetServiceNodePauseHeight(actor.Address, 0); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeServiceNodesPausedBefore(1); err != nil {
		t.Fatal(err)
	}
	if err := ctx.UnstakeServiceNodesThatAreReady(); err != nil {
		t.Fatal(err)
	}
	if len(GetAllTestingServiceNodes(t, ctx)) != 0 {
		t.Fatal("actor still exists after unstake that are ready() call")
	}
}

func TestUtilityContext_UpdateServiceNode(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingServiceNodes(t, ctx)[0]
	newAmountBig := big.NewInt(9999999999999999)
	newAmount := types.BigIntToString(newAmountBig)
	oldAmount := actor.StakedTokens
	oldAmountBig, err := types.StringToBigInt(oldAmount)
	require.NoError(t, err)
	expectedAmountBig := newAmountBig.Add(newAmountBig, oldAmountBig)
	expectedAmount := types.BigIntToString(expectedAmountBig)
	if err := ctx.UpdateServiceNode(actor.Address, actor.ServiceUrl, newAmount, actor.Chains); err != nil {
		t.Fatal(err)
	}
	actor = GetAllTestingServiceNodes(t, ctx)[0]
	if actor.StakedTokens != expectedAmount {
		t.Fatalf("updated amount is incorrect; expected %s got %s", expectedAmount, actor.StakedTokens)
	}
}

func GetAllTestingServiceNodes(t *testing.T, ctx utility.UtilityContext) []*genesis.ServiceNode {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllServiceNodes(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
