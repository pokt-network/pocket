package utility

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestSession_NewSession(t *testing.T) {
	session, err := testUtilityMod.GetSession("app", 1, coreTypes.RelayChain_ETHEREUM, "geo")
	require.NoError(t, err)

	require.Equal(t, session.Application.Address, "app")
}

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
