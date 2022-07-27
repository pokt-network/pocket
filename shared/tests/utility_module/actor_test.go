package utility_module

import (
	"bytes"
	"math"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/pre_persistence"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	pubKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	outputAddress, err := crypto.GenerateAddress()
	require.NoError(t, err)

	err = ctx.SetAccountAmount(outputAddress, defaultAmount)
	require.NoError(t, err, "error setting account amount error")

	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: outputAddress,
		Signer:        outputAddress,
	}

	er := ctx.HandleStakeMessage(msg)
	require.NoError(t, er, "handle stake message")

	actor := GetAppByAddress(t, ctx, pubKey.Address().Bytes())

	require.Equal(t, actor.Address, pubKey.Address().Bytes(), "incorrect actor address")
	require.Equal(t, actor.Status, int32(typesUtil.StakedStatus), "incorrect actor  status")
	require.Equal(t, actor.Chains, msg.Chains, "incorrect actor chains")
	require.False(t, actor.Paused, "incorrect actor paused status")
	require.Equal(t, actor.PausedHeight, types.HeightNotUsed, "incorrect actor height")
	require.Equal(t, actor.StakedTokens, defaultAmountString, "incorrect actor stake amount")
	require.Equal(t, actor.UnstakingHeight, types.HeightNotUsed, "incorrect actor unstaking height")
	require.Equal(t, actor.Output, outputAddress.Bytes(), "incorrect actor output address")
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	msg := &typesUtil.MessageEditStake{
		Address:   actor.Address,
		Chains:    defaultTestingChains,
		Amount:    defaultAmountString,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}

	msgChainsEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
	msgChainsEdited.Chains = defaultTestingChainsEdited

	err := ctx.HandleEditStakeMessage(msgChainsEdited)
	require.NoError(t, err, "handle edit stake message")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.False(t, actor.Paused, "incorrect paused status")
	require.Equal(t, actor.PausedHeight, types.HeightNotUsed, "incorrect paused height")
	require.Equal(t, actor.Chains, msgChainsEdited.Chains, "incorrect edited chains")
	require.Equal(t, actor.StakedTokens, defaultAmountString, "incorrect staked tokens")
	require.Equal(t, actor.UnstakingHeight, types.HeightNotUsed, "incorrect unstaking height")

	amountEdited := defaultAmount.Add(defaultAmount, big.NewInt(1))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
	msgAmountEdited.Amount = amountEditedString

	err = ctx.HandleEditStakeMessage(msgAmountEdited)
	require.NoError(t, err, "handle edit stake message")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.Equal(t, actor.StakedTokens, types.BigIntToString(amountEdited), "incorrect staked amount")
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	err := ctx.Context.SetAppMinimumPauseBlocks(0)
	require.NoError(t, err, "error setting minimum pause blocks")

	actor := GetFirstApp(t, ctx)
	err = ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1)
	require.NoError(t, err, "error setting pause height")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.True(t, actor.Paused, "actor should be paused")

	msgUnpauseActor := &typesUtil.MessageUnpause{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}

	err = ctx.HandleUnpauseMessage(msgUnpauseActor)
	require.NoError(t, err, "handle unpause message")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.False(t, actor.Paused, "actor should not be paused")
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	err := ctx.Context.SetAppMinimumPauseBlocks(0)
	require.NoError(t, err, "error setting minimum pause blocks")

	actor := GetFirstApp(t, ctx)
	msg := &typesUtil.MessageUnstake{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}

	err = ctx.HandleUnstakeMessage(msg)
	require.NoError(t, err, "handle unstake message")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.Equal(t, actor.Status, int32(typesUtil.UnstakingStatus), "actor should be unstaking")
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	actor := GetFirstApp(t, ctx)

	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)

	err = ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0)
	require.NoError(t, err, "error setting actor pause height")

	err = ctx.BeginUnstakingMaxPaused()
	require.NoError(t, err, "error beginning unstaking max paused actors")

	status, err := ctx.GetActorStatus(typesUtil.ActorType_App, actor.Address)
	require.Equal(t, status, typesUtil.UnstakingStatus, "actor should be unstaking")
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	actor := GetFirstApp(t, ctx)

	newMaxRelays, err := ctx.CalculateAppRelays(actor.StakedTokens)
	require.NoError(t, err)

	require.Equal(t, actor.MaxRelays, newMaxRelays, "relay calculation incorrect")
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)

	unstakingHeight, err := ctx.GetUnstakingHeight(typesUtil.ActorType_App)
	require.NoError(t, err)

	require.Equal(t, unstakingBlocks, unstakingHeight, "unexpected unstaking height")
}

func TestUtilityContext_Delete(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	err := ctx.DeleteActor(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err, "error deleting actor")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.Nil(t, actor, "actor should be deleted")
}

