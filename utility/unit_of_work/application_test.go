package unit_of_work

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

const (
	DefaultAppSessionTokens = "100000000000000"
)

func TestUtilityUnitOfWork_CalculateMaxAppRelays(t *testing.T) {
	uow := newTestingUtilityUnitOfWork(t, 1)
	actor := getFirstActor(t, uow, coreTypes.ActorType_ACTOR_TYPE_APP)
	appSessionTokens, err := uow.calculateAppSessionTokens(actor.StakedAmount)
	require.NoError(t, err)
	// TODO: These are hardcoded values based on params from the genesis file. Expand on tests
	// when implementing the Application protocol.
	require.Equal(t, DefaultAppSessionTokens, appSessionTokens)
}
