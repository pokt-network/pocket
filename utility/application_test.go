package utility

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

const (
	DefaultAppSessionTokens = "100000000000000"
)

func TestUtilityContext_CalculateMaxAppRelays(t *testing.T) {
	ctx := newTestingUtilityContext(t, 1)
	actor := getFirstActor(t, ctx, coreTypes.ActorType_ACTOR_TYPE_APP)
	appSessionTokens, err := ctx.calculateAppSessionTokens(actor.StakedAmount)
	require.NoError(t, err)
	// TODO: These are hardcoded values based on params from the genesis file. Expand on tests
	// when implementing the Application protocol.
	require.Equal(t, DefaultAppSessionTokens, appSessionTokens)
}
