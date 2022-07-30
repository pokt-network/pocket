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

// INVESTIGATE: Is there a better way to implement this than to simply have an actors forloop in each test?

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageStake", actorType.GetActorName()), func(t *testing.T) {
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
				ServiceUrl:    "https://localhost.com",
				OutputAddress: outputAddress,
				Signer:        outputAddress,
				ActorType:     actorType,
			}

			er := ctx.HandleStakeMessage(msg)
			require.NoError(t, er, "handle stake message")

			actor := GetActorByAddr(t, ctx, pubKey.Address().Bytes(), actorType)

			require.Equal(t, actor.GetAddress(), pubKey.Address().Bytes(), "incorrect actor address")
			require.Equal(t, actor.GetStatus(), int32(typesUtil.StakedStatus), "incorrect actor  status")
			if actorType != typesUtil.ActorType_Val {
				require.Equal(t, actor.GetChains(), msg.Chains, "incorrect actor chains")
			}
			require.False(t, actor.GetPaused(), "incorrect actor paused status")
			require.Equal(t, actor.GetPausedHeight(), types.HeightNotUsed, "incorrect actor height")
			require.Equal(t, actor.GetStakedTokens(), defaultAmountString, "incorrect actor stake amount")
			require.Equal(t, actor.GetUnstakingHeight(), types.HeightNotUsed, "incorrect actor unstaking height")
			require.Equal(t, actor.GetOutput(), outputAddress.Bytes(), "incorrect actor output address")
		})
	}
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageEditStake", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := GetFirstActor(t, ctx, actorType)
			msg := &typesUtil.MessageEditStake{
				Address:   actor.GetAddress(),
				Chains:    defaultTestingChains,
				Amount:    defaultAmountString,
				Signer:    actor.GetAddress(),
				ActorType: actorType,
			}

			msgChainsEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgChainsEdited.Chains = defaultTestingChainsEdited

			err := ctx.HandleEditStakeMessage(msgChainsEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.False(t, actor.GetPaused(), "incorrect paused status")
			require.Equal(t, actor.GetPausedHeight(), types.HeightNotUsed, "incorrect paused height")
			if actorType != typesUtil.ActorType_Val {
				require.Equal(t, actor.GetChains(), msgChainsEdited.Chains, "incorrect edited chains")
			}
			require.Equal(t, actor.GetStakedTokens(), defaultAmountString, "incorrect staked tokens")
			require.Equal(t, actor.GetUnstakingHeight(), types.HeightNotUsed, "incorrect unstaking height")

			amountEdited := defaultAmount.Add(defaultAmount, big.NewInt(1))
			amountEditedString := types.BigIntToString(amountEdited)
			msgAmountEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgAmountEdited.Amount = amountEditedString

			err = ctx.HandleEditStakeMessage(msgAmountEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.Equal(t, actor.GetStakedTokens(), types.BigIntToString(amountEdited), "incorrect staked amount")
		})
	}
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnpause", actorType.GetActorName()), func(t *testing.T) {

			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.ActorType_Val:
				err = ctx.Context.SetValidatorMinimumPauseBlocks(0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetServiceNodeMinimumPauseBlocks(0)
			case typesUtil.ActorType_App:
				err = ctx.Context.SetAppMinimumPauseBlocks(0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetFishermanMinimumPauseBlocks(0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := GetFirstActor(t, ctx, actorType)
			err = ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 1)
			require.NoError(t, err, "error setting pause height")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.True(t, actor.GetPaused(), "actor should be paused")

			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   actor.GetAddress(),
				Signer:    actor.GetAddress(),
				ActorType: actorType,
			}

			err = ctx.HandleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err, "handle unpause message")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.False(t, actor.GetPaused(), "actor should not be paused")
		})
	}
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnstake", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.ActorType_App:
				err = ctx.Context.SetAppMinimumPauseBlocks(0)
			case typesUtil.ActorType_Val:
				err = ctx.Context.SetValidatorMinimumPauseBlocks(0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetFishermanMinimumPauseBlocks(0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetServiceNodeMinimumPauseBlocks(0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := GetFirstActor(t, ctx, actorType)
			msg := &typesUtil.MessageUnstake{
				Address:   actor.GetAddress(),
				Signer:    actor.GetAddress(),
				ActorType: actorType,
			}

			err = ctx.HandleUnstakeMessage(msg)
			require.NoError(t, err, "handle unstake message")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.Equal(t, actor.GetStatus(), int32(typesUtil.UnstakingStatus), "actor should be unstaking")
		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPaused", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := GetFirstActor(t, ctx, actorType)
			var err error
			switch actorType {
			case typesUtil.ActorType_App:
				err = ctx.Context.SetAppMaxPausedBlocks(0)
			case typesUtil.ActorType_Val:
				err = ctx.Context.SetValidatorMaxPausedBlocks(0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetFishermanMaxPausedBlocks(0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetServiceNodeMaxPausedBlocks(0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 0)
			require.NoError(t, err, "error setting actor pause height")

			err = ctx.BeginUnstakingMaxPaused()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			status, err := ctx.GetActorStatus(actorType, actor.GetAddress())
			require.Equal(t, status, typesUtil.UnstakingStatus, "actor should be unstaking")
		})
	}
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)

	actor := GetFirstActor(t, ctx, typesUtil.ActorType_App)

	newMaxRelays, err := ctx.CalculateAppRelays(actor.GetStakedTokens())
	require.NoError(t, err)

	require.Equal(t, actor.GetGenericParam(), newMaxRelays, "relay calculation incorrect")
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.CalculateUnstakingHeight", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			var unstakingBlocks int64
			var err error
			switch actorType {
			case typesUtil.ActorType_Val:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.ActorType_Node:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.ActorType_Fish:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error getting unstaking blocks")

			unstakingHeight, err := ctx.GetUnstakingHeight(actorType)
			require.NoError(t, err)

			require.Equal(t, unstakingBlocks, unstakingHeight, "unexpected unstaking height")
		})
	}
}