func TestUtilityContext_GetExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)
	randAddr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	exists, err := ctx.GetActorExists(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, exists, "actor that should exist does not")

	exists, err = ctx.GetActorExists(typesUtil.ActorType_App, randAddr)
	require.NoError(t, err)
	require.False(t, exists, "actor that shouldn't exist does")
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	outputAddress, err := ctx.GetActorOutputAddress(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)

	require.Equal(t, outputAddress, actor.Output, "unexpected output address")
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	pauseHeight := int64(100)
	actor := GetFirstApp(t, ctx)

	err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, pauseHeight)
	require.NoError(t, err, "error setting actor pause height")

	gotPauseHeight, err := ctx.GetPauseHeight(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.Equal(t, pauseHeight, gotPauseHeight, "unable to get pause height from the actor")

	randAddr, er := crypto.GenerateAddress()
	require.NoError(t, er)

	_, err = ctx.GetPauseHeight(typesUtil.ActorType_App, randAddr)
	require.Error(t, err, "non existent actor should error")
}

func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	msgEditStake := &typesUtil.MessageEditStake{
		Address: actor.Address,
		Chains:  defaultTestingChains,
		Amount:  defaultAmountString,
	}

	candidates, err := ctx.GetMessageEditStakeSignerCandidates(msgEditStake)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 2, "unexpected number of candidates")
	require.Equal(t, candidates[0], actor.Output, "incorrect output candidate")
	require.Equal(t, candidates[1], actor.Address, "incorrect addr candidate")
}

func TestUtilityContext_GetMessageStakeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	pubKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	addr := pubKey.Address()
	out, err := crypto.GenerateAddress()
	require.NoError(t, err)

	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
	}

	candidates, err := ctx.GetMessageStakeSignerCandidates(msg)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 2, "unexpected number of candidates")
	require.Equal(t, candidates[0], out.Bytes(), "incorrect output candidate")
	require.Equal(t, candidates[1], addr.Bytes(), "incorrect addr candidate")
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	msg := &typesUtil.MessageUnpause{
		Address: actor.Address,
	}

	candidates, err := ctx.GetMessageUnpauseSignerCandidates(msg)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 2, "unexpected number of candidates")
	require.Equal(t, candidates[0], actor.Output, "incorrect output candidate")
	require.Equal(t, candidates[1], actor.Address, "incorrect addr candidate")
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	actor := GetFirstApp(t, ctx)

	msg := &typesUtil.MessageUnstake{
		Address: actor.Address,
	}
	candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
	require.NoError(t, err)
	require.Equal(t, len(candidates), 2, "unexpected number of candidates")
	require.Equal(t, candidates[0], actor.Output, "incorrect output candidate")
	require.Equal(t, candidates[1], actor.Address, "incorrect addr candidate")
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	actor := GetFirstApp(t, ctx)
	require.Equal(t, actor.Status, int32(typesUtil.StakedStatus), "wrong starting status")

	err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0)
	require.NoError(t, err, "error setting actor pause height")

	er := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, er)
	err = ctx.UnstakeActorPausedBefore(0, typesUtil.ActorType_App)
	require.NoError(t, err, "error unstaking actor pause before")

	err = ctx.UnstakeActorPausedBefore(1, typesUtil.ActorType_App)
	require.NoError(t, err, "error unstaking actor pause before height 1")

	actor = GetAppByAddress(t, ctx, actor.Address)
	require.Equal(t, actor.Status, int32(typesUtil.UnstakingStatus), "status does not equal unstaking")

	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, actor.UnstakingHeight, unstakingBlocks+1, "incorrect unstaking height")
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	ctx.SetPoolAmount(genesis.AppStakePoolName, big.NewInt(math.MaxInt64))
	err := ctx.Context.SetAppUnstakingBlocks(0)
	require.NoError(t, err, "error setting unstaking blocks")

	err = ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err, "error setting max pause blocks")

	actors := GetAllTestingApps(t, ctx)

	for _, actor := range actors {
		require.Equal(t, actor.Status, int32(typesUtil.StakedStatus), "wrong starting staked status")
		err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1)
		require.NoError(t, err, "error setting actor pause height")
	}

	err = ctx.UnstakeActorPausedBefore(2, typesUtil.ActorType_App)
	require.NoError(t, err, "error setting actor pause before")

	err = ctx.UnstakeActorsThatAreReady()
	require.NoError(t, err, "error unstaking actors that are ready")

	require.Zero(t, len(GetAllTestingApps(t, ctx)), "apps still exists after unstake that are ready() call")
}

// Helpers

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []*genesis.App {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetFirstApp(t *testing.T, ctx utility.UtilityContext) *genesis.App {
	return GetAllTestingApps(t, ctx)[0]
}

func GetAppByAddress(t *testing.T, ctx utility.UtilityContext, addr []byte) (actor *genesis.App) {
	actors := GetAllTestingApps(t, ctx)
	for _, a := range actors {
		if bytes.Equal(a.Address, addr) {
			actor = a
			break
		}
	}
	return
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Validator {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
