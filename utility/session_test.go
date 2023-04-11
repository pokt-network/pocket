package utility

import (
	"encoding/hex"
	"fmt"
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
	session, err := testUtilityMod.GetSession(hex.EncodeToString([]byte("app")), 1, coreTypes.RelayChain_ETHEREUM, "geo")
	require.NoError(t, err)
	fmt.Println(session)

	// require.Equal(t, session.Application.Address, "app")
}

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
