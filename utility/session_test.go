package utility

import (
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

// TECHDEBT: Geozones are not current implemented, used or tested

// One fishermen and one servicer happy base case
func TestSession_NewSession_BaseCase(t *testing.T) {
	// Test parameters
	height := int64(1)
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"
	numFishermen := 1
	numServicers := 1
	expectedSessionId := "3545185ff1519bf7706ec8f828d16525830d3c0dcc2425c40db597ee6b67b8bc" // needs to be manually updated if business logic changes

	runtimeCfg, utilityMod, _ := prepareEnvironment(t, 5, numServicers, 1, numFishermen)

	// Sanity check genesis
	require.Len(t, runtimeCfg.GetGenesis().Applications, 1)
	app := runtimeCfg.GetGenesis().Applications[0]
	require.Len(t, runtimeCfg.GetGenesis().Fishermen, 1)
	fish := runtimeCfg.GetGenesis().Fishermen[0]
	require.Len(t, runtimeCfg.GetGenesis().Servicers, 1)
	servicer := runtimeCfg.GetGenesis().Servicers[0]

	session, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
	require.NoError(t, err)
	require.Equal(t, expectedSessionId, session.Id)
	require.Equal(t, height, session.Height)
	require.Equal(t, relayChain, session.RelayChain)
	require.Equal(t, geoZone, session.GeoZone)
	require.Equal(t, app.Address, session.Application.Address)
	require.Len(t, session.Servicers, numServicers)
	require.Equal(t, servicer.Address, session.Servicers[0].Address)
	require.Len(t, session.Fishermen, numFishermen)
	require.Equal(t, fish.Address, session.Fishermen[0].Address)
}

func TestSession_ServicersAndFishermanCounts_TotalAvailability(t *testing.T) {
	numServicers := 100
	numFishermen := 100
	runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, numServicers, 1, numFishermen)

	app := runtimeCfg.GetGenesis().Applications[0]
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"

	tests := []struct {
		name                   string
		numServicersPerSession int64
		numFishermanPerSession int64
		wantServicerCount      int
		wantFishermanCount     int
	}{
		{
			name:                   "more actors per session than available in network",
			numServicersPerSession: int64(numServicers) * 10,
			numFishermanPerSession: int64(numFishermen) * 10,
			wantServicerCount:      numServicers,
			wantFishermanCount:     numFishermen,
		},
		{
			name:                   "less actors per session than available in network",
			numServicersPerSession: int64(numServicers) / 2,
			numFishermanPerSession: int64(numFishermen) / 2,
			wantServicerCount:      numServicers / 2,
			wantFishermanCount:     numFishermen / 2,
		},
		{
			name:                   "same number of actors per session as available in network",
			numServicersPerSession: int64(numServicers),
			numFishermanPerSession: int64(numFishermen),
			wantServicerCount:      numServicers,
			wantFishermanCount:     numFishermen,
		},
		{
			name:                   "more than enough servicers but not enough fishermen",
			numServicersPerSession: int64(numServicers) / 2,
			numFishermanPerSession: int64(numFishermen) * 10,
			wantServicerCount:      numServicers / 2,
			wantFishermanCount:     numFishermen,
		},
		{
			name:                   "more than enough fishermen but not enough servicers",
			numServicersPerSession: int64(numServicers) * 10,
			numFishermanPerSession: int64(numFishermen) / 2,
			wantServicerCount:      numServicers,
			wantFishermanCount:     numFishermen / 2,
		},
	}

	updateParamsHeight := int64(1)
	querySessionHeight := int64(2)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persistenceMod.HandleDebugMessage(&messaging.DebugMessage{
				Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
				Message: nil,
			})
			require.NoError(t, err)

			writeCtx, err := persistenceMod.NewRWContext(updateParamsHeight)
			require.NoError(t, err)
			defer writeCtx.Release()

			err = writeCtx.SetParam(types.ServicersPerSessionParamName, tt.numServicersPerSession)
			require.NoError(t, err)
			err = writeCtx.SetParam(types.FishermanPerSessionParamName, tt.numFishermanPerSession)
			require.NoError(t, err)
			err = writeCtx.Commit([]byte("empty_proposed_addr"), []byte("empty_quorum_cert"))
			require.NoError(t, err)

			session, err := utilityMod.GetSession(app.Address, querySessionHeight, relayChain, geoZone)
			require.NoError(t, err)
			require.Equal(t, tt.wantServicerCount, len(session.Servicers))
			require.Equal(t, tt.wantFishermanCount, len(session.Fishermen))
		})
	}
}

