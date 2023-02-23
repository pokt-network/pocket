package utility

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestUtilityContext_CalculateMaxAppRelays(t *testing.T) {
	ctx := newTestingUtilityContext(t, 1)
	actor := getFirstActor(t, ctx, coreTypes.ActorType_ACTOR_TYPE_APP)
	newMaxRelays, err := ctx.calculateMaxAppRelays(actor.GetStakedAmount())
	require.NoError(t, err)
	require.Equal(t, actor.GetServiceUrl(), newMaxRelays)
}
