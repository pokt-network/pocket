package utility_module

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/persistence"
	"github.com/stretchr/testify/require"
	"math"
	"math/big"
	"reflect"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
)

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	pubKey, _ := crypto.GeneratePublicKey()
	out, _ := crypto.GenerateAddress()
	require.NoError(t, ctx.SetAccountAmount(out, defaultAmount), "set account amount")
	msg := &typesUtil.MessageStake{
		PublicKey:     pubKey.Bytes(),
		Chains:        defaultTestingChains,
		Amount:        defaultAmountString,
		OutputAddress: out,
		Signer:        out,
	}
	require.NoError(t, ctx.HandleStakeMessage(msg), "handle stake message")
	actors := GetAllTestingApps(t, ctx)
	var actor *genesis.App
	for _, a := range actors {
		if bytes.Equal(a.PublicKey, msg.PublicKey) {
			actor = a
			break
		}
	}
	require.True(t, bytes.Equal(actor.Address, pubKey.Address()), fmt.Sprintf("incorrect address, expected %v, got %v", pubKey.Address(), actor.Address))
	require.True(t, actor.Status == typesUtil.StakedStatus, fmt.Sprintf("incorrect status, expected %v, got %v", typesUtil.StakedStatus, actor.Status))
	require.Equal(t, actor.Chains, msg.Chains, fmt.Sprintf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains))
	require.False(t, actor.Paused, fmt.Sprintf("incorrect paused status, expected %v, got %v", false, actor.Paused))
	require.True(t, actor.PausedHeight == types.HeightNotUsed, fmt.Sprintf("incorrect paused height, expected %v, got %v", types.HeightNotUsed, actor.PausedHeight))
	require.True(t, actor.StakedTokens == defaultAmountString, fmt.Sprintf("incorrect stake amount, expected %v, got %v", defaultAmountString, actor.StakedTokens))
	require.True(t, actor.UnstakingHeight == types.HeightNotUsed, fmt.Sprintf("incorrect unstaking height, expected %v, got %v", types.HeightNotUsed, actor.UnstakingHeight))
	require.True(t, bytes.Equal(actor.Output, out), fmt.Sprintf("incorrect output address, expected %v, got %v", actor.Output, out))
	ctx.Context.Release()
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
	msgChainsEdited := msg
	msgChainsEdited.Chains = defaultTestingChainsEdited
	require.NoError(t, ctx.HandleEditStakeMessage(msgChainsEdited), "handle edit stake message")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, reflect.DeepEqual(actor.Chains, msg.Chains), fmt.Sprintf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains))
	require.True(t, actor.Paused == false, fmt.Sprintf("incorrect paused status, expected %v, got %v", false, actor.Paused))
	require.True(t, actor.PausedHeight == types.HeightNotUsed, fmt.Sprintf("incorrect paused height, expected %v, got %v", types.HeightNotUsed, actor.PausedHeight))
	require.True(t, reflect.DeepEqual(actor.Chains, msgChainsEdited.Chains), fmt.Sprintf("incorrect chains, expected %v, got %v", msg.Chains, actor.Chains))
	require.True(t, actor.StakedTokens == defaultAmountString, fmt.Sprintf("incorrect staked tokens, expected %v, got %v", defaultAmountString, actor.StakedTokens))
	require.True(t, actor.UnstakingHeight == types.HeightNotUsed, fmt.Sprintf("incorrect unstaking height, expected %v, got %v", types.HeightNotUsed, actor.UnstakingHeight))
	amountEdited := defaultAmount.Add(defaultAmount, big.NewInt(1))
	amountEditedString := types.BigIntToString(amountEdited)
	msgAmountEdited := msg
	msgAmountEdited.Amount = amountEditedString
	require.NoError(t, ctx.HandleEditStakeMessage(msgAmountEdited), "handle edit stake message")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.StakedTokens == types.BigIntToString(amountEdited), fmt.Sprintf("incorrect amount status, expected %v, got %v", amountEdited, actor.StakedTokens))
	ctx.Context.Release()
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	require.NoError(t, ctx.Context.SetAppMinimumPauseBlocks(0), "set minimum pause blocks")
	actor := GetAllTestingApps(t, ctx)[0]
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1), "set pause height")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Paused, fmt.Sprintf("actor isn't paused after"))
	msgU := &typesUtil.MessageUnpause{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	require.NoError(t, ctx.HandleUnpauseMessage(msgU), "handle unpause message")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, !actor.Paused, fmt.Sprintf("actor is paused after"))
	ctx.Context.Release()
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	require.NoError(t, ctx.Context.SetAppMinimumPauseBlocks(0), "set min pause blocks")
	actor := GetAllTestingApps(t, ctx)[0]
	msg := &typesUtil.MessageUnstake{
		Address:   actor.Address,
		Signer:    actor.Address,
		ActorType: typesUtil.ActorType_App,
	}
	require.NoError(t, ctx.HandleUnstakeMessage(msg), "handle unstake message")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.UnstakingStatus, "actor isn't unstaking")
	ctx.Context.Release()
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0), "set actor pause height")
	require.NoError(t, ctx.BeginUnstakingMaxPaused(), "begin unstaking max paused")
	status, err := ctx.GetActorStatus(typesUtil.ActorType_App, actor.Address)
	require.True(t, status == 1, fmt.Sprintf("incorrect status; expected %d got %d", 1, actor.Status))
	ctx.Context.Release()
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newMaxRelays, err := ctx.CalculateAppRelays(actor.StakedTokens)
	require.NoError(t, err)
	require.True(t, actor.MaxRelays == newMaxRelays, fmt.Sprintf("unexpected max relay calculation; got %v wanted %v", actor.MaxRelays, newMaxRelays))
	ctx.Context.Release()
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	unstakingHeight, err := ctx.GetUnstakingHeight(typesUtil.ActorType_App)
	require.NoError(t, err)
	require.True(t, unstakingBlocks == unstakingHeight, fmt.Sprintf("unexpected unstakingHeight; got %d expected %d", unstakingBlocks, unstakingHeight))
	ctx.Context.Release()
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
	ctx.Context.Release()
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	outputAddress, err := ctx.GetActorOutputAddress(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, bytes.Equal(outputAddress, actor.Output), fmt.Sprintf("unexpected output address, expected %v got %v", actor.Output, outputAddress))
	ctx.Context.Release()
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	actor := GetAllTestingApps(t, ctx)[0]
	pauseHeight := int64(100)
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, pauseHeight), "set actor pause height")
	gotPauseHeight, err := ctx.GetPauseHeight(typesUtil.ActorType_App, actor.Address)
	require.NoError(t, err)
	require.True(t, pauseHeight == gotPauseHeight, fmt.Sprintf("unable to get pause height from the actor"))
	addr, _ := crypto.GenerateAddress()
	_, err = ctx.GetPauseHeight(typesUtil.ActorType_App, addr)
	require.Error(t, err, "no error on non-existent actor pause height")
	ctx.Context.Release()
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
	ctx.Context.Release()
}

