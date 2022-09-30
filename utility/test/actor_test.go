package test

import (
	"encoding/hex"
	"fmt"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/test_artifacts"
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
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
			if actorType != typesUtil.UtilActorType_Val {
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
			if actorType != typesUtil.UtilActorType_Val {
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

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnpause", actorType.GetActorName()), func(t *testing.T) {

			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.UtilActorType_Val:
				err = ctx.Context.SetParam(typesUtil.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Node:
				err = ctx.Context.SetParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_App:
				err = ctx.Context.SetParam(typesUtil.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Fish:
				err = ctx.Context.SetParam(typesUtil.FishermanMinimumPauseBlocksParamName, 0)
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
			require.Equal(t, int64(1), actor.GetPausedHeight())

			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			err = ctx.HandleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err, "handle unpause message")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, int64(-1), actor.GetPausedHeight())
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.HandleMessageUnstake", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)
			var err error
			switch actorType {
			case typesUtil.UtilActorType_App:
				err = ctx.Context.SetParam(typesUtil.AppMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Val:
				err = ctx.Context.SetParam(typesUtil.ValidatorMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Fish:
				err = ctx.Context.SetParam(typesUtil.FishermanMinimumPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Node:
				err = ctx.Context.SetParam(typesUtil.ServiceNodeMinimumPauseBlocksParamName, 0)
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
			require.Equal(t, defaultUnstaking, actor.GetUnstakingHeight(), "actor should be unstaking")
			test_artifacts.CleanupTest(ctx)
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
			case typesUtil.UtilActorType_App:
				err = ctx.Context.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Val:
				err = ctx.Context.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.UtilActorType_Fish:
				err = ctx.Context.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Node:
				err = ctx.Context.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err)

			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			err = ctx.BeginUnstakingMaxPaused()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			status, err := ctx.GetActorStatus(actorType, addrBz)
			require.Equal(t, typesUtil.UnstakingStatus, status, "actor should be unstaking")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_CalculateRelays(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	newMaxRelays, err := ctx.CalculateAppRelays(actor.GetStakedAmount())
	require.NoError(t, err)
	require.Equal(t, actor.GetGenericParam(), newMaxRelays)
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_CalculateUnstakingHeight(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.CalculateUnstakingHeight", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 0)
			var unstakingBlocks int64
			var err error
			switch actorType {
			case typesUtil.UtilActorType_Val:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.UtilActorType_Node:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.UtilActorType_Fish:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
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
			test_artifacts.CleanupTest(ctx)
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

			require.Equal(t, actor.GetOutput(), hex.EncodeToString(outputAddress), "unexpected output address")
			test_artifacts.CleanupTest(ctx)
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
			test_artifacts.CleanupTest(ctx)
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
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_UnstakesPausedBefore(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 1)
	actor := GetAllTestingApps(t, ctx)[0]
	addrBz, err := hex.DecodeString(actor.GetAddress())
	require.NoError(t, err)
	require.Equal(t, -1, actor.GetUnstakingHeight())
	require.NoError(t, ctx.SetActorPauseHeight(typesUtil.UtilActorType_App, addrBz, 0), "set actor pause height")
	err = ctx.Context.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
	require.NoError(t, err)
	require.NoError(t, ctx.UnstakeActorPausedBefore(0, typesUtil.UtilActorType_App), "unstake actor pause before")
	require.NoError(t, ctx.UnstakeActorPausedBefore(1, typesUtil.UtilActorType_App), "unstake actor pause before height 1")
	actor = GetAllTestingApps(t, ctx)[0]
	require.Equal(t, -1, actor.GetUnstakingHeight())
	unstakingBlocks, err := ctx.GetAppUnstakingBlocks()
	require.NoError(t, err)
	require.Equal(t, unstakingBlocks+1, actor.GetUnstakingHeight())
	test_artifacts.CleanupTest(ctx)
}

func TestUtilityContext_UnstakesThatAreReady(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)
	ctx.SetPoolAmount("AppStakePool", big.NewInt(math.MaxInt64))
	err := ctx.Context.SetParam(typesUtil.AppUnstakingBlocksParamName, 0)
	require.NoError(t, err)
	actors := GetAllTestingApps(t, ctx)
	for _, actor := range actors {
		addrBz, err := hex.DecodeString(actor.GetAddress())
		require.NoError(t, err)
		require.Equal(t, -1, actor.GetUnstakingHeight())
		require.NoError(t, ctx.SetActorPauseHeight(typesUtil.UtilActorType_App, addrBz, 1), "set actor pause height")
	}
	require.NoError(t, ctx.UnstakeActorPausedBefore(2, typesUtil.UtilActorType_App), "set actor pause before")
	require.NoError(t, ctx.UnstakeActorsThatAreReady(), "unstake actors that are ready")
	appAfter := GetAllTestingApps(t, ctx)[0]
	require.Equal(t, 0, appAfter.GetUnstakingHeight())
	// TODO (Team) we need to better define what 'deleted' really is in the postgres world.
	// We might not need to 'unstakeActorsThatAreReady' if we are already filtering by unstakingHeight
	test_artifacts.CleanupTest(ctx)
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
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
			test_artifacts.CleanupTest(ctx)
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
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_UnstakePausedBefore(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		t.Run(fmt.Sprintf("%s.UnstakePausedBefore", actorType.GetActorName()), func(t *testing.T) {
			ctx := NewTestingUtilityContext(t, 1)

			actor := GetFirstActor(t, ctx, actorType)
			require.Equal(t, actor.GetUnstakingHeight(), int64(-1), "wrong starting status")
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)
			err = ctx.SetActorPauseHeight(actorType, addrBz, 0)
			require.NoError(t, err, "error setting actor pause height")

			var er error
			switch actorType {
			case typesUtil.UtilActorType_App:
				er = ctx.Context.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Val:
				er = ctx.Context.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
			case typesUtil.UtilActorType_Fish:
				er = ctx.Context.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
			case typesUtil.UtilActorType_Node:
				er = ctx.Context.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, er, "error setting max paused blocks")

			err = ctx.UnstakeActorPausedBefore(0, actorType)
			require.NoError(t, err, "error unstaking actor pause before")

			err = ctx.UnstakeActorPausedBefore(1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")

			actor = GetActorByAddr(t, ctx, addrBz, actorType)
			require.Equal(t, actor.GetUnstakingHeight(), defaultUnstaking, "status does not equal unstaking")

			var unstakingBlocks int64
			switch actorType {
			case typesUtil.UtilActorType_Val:
				unstakingBlocks, err = ctx.GetValidatorUnstakingBlocks()
			case typesUtil.UtilActorType_Node:
				unstakingBlocks, err = ctx.GetServiceNodeUnstakingBlocks()
			case actorType:
				unstakingBlocks, err = ctx.GetAppUnstakingBlocks()
			case typesUtil.UtilActorType_Fish:
				unstakingBlocks, err = ctx.GetFishermanUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.GetActorName())
			}
			require.NoError(t, err, "error getting unstaking blocks")
			require.Equal(t, unstakingBlocks+1, actor.GetUnstakingHeight(), "incorrect unstaking height")
			test_artifacts.CleanupTest(ctx)
		})
	}
}

func TestUtilityContext_UnstakeActorsThatAreReady(t *testing.T) {
	for _, actorType := range typesUtil.ActorTypes {
		ctx := NewTestingUtilityContext(t, 1)

		poolName := actorType.GetActorPoolName()
		var err1, err2 error
		switch actorType {
		case typesUtil.UtilActorType_App:
			err1 = ctx.Context.SetParam(typesUtil.AppUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(typesUtil.AppMaxPauseBlocksParamName, 0)
		case typesUtil.UtilActorType_Val:
			err1 = ctx.Context.SetParam(typesUtil.ValidatorUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(typesUtil.ValidatorMaxPausedBlocksParamName, 0)
		case typesUtil.UtilActorType_Fish:
			err1 = ctx.Context.SetParam(typesUtil.FishermanUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(typesUtil.FishermanMaxPauseBlocksParamName, 0)
		case typesUtil.UtilActorType_Node:
			err1 = ctx.Context.SetParam(typesUtil.ServiceNodeUnstakingBlocksParamName, 0)
			err2 = ctx.Context.SetParam(typesUtil.ServiceNodeMaxPauseBlocksParamName, 0)
		default:
			t.Fatalf("unexpected actor type %s", actorType.GetActorName())
		}

		err := ctx.SetPoolAmount(poolName, big.NewInt(math.MaxInt64))
		require.NoError(t, err)
		require.NoError(t, err1, "error setting unstaking blocks")
		require.NoError(t, err2, "error setting max pause blocks")

		actors := GetAllTestingActors(t, ctx, actorType)
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

func GetAllTestingActors(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.UtilActorType) (actors []modules.Actor) {
	actors = make([]modules.Actor, 0)
	switch actorType {
	case typesUtil.UtilActorType_App:
		apps := GetAllTestingApps(t, ctx)
		for _, a := range apps {
			actors = append(actors, a)
		}
	case typesUtil.UtilActorType_Node:
		nodes := GetAllTestingNodes(t, ctx)
		for _, a := range nodes {
			actors = append(actors, a)
		}
	case typesUtil.UtilActorType_Val:
		vals := GetAllTestingValidators(t, ctx)
		for _, a := range vals {
			actors = append(actors, a)
		}
	case typesUtil.UtilActorType_Fish:
		fish := GetAllTestingFish(t, ctx)
		for _, a := range fish {
			actors = append(actors, a)
		}
	default:
		t.Fatalf("unexpected actor type %s", actorType.GetActorName())
	}

	return
}

func GetFirstActor(t *testing.T, ctx utility.UtilityContext, actorType typesUtil.UtilActorType) modules.Actor {
	return GetAllTestingActors(t, ctx, actorType)[0]
}

func GetActorByAddr(t *testing.T, ctx utility.UtilityContext, addr []byte, actorType typesUtil.UtilActorType) (actor modules.Actor) {
	actors := GetAllTestingActors(t, ctx, actorType)
	for _, a := range actors {
		if a.GetAddress() == hex.EncodeToString(addr) {
			return a
		}
	}
	return
}

func GetAllTestingApps(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllApps(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingValidators(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllValidators(ctx.LatestHeight)
	require.NoError(t, err)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].GetAddress() < actors[j].GetAddress()
	})
	return actors
}

func GetAllTestingFish(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllFishermen(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}

func GetAllTestingNodes(t *testing.T, ctx utility.UtilityContext) []modules.Actor {
	actors, err := (ctx.Context.PersistenceRWContext).GetAllServiceNodes(ctx.LatestHeight)
	require.NoError(t, err)
	return actors
}
