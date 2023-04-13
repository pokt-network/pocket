package utility

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/require"
)

// TECH_DEBT_IDENTIFIED_IN_THIS_COMMIT:
// 1. Replace []byte with string
// 2. Remove height from Write context in persistence
// 3. Need to add geozone to actors
// 4. Need to generalize persitence functions based on actor type
// 5. Need different protos for each actor

func TestSession_NewSession(t *testing.T) {
	height := int64(1)
	relayChain := coreTypes.RelayChain_ETHEREUM
	geoZone := "geo"

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

func TestSession_SessionHeight(t *testing.T) {
	_, _, persistenceMod := prepareEnvironment(t, 5, 1, 1, 1)

	readCtx, err := persistenceMod.NewReadContext(-1)
	require.NoError(t, err)
	defer readCtx.Release()

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

// not enough servicers to choose from

// no fisherman available

// validate application dispatch

// Not enough servicers

// What if someone paused mid session?

// stake a new servicer -> do I get them?

// Invalid application

// Not enough servicers in region

// No fisherman available

// Check session block height

// Configurable number of geo zones per session
// above max
// below max
// at max

// invalid relay chain
// valid relay chain
// application is not staked for relay chain

// dispatching session in the future
// dispatching session in the past

// generate session id
// validate entropy and randomness

func TestSession_MatchNewSession(t *testing.T) {
}
