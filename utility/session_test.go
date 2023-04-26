package utility

import (
	"fmt"
	"math"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/utility/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/stat/combin"
)

// TECHDEBT: Geozones are not current implemented, used or tested

func TestSession_GetSession_SingleFishermanSingleServicerBaseCase(t *testing.T) {
	// Test parameters
	height := int64(1)
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"
	numFishermen := 1
	numServicers := 1
	expectedSessionId := "5acf559f1a3faf3bea7eb692fe51bc1e2e5fb687ede0a6daa7d42399da4aa82b" // needs to be manually updated if business logic changes

	runtimeCfg, utilityMod, _ := prepareEnvironment(t, 5, numServicers, 1, numFishermen)

	// Sanity check genesis
	require.Len(t, runtimeCfg.GetGenesis().Applications, 1)
	app := runtimeCfg.GetGenesis().Applications[0]
	require.Len(t, runtimeCfg.GetGenesis().Fishermen, 1)
	fish := runtimeCfg.GetGenesis().Fishermen[0]
	require.Len(t, runtimeCfg.GetGenesis().Servicers, 1)
	servicer := runtimeCfg.GetGenesis().Servicers[0]

	// Verify some of the session defaults
	session, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
	require.NoError(t, err)
	require.Equal(t, expectedSessionId, session.Id)
	require.Equal(t, height, session.SessionHeight)
	require.Equal(t, int64(1), session.SessionNumber)
	require.Equal(t, int64(1), session.NumSessionBlocks)
	require.Equal(t, relayChain, session.RelayChain)
	require.Equal(t, geoZone, session.GeoZone)
	require.Equal(t, app.Address, session.Application.Address)
	require.Len(t, session.Servicers, numServicers)
	require.Equal(t, servicer.Address, session.Servicers[0].Address)
	require.Len(t, session.Fishermen, numFishermen)
	require.Equal(t, fish.Address, session.Fishermen[0].Address)
}

func TestSession_GetSession_ApplicationInvalid(t *testing.T) {
	runtimeCfg, utilityMod, _ := prepareEnvironment(t, 5, 1, 1, 1)

	// Verify there's only 1 app
	require.Len(t, runtimeCfg.GetGenesis().Applications, 1)
	app := runtimeCfg.GetGenesis().Applications[0]

	// Create a new app address
	pk, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Verify that the one app in the genesis is not the one we just generated
	addr := pk.Address().String()
	require.NotEqual(t, app.Address, addr)

	// Expect no error trying to get a session for the real application
	_, err = utilityMod.GetSession(app.Address, 1, test_artifacts.DefaultChains[0], "unused_geo")
	require.NoError(t, err)

	// Expect an error trying to get a session for an unstaked chain
	_, err = utilityMod.GetSession(addr, 1, "chain", "unused_geo")
	require.Error(t, err)

	// Expect an error trying to get a session for a non-existent application
	_, err = utilityMod.GetSession(addr, 1, test_artifacts.DefaultChains[0], "unused_geo")
	require.Error(t, err)
}

func TestSession_GetSession_InvalidFutureSession(t *testing.T) {
	runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, 1, 1, 1)

	// Test parameters
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"
	app := runtimeCfg.GetGenesis().Applications[0]

	// Local variable to keep track of the height we're getting a session for
	currentHeight := int64(0)

	// Successfully get a session for 1 block ahead of the latest committed height
	session, err := utilityMod.GetSession(app.Address, currentHeight+1, relayChain, geoZone)
	require.NoError(t, err)
	require.Equal(t, currentHeight+1, session.SessionHeight)

	// Expect an error for a few heights into the future
	for height := currentHeight + 2; height < 10; height++ {
		_, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
		require.Error(t, err)
	}

	// Commit new blocks for all the heights that failed above
	for ; currentHeight < 10; currentHeight++ {
		writeCtx, err := persistenceMod.NewRWContext(currentHeight + 1)
		require.NoError(t, err)
		err = writeCtx.Commit([]byte(fmt.Sprintf("proposer_height_%d", currentHeight)), []byte(fmt.Sprintf("quorum_cert_height_%d", currentHeight)))
		require.NoError(t, err)
		writeCtx.Release()
	}

	// Expect no errors since those blocks exist now
	// Note that we can get the session for latest_committed + 1
	for height := int64(1); height <= currentHeight+1; height++ {
		_, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
		require.NoError(t, err)
	}

	// Verify that currentHeight + 2 fails
	_, err = utilityMod.GetSession(app.Address, currentHeight+2, relayChain, geoZone)
	require.Error(t, err)
}

