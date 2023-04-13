package utility

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

// TECH_DEBT_IDENTIFIED_IN_THIS_COMMIT:
// 1. Replace []byte with string
// 2. Remove height from Write context in persistence
// 3. Need to add geozone to actors
// 4. Need to generalize persitence functions based on actor type
// 5. Need different protos for each actor

func TestSession_NewSession(t *testing.T) {
	runtimeCfg, utilityMod, _ := prepareEnvironment(t, 5, 1, 1, 1)

	// Loop over these
	app := runtimeCfg.GetGenesis().Applications[0]
	height := int64(1)
	relayChain := coreTypes.RelayChain_ETHEREUM
	geoZone := "geo"

	session, err := utilityMod.GetSession(app.Address, height, relayChain, geoZone)
	require.NoError(t, err)
	require.Equal(t, "8b50d1f751029a06d0b860e3b900163b3c6532fc48df2e11f848600019df5483", session.Id)
	require.Equal(t, height, session.Height)
	require.Equal(t, relayChain, session.RelayChain)
	require.Equal(t, geoZone, session.GeoZone)
	require.Equal(t, session.Application.Address, app.Address)
	require.Equal(t, "c7832263600476fd6ff4c5cb0a86080d0e5f48b2", session.Servicers[0].Address)
	require.Equal(t, "fisherman", session.Fishermen[0].Address)
}

func TestSession_SessionHeight(t *testing.T) {
	// for i := 0; i < 100; i++ {

	// BlocksPerSessionParamName = 4
	// blockHeigh = 4
	// % 4 = 0
	// % 4 = prevSessionHeight
	// % 4 = nextSessionHeight

	// What if session height changes mid session
	//

	// testUtilityMod := newTestUtilityModule(bus)
	// testUtilityMod.Start()
	// defer testUtilityMod.Stop()

	// BlocksPerSessionParamName
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