func TestUtilityContext_Delete(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.Delete", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)

			err := ctx.DeleteActor(actorType, actor.GetAddress())
			require.NoError(t, err, "error deleting actor")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.Nil(t, actor, "actor should be deleted")
		})
	}
}

func TestUtilityContext_GetExists(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetExists", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
			randAddr, err := crypto.GenerateAddress()
			require.NoError(t, err)

			exists, err := ctx.GetActorExists(actorType, actor.GetAddress())
			require.NoError(t, err)
			require.True(t, exists, "actor that should exist does not")

			exists, err = ctx.GetActorExists(actorType, randAddr)
			require.NoError(t, err)
			require.False(t, exists, "actor that shouldn't exist does")
		})
	}
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetOutputAddress", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)

			outputAddress, err := ctx.GetActorOutputAddress(actorType, actor.GetAddress())
			require.NoError(t, err)

			require.Equal(t, outputAddress, actor.GetOutput(), "unexpected output address")
		})
	}
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetPauseHeightIfExists", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			pauseHeight := int64(100)
			actor := GetFirstActor(t, ctx, actorType)

			err := ctx.SetActorPauseHeight(actorType, actor.GetAddress(), pauseHeight)
			require.NoError(t, err, "error setting actor pause height")

			gotPauseHeight, err := ctx.GetPauseHeight(actorType, actor.GetAddress())
			require.NoError(t, err)
			require.Equal(t, pauseHeight, gotPauseHeight, "unable to get pause height from the actor")

			randAddr, er := crypto.GenerateAddress()
			require.NoError(t, er)

			_, err = ctx.GetPauseHeight(actorType, randAddr)
			require.Error(t, err, "non existent actor should error")
		})
	}
}

