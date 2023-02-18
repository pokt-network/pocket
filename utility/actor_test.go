package utility

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	typesUtil "github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestUtilityContext_HandleMessageStake(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.HandleMessageStake", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)

			pubKey, err := crypto.GeneratePublicKey()
			require.NoError(t, err)

			outputAddress, err := crypto.GenerateAddress()
			require.NoError(t, err)

			err = ctx.setAccountAmount(outputAddress, test_artifacts.DefaultAccountAmount)
			require.NoError(t, err, "error setting account amount error")

			msg := &typesUtil.MessageStake{
				PublicKey:     pubKey.Bytes(),
				Chains:        test_artifacts.DefaultChains,
				Amount:        test_artifacts.DefaultStakeAmountString,
				ServiceUrl:    test_artifacts.DefaultServiceURL,
				OutputAddress: outputAddress,
				Signer:        outputAddress,
				ActorType:     actorType,
			}

			err = ctx.handleStakeMessage(msg)
			require.NoError(t, err)

			actor := getActorByAddr(t, ctx, actorType, pubKey.Address().String())
			require.Equal(t, actor.GetAddress(), pubKey.Address().String(), "incorrect actor address")
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetPausedHeight(), "incorrect actorpaused height")
			require.Equal(t, test_artifacts.DefaultStakeAmountString, actor.GetStakedAmount(), "incorrect actor stake amount")
			require.Equal(t, outputAddress.String(), actor.GetOutput(), "incorrect actor output address")
			if actorType != coreTypes.ActorType_ACTOR_TYPE_VAL {
				require.Equal(t, msg.Chains, actor.GetChains(), "incorrect actor chains")
			}
		})
	}
}

func TestUtilityContext_HandleMessageEditStake(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.HandleMessageEditStake", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			// Edit the staked chains
			msg := &typesUtil.MessageEditStake{
				Address:   addrBz,
				Chains:    test_artifacts.DefaultChains,
				Amount:    test_artifacts.DefaultStakeAmountString,
				Signer:    addrBz,
				ActorType: actorType,
			}
			msgChainsEdited := codec.GetCodec().Clone(msg).(*typesUtil.MessageEditStake)
			msgChainsEdited.Chains = []string{"0002"}
			require.NotEqual(t, msgChainsEdited.Chains, test_artifacts.DefaultChains) // sanity check to make sure the test makes sense

			err = ctx.handleEditStakeMessage(msgChainsEdited)
			require.NoError(t, err)

			// Verify the chains were edited
			actor = getActorByAddr(t, ctx, actorType, addr)
			if actorType != coreTypes.ActorType_ACTOR_TYPE_VAL {
				require.NotEqual(t, test_artifacts.DefaultChains, actor.GetChains(), "incorrect edited chains")
				require.Equal(t, msgChainsEdited.Chains, actor.GetChains(), "incorrect edited chains")
			}

			// Edit the staked amount
			amountEdited := test_artifacts.DefaultAccountAmount.Add(test_artifacts.DefaultAccountAmount, big.NewInt(1))
			amountEditedString := converters.BigIntToString(amountEdited)

			msgAmountEdited := codec.GetCodec().Clone(msg).(*typesUtil.MessageEditStake)
			msgAmountEdited.Amount = amountEditedString

			// Verify the staked amount was edited
			err = ctx.handleEditStakeMessage(msgAmountEdited)
			require.NoError(t, err, "handle edit stake message")

			actor = getActorByAddr(t, ctx, actorType, addr)
			require.NotEqual(t, test_artifacts.DefaultStakeAmountString, actor.GetStakedAmount(), "incorrect edited amount staked")
			require.Equal(t, amountEditedString, actor.StakedAmount, "incorrect edited amount staked")
		})
	}
}

