package utility

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

// CLEANUP: Move `App` specific tests to `app_test.go`

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageStake", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			pubKey, err := crypto.GeneratePublicKey()
			require.NoError(t, err)

			outputAddress, err := crypto.GenerateAddress()
			require.NoError(t, err)

			err = ctx.SetAccountAmount(outputAddress, test_artifacts.DefaultAccountAmount)
			require.NoError(t, err, "error setting account amount error")

			msg := &typesUtil.MessageStake{
				PublicKey:     pubKey.Bytes(),
				Chains:        test_artifacts.DefaultChains,
				Amount:        test_artifacts.DefaultStakeAmountString,
				ServiceUrl:    "https://localhost.com",
				OutputAddress: outputAddress,
				Signer:        outputAddress,
				ActorType:     actorType,
			}

			er := ctx.HandleStakeMessage(msg)
			require.NoError(t, er, "handle stake message")

			actor := getActorByAddr(t, ctx, actorType, pubKey.Address().String())

			require.Equal(t, actor.GetAddress(), pubKey.Address().String(), "incorrect actor address")
			if actorType != coreTypes.ActorType_ACTOR_TYPE_VAL {
				require.Equal(t, msg.Chains, actor.GetChains(), "incorrect actor chains")
			}
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetPausedHeight(), "incorrect actor height")
			require.Equal(t, test_artifacts.DefaultStakeAmountString, actor.GetStakedAmount(), "incorrect actor stake amount")
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "incorrect actor unstaking height")
			require.Equal(t, outputAddress.String(), actor.GetOutput(), "incorrect actor output address")

		})
	}
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageEditStake", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			msg := &typesUtil.MessageEditStake{
				Address:   addrBz,
				Chains:    test_artifacts.DefaultChains,
				Amount:    test_artifacts.DefaultStakeAmountString,
				Signer:    addrBz,
				ActorType: actorType,
			}
			msgChainsEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgChainsEdited.Chains = []string{"0002"}

			err = ctx.HandleEditStakeMessage(msgChainsEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = getActorByAddr(t, ctx, actorType, addr)
			if actorType != coreTypes.ActorType_ACTOR_TYPE_VAL {
				require.Equal(t, msgChainsEdited.Chains, actor.GetChains(), "incorrect edited chains")
			}
			require.Equal(t, test_artifacts.DefaultStakeAmountString, actor.GetStakedAmount(), "incorrect staked tokens")
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "incorrect unstaking height")

			amountEdited := test_artifacts.DefaultAccountAmount.Add(test_artifacts.DefaultAccountAmount, big.NewInt(1))
			amountEditedString := typesUtil.BigIntToString(amountEdited)
			msgAmountEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgAmountEdited.Amount = amountEditedString

			err = ctx.HandleEditStakeMessage(msgAmountEdited)
			require.NoError(t, err, "handle edit stake message")

		})
	}
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnpause", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			var err error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				err = ctx.persistenceContext.SetParam(typesUtil.ValidatorMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				err = ctx.persistenceContext.SetParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				err = ctx.persistenceContext.SetParam(typesUtil.AppMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				err = ctx.persistenceContext.SetParam(typesUtil.FishermanMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting pause height")

			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, int64(1), actor.GetPausedHeight())

			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err, "handle unpause message")

			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, int64(-1), actor.GetPausedHeight())

		})
	}
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnstake", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			var err error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				err = ctx.persistenceContext.SetParam(typesUtil.AppMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				err = ctx.persistenceContext.SetParam(typesUtil.ValidatorMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				err = ctx.persistenceContext.SetParam(typesUtil.FishermanMinimumPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				err = ctx.persistenceContext.SetParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnstakeMessage(msg)
			require.NoError(t, err, "handle unstake message")

			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, defaultUnstakingHeight, actor.GetUnstakingHeight(), "actor should be unstaking")

		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPaused", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				err = ctx.persistenceContext.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				err = ctx.persistenceContext.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				err = ctx.persistenceContext.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				err = ctx.persistenceContext.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			err = ctx.BeginUnstakingMaxPaused()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			status, err := ctx.GetActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, int32(typesUtil.StakeStatus_Unstaking), status, "actor should be unstaking")
		})
	}
}