func TestSession_ServicersAndFishermanCounts_ChainAvailability(t *testing.T) {
	// numServicers := 100
	// numFishermen := 100
	// runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, numServicers, 1, numFishermen)

	// app := runtimeCfg.GetGenesis().Applications[0]
	// relayChain := test_artifacts.DefaultChains[0]
	// geoZone := "unused_geo"

	// tests := []struct {
	// 	name                   string
	// 	numServicersPerSession int64
	// 	numFishermanPerSession int64
	// 	wantServicerCount      int
	// 	wantFishermanCount     int
	// }{
	// 	{
	// 		name:                   "more actors per session than available in network",
	// 		numServicersPerSession: int64(numServicers) * 10,
	// 		numFishermanPerSession: int64(numFishermen) * 10,
	// 		wantServicerCount:      numServicers,
	// 		wantFishermanCount:     numFishermen,
	// 	},
	// 	{
	// 		name:                   "less actors per session than available in network",
	// 		numServicersPerSession: int64(numServicers) / 2,
	// 		numFishermanPerSession: int64(numFishermen) / 2,
	// 		wantServicerCount:      numServicers / 2,
	// 		wantFishermanCount:     numFishermen / 2,
	// 	},
	// 	{
	// 		name:                   "same number of actors per session as available in network",
	// 		numServicersPerSession: int64(numServicers),
	// 		numFishermanPerSession: int64(numFishermen),
	// 		wantServicerCount:      numServicers,
	// 		wantFishermanCount:     numFishermen,
	// 	},
	// }

	// updateParamsHeight := int64(1)
	// querySessionHeight := int64(2)

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		err := persistenceMod.HandleDebugMessage(&messaging.DebugMessage{
	// 			Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
	// 			Message: nil,
	// 		})
	// 		require.NoError(t, err)

	// 		writeCtx, err := persistenceMod.NewRWContext(updateParamsHeight)
	// 		require.NoError(t, err)
	// 		defer writeCtx.Release()

	// 		err = writeCtx.SetParam(types.ServicersPerSessionParamName, tt.numServicersPerSession)
	// 		require.NoError(t, err)
	// 		err = writeCtx.SetParam(types.FishermanPerSessionParamName, tt.numFishermanPerSession)
	// 		require.NoError(t, err)
	// 		err = writeCtx.Commit([]byte("empty_proposed_addr"), []byte("empty_quorum_cert"))
	// 		require.NoError(t, err)

	// 		session, err := utilityMod.GetSession(app.Address, querySessionHeight, relayChain, geoZone)
	// 		require.NoError(t, err)
	// 		require.Equal(t, tt.wantServicerCount, len(session.Servicers))
	// 		require.Equal(t, tt.wantFishermanCount, len(session.Fishermen))
	// 	})
	// }
}

// Not enough servicers in region
// Not enough fisherman in region
// Not enough servicers per chain
// Not enough fisherman per chain

// func TestSession_NewSession_BaseCase(t *testing.T) {

// dispatching session in the future
// dispatching session in the past

// not enough servicers to choose from
// no fisherman available
// validate application dispatch

// invalid app
// unstaked app
// non-existent app

// invalid chain
// unused chain
// non-existent chain

// invalid geozone
// unused geozone
// non-existent geozone

// generate session id

func TestSession_ServicersAndFishermanRandomness(t *testing.T) {
	// validate entropy and randomness
	// different height
	// different chain
}

func TestSession_SessionHeightAndNumber_StaticBlocksPerSession(t *testing.T) {
	_, _, persistenceMod := prepareEnvironment(t, 5, 1, 1, 1)

	writeCtx, err := persistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer writeCtx.Release()

	tests := []struct {
		name                string
		numBlocksPerSession int64
		haveBlockHeight     int64
		wantSessionHeight   int64
		wantSessionNumber   int64
	}{
		{
			name:                "block is at start of first session",
			numBlocksPerSession: 5,
			haveBlockHeight:     5,
			wantSessionHeight:   5,
			wantSessionNumber:   1,
		},
		{
			name:                "block is right before start of first session",
			numBlocksPerSession: 5,
			haveBlockHeight:     4,
			wantSessionHeight:   0,
			wantSessionNumber:   0,
		},
		{
			name:                "block is right after start of first session",
			numBlocksPerSession: 5,
			haveBlockHeight:     6,
			wantSessionHeight:   5,
			wantSessionNumber:   1,
		},
		{
			name:                "block is at start of second session",
			numBlocksPerSession: 5,
			haveBlockHeight:     10,
			wantSessionHeight:   10,
			wantSessionNumber:   2,
		},
		// TODO: Different blocks per session
		// What if we change the num blocks -> gets complex
		// -> Need to enforce waiting until the end of the current sessions
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeCtx.SetParam(types.BlocksPerSessionParamName, tt.numBlocksPerSession)
			require.NoError(t, err)
			// require.NoError(t, writeCtx.Commit([]byte(""), []byte("")))

			sessionHeight, sessionNumber, err := getSessionHeight(writeCtx, tt.haveBlockHeight)
			require.NoError(t, err)
			require.Equal(t, tt.wantSessionHeight, sessionHeight)
			require.Equal(t, tt.wantSessionNumber, sessionNumber)
		})
	}
}

func TestSession_SessionHeightAndNumber_DynamicBlocksPerSession(t *testing.T) {

}

func TestSession_MatchNewSession(t *testing.T) {
}

func TestSession_RelayChainVariability(t *testing.T) {
	// invalid relay chain
	// valid relay chain

}

func TestSession_ActorReplacement(t *testing.T) {
	// What if a servicers/fisherman paused mid session?
	// -> Need to replace them

	// What if a new servicers/fisherman staked mid session?
	// -> They could potentially get selected
}

func TestSession_InvalidApplication(t *testing.T) {
	// TODO: What if the application pauses mid session?
	// TODO: What if the application has no stake?
}

// Potential: Changing num blocks per session must wait until current session ends -> easily fixes things
// New servicers / fisherman -> need to wait until current session ends -> easily fixes things

// Configurable number of geo zones per session
// above max
// below max
// at max
// application is not staked for relay chain
