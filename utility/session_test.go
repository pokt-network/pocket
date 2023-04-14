package utility

import (
	"testing"

	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

func TestSession_NewSession_SimpleCase(t *testing.T) {
	height := int64(1)
	relayChain := "0001"
	geoZone := "unused_geo"

	runtimeCfg, utilityMod, _ := prepareEnvironment(t, 5, 1, 1, 1)
	require.Len(t, runtimeCfg.GetGenesis().Applications, 1)
	app := runtimeCfg.GetGenesis().Applications[0]

	session, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
	require.NoError(t, err)
	require.Equal(t, "8b50d1f751029a06d0b860e3b900163b3c6532fc48df2e11f848600019df5483", session.Id)
	require.Equal(t, height, session.Height)
	require.Equal(t, relayChain, session.RelayChain)
	require.Equal(t, geoZone, session.GeoZone)
	require.Equal(t, session.Application.Address, app.Address)
	require.Equal(t, "c7832263600476fd6ff4c5cb0a86080d0e5f48b2", session.Servicers[0].Address)
	require.Equal(t, "a6e7b6810df8120580f2a81710e228f454f99c97", session.Fishermen[0].Address)
}

// dispatching session in the future
// dispatching session in the past

// not enough servicers to choose from
// no fisherman available
// validate application dispatch

func TestSession_ServicersAndFishermanCount(t *testing.T) {
	// Prepare an environment with lots of servicers and fisherman
	_, _, persistenceMod := prepareEnvironment(t, 5, 100, 1, 100)

	writeCtx, err := persistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer writeCtx.Release()

	tests := []struct {
		name                   string
		numServicersPerSession int64
		numFishermanPerSession int64
		wantServicerCount      int64
		wantFishermanCount     int64
	}{
		{
			name:                   "block is at start of first session",
			numServicersPerSession: 5,
			numFishermanPerSession: 5,
			wantServicerCount:      5,
			wantFishermanCount:     1,
		},
		{
			name:                   "block is right before start of first session",
			numServicersPerSession: 5,
			numFishermanPerSession: 4,
			wantServicerCount:      0,
			wantFishermanCount:     0,
		},
		{
			name:                   "block is right after start of first session",
			numServicersPerSession: 5,
			numFishermanPerSession: 6,
			wantServicerCount:      5,
			wantFishermanCount:     1,
		},
		{
			name:                   "block is at start of second session",
			numServicersPerSession: 5,
			numFishermanPerSession: 10,
			wantServicerCount:      10,
			wantFishermanCount:     2,
		},
		// Not enough servicers in region
		// Not enough fisherman in region
		// Not enough servicers per chain
		// Not enough fisherman per chain
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeCtx.SetParam(types.BlocksPerSessionParamName, tt.numBlocksPerSession)
			// require.NoError(t, writeCtx.Commit([]byte(""), []byte("")))
			sessionHeight, sessionNumber, err := getSessionHeight(writeCtx, tt.haveBlockHeight)
			require.NoError(t, err)
			require.Equal(t, tt.wantSessionHeight, sessionHeight)
			require.Equal(t, tt.wantSessionNumber, sessionNumber)
		})
	}
}

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
			writeCtx.SetParam(types.BlocksPerSessionParamName, tt.numBlocksPerSession)
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

// Potential: Changing num blocks per session must wait until current session ends -> easily fixes things
// New servicers / fisherman -> need to wait until current session ends -> easily fixes things

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

// Configurable number of geo zones per session
// above max
// below max
// at max
// application is not staked for relay chain
