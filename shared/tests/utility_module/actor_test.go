package utility_module

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/tests"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"github.com/pokt-network/pocket/utility"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// CLEANUP: Move `App` specific tests to `app_test.go`

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageStake", actorType.GetActorName()), func(t *testing.T) {
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

			actor := GetActorByAddr(t, ctx, pubKey.Address().Bytes(), actorType)

			require.Equal(t, actor.GetAddress(), pubKey.Address().String(), "incorrect actor address")
			if actorType != typesUtil.ActorType_Val {
				require.Equal(t, actor.GetChains(), msg.Chains, "incorrect actor chains")
			}
			require.Equal(t, actor.GetPausedHeight(), types.HeightNotUsed, "incorrect actor height")
			require.Equal(t, actor.GetStakedAmount(), test_artifacts.DefaultStakeAmountString, "incorrect actor stake amount")
			require.Equal(t, actor.GetUnstakingHeight(), types.HeightNotUsed, "incorrect actor unstaking height")
			require.Equal(t, actor.GetOutput(), outputAddress.String(), "incorrect actor output address")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageEditStake", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			actor := GetFirstActor(t, ctx, actorType)
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

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			if actorType != typesUtil.ActorType_Val {
				require.Equal(t, actor.GetChains(), msgChainsEdited.Chains, "incorrect edited chains")
			}
			require.Equal(t, actor.GetStakedAmount(), test_artifacts.DefaultStakeAmountString, "incorrect staked tokens")
			require.Equal(t, actor.GetUnstakingHeight(), types.HeightNotUsed, "incorrect unstaking height")

			amountEdited := test_artifacts.DefaultAccountAmount.Add(test_artifacts.DefaultAccountAmount, big.NewInt(1))
			amountEditedString := types.BigIntToString(amountEdited)
			msgAmountEdited := proto.Clone(msg).(*typesUtil.MessageEditStake)
			msgAmountEdited.Amount = amountEditedString

			err = ctx.HandleEditStakeMessage(msgAmountEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
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
				err = ctx.Context.SetParam(types.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetParam(types.ServiceNodeMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_App:
				err = ctx.Context.SetParam(types.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetParam(types.FishermanMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting pause height")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.GetPausedHeight(), int64(1))

			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err, "handle unpause message")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.PausedHeight, int64(-1))
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
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
				err = ctx.Context.SetParam(types.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Val:
				err = ctx.Context.SetParam(types.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetParam(types.FishermanMinimumPauseBlocksParamName, 0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetParam(types.ServiceNodeMinimumPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnstakeMessage(msg)
			require.NoError(t, err, "handle unstake message")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.UnstakingHeight, defaultUnstaking, "actor should be unstaking")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPaused(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPaused", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			switch actorType {
			case typesUtil.ActorType_App:
				err = ctx.Context.SetParam(types.AppMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Val:
				err = ctx.Context.SetParam(types.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.ActorType_Fish:
				err = ctx.Context.SetParam(types.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Node:
				err = ctx.Context.SetParam(types.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			err = ctx.BeginUnstakingMaxPaused()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			status, err := ctx.GetActorStatus(actorType, addrBz)
			require.Equal(t, status, typesUtil.UnstakingStatus, "actor should be unstaking")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newMaxRelays, err := ctx.CalculateAppRelays(actor.StakedAmount)
	require.NoError(t, err)
	require.True(t, actor.GenericParam == newMaxRelays, fmt.Sprintf("unexpected max relay calculation; got %v wanted %v", actor.GenericParam, newMaxRelays))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
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
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_Delete(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.Delete", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.DeleteActor(actorType, addrBz)
			require.NoError(t, err, "error deleting actor")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			// TODO Delete actor is currently a NO-OP. We need to better define
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
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
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			exists, err := ctx.GetActorExists(actorType, addrBz)
			require.NoError(t, err)
			require.True(t, exists, "actor that should exist does not")

			exists, err = ctx.GetActorExists(actorType, randAddr)
			require.NoError(t, err)
			require.False(t, exists, "actor that shouldn't exist does")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetOutputAddress", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			outputAddress, err := ctx.GetActorOutputAddress(actorType, addrBz)
			require.NoError(t, err)

			require.Equal(t, hex.EncodeToString(outputAddress), actor.GetOutput(), "unexpected output address")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetPauseHeightIfExists", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			pauseHeight := int64(100)
			actor := GetFirstActor(t, ctx, actorType)
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
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageEditStakeSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
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
			require.Equal(t, hex.EncodeToString(candidates[0]), actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, hex.EncodeToString(candidates[1]), actor.GetAddress(), "incorrect addr candidate")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_UnstakesPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	addrBz, err := hex.DecodeString(actor.GetAddress())
	require.NoError(t, err)
	require.True(t, actor.UnstakingHeight == -1, fmt.Sprintf("wrong starting status"))
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, addrBz, 0), "set actor pause height")
	err = ctx.Context.SetAppMaxPausedBlocks(0)
	require.NoError(t, err)
	require.NoError(t, ctx.UnstakeActorPausedBefore(0, typesUtil.ActorType_App), "unstake actor pause before")
	require.NoError(t, ctx.UnstakeActorPausedBefore(1, typesUtil.ActorType_App), "unstake actor pause before height 1")
	actor = GetAllTestingApps(t, ctx)[0]
	require.True(t, actor.UnstakingHeight != -1, fmt.Sprintf("status does not equal unstaking"))
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.True(t, actor.UnstakingHeight == unstakingBlocks+1, fmt.Sprintf("incorrect unstaking height"))
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func TestUtilityContext_UnstakesThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount(genesis.Pool_Names_AppStakePool.String(), big.NewInt(math.MaxInt64))
	require.NoError(t, ctx.Context.SetAppUnstakingBlocks(0), "set unstaking blocks")
	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		addrBz, err := hex.DecodeString(actor.GetAddress())
		require.NoError(t, err)
		require.True(t, actor.UnstakingHeight == -1, fmt.Sprintf("wrong starting status"))
		require.NoError(t, ctx.SetActorPauseHeight(typesUtil.ActorType_App, addrBz, 1), "set actor pause height")
	}
	require.NoError(t, ctx.UnstakeActorPausedBefore(2, typesUtil.ActorType_App), "set actor pause before")
	require.NoError(t, ctx.UnstakeActorsThatAreReady(), "unstake actors that are ready")
	appAfter := GetAllTestingApps(t, ctx)[0]
	require.True(t, appAfter.UnstakingHeight == 0, fmt.Sprintf("apps still exists after unstake that are ready() call"))
	// TODO (Team) we need to better define what 'deleted' really is in the postgres world.
	// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight
	ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
	tests.CleanupTest()
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnpauseSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			msg := &typesUtil.MessageUnpause{
				Address:   addrBz,
				ActorType: actorType,
			}

			candidates, err := ctx.GetMessageUnpauseSignerCandidates(msg)
			require.NoError(t, err)
			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, hex.EncodeToString(candidates[0]), actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, hex.EncodeToString(candidates[1]), actor.GetAddress(), "incorrect addr candidate")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.GetMessageUnstakeSignerCandidates", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)

			actor := GetFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				ActorType: actorType,
			}
			candidates, err := ctx.GetMessageUnstakeSignerCandidates(msg)
			require.NoError(t, err)
			require.Equal(t, len(candidates), 2, "unexpected number of candidates")
			require.Equal(t, hex.EncodeToString(candidates[0]), actor.GetOutput(), "incorrect output candidate")
			require.Equal(t, hex.EncodeToString(candidates[1]), actor.GetAddress(), "incorrect addr candidate")
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
		})
	}
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.UnstakePausedBefore", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := GetFirstActor(t, ctx, actorType)
			require.Equal(t, actor.UnstakingHeight, int64(-1), "wrong starting status")
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			var er error
			switch actorType {
			case typesUtil.ActorType_App:
				er = ctx.Context.SetParam(types.AppMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Val:
				er = ctx.Context.SetParam(types.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.ActorType_Fish:
				er = ctx.Context.SetParam(types.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.ActorType_Node:
				er = ctx.Context.SetParam(types.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, er, "error setting max paused blocks")

			err = ctx.UnstakeActorPausedBefore(0, actorType)
			require.NoError(t, err, "error unstaking actor pause before")

			err = ctx.UnstakeActorPausedBefore(1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.UnstakingHeight, defaultUnstaking, "status does not equal unstaking")

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
			ctx.Context.Release() // TODO (team) need a golang specific solution for teardown
			tests.CleanupTest()
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
			err1 = ctx.Context.SetParam(types.AppUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(types.AppMaxPauseBlocksParamName, 0)
		case typesUtil.ActorType_Val:
			err1 = ctx.Context.SetParam(types.ValidatorUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(types.ValidatorMaxPausedBlocksParamName, 0)
		case typesUtil.ActorType_Fish:
			err1 = ctx.Context.SetParam(types.FishermanUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(types.FishermanMaxPauseBlocksParamName, 0)
		case typesUtil.ActorType_Node:
			err1 = ctx.Context.SetParam(types.ServiceNodeUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(types.ServiceNodeMaxPauseBlocksParamName, 0)
		default:
			t.Fatalf("unexpected actor type %s", actorType.GetActorName())
		}

		ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		require.NoError(t, err1, "error setting unstaking blocks")
		require.NoError(t, err2, "error setting max pause blocks")

		actors := GetAllTestingActors(t, ctx, actorType)
		for _, actor := range actors {
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			require.Equal(t, actor.UnstakingHeight, int64(-1), "wrong starting staked status")
			err = ctx.SetActorPauseHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting actor pause height")
		}

		err := ctx.UnstakeActorPausedBefore(2, actorType)
		require.NoError(t, err, "error setting actor pause before")

		err = ctx.UnstakeActorsThatAreReady()
		require.NoError(t, err, "error unstaking actors that are ready")
		// TODO Delete() is no op
		tests.CleanupTest(ctx)
	}
}

// Helpers

func GetAllTestingActors(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) (actors []*genesis.Actor) {
	actors = make([]*genesis.Actor, 0)
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

func GetFirstActor(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.ActorType) *genesis.Actor {
	return GetAllTestingActors(t, ctx, actorType)[0]
}

func GetActorByAddr(t *testing.T, ctx utility.UtilityContext, addr []byte, actorType typesUtil.ActorType) (actor *genesis.Actor) {
	actors := GetAllTestingActors(t, ctx, actorType)
	for _, a := range actors {
		if a.GetAddress() == hex.EncodeToString(addr) {
			return a
		}
	}
	return
}

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []*genesis.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []*genesis.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].Address < actors[j].Address
	})
	return actors
}

func GetAllTestingFish(t *testing.T, ctx utility.UtilityContext) []*genesis.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllFishermen(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingNodes(t *testing.T, ctx utility.UtilityContext) []*genesis.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).(persistence.PostgresContext).GetAllServiceNodes(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