func TestUtilityContext_CalculateMaxAppRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := getFirstActor(t, ctx, coreTypes.ActorType_ACTOR_TYPE_APP)
	newMaxRelays, err := ctx.CalculateAppRelays(actor.GetStakedAmount())
	require.NoError(t, err)
	require.Equal(t, actor.GetGenericParam(), newMaxRelays)
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.CalculateUnstakingHeight", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			var unstakingBlocks int64
			var err error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error getting unstaking blocks")

			unstakingHeight, err := ctx.GetUnstakingHeight(actorType)
			require.NoError(t, err)
			require.Equal(t, unstakingBlocks, unstakingHeight, "unexpected unstaking height")

		})
	}
}

func TestUtilityContext_GetExists(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetExists", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := getFirstActor(t, ctx, actorType)
			randAddr, err := crypto.GenerateAddress()
			require.NoError(t, err)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			exists, err := ctx.GetActorExists(actorType, addrBz)
			require.NoError(t, err)
			require.True(t, exists, "actor that should exist does not")

			exists, err = ctx.GetActorExists(actorType, randAddr)
			require.NoError(t, err)
			require.False(t, exists, "actor that shouldn't exist does")

		})
	}
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetOutputAddress", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			outputAddress, err := ctx.GetActorOutputAddress(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(outputAddress), "unexpected output address")

		})
	}
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetPauseHeightIfExists", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			pauseHeight := int64(100)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, pauseHeight)
			require.NoError(t, err, "error setting actor pause height")

			gotPauseHeight, err := ctx.GetPauseHeight(actorType, addrBz)
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
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageEditStakeSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msgEditStake := &typesUtil.MessageEditStake{
				Address:   addrBz,
				Chains:    test_artifacts.DefaultChains,
				Amount:    test_artifacts.DefaultStakeAmountString,
				ActorType: actorType,
			}
			candidates, err := ctx.GetMessageEditStakeSignerCandidates(msgEditStake)
			require.NoError(t, err)

			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")

		})
	}
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnpauseSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msg := &typesUtil.MessageUnpause{
				Address:   addrBz,
				ActorType: actorType,
			}
			candidates, err := ctx.GetMessageUnpauseSignerCandidates(msg)
			require.NoError(t, err)

			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")

		})
	}
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnstakeSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				ActorType: actorType,
			}
			candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
			require.NoError(t, err)

			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")

		})
	}
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.UnstakePausedBefore", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := getFirstActor(t, ctx, actorType)
			require.Equal(t, actor.GetUnstakingHeight(), int64(-1), "wrong starting status")

			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			var er error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				er = ctx.persistenceContext.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				er = ctx.persistenceContext.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				er = ctx.persistenceContext.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				er = ctx.persistenceContext.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, er, "error setting max paused blocks")

			err = ctx.UnstakeActorPausedBefore(0, actorType)
			require.NoError(t, err, "error unstaking actor pause before")

			err = ctx.UnstakeActorPausedBefore(1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")

			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, actor.GetUnstakingHeight(), defaultUnstakingHeight, "status does not equal unstaking")

			var unstakingBlocks int64
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error getting unstaking blocks")
			require.Equal(t, unstakingBlocks+1, actor.GetUnstakingHeight(), "incorrect unstaking height")

		})
	}
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.UnstakeActorsThatAreReady", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			var poolName string
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				poolName = coreTypes.Pools_POOLS_APP_STAKE.FriendlyName()
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				poolName = coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName()
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				poolName = coreTypes.Pools_POOLS_FISHERMAN_STAKE.FriendlyName()
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				poolName = coreTypes.Pools_POOLS_SERVICE_NODE_STAKE.FriendlyName()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))

			err := ctx.persistenceContext.SetParam(typesUtil.AppUnstakingBlocksParamName, 0)
			require.NoError(t, err)

			err = ctx.persistenceContext.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			require.NoError(t, err)

			actors := getAllTestingActors(t, ctx, actorType)
			for _, actor := range actors {
				// require.Equal(t, int32(typesUtil.StakedStatus), actor.GetStatus(), "wrong starting status")
				addrBz, er := hex.DecodeString(actor.GetAddress())
				require.NoError(t, er)
				er = ctx.SetActorPauseHeight(actorType, addrBz, 1)
				require.NoError(t, er)
			}

			err = ctx.UnstakeActorPausedBefore(2, actorType)
			require.NoError(t, err)

			err = ctx.UnstakeActorsThatAreReady()
			require.NoError(t, err)

			actors = getAllTestingActors(t, ctx, actorType)
			require.NotEqual(t, actors[0].GetUnstakingHeight(), -1, "validators still exists after unstake that are ready() call")

			// TODO: We need to better define what 'deleted' really is in the postgres world.
			// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight
		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	for _, actorType := range coreTypes.ActorTypes {
		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPausedActors", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)
			actor := getFirstActor(t, ctx, actorType)

			var err error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				err = ctx.persistenceContext.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				err = ctx.persistenceContext.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				err = ctx.persistenceContext.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
			case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
				err = ctx.persistenceContext.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err)

			addrBz, er := hex.DecodeString(actor.GetAddress())
			require.NoError(t, er)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err)

			err = ctx.BeginUnstakingMaxPaused()
			require.NoError(t, err)

			status, err := ctx.GetActorStatus(actorType, addrBz)
			require.Equal(t, int32(typesUtil.StakeStatus_Unstaking), status, "incorrect status")

		})
	}
}