func TestUtilityContext_HandleMessageUnstake(t *testing.T) {
	// The gov param for each actor will be set to this value
	numUnstakingBlocks := 5

	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.HandleMessageUnstake", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 1)

			var paramName string
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				paramName = typesUtil.AppUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				paramName = typesUtil.FishermanUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				paramName = typesUtil.ServicerUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				paramName = typesUtil.ValidatorUnstakingBlocksParamName
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			err := ctx.persistenceContext.SetParam(paramName, numUnstakingBlocks)
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}

			// Unstake the actor
			err = ctx.handleUnstakeMessage(msg)
			require.NoError(t, err, "handle unstake message")

			// Verify the unstaking height is correct
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, int64(numUnstakingBlocks)+1, actor.GetUnstakingHeight(), "actor should be unstaking")
		})
	}
}

func TestUtilityContext_HandleMessageUnpause(t *testing.T) {
	// The gov param for each actor will be set to this value
	minPauseBlocksNumber := 5

	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.HandleMessageUnpause", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 1)

			var paramName string
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				paramName = typesUtil.AppMinimumPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				paramName = typesUtil.FishermanMinimumPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				paramName = typesUtil.ServicerMinimumPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				paramName = typesUtil.ValidatorMinimumPauseBlocksParamName
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			err := ctx.persistenceContext.SetParam(paramName, minPauseBlocksNumber)
			require.NoError(t, err, "error setting minimum pause blocks")

			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			// Pause the actor
			err = ctx.setActorPausedHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting pause height")

			// Verify the actor is paused
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, int64(1), actor.GetPausedHeight())

			// Try to unpause the actor and verify that it fails
			msgUnpauseActor := &typesUtil.MessageUnpause{
				Address:   addrBz,
				Signer:    addrBz,
				ActorType: actorType,
			}
			err = ctx.handleUnpauseMessage(msgUnpauseActor)
			require.Error(t, err)
			require.ErrorContains(t, err, "minimum number of blocks hasn't passed since pausing")

			// Start a new context when the actor still cannot be unpaused
			require.NoError(t, ctx.Commit([]byte("empty qc")))
			require.NoError(t, ctx.Release())
			ctx = newTestingUtilityContext(t, int64(minPauseBlocksNumber)-1)

			// Try to unpause the actor
			err = ctx.handleUnpauseMessage(msgUnpauseActor)
			require.Error(t, err)
			require.ErrorContains(t, err, "minimum number of blocks hasn't passed since pausing")

			// Verify the actor is still paused
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.NotEqual(t, typesUtil.HeightNotUsed, actor.GetPausedHeight())

			// Start a new context when the actor can be unpaused
			require.Error(t, ctx.Commit([]byte("empty qc"))) // Nothing to commit so we expect an error
			require.NoError(t, ctx.Release())
			ctx = newTestingUtilityContext(t, int64(minPauseBlocksNumber)+1)

			// Try to unpause the actor
			err = ctx.handleUnpauseMessage(msgUnpauseActor)
			require.NoError(t, err)

			// Verify the actor is still paused
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetPausedHeight())
		})
	}
}

func TestUtilityContext_GetUnbondingHeight(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.CalculateUnstakingHeight", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)

			var unstakingBlocks int64
			var err error
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				unstakingBlocks, err = ctx.getAppUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				unstakingBlocks, err = ctx.getFishermanUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				unstakingBlocks, err = ctx.getServicerUnstakingBlocks()
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				unstakingBlocks, err = ctx.getValidatorUnstakingBlocks()
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			require.NoError(t, err, "error getting unstaking blocks")

			unbondingHeight, err := ctx.getUnbondingHeight(actorType)
			require.NoError(t, err)
			require.Equal(t, unstakingBlocks, unbondingHeight, "unexpected unstaking height")
		})
	}
}