func TestSession_GetSession_ServicersAndFishermenCounts_TotalAvailability(t *testing.T) {
	// Prepare an environment with a lot of servicers and fishermen
	numServicers := 100
	numFishermen := 100
	runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, numServicers, 1, numFishermen)

	// Vary the number of actors per session using gov params and check that the session is populated with the correct number of actorss
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

	// Test constant parameters
	updateParamsHeight := int64(1)
	querySessionHeight := int64(2)

	app := runtimeCfg.GetGenesis().Applications[0]
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to genesis
			err := persistenceMod.HandleDebugMessage(&messaging.DebugMessage{
				Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
				Message: nil,
			})
			require.NoError(t, err)

			// Update the number of servicers and fishermen per session gov params
			writeCtx, err := persistenceMod.NewRWContext(updateParamsHeight)
			require.NoError(t, err)
			defer writeCtx.Release()

			err = writeCtx.SetParam(types.ServicersPerSessionParamName, tt.numServicersPerSession)
			require.NoError(t, err)
			err = writeCtx.SetParam(types.FishermanPerSessionParamName, tt.numFishermanPerSession)
			require.NoError(t, err)
			err = writeCtx.Commit([]byte("empty_proposed_addr"), []byte("empty_quorum_cert"))
			require.NoError(t, err)

			// Verify that the session is populated with the correct number of actors
			session, err := utilityMod.GetSession(app.Address, querySessionHeight, relayChain, geoZone)
			require.NoError(t, err)
			require.Equal(t, tt.wantServicerCount, len(session.Servicers))
			require.Equal(t, tt.wantFishermanCount, len(session.Fishermen))
		})
	}
}

func TestSession_GetSession_ServicersAndFishermenCounts_ChainAvailability(t *testing.T) {
	// Test constant parameters
	numServicersPerSession := 10
	numFishermenPerSession := 2

	// Make sure there are MORE THAN ENOUGH servicers and fishermen in the network for each session for chain 1
	servicersChain1, servicerKeysChain1 := test_artifacts.NewActors(coreTypes.ActorType_ACTOR_TYPE_SERVICER, numServicersPerSession*2, []string{"chn1"})
	fishermenChain2, fishermenKeysChain2 := test_artifacts.NewActors(coreTypes.ActorType_ACTOR_TYPE_FISH, numFishermenPerSession*2, []string{"chn1"})

	// Make sure there are NOT ENOUGH servicers and fishermen in the network for each session for chain 2
	servicersChain2, servicerKeysChain2 := test_artifacts.NewActors(coreTypes.ActorType_ACTOR_TYPE_SERVICER, numServicersPerSession/2, []string{"chn2"})
	fishermenChain1, fishermenKeysChain1 := test_artifacts.NewActors(coreTypes.ActorType_ACTOR_TYPE_FISH, numFishermenPerSession/2, []string{"chn2"})

	application, applicationKey := test_artifacts.NewActors(coreTypes.ActorType_ACTOR_TYPE_APP, 1, []string{"chn1", "chn2", "chn3"})

	//nolint:gocritic // intentionally not appending result to a new slice
	actors := append(application, append(servicersChain1, append(servicersChain2, append(fishermenChain1, fishermenChain2...)...)...)...)
	//nolint:gocritic // intentionally not appending result to a new slice
	keys := append(applicationKey, append(servicerKeysChain1, append(servicerKeysChain2, append(fishermenKeysChain1, fishermenKeysChain2...)...)...)...)

	// Prepare the environment
	runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, 0, 0, 0, test_artifacts.WithActors(actors, keys))

	// Vary the chain and check the number of fishermen and servicers returned for each one
	tests := []struct {
		name               string
		chain              string
		wantServicerCount  int
		wantFishermanCount int
	}{
		{
			name:               "chn1 has enough servicers and fishermen",
			chain:              "chn1",
			wantServicerCount:  numServicersPerSession,
			wantFishermanCount: numFishermenPerSession,
		},
		{
			name:               "chn2 does not have enough servicers and fishermen",
			chain:              "chn2",
			wantServicerCount:  numServicersPerSession / 2,
			wantFishermanCount: numFishermenPerSession / 2,
		},
		{
			name:               "chn3 has no servicers and fishermen",
			chain:              "chn3",
			wantServicerCount:  0,
			wantFishermanCount: 0,
		},
	}

	// Reset to genesis
	err := persistenceMod.HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	})
	require.NoError(t, err)

	// Update the number of servicers and fishermen per session gov params
	writeCtx, err := persistenceMod.NewRWContext(1)
	require.NoError(t, err)
	err = writeCtx.SetParam(types.ServicersPerSessionParamName, numServicersPerSession)
	require.NoError(t, err)
	err = writeCtx.SetParam(types.FishermanPerSessionParamName, numFishermenPerSession)
	require.NoError(t, err)
	err = writeCtx.Commit([]byte("empty_proposed_addr"), []byte("empty_quorum_cert"))
	require.NoError(t, err)
	defer writeCtx.Release()

	// Test parameters
	app := runtimeCfg.GetGenesis().Applications[0]
	geoZone := "unused_geo"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := utilityMod.GetSession(app.Address, 2, tt.chain, geoZone)
			require.NoError(t, err)
			require.Len(t, session.Servicers, tt.wantServicerCount)
			require.Len(t, session.Fishermen, tt.wantFishermanCount)
		})
	}
}