func TestUtilityContext_GetMessageStakeSignerCandidates(t *testing.T) {
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
	require.False(t, !bytes.Equal(candidates[0], out) || !bytes.Equal(candidates[1], addr), "incorrect signer candidates")
	ctx.Context.Release()
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
	ctx.Context.Release()
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
	ctx.Context.Release()
}

func TestUtilityContext_UnstakesPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.StakedStatus, fmt.Sprintf("wrong starting status"))
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 0), "set actor pause height")
	err := ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	require.NoError(t, ctx.UnstakeActorPausedBefore(0, typesUtil.ActorType_App), "unstake actor pause before")
	require.NoError(t, ctx.UnstakeActorPausedBefore(1, typesUtil.ActorType_App), "unstake actor pause before height 1")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.Status == typesUtil.UnstakingStatus, fmt.Sprintf("status does not equal unstaking"))
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.True(t, actor.UnstakingHeight == unstakingBlocks+1, fmt.Sprintf("incorrect unstaking height"))
	ctx.Context.Release()
}

func TestUtilityContext_UnstakesThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount(genesis.AppStakePoolName, big.NewInt(math.MaxInt64))
	require.NoError(t, ctx.Context.SetAppUnstakingBlocks(0), "set unstaking blocks")
	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		require.True(t, actor.Status == typesUtil.StakedStatus, fmt.Sprintf("wrong starting status"))
		require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, actor.Address, 1), "set actor pause height")
	}
	require.NoError(t, ctx.UnstakeActorPausedBefore(2, typesUtil.ActorType_App), "set actor pause before")
	require.NoError(t, ctx.UnstakeActorsThatAreReady(), "unstake actors that are ready")
	appAfter := GetAllTestingApps(t, ctx)[0]
	require.True(t, appAfter.UnstakingHeight == 0, fmt.Sprintf("apps still exists after unstake that are ready() call"))
	// TODO (Team) we need to better define what 'deleted' really is in the postgres world.
	// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight
	ctx.Context.Release()
}

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []*genesis.App {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllApps(ctx.LatestHeight)
	sort.Slice(actors, func(i, j int) bool {
		return hex.EncodeToString(actors[i].Address) < hex.EncodeToString(actors[j].Address)
	})
	require.NoError(t, err)
	return actors
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Validator {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllValidators(ctx.LatestHeight)
	sort.Slice(actors, func(i, j int) bool {
		return hex.EncodeToString(actors[i].Address) < hex.EncodeToString(actors[j].Address)
	})
	require.NoError(t, err)
	return actors
}
