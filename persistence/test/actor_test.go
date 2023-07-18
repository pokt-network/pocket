package test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestGetValidatorSetHash(t *testing.T) {
	currHeight := int64(0)
	db := NewTestPostgresContext(t, currHeight)
	validators, err := db.GetAllValidators(currHeight)
	require.NoError(t, err)
	require.Len(t, validators, genesisStateNumValidators)

	buf := new(bytes.Buffer)
	for _, val := range validators {
		_, err := buf.WriteString(val.PublicKey)
		require.NoError(t, err)
	}
	expectedHash := crypto.SHA3Hash(buf.Bytes())

	actualHash, err := db.GetValidatorSetHash(currHeight)
	require.NoError(t, err)
	require.Equal(t, expectedHash, actualHash)

	// ensure deterministic
	require.Equal(t, hex.EncodeToString(expectedHash), "8bc3057b495c988e831052a28946cd5e291e3b4ec51a22680fde12657a277f79")
	require.Equal(t, hex.EncodeToString(actualHash), "8bc3057b495c988e831052a28946cd5e291e3b4ec51a22680fde12657a277f79")

	// ensure hash remains the same with no changes
	currHeight = int64(1)
	db.Height = currHeight

	actualHash, err = db.GetValidatorSetHash(currHeight)
	require.NoError(t, err)
	require.Equal(t, expectedHash, actualHash)

	// ensure hash changes with changes
	currHeight = int64(2)
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

	actualHash, err = db.GetValidatorSetHash(currHeight)
	require.NoError(t, err)
	require.NotEqual(t, expectedHash, actualHash)
	require.Equal(t, hex.EncodeToString(actualHash), "560e954828499c90c5431576b78927cb0f3613435e8d302f0302113806b11d25")
}
