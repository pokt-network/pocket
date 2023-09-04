package ibc

import (
	"errors"
	"testing"
	"time"

	ics23 "github.com/cosmos/ics23/go"
	light_client_types "github.com/pokt-network/pocket/ibc/client/light_clients/types"
	client_types "github.com/pokt-network/pocket/ibc/client/types"
	"github.com/pokt-network/pocket/ibc/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHost_GetCurrentHeight(t *testing.T) {
	_, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	// get the current height
	height, err := cm.GetCurrentHeight()
	require.NoError(t, err)
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(0), height.GetRevisionHeight())

	// increment the height
	publishNewHeightEvent(t, ibcMod.GetBus(), 1)

	height, err = cm.GetCurrentHeight()
	require.NoError(t, err)
	require.Equal(t, uint64(1), height.GetRevisionNumber())
	require.Equal(t, uint64(1), height.GetRevisionHeight())
}

func TestHost_GetHostConsensusState(t *testing.T) {
	_, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	consState, err := cm.GetHostConsensusState(&client_types.Height{RevisionNumber: 1, RevisionHeight: 0})
	require.NoError(t, err)

	require.Equal(t, "08-wasm", consState.ClientType())
	require.NoError(t, consState.ValidateBasic())
	require.Less(t, consState.GetTimestamp(), uint64(time.Now().UnixNano()))

	pocketConState := new(light_client_types.PocketConsensusState)
	err = codec.GetCodec().Unmarshal(consState.GetData(), pocketConState)
	require.NoError(t, err)

	blockstore := ibcMod.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockstore.GetBlock(0)
	require.NoError(t, err)

	require.Equal(t, block.BlockHeader.Timestamp, pocketConState.Timestamp)
	require.Equal(t, block.BlockHeader.StateHash, pocketConState.StateHash)
	require.Equal(t, block.BlockHeader.StateTreeHashes, pocketConState.StateTreeHashes)
	require.Equal(t, block.BlockHeader.NextValSetHash, pocketConState.NextValSetHash)
}

func TestHost_GetHostClientState(t *testing.T) {
	_, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	clientState, err := cm.GetHostClientState(&client_types.Height{RevisionNumber: 1, RevisionHeight: 0})
	require.NoError(t, err)
	require.Equal(t, "08-wasm", clientState.ClientType())

	pocketClientState := new(light_client_types.PocketClientState)
	err = codec.GetCodec().Unmarshal(clientState.GetData(), pocketClientState)
	require.NoError(t, err)

	blockstore := ibcMod.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockstore.GetBlock(0)
	require.NoError(t, err)

	require.Equal(t, pocketClientState.NetworkId, block.BlockHeader.NetworkId)
	require.Equal(t, pocketClientState.TrustLevel, &light_client_types.Fraction{Numerator: 2, Denominator: 3})
	require.Equal(t, pocketClientState.TrustingPeriod.AsDuration().Nanoseconds(), int64(1814400000000000))
	require.Equal(t, pocketClientState.UnbondingPeriod.AsDuration().Nanoseconds(), int64(1814400000000000))
	require.Equal(t, pocketClientState.MaxClockDrift.AsDuration().Nanoseconds(), int64(900000000000))
	require.Equal(t, pocketClientState.LatestHeight, &client_types.Height{RevisionNumber: 1, RevisionHeight: 0})
	require.True(t, pocketClientState.ProofSpec.ConvertToIcs23ProofSpec().SpecEquals(ics23.SmtSpec))
}

func TestHost_VerifyHostClientState(t *testing.T) {
	_, _, _, persistenceMod, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	cm := ibcMod.GetBus().GetClientManager()

	approxTime := time.Minute * 15
	unbondingPeriod := time.Duration(1814400000000000) * approxTime
	blockstore := ibcMod.GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockstore.GetBlock(0)
	require.NoError(t, err)

	publishNewHeightEvent(t, ibcMod.GetBus(), 1)

	rwCtx, err := persistenceMod.NewRWContext(1)
	require.NoError(t, err)
	defer rwCtx.Release()
	err = rwCtx.Commit(nil, nil)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		pcs         *light_client_types.PocketClientState
		expectedErr error
	}{
		{
			name: "invalid: frozen client",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec:    types.SmtSpec,
				FrozenHeight: 1,
			},
			expectedErr: errors.New("counterparty client state is frozen"),
		},
		{
			name: "invalid: different network id",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       "not correct",
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state has a different network id"),
		},
		{
			name: "invalid: different revision number",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 0,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state has a different revision number"),
		},
		{
			name: "invalid: equal height",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 1,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state has a height greater than or equal to the host client state"),
		},
		{
			name: "invalid: wrong trust level",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 1, Denominator: 4},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state trust level is not in the accepted range"),
		},
		{
			name: "invalid: different proof spec",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.ConvertFromIcs23ProofSpec(ics23.IavlSpec),
			},
			expectedErr: errors.New("counterparty client state has different proof spec"),
		},
		{
			name: "invalid: different unbonding period",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod + 1),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state has different unbonding period"),
		},
		{
			name: "invalid: unbonding period less than trusting period",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod - 1),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: errors.New("counterparty client state unbonding period is less than trusting period"),
		},
		{
			name: "valid client state",
			pcs: &light_client_types.PocketClientState{
				NetworkId:       block.BlockHeader.NetworkId,
				TrustLevel:      &light_client_types.Fraction{Numerator: 2, Denominator: 3},
				TrustingPeriod:  durationpb.New(unbondingPeriod),
				UnbondingPeriod: durationpb.New(unbondingPeriod),
				MaxClockDrift:   durationpb.New(approxTime),
				LatestHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
				ProofSpec: types.SmtSpec,
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bz, err := codec.GetCodec().Marshal(tc.pcs)
			require.NoError(t, err)
			clientState := &client_types.ClientState{
				Data: bz,
				RecentHeight: &client_types.Height{
					RevisionNumber: 1,
					RevisionHeight: 0,
				},
			}
			err = cm.VerifyHostClientState(clientState)
			require.ErrorAs(t, err, &tc.expectedErr)
		})
	}
}