// Helpers

func getAllTestingActors(t *testing.T, ctx utilityContext, actorType coreTypes.ActorType) (actors []*coreTypes.Actor) {
	actors = make([]*coreTypes.Actor, 0)
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		apps := getAllTestingApps(t, ctx)
		for _, a := range apps {
			actors = append(actors, a)
		}
	case coreTypes.ActorType_ACTOR_TYPE_SERVICENODE:
		nodes := getAllTestingNodes(t, ctx)
		for _, a := range nodes {
			actors = append(actors, a)
		}
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		vals := getAllTestingValidators(t, ctx)
		for _, a := range vals {
			actors = append(actors, a)
		}
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		fish := getAllTestingFish(t, ctx)
		for _, a := range fish {
			actors = append(actors, a)
		}
	default:
		t.Fatalf("unexpected actor type %s", actorType.String())
	}

	return
}

func getFirstActor(t *testing.T, ctx utilityContext, actorType coreTypes.ActorType) *coreTypes.Actor {
	return getAllTestingActors(t, ctx, actorType)[0]
}

func getActorByAddr(t *testing.T, ctx utilityContext, actorType coreTypes.ActorType, addr string) (actor *coreTypes.Actor) {
	actors := getAllTestingActors(t, ctx, actorType)
	idx := slices.IndexFunc(actors, func(a *coreTypes.Actor) bool { return a.GetAddress() == addr })
	return actors[idx]
}

func getAllTestingApps(t *testing.T, ctx utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllApps(ctx.height)
	require.NoError(t, err)
	return actors
}

func getAllTestingValidators(t *testing.T, ctx utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllValidators(ctx.height)
	require.NoError(t, err)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].GetAddress() < actors[j].GetAddress()
	})
	return actors
}

func getAllTestingFish(t *testing.T, ctx utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllFishermen(ctx.height)
	require.NoError(t, err)
	return actors
}

func getAllTestingNodes(t *testing.T, ctx utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllServiceNodes(ctx.height)
	require.NoError(t, err)
	return actors
}