func TestSession_GetSession_SessionHeightAndNumber_StaticBlocksPerSession(t *testing.T) {
	// Prepare the environment
	_, _, persistenceMod := prepareEnvironment(t, 5, 1, 1, 1)

	// Note that we are using an ephemeral write context at the genesis block (height=0).
	// This cannot be committed but useful for the test.
	writeCtx, err := persistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer writeCtx.Release()

	s := &sessionHydrator{
		session: &coreTypes.Session{},
		readCtx: writeCtx,
	}

	tests := []struct {
		name                   string
		setNumBlocksPerSession int64
		provideBlockHeight     int64
		wantSessionHeight      int64
		wantSessionNumber      int64
	}{
		{
			name:                   "genesis block",
			setNumBlocksPerSession: 5,
			provideBlockHeight:     0,
			wantSessionHeight:      0,
			wantSessionNumber:      0,
		},
		{
			name:                   "block is at start of first session",
			setNumBlocksPerSession: 5,
			provideBlockHeight:     5,
			wantSessionHeight:      5,
			wantSessionNumber:      1,
		},
		{
			name:                   "block is right before start of first session",
			setNumBlocksPerSession: 5,
			provideBlockHeight:     4,
			wantSessionHeight:      0,
			wantSessionNumber:      0,
		},
		{
			name:                   "block is right after start of first session",
			setNumBlocksPerSession: 5,
			provideBlockHeight:     6,
			wantSessionHeight:      5,
			wantSessionNumber:      1,
		},
		{
			name:                   "block is at start of second session",
			setNumBlocksPerSession: 5,
			provideBlockHeight:     10,
			wantSessionHeight:      10,
			wantSessionNumber:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeCtx.SetParam(types.BlocksPerSessionParamName, tt.setNumBlocksPerSession)
			require.NoError(t, err)

			s.blockHeight = tt.provideBlockHeight
			err = s.hydrateSessionMetadata()
			require.NoError(t, err)
			require.Equal(t, tt.setNumBlocksPerSession, s.session.NumSessionBlocks)
			require.Equal(t, tt.wantSessionHeight, s.session.SessionHeight)
			require.Equal(t, tt.wantSessionNumber, s.session.SessionNumber)
		})
	}
}