func TestUtilityContext_BeginUnstakingMaxPausedActors(t *testing.T) {
	// The gov param for each actor will be set to this value
	maxPausedBlocks := 5

	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.BeginUnstakingMaxPausedActors", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 1)

			var paramName string
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				paramName = typesUtil.AppMaxPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				paramName = typesUtil.FishermanMaxPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				paramName = typesUtil.ServicerMaxPauseBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				paramName = typesUtil.ValidatorMaxPausedBlocksParamName
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}
			err := ctx.persistenceContext.SetParam(paramName, maxPausedBlocks)
			require.NoError(t, err)

			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			// Pause the actor at height 0
			err = ctx.setActorPausedHeight(actorType, addrBz, 1)
			require.NoError(t, err, "error setting actor pause height")

			// Start unstaking paused actors at the current height
			err = ctx.beginUnstakingMaxPausedActors()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			// Verify that the actor is still staked
			status, err := ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Staked, status, "actor should be staked")

			// Start a new context when the actor still shouldn't be unstaked
			require.NoError(t, ctx.Commit([]byte("empty qc")))
			require.NoError(t, ctx.Release())
			ctx = newTestingUtilityContext(t, int64(maxPausedBlocks)-1)

			// Start unstaking paused actors at the current height
			err = ctx.beginUnstakingMaxPausedActors()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			// Verify that the actor is still staked
			status, err = ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Staked, status, "actor should be staked")

			// Start a new context when the actor should be unstaked
			require.NoError(t, ctx.Release())
			ctx = newTestingUtilityContext(t, int64(maxPausedBlocks)+2)

			// Start unstaking paused actors at the current height
			err = ctx.beginUnstakingMaxPausedActors()
			require.NoError(t, err, "error beginning unstaking max paused actors")

			// Verify that the actor is still staked
			status, err = ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Unstaking, status, "actor should be staked")
		})
	}
}

func TestUtilityContext_BeginUnstakingActorsPausedBefore_UnbondUnstakingActors(t *testing.T) {
	// The gov param for each actor will be set to this value
	maxPausedBlocks := 5
	unstakingBlocks := 10

	poolInitAMount := big.NewInt(10000000000000)
	pauseHeight := int64(2)

	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.BeginUnstakingActorsPausedBefore", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 1)

			var poolName string
			var paramName1 string
			var paramName2 string
			switch actorType {
			case coreTypes.ActorType_ACTOR_TYPE_APP:
				poolName = coreTypes.Pools_POOLS_APP_STAKE.FriendlyName()
				paramName1 = typesUtil.AppMaxPauseBlocksParamName
				paramName2 = typesUtil.AppUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_FISH:
				poolName = coreTypes.Pools_POOLS_FISHERMAN_STAKE.FriendlyName()
				paramName1 = typesUtil.FishermanMaxPauseBlocksParamName
				paramName2 = typesUtil.FishermanUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
				poolName = coreTypes.Pools_POOLS_SERVICE_NODE_STAKE.FriendlyName()
				paramName1 = typesUtil.ServicerMaxPauseBlocksParamName
				paramName2 = typesUtil.ServicerUnstakingBlocksParamName
			case coreTypes.ActorType_ACTOR_TYPE_VAL:
				poolName = coreTypes.Pools_POOLS_VALIDATOR_STAKE.FriendlyName()
				paramName1 = typesUtil.ValidatorMaxPausedBlocksParamName
				paramName2 = typesUtil.ValidatorUnstakingBlocksParamName
			default:
				t.Fatalf("unexpected actor type %s", actorType.String())
			}

			er := ctx.persistenceContext.SetParam(paramName1, maxPausedBlocks)
			require.NoError(t, er, "error setting max paused blocks")
			er = ctx.persistenceContext.SetParam(paramName2, unstakingBlocks)
			require.NoError(t, er, "error setting max paused blocks")
			er = ctx.setPoolAmount(poolName, poolInitAMount)
			require.NoError(t, er)

			// Validate the actor is not unstaking
			actor := getFirstActor(t, ctx, actorType)
			addr := actor.GetAddress()
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "wrong starting status")

			addrBz, err := hex.DecodeString(addr)
			require.NoError(t, err)

			// Set the actor to be paused at height 1
			err = ctx.setActorPausedHeight(actorType, addrBz, pauseHeight)
			require.NoError(t, err, "error setting actor pause height")

			// Check that the actor is still not unstaking
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "incorrect unstaking height")

			// Verify that the actor is still not unstaking
			err = ctx.beginUnstakingActorsPausedBefore(pauseHeight-1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 0")
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, typesUtil.HeightNotUsed, actor.GetUnstakingHeight(), "incorrect unstaking height")

			unbondingHeight := pauseHeight - 1 + int64(unstakingBlocks)

			// Verify that the actor is now unstaking
			err = ctx.beginUnstakingActorsPausedBefore(pauseHeight+1, actorType)
			require.NoError(t, err, "error unstaking actor pause before height 1")
			actor = getActorByAddr(t, ctx, actorType, addr)
			require.Equal(t, unbondingHeight, actor.GetUnstakingHeight(), "incorrect unstaking height")

			status, err := ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Unstaking, status, "actor should be unstaking")

			// Commit the context and start a new one while the actor is still unstaking
			require.NoError(t, ctx.Commit([]byte("empty QC")))
			ctx = newTestingUtilityContext(t, unbondingHeight-1)

			status, err = ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Unstaking, status, "actor should be unstaking")

			// Release the context since there's nothing to commit and start a new one where the actors can be unbound
			require.NoError(t, ctx.Release())
			ctx = newTestingUtilityContext(t, unbondingHeight)

			// Before unbonding, the pool amount should be unchanged
			amount, err := ctx.getPoolAmount(poolName)
			require.NoError(t, err)
			require.Equal(t, poolInitAMount, amount, "pool amount should be unchanged")

			err = ctx.unbondUnstakingActors()
			require.NoError(t, err)

			// Before unbonding, the money from the staked actor should go to the pool
			amount, err = ctx.getPoolAmount(poolName)
			require.NoError(t, err)

			stakedAmount, err := converters.StringToBigInt(actor.StakedAmount)
			require.NoError(t, err)
			expectedAmount := big.NewInt(0).Sub(poolInitAMount, stakedAmount)
			require.Equalf(t, expectedAmount, amount, "pool amount should be unchanged for %s", poolName)

			// Status should be changed from Unstaking to Unstaked
			status, err = ctx.getActorStatus(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.StakeStatus_Unstaked, status, "actor should be unstaking")
		})
	}
}

