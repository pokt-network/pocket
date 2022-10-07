package test

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// CLEANUP: Move `App` specific tests to `app_test.go`

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	for _, actorType := range actorTypes {
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

			actor := getActorByAddr(t, ctx, pubKey.Address().Bytes(), actorType)

			require.Equal(t, actor.GetAddress(), pubKey.Address().String(), "incorrect actor address")
			if actorType != typesUtil.ActorType_Validator {
				require.Equal(t, msg.Chains, actor.GetChains(), "incorrect actor chains")
			}
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetPausedHeight(), "incorrect actor height")
			require.Equal(t, test_artifacts.DefaultStakeAmountString, actor.GetStakedAmount(), "incorrect actor stake amount")
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "incorrect actor unstaking height")
			require.Equal(t, outputAddress.String(), actor.GetOutput(), "incorrect actor output address")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageEditStake", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			msg := &typesUtil.MessageEditStake{
				Address:   addrBz,
				Chains:    test_artifacts.DefaultChains,
				Amount:    test_artifacts.DefaultStakeAmountString,
				Signer:    addrBz,
				ActorType: actorType,
			}

			msgChainsEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgChainsEdited.Chains = defaultTestingChainsEdited

			err = ctx.HandleEditStakeMessage(msgChainsEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			if actorType != typesUtil.ActorType_Validator {
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

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnpause", actorType.String()), func(t *testing.T) {

			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.ActorType_Validator:
				err = ctx.Context.SetParam(modules.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_ServiceNode:
				err = ctx.Context.SetParam(modules.ServiceNodeMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_App:
				err = ctx.Context.SetParam(modules.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Fisherman:
				err = ctx.Context.SetParam(modules.FishermanMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting pause height")

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, int64(1), actor.GetPausedHeight())

			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err, "handle unpause message")

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, int64(-1), actor.GetPausedHeight())
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnstake", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.ActorType_App:
				err = ctx.Context.SetParam(modules.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Validator:
				err = ctx.Context.SetParam(modules.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Fisherman:
				err = ctx.Context.SetParam(modules.FishermanMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_ServiceNode:
				err = ctx.Context.SetParam(modules.ServiceNodeMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnstakeMessage(msg)
			require.NoError(t, err, "handle unstake message")

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, defaultUnstaking, actor.GetUnstakingHeight(), "actor should be unstaking")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPaused", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			switch actorType {
			case typesUtil.ActorType_App:
				err = ctx.Context.SetParam(modules.AppMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Validator:
				err = ctx.Context.SetParam(modules.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.ActorType_Fisherman:
				err = ctx.Context.SetParam(modules.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_ServiceNode:
				err = ctx.Context.SetParam(modules.ServiceNodeMaxPauseBlocksParamName, 0)
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_CalculateMaxAppRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := getFirstActor(t, ctx, typesUtil.ActorType_App)
	newMaxRelays, err := ctx.CalculateAppRelays(actor.GetStakedAmount())
	require.NoError(t, err)
	require.Equal(t, actor.GetGenericParam(), newMaxRelays)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.CalculateUnstakingHeight", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			var unstakingBlocks int64
			var err error
			switch actorType {
			case typesUtil.ActorType_Validator:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.ActorType_ServiceNode:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.ActorType_Fisherman:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error getting unstaking blocks")

			unstakingHeight, err := ctx.GetUnstakingHeight(actorType)
			require.NoError(t, err)

			require.Equal(t, unstakingBlocks, unstakingHeight, "unexpected unstaking height")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetExists(t *testing.T) {
	for _, actorType := range actorTypes {
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.GetOutputAddress", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			outputAddress, err := ctx.GetActorOutputAddress(actorType, addrBz)
			require.NoError(t, err)

			require.Equal(t, actor.GetOutput(), hex.EncodeToString(outputAddress), "unexpected output address")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	for _, actorType := range actorTypes {
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	for _, actorType := range actorTypes {
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	for _, actorType := range actorTypes {
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	for _, actorType := range actorTypes {
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
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	for _, actorType := range actorTypes {
		t.Run(fmt.Sprintf("%s.UnstakePausedBefore", actorType.String()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := getFirstActor(t, ctx, actorType)
			require.Equal(t, actor.GetUnstakingHeight(), int64(-1), "wrong starting status")
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			var er error
			switch actorType {
			case typesUtil.ActorType_App:
				er = ctx.Context.SetParam(modules.AppMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Validator:
				er = ctx.Context.SetParam(modules.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.ActorType_Fisherman:
				er = ctx.Context.SetParam(modules.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_ServiceNode:
				er = ctx.Context.SetParam(modules.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, er, "error setting max paused blocks")

			err = ctx.UnstakeActorPausedBefore(0, actorType)
			require.NoError(t, err, "error unstaking actor pause before")

			err = ctx.UnstakeActorPausedBefore(1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")

			actor = getActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.GetUnstakingHeight(), defaultUnstaking, "status does not equal unstaking")

			var unstakingBlocks int64
			switch actorType {
			case typesUtil.ActorType_Validator:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.ActorType_ServiceNode:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.ActorType_Fisherman:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error getting unstaking blocks")
			require.Equal(t, unstakingBlocks+1, actor.GetUnstakingHeight(), "incorrect unstaking height")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	for _, actorType := range actorTypes {
		ctx := NewTestingUtilityContext(t, 1)

		poolName := ""
		var err1, err2 error
		switch actorType {
		case typesUtil.ActorType_App:
			err1 = ctx.Context.SetParam(modules.AppUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(modules.AppMaxPauseBlocksParamName, 0)
			poolName = typesUtil.PoolNames_AppStakePool.String()
		case typesUtil.ActorType_Validator:
			err1 = ctx.Context.SetParam(modules.ValidatorUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(modules.ValidatorMaxPausedBlocksParamName, 0)
			poolName = typesUtil.PoolNames_ValidatorStakePool.String()
		case typesUtil.ActorType_Fisherman:
			err1 = ctx.Context.SetParam(modules.FishermanUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(modules.FishermanMaxPauseBlocksParamName, 0)
			poolName = typesUtil.PoolNames_FishermanStakePool.String()
		case typesUtil.ActorType_ServiceNode:
			err1 = ctx.Context.SetParam(modules.ServiceNodeUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(modules.ServiceNodeMaxPauseBlocksParamName, 0)
			poolName = typesUtil.PoolNames_ServiceNodeStakePool.String()
		default:
			t.Fatalf("unexpected actor type %s", actorType.String())
		}

		err := ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		require.NoError(t, err)
		require.NoError(t, err1, "error setting unstaking blocks")
		require.NoError(t, err2, "error setting max pause blocks")

		actors := getAllTestingActors(t, ctx, actorType)
		for _, actor := range actors {
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			require.Equal(t, int64(-1), actor.GetUnstakingHeight(), "wrong starting staked status")
			err = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting actor pause height")
		}

		err = ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err, "error setting actor pause before")

		accountAmountsBefore := make([]*big.Int, 0)

		for _, actor := range actors {
			// get the output address account amount before the 'unstake'
			outputAddressString := actor.GetOutput()
			outputAddress, err := hex.DecodeString(outputAddressString)
			require.NoError(t, err)
			outputAccountAmount, err := ctx.GetAccountAmount(outputAddress)
			require.NoError(t, err)
			// capture the amount before
			accountAmountsBefore = append(accountAmountsBefore, outputAccountAmount)
		}
		// capture the pool amount before
		poolAmountBefore, err := ctx.GetPoolAmount(poolName)
		require.NoError(t, err)

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err, "error unstaking actors that are ready")

		for i, actor := range actors {
			// get the output address account amount after the 'unstake'
			outputAddressString := actor.GetOutput()
			outputAddress, err := hex.DecodeString(outputAddressString)
			require.NoError(t, err)
			outputAccountAmount, err := ctx.GetAccountAmount(outputAddress)
			require.NoError(t, err)
			// ensure the stake amount went to the output address
			outputAccountAmountDelta := new(big.Int).Sub(outputAccountAmount, accountAmountsBefore[i])
			require.Equal(t, outputAccountAmountDelta, test_artifacts.DefaultStakeAmount)
		}
		// ensure the staking pool is `# of readyToUnstake actors * default stake` less than before the unstake
		poolAmountAfter, err := ctx.GetPoolAmount(poolName)
		require.NoError(t, err)
		actualPoolDelta := new(big.Int).Sub(poolAmountBefore, poolAmountAfter)
		expectedPoolDelta := new(big.Int).Mul(big.NewInt(int64(len(actors))), test_artifacts.DefaultStakeAmount)
		require.Equal(t, expectedPoolDelta, actualPoolDelta)

		test_artifacts.CleanupTest(ctx)
	}
}

// Helpers

func getAllTestingActors(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) (actors []modules.Actor) {
	actors = make([]modules.Actor, 0)
	switch actorType {
	case typesUtil.ActorType_App:
		apps := getAllTestingApps(t, ctx)
		for _, a := range apps {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_ServiceNode:
		nodes := getAllTestingNodes(t, ctx)
		for _, a := range nodes {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_Validator:
		vals := getAllTestingValidators(t, ctx)
		for _, a := range vals {
			actors = append(actors, a)
		}
	case typesUtil.ActorType_Fisherman:
		fish := getAllTestingFish(t, ctx)
		for _, a := range fish {
			actors = append(actors, a)
		}
	default:
		t.Fatalf("unexpected actor type %s", actorType.String())
	}

	return
}

func getFirstActor(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) modules.Actor {
	return getAllTestingActors(t, ctx, actorType)[0]
}

func getActorByAddr(t *testing.T, ctx utility.UtilityContext, addr []byte, actorType typesUtil.ActorType) (actor modules.Actor) {
	actors := getAllTestingActors(t, ctx, actorType)
	for _, a := range actors {
		if a.GetAddress() == hex.EncodeToString(addr) {
			return a
		}
	}
	return
}

func getAllTestingApps(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func getAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].GetAddress() < actors[j].GetAddress()
	})
	return actors
}

func getAllTestingFish(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllFishermen(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func getAllTestingNodes(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllServiceNodes(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
