package test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
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
		case coreTypes.ActorType_ACTOR_TYPE_VAL:
			actualValidators++
		case coreTypes.ActorType_ACTOR_TYPE_SERVICER:
			actualServicers++
		case coreTypes.ActorType_ACTOR_TYPE_APP:
			actualApplications++
		case coreTypes.ActorType_ACTOR_TYPE_FISH:
			actualFishermen++
		}
	}
	require.Equal(t, genesisStateNumValidators, actualValidators)
	require.Equal(t, genesisStateNumServicers, actualServicers)
	require.Equal(t, genesisStateNumApplications, actualApplications)
	require.Equal(t, genesisStateNumFishermen, actualFishermen)
}

func TestPostgresContext_GetValidatorSet(t *testing.T) {
	expectedHashes := []string{
		"5831e3e6a8d3beda0adb6027126a4bc3b0181836eebcc45a83dc2970ee9b4468",
		"1f6faa8782a4608341fef83ee72c9e9cd3f96e042f2ddfdeebda8d971bb2ac13",
	}

	// Ensure genesis next val set hash is correct
	db := NewTestPostgresContext(t, 0)
	nextValSet, err := db.GetValidatorSet(0)
	require.NoError(t, err)
	nextValSetHash := hashValSet(t, nextValSet)
	require.Equal(t, expectedHashes[0], nextValSetHash)

	// Ensure next val set hash for genesis == curr val set hash for height 1
	// and next val set hash remains the same with no changes
	currHeight := int64(1)
	db.Height = currHeight

	currValSet, err := db.GetValidatorSet(currHeight - 1)
	require.NoError(t, err)
	currValSetHash := hashValSet(t, currValSet)
	nextValSet, err = db.GetValidatorSet(currHeight)
	require.NoError(t, err)
	nextValSetHash = hashValSet(t, nextValSet)

	require.Equal(t, expectedHashes[0], currValSetHash)
	require.Equal(t, expectedHashes[0], nextValSetHash)

	// ensure both hashes remain the same with no changes
	currHeight = int64(2)
	db.Height = currHeight

	currValSet, err = db.GetValidatorSet(currHeight - 1)
	require.NoError(t, err)
	currValSetHash = hashValSet(t, currValSet)
	nextValSet, err = db.GetValidatorSet(currHeight)
	require.NoError(t, err)
	nextValSetHash = hashValSet(t, nextValSet)

	require.Equal(t, expectedHashes[0], currValSetHash)
	require.Equal(t, expectedHashes[0], nextValSetHash)

	// ensure next val set hash changes when we add a new validator  at current
	// height but the current val set hash remains the same
	currHeight = int64(3)
	db.Height = currHeight

	err = db.InsertValidator(
		[]byte("address"),
		[]byte("publickey"),
		[]byte("output"),
		false, 0,
		"serviceurl",
		"1000000000",
		0,
		0,
	)
	require.NoError(t, err)

	currValSet, err = db.GetValidatorSet(currHeight - 1)
	require.NoError(t, err)
	currValSetHash = hashValSet(t, currValSet)
	nextValSet, err = db.GetValidatorSet(currHeight)
	require.NoError(t, err)
	nextValSetHash = hashValSet(t, nextValSet)

	require.Equal(t, expectedHashes[0], currValSetHash)
	require.Equal(t, expectedHashes[1], nextValSetHash)
}

func hashValSet(t *testing.T, valSet *coreTypes.ValidatorSet) string {
	t.Helper()

	bz, err := codec.GetCodec().Marshal(valSet)
	require.NoError(t, err)

	hash := crypto.SHA3Hash(bz)
	return hex.EncodeToString(hash)
}
