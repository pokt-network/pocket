package utility_module

import (
	"bytes"
	"fmt"
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
	require.NoError(t, err, "set account amount error")

	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: outputAddress,
		Signer:        outputAddress,
	}

	er := ctx.HandleStakeMessage(msg)
	require.NoError(t, er, "handle stake message")

	actors := GetAllTestingApps(t, ctx)
	var actor *genesis.App
	for _, a := range actors {
		if bytes.Equal(a.PublicKey, msg.PublicKey) {
			actor = a
			break
		}
	}

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

	actor := GetAllTestingApps(t, ctx)[0]

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

	actor = GetAllTestingApps(t, ctx)[0]
	require.False(t, actor.Paused, "incorrect paused status")
	require.Equal(t, actor.PausedHeight, types.HeightNotUsed, "incorrect paused height")
	require.Equal(t, actor.Chains, msgChainsEdited.Chains, "incorrect edited chains")
	require.Equal(t, actor.StakedTokens, defaultAmountString, "incorrect staked tokens")
	require.Equal(t, actor.UnstakingHeight, types.HeightNotUsed, "incorrect unstaking height")

	amountEdited := defaultAmount.Add(defaultAmount, big.NewInt(1))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := msg
	msgAmountEdited.Amount = amountEditedString
	err = ctx.HandleEditStakeMessage(msgAmountEdited)
	require.NoError(t, err, "handle edit stake message")

	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.StakedTokens == types.BigIntToString(amountEdited), fmt.Sprintf("incorrect amount status, expected %v, got %v", amountEdited, actor.StakedTokens))
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	err := ctx.Context.SetAppMinimumPauseBlocks(0)
	require.NoError(t, err, "set minimum pause blocks")

	actor := GetAllTestingApps(t, ctx)[0]
	err = ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1)
	require.NoError(t, err, "set pause height")

	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Paused, "actor isn't paused after")
	msgU := &typesUtil.MessageUnpause{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	err = ctx.HandleUnpauseMessage(msgU)
	require.NoError(t, err, "handle unpause message")

	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, !actor.Paused, "actor is paused after")
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	err := ctx.Context.SetAppMinimumPauseBlocks(0)
	require.NoError(t, err, "set min pause blocks")

	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessageUnstake{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	err = ctx.HandleUnstakeMessage(msg)
	require.NoError(t, err, "handle unstake message")

	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.UnstakingStatus, "actor isn't unstaking")
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	err = ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0)
	require.NoError(t, err, "set actor pause height")

	err = ctx.BeginUnstakingMaxPaused()
	require.NoError(t, err, "begin unstaking max paused")

	status, err := ctx.GetActorStatus(typesUtil.ActorType_App, actor.Address)
	require.True(t, status == 1, fmt.Sprintf("incorrect status; expected %d got %d", 1, actor.Status))
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newMaxRelays, err := ctx.CalculateAppRelays(actor.StakedTokens)
	require.NoError(t, err)
	require.True(t, actor.MaxRelays == newMaxRelays, fmt.Sprintf("unexpected max relay calculation; got %v wanted %v", actor.MaxRelays, newMaxRelays))
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.GetUnstakingHeight(typesUtil.ActorType_App)
	require.NoError(t, err)
	require.True(t, unstakingBlocks == unstakingHeight, fmt.Sprintf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight))
}

func TestUtilityContext_Delete(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	err := ctx.DeleteActor(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err, "delete actor")

	require.False(t, len(GetAllTestingApps(t, ctx)) > 0, fmt.Sprintf("deletion unsuccessful"))
}

func TestUtilityContext_GetExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	randAddr, _ := crypto.GenerateAddress()
	actor := GetAllTestingApps(t, ctx)[0]
	exists, err := ctx.GetActorExists(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, exists, fmt.Sprintf("actor that should exist does not"))
	exists, err = ctx.GetActorExists(typesUtil.ActorType_App, randAddr)
	require.NoError(t, err)
	require.True(t, !exists, fmt.Sprintf("actor that shouldn't exist does"))
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	outputAddress, err := ctx.GetActorOutputAddress(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, bytes.Equal(outputAddress, actor.Output), fmt.Sprintf("unexpected output address, expected %v got %v", actor.Output, outputAddress))
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	pauseHeight := int64(100)
	err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, pauseHeight)
	require.NoError(t, err, "set actor pause height")

	gotPauseHeight, err := ctx.GetPauseHeight(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, pauseHeight == gotPauseHeight, fmt.Sprintf("unable to get pause height from the actor"))
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetPauseHeight(typesUtil.ActorType_App, addr)
	require.Error(t, err, "no error on non-existent actor pause height")
}

func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msgEditStake := &typesUtil.MessageEditStake{
		Address: actors[0].Address,
		Chains:  defaultTestingChains,
		Amount:  defaultAmountString,
	}
	candidates, err := ctx.GetMessageEditStakeSignerCandidates(msgEditStake)
	require.NoError(t, err)
	require.False(t, !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address), "incorrect signer candidates")
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
	require.False(t, !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr), "incorrect signer candidates")
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnpause{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnpauseSignercandidates(msg)
	require.NoError(t, err)
	require.False(t, !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address), "incorrect signer candidates")
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actors := GetAllTestingApps(t, ctx)
	msg := &typesUtil.MessageUnstake{
		Address: actors[0].Address,
	}
	candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
	require.NoError(t, err)
	require.False(t, !bytes.Equal(candidates[0], actors[0].Output) || !bytes.Equal(candidates[1], actors[0].Address), "incorrect signer candidates")
}

func TestUtilityContext_UnstakesPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.StakedStatus, fmt.Sprintf("wrong starting status"))
	err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0)
	require.NoError(t, err, "set actor pause height")

	er := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, er)
	err = ctx.UnstakeActorPausedBefore(0, typesUtil.ActorType_App)
	require.NoError(t, err, "unstake actor pause before")

	err = ctx.UnstakeActorPausedBefore(1, typesUtil.ActorType_App)
	require.NoError(t, err, "unstake actor pause before height 1")

	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.UnstakingStatus, fmt.Sprintf("status does not equal unstaking"))
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.True(t, actor.UnstakingHeight == unstakingBlocks+1, fmt.Sprintf("incorrect unstaking height"))
}

func TestUtilityContext_UnstakesThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	ctx.SetPoolAmount(genesis.AppStakePoolName, big.NewInt(math.MaxInt64))
	err := ctx.Context.SetAppUnstakingBlocks(0)
	require.NoError(t, err, "set unstaking blocks")

	err = ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err, "set max pause blocks")

	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		require.True(t, actor.Status == typesUtil.StakedStatus, fmt.Sprintf("wrong starting status"))
		err := ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1)
		require.NoError(t, err, "set actor pause height")

	}
	err = ctx.UnstakeActorPausedBefore(2, typesUtil.ActorType_App)
	require.NoError(t, err, "set actor pause before")

	err = ctx.UnstakeActorsThatAreReady()
	require.NoError(t, err, "unstake actors that are ready")

	require.True(t, len(GetAllTestingApps(t, ctx)) == 0, fmt.Sprintf("apps still exists after unstake that are ready() call"))
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