func TestSession_GetSession_ServicersAndFishermanEntropy(t *testing.T) {
	// Prepare an environment with a lot of servicers and fishermen
	numServicers := 1000
	numFishermen := 1000 // make them equal for simplicity
	numServicersPerSession := 10
	numFishermenPerSession := 10 // make them equal for simplicity
	numApplications := 3
	numBlocksPerSession := 2 // expect a different every other height

	// Determine probability of overlap using combinatorics
	numChoices := combin.GeneralizedBinomial(float64(numServicers), float64(numServicersPerSession))                                 // (numServicers) C (numServicersPerSession)
	numChoicesRemaining := combin.GeneralizedBinomial(float64(numServicers-numServicersPerSession), float64(numServicersPerSession)) // (numServicers - numServicersPerSession) C (numServicersPerSession)
	probabilityOfOverlap := (numChoices - numChoicesRemaining) / numChoices

	// Prepare the environment
	runtimeCfg, utilityMod, persistenceMod := prepareEnvironment(t, 5, numServicers, numApplications, numFishermen)

	// Set the number of servicers and fishermen per session gov params
	writeCtx, err := persistenceMod.NewRWContext(1)
	require.NoError(t, err)
	err = writeCtx.SetParam(types.ServicersPerSessionParamName, numServicersPerSession)
	require.NoError(t, err)
	err = writeCtx.SetParam(types.FishermanPerSessionParamName, numFishermenPerSession)
	require.NoError(t, err)
	err = writeCtx.SetParam(types.BlocksPerSessionParamName, numBlocksPerSession)
	require.NoError(t, err)
	err = writeCtx.Commit([]byte(fmt.Sprintf("proposer_height_%d", 1)), []byte(fmt.Sprintf("quorum_cert_height_%d", 1)))
	require.NoError(t, err)
	writeCtx.Release()

	// Keep the relay chain and geoZone static, but vary the app and height to verify that the servicers and fishermen vary
	relayChain := test_artifacts.DefaultChains[0]
	geoZone := "unused_geo"

	// Sanity check we have 3 apps
	require.Len(t, runtimeCfg.GetGenesis().Applications, numApplications)
	app1 := runtimeCfg.GetGenesis().Applications[0]
	app2 := runtimeCfg.GetGenesis().Applications[1]
	app3 := runtimeCfg.GetGenesis().Applications[2]

	// Keep track of the actors from the session at the previous height to verify a delta
	var app1PrevServicers, app2PrevServicers, app3PrevServicers []*coreTypes.Actor
	var app1PrevFishermen, app2PrevFishermen, app3PrevFishermen []*coreTypes.Actor

	// The number of blocks to increase until we expect a different set of servicers and fishermen; see numBlocksPerSession
	numBlocksUntilChange := 0

	// Commit new blocks for all the heights that failed above
	for height := int64(2); height < 10; height++ {
		session1, err := utilityMod.GetSession(app1.Address, height, relayChain, geoZone)
		require.NoError(t, err)
		session2, err := utilityMod.GetSession(app2.Address, height, relayChain, geoZone)
		require.NoError(t, err)
		session3, err := utilityMod.GetSession(app3.Address, height, relayChain, geoZone)
		require.NoError(t, err)

		// All the sessions have the same number of servicers
		require.Len(t, session1.Servicers, numServicersPerSession)
		require.Equal(t, len(session1.Servicers), len(session2.Servicers))
		require.Equal(t, len(session1.Servicers), len(session3.Servicers))

		// All the sessions have the same number of fishermen
		require.Len(t, session1.Fishermen, numFishermenPerSession)
		require.Equal(t, len(session1.Fishermen), len(session2.Fishermen))
		require.Equal(t, len(session1.Fishermen), len(session3.Fishermen))

		// Assert different services between apps
		assertActorsDifference(t, session1.Servicers, session2.Servicers, probabilityOfOverlap)
		assertActorsDifference(t, session1.Servicers, session3.Servicers, probabilityOfOverlap)

		// Assert different fishermen between apps
		assertActorsDifference(t, session1.Fishermen, session2.Fishermen, probabilityOfOverlap)
		assertActorsDifference(t, session1.Fishermen, session3.Fishermen, probabilityOfOverlap)

		if numBlocksUntilChange == 0 {
			// Assert different servicers between heights for the same app
			assertActorsDifference(t, app1PrevServicers, session1.Servicers, probabilityOfOverlap)
			assertActorsDifference(t, app2PrevServicers, session2.Servicers, probabilityOfOverlap)
			assertActorsDifference(t, app3PrevServicers, session3.Servicers, probabilityOfOverlap)

			// Assert different fishermen between heights for the same app
			assertActorsDifference(t, app1PrevFishermen, session1.Fishermen, probabilityOfOverlap)
			assertActorsDifference(t, app2PrevFishermen, session2.Fishermen, probabilityOfOverlap)
			assertActorsDifference(t, app3PrevFishermen, session3.Fishermen, probabilityOfOverlap)

			// Store the new servicers and fishermen for the next height
			app1PrevServicers = session1.Servicers
			app2PrevServicers = session2.Servicers
			app3PrevServicers = session3.Servicers
			app1PrevFishermen = session1.Fishermen
			app2PrevFishermen = session2.Fishermen
			app3PrevFishermen = session3.Fishermen

			// Reset the number of blocks until we expect a different set of servicers and fishermen
			numBlocksUntilChange = numBlocksPerSession - 1
		} else {
			// Assert the same servicers between heights for the same app
			require.ElementsMatch(t, app1PrevServicers, session1.Servicers)
			require.ElementsMatch(t, app2PrevServicers, session2.Servicers)
			require.ElementsMatch(t, app3PrevServicers, session3.Servicers)

			// Assert the same fishermen between heights for the same app
			require.ElementsMatch(t, app1PrevFishermen, session1.Fishermen)
			require.ElementsMatch(t, app2PrevFishermen, session2.Fishermen)
			require.ElementsMatch(t, app3PrevFishermen, session3.Fishermen)

			numBlocksUntilChange--
		}

		// Advance block height
		writeCtx, err := persistenceMod.NewRWContext(height)
		require.NoError(t, err)
		err = writeCtx.Commit([]byte(fmt.Sprintf("proposer_height_%d", height)), []byte(fmt.Sprintf("quorum_cert_height_%d", height)))
		require.NoError(t, err)
		writeCtx.Release()
	}
}

