package test

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"

	"github.com/pokt-network/pocket/persistence"
	"github.com/stretchr/testify/require"
)

// TODO(andrew): Be consistent with `GenericParam` and `ActorSpecificParam` throughout the codebase; preferably the latter.

func getGenericActor[T any](
	protocolActorSchema types.ProtocolActorSchema,
	getActor func(*persistence.PostgresContext, []byte) (T, error),
) func(*persistence.PostgresContext, string) (*coreTypes.Actor, error) {
	return func(db *persistence.PostgresContext, address string) (*coreTypes.Actor, error) {
		addr, err := hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
		actor, err := getActor(db, addr)
		if err != nil {
			return nil, err
		}
		baseActor := getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor)))
		return baseActor, nil
	}
}

func newTestGenericActor[T any](protocolActorSchema types.ProtocolActorSchema, newActor func() (T, error)) func() (*coreTypes.Actor, error) {
	return func() (*coreTypes.Actor, error) {
		actor, err := newActor()
		if err != nil {
			return nil, err
		}
		return getActorValues(protocolActorSchema, reflect.Indirect(reflect.ValueOf(actor))), nil
	}
}

func getAllActorsTest[T any](
	t *testing.T,
	db *persistence.PostgresContext,
	getAllActors func(height int64) ([]T, error),
	createTestActor func(*persistence.PostgresContext) (T, error),
	updateActor func(*persistence.PostgresContext, T) error,
	initialCount int,
) {
	// The default test state contains `initialCount` actors
	actors, err := getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Add 2 actors at height 1
	db.Height++
	_, err = createTestActor(db)
	require.NoError(t, err)
	_, err = createTestActor(db)
	require.NoError(t, err)

	// Check height 0
	actors, err = getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Add 1 actor at height 3
	db.Height++
	db.Height++
	_, err = createTestActor(db)
	require.NoError(t, err)

	// Check height 0
	actors, err = getAllActors(0)
	require.NoError(t, err)
	require.Len(t, actors, initialCount)

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 2
	actors, err = getAllActors(2)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 3
	actors, err = getAllActors(3)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+3)

	// Update the servicers at different heights and confirm that count does not change
	for _, actor := range actors {
		db.Height++
		err = updateActor(db, actor)
		require.NoError(t, err)

		// Check that count did not change
		actors, err := getAllActors(db.Height)
		require.NoError(t, err)
		require.Len(t, actors, initialCount+3)
	}

	// Check height 1
	actors, err = getAllActors(1)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+2)

	// Check height 10
	actors, err = getAllActors(10)
	require.NoError(t, err)
	require.Len(t, actors, initialCount+3)
}

func getTestGetSetStakeAmountTest[T any](
	t *testing.T,
	db *persistence.PostgresContext,
	createTestActor func(*persistence.PostgresContext) (*T, error),
	getActorStake func(int64, []byte) (string, error),
	setActorStake func([]byte, string) error,
	height int64,
) {
	var newStakeAmount = "new_stake_amount"

	actor, err := createTestActor(db)
	require.NoError(t, err)
	addrStr := reflect.ValueOf(*actor).FieldByName("Address").String()

	addr, err := hex.DecodeString(addrStr)
	require.NoError(t, err)

	// Check stake amount before
	stakeAmount, err := getActorStake(height, addr)
	require.NoError(t, err)
	require.Equal(t, DefaultStake, stakeAmount, "unexpected beginning stakeAmount")

	// Check stake amount after
	err = setActorStake(addr, newStakeAmount)
	require.NoError(t, err)

	stakeAmountAfter, err := getActorStake(height, addr)
	require.NoError(t, err)
	require.Equal(t, newStakeAmount, stakeAmountAfter, "unexpected status")
}

func getAllActorsUpdatedAtHeightTest[T any](
	t *testing.T,
	createAndInsertTestActor func(*persistence.PostgresContext) (*T, error),
	getActorsUpdated func(*persistence.PostgresContext, int64) ([]*T, error),
	numActorsInTestGenesis int,
) {
	db := NewTestPostgresContext(t, 0)

	// Check num actors in genesis
	accs, err := getActorsUpdated(db, 0)
	require.NoError(t, err)
	require.Equal(t, numActorsInTestGenesis, len(accs))

	// Insert a new actor at height 0
	_, err = createAndInsertTestActor(db)
	require.NoError(t, err)

	// Verify that num actors incremented by 1
	accs, err = getActorsUpdated(db, 0)
	require.NoError(t, err)
	require.Equal(t, numActorsInTestGenesis+1, len(accs))

	// Close context at height 0 without committing new Pool
	require.NoError(t, db.Close())
	// start a new context at height 1
	db = NewTestPostgresContext(t, 1)

	// Verify that num actors at height 0 is genesis because the new one was not committed
	accs, err = getActorsUpdated(db, 0)
	require.NoError(t, err)
	require.Equal(t, numActorsInTestGenesis, len(accs))

	// Insert a new actor at height 1
	_, err = createAndInsertTestActor(db)
	require.NoError(t, err)

	// Verify that num actors updated height 1 is 1
	accs, err = getActorsUpdated(db, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(accs))

	// Commit & close the context at height 1
	require.NoError(t, db.Commit(nil, nil))
	// start a new context at height 2
	db = NewTestPostgresContext(t, 2)

	// Verify only 1 actor was updated at height 1
	accs, err = getActorsUpdated(db, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(accs))
}

func getActorValues(_ types.ProtocolActorSchema, actorValue reflect.Value) *coreTypes.Actor {
	chains := make([]string, 0)
	if actorValue.FieldByName("Chains").Kind() != 0 {
		chains = actorValue.FieldByName("Chains").Interface().([]string)
	}

	return &coreTypes.Actor{
		Address:         actorValue.FieldByName("Address").String(),
		PublicKey:       actorValue.FieldByName("PublicKey").String(),
		StakedAmount:    actorValue.FieldByName("StakedAmount").String(),
		GenericParam:    actorValue.FieldByName("GenericParam").String(),
		Output:          actorValue.FieldByName("Output").String(),
		PausedHeight:    actorValue.FieldByName("PausedHeight").Int(),
		UnstakingHeight: actorValue.FieldByName("UnstakingHeight").Int(),
		Chains:          chains,
	}
}
