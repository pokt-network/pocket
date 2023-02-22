package test

import (
	"testing"

	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestGetAllStakedActors(t *testing.T) {
	db := NewTestPostgresContext(t, 0)
	expectedActorCount := genesisStateNumValidators + genesisStateNumServicers + genesisStateNumApplications + genesisStateNumFishermen

	actors, err := db.GetAllStakedActors(0)
	require.NoError(t, err)
	require.Equal(t, expectedActorCount, len(actors))

	actualValidators := 0
	actualServicers := 0
	actualApplications := 0
	actualFishermen := 0
	for _, actor := range actors {
		switch actor.ActorType {
		case types.ActorType_ACTOR_TYPE_VAL:
			actualValidators++
		case types.ActorType_ACTOR_TYPE_SERVICER:
			actualServicers++
		case types.ActorType_ACTOR_TYPE_APP:
			actualApplications++
		case types.ActorType_ACTOR_TYPE_FISH:
			actualFishermen++
		}
	}
	require.Equal(t, genesisStateNumValidators, actualValidators)
	require.Equal(t, genesisStateNumServicers, actualServicers)
	require.Equal(t, genesisStateNumApplications, actualApplications)
	require.Equal(t, genesisStateNumFishermen, actualFishermen)
}