func TestSession_GetSession_ApplicationUnbonds(t *testing.T) {
	// TODO: What if an Application unbonds (unstaking period elapses) mid session?
}

func TestSession_GetSession_ServicersAndFishermenCounts_GeoZoneAvailability(t *testing.T) {
	// TODO(#697): Once GeoZones are implemented, the tests need to be added as well
	// Cases: Invalid, unused, non-existent, empty, insufficiently complete, etc...
}

func TestSession_GetSession_ActorReplacement(t *testing.T) {
	// TODO: Since sessions last multiple blocks, we need to design what happens when an actor is (un)jailed, (un)stakes, (un)bonds, (un)pauses
	// mid session. There are open design questions that need to be made.
}

func TestSession_GetSession_SessionHeightAndNumber_ModifiedBlocksPerSession(t *testing.T) {
	// RESEARCH: Need to design what happens (actor replacement, session numbers, etc...) when the number
	// of blocks per session changes mid session. For example, all existing sessions could go to completion
	// until the new parameter takes effect. There are open design questions that need to be made.
}

func assertActorsDifference(t *testing.T, actors1, actors2 []*coreTypes.Actor, maxSimilarityThreshold float64) {
	slice1 := actorsToAddrs(actors1)
	slice2 := actorsToAddrs(actors2)
	var commonCount float64
	for _, s1 := range slice1 {
		for _, s2 := range slice2 {
			if s1 == s2 {
				commonCount++
				break
			}
		}
	}
	maxCommonCount := math.Round(maxSimilarityThreshold * float64(len(slice1)))
	assert.LessOrEqual(t, commonCount, maxCommonCount, "Slices have more similarity than expected: %v vs max %v", slice1, slice2)
}

func actorsToAddrs(actors []*coreTypes.Actor) []string {
	addresses := make([]string, len(actors))
	for i, actor := range actors {
		addresses[i] = actor.Address
	}
	return addresses
}