func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageEditStakeSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)

			msgEditStake := &typesUtil.MessageEditStake{
				Address:   actor.GetAddress(),
				Chains:    defaultTestingChains,
				Amount:    defaultAmountString,
				ActorType: actorType,
			}

			candidates, err := ctx.GetMessageEditStakeSignerCandidates(msgEditStake)
			require.NoError(t, err)
			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, candidates[0], actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, candidates[1], actor.GetAddress(), "incorrect addr candidate")
		})
	}
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
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnpauseSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)

			msg := &typesUtil.MessageUnpause{
				Address:   actor.GetAddress(),
				ActorType: actorType,
			}

			candidates, err := ctx.GetMessageUnpauseSignerCandidates(msg)
			require.NoError(t, err)
			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, candidates[0], actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, candidates[1], actor.GetAddress(), "incorrect addr candidate")
		})
	}
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnstakeSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)

			msg := &typesUtil.MessageUnstake{
				Address:   actor.GetAddress(),
				ActorType: actorType,
			}
			candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
			require.NoError(t, err)
			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, candidates[0], actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, candidates[1], actor.GetAddress(), "incorrect addr candidate")
		})
	}
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.UnstakePausedBefore", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := GetFirstActor(t, ctx, actorType)
			require.Equal(t, actor.GetStatus(), int32(typesUtil.StakedStatus), "wrong starting status")

			err := ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 0)
			require.NoError(t, err, "error setting actor pause height")

			var er error
			switch actorType {
			case typesUtil.ActorType_App:
				er = ctx.Context.SetAppMaxPausedBlocks(0)
			case typesUtil.ActorType_Val:
				er = ctx.Context.SetValidatorMaxPausedBlocks(0)
			case typesUtil.ActorType_Fish:
				er = ctx.Context.SetFishermanMaxPausedBlocks(0)
			case typesUtil.ActorType_Node:
				er = ctx.Context.SetServiceNodeMaxPausedBlocks(0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, er, "error setting max paused blocks")

			err = ctx.UnstakeActorPausedBefore(0, actorType)
			require.NoError(t, err, "error unstaking actor pause before")

			err = ctx.UnstakeActorPausedBefore(1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")

			actor = GetActorByAddr(t, ctx, actor.GetAddress(), actorType)
			require.Equal(t, actor.GetStatus(), int32(typesUtil.UnstakingStatus), "status does not equal unstaking")

			var unstakingBlocks int64
			switch actorType {
			case typesUtil.ActorType_Val:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.ActorType_Node:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.ActorType_Fish:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error getting unstaking blocks")
			require.Equal(t, actor.GetUnstakingHeight(), unstakingBlocks+1, "incorrect unstaking height")
		})
	}
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)

		poolName := actorType.GetActorPoolName()
		var err1, err2 error
		switch actorType {
		case typesUtil.ActorType_App:
			err1 = ctx.Context.SetAppUnstakingBlocks(0)
			err2 = ctx.Context.SetAppMaxPausedBlocks(0)
		case typesUtil.ActorType_Val:
			err1 = ctx.Context.SetValidatorUnstakingBlocks(0)
			err2 = ctx.Context.SetValidatorMaxPausedBlocks(0)
		case typesUtil.ActorType_Fish:
			err1 = ctx.Context.SetFishermanUnstakingBlocks(0)
			err2 = ctx.Context.SetFishermanMaxPausedBlocks(0)
		case typesUtil.ActorType_Node:
			err1 = ctx.Context.SetServiceNodeUnstakingBlocks(0)
			err2 = ctx.Context.SetServiceNodeMaxPausedBlocks(0)
		default:
			t.Fatalf("unexpected actor type %s", actorType.GetActorName())
		}

		ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		require.NoError(t, err1, "error setting unstaking blocks")
		require.NoError(t, err2, "error setting max pause blocks")

		actors := GetAllTestingActors(t, ctx, actorType)
		for _, actor := range actors {
			require.Equal(t, actor.GetStatus(), int32(typesUtil.StakedStatus), "wrong starting staked status")
			err := ctx.SetActorPauseHeight(actorType, actor.GetAddress(), 1)
			require.NoError(t, err, "error setting actor pause height")
		}

		err := ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err, "error setting actor pause before")

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err, "error unstaking actors that are ready")

		require.Zero(t, len(GetAllTestingActors(t, ctx, actorType)), "actors still exists after unstake that are ready() call")
	}
}

// Helpers

func GetAllTestingActors(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) (actors []genesis.Actor) {
	actors = make([]genesis.Actor, 0)
	switch actorType {
	case typesUtil.ActorType_App:
		apps := GetAllTestingApps(t, ctx)
		for _, a := range apps {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_Node:
		nodes := GetAllTestingNodes(t, ctx)
		for _, a := range nodes {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_Val:
		vals := GetAllTestingValidators(t, ctx)
		for _, a := range vals {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_Fish:
		fish := GetAllTestingFish(t, ctx)
		for _, a := range fish {
			actors = append(actors, a)
		}
	default:
		t.Fatalf("unexpected actor type %s", actorType.GetActorName())
	}

	return
}

func GetFirstActor(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) genesis.Actor {
	return GetAllTestingActors(t, ctx, actorType)[0]
}

func GetActorByAddr(t *testing.T, ctx utility.UtilityContext, addr []byte, actorType typesUtil.ActorType) (actor genesis.Actor) {
	actors := GetAllTestingActors(t, ctx, actorType)
	for _, a := range actors {
		if bytes.Equal(a.GetAddress(), addr) {
			return a
		}
	}
	return
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

func GetAllTestingFish(t *testing.T, ctx utility.UtilityContext) []*genesis.Fisherman {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllFishermen(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingNodes(t *testing.T, ctx utility.UtilityContext) []*genesis.ServiceNode {
	actors, err := (ctx.Context.PersistenceContext).(*pre_persistence.PrePersistenceContext).GetAllServiceNodes(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