func TestUtilityContext_GetExists(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetExists", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)

			actor := getFirstActor(t, ctx, actorType)
			randAddr, err := crypto.GenerateAddress()
			require.NoError(t, err)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			exists, err := ctx.getActorExists(actorType, addrBz)
			require.NoError(t, err)
			require.True(t, exists, "actor that should exist does not")

			exists, err = ctx.getActorExists(actorType, randAddr)
			require.NoError(t, err)
			require.False(t, exists, "actor that shouldn't exist does")
		})
	}
}

func TestUtilityContext_GetOutputAddress(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetOutputAddress", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)

			actor := getFirstActor(t, ctx, actorType)
			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			outputAddress, err := ctx.getActorOutputAddress(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(outputAddress), "unexpected output address")
		})
	}
}

func TestUtilityContext_GetPauseHeightIfExists(t *testing.T) {
	pauseHeight := int64(100)

	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetPauseHeightIfExists", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			// Paused height should not exist here
			gotPauseHeight, err := ctx.getPausedHeightIfExists(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, typesUtil.HeightNotUsed, gotPauseHeight)

			// Paused height should be set after this
			err = ctx.setActorPausedHeight(actorType, addrBz, pauseHeight)
			require.NoError(t, err, "error setting actor pause height")

			gotPauseHeight, err = ctx.getPausedHeightIfExists(actorType, addrBz)
			require.NoError(t, err)
			require.Equal(t, pauseHeight, gotPauseHeight, "unable to get pause height from the actor")

			// Random address shouldn't have a paused height
			randAddr, er := crypto.GenerateAddress()
			require.NoError(t, er)

			_, err = ctx.getPausedHeightIfExists(actorType, randAddr)
			require.Error(t, err, "non existent actor should error")
		})
	}
}
func TestUtilityContext_GetMessageEditStakeSignerCandidates(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetMessageEditStakeSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msgEditStake := &typesUtil.MessageEditStake{
				Address:   addrBz,
				Chains:    test_artifacts.DefaultChains,
				Amount:    test_artifacts.DefaultStakeAmountString,
				ActorType: actorType,
			}
			candidates, err := ctx.getMessageEditStakeSignerCandidates(msgEditStake)
			require.NoError(t, err)

			require.Equal(t, 2, len(candidates), "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
		})
	}
}

func TestUtilityContext_GetMessageUnpauseSignerCandidates(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetMessageUnpauseSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msg := &typesUtil.MessageUnpause{
				Address:   addrBz,
				ActorType: actorType,
			}
			candidates, err := ctx.getMessageUnpauseSignerCandidates(msg)
			require.NoError(t, err)

			require.Equal(t, 2, len(candidates), "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
		})
	}
}

func TestUtilityContext_GetMessageUnstakeSignerCandidates(t *testing.T) {
	for actorTypeNum := range coreTypes.ActorType_name {
		if actorTypeNum == 0 { // ACTOR_TYPE_UNSPECIFIED
			continue
		}
		actorType := coreTypes.ActorType(actorTypeNum)

		t.Run(fmt.Sprintf("%s.GetMessageUnstakeSignerCandidates", actorType.String()), func(t *testing.T) {
			ctx := newTestingUtilityContext(t, 0)
			actor := getFirstActor(t, ctx, actorType)

			addrBz, err := hex.DecodeString(actor.GetAddress())
			require.NoError(t, err)

			msg := &typesUtil.MessageUnstake{
				Address:   addrBz,
				ActorType: actorType,
			}
			candidates, err := ctx.getMessageUnstakeSignerCandidates(msg)
			require.NoError(t, err)

			require.Equal(t, 2, len(candidates), "unexpected number of candidates")
			require.Equal(t, actor.GetOutput(), hex.EncodeToString(candidates[0]), "incorrect output candidate")
			require.Equal(t, actor.GetAddress(), hex.EncodeToString(candidates[1]), "incorrect addr candidate")
		})
	}
}

// Helpers

func getAllTestingActors(t *testing.T, ctx *utilityContext, actorType coreTypes.ActorType) (actors []*coreTypes.Actor) {
	actors = make([]*coreTypes.Actor, 0)
	switch actorType {
	case coreTypes.ActorType_ACTOR_TYPE_APP:
		apps := getAllTestingApps(t, ctx)
		actors = append(actors, apps...)
	case coreTypes.ActorType_ACTOR_TYPE_FISH:
		fish := getAllTestingFish(t, ctx)
		actors = append(actors, fish...)
	case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
		nodes := getAllTestingServicers(t, ctx)
		actors = append(actors, nodes...)
	case coreTypes.ActorType_ACTOR_TYPE_VAL:
		vals := getAllTestingValidators(t, ctx)
		actors = append(actors, vals...)
	default:
		t.Fatalf("unexpected actor type %s", actorType.String())
	}

	return
}

func getFirstActor(t *testing.T, ctx *utilityContext, actorType coreTypes.ActorType) *coreTypes.Actor {
	return getAllTestingActors(t, ctx, actorType)[0]
}

func getActorByAddr(t *testing.T, ctx *utilityContext, actorType coreTypes.ActorType, addr string) (actor *coreTypes.Actor) {
	actors := getAllTestingActors(t, ctx, actorType)
	idx := slices.IndexFunc(actors, func(a *coreTypes.Actor) bool { return a.GetAddress() == addr })
	return actors[idx]
}

func getAllTestingApps(t *testing.T, ctx *utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllApps(ctx.height)
	require.NoError(t, err)
	return actors
}

func getAllTestingValidators(t *testing.T, ctx *utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllValidators(ctx.height)
	require.NoError(t, err)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].GetAddress() < actors[j].GetAddress()
	})
	return actors
}

func getAllTestingFish(t *testing.T, ctx *utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllFishermen(ctx.height)
	require.NoError(t, err)
	return actors
}

func getAllTestingServicers(t *testing.T, ctx *utilityContext) []*coreTypes.Actor {
	actors, err := (ctx.persistenceContext).GetAllServicers(ctx.height)
	require.NoError(t, err)
	return actors
}
