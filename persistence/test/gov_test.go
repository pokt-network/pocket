package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
)

func TestInitParams(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}
	err := db.InitParams()
	require.NoError(t, err)
}

func TestGetSetIntParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newMaxChains := 42

	err = persistence.SetParam(db, types.AppMaxChainsParamName, newMaxChains)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	//maxChains, err := db.GetMaxAppChains(height)
	maxChains, err := db.GetIntParam(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)
}

func TestGetSetStringParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newServiceNodeMinimumStake := "99999999"

	err = persistence.SetParam(db, types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	//serviceNodeMinimumStake, err := db.GetParamServiceNodeMinimumStake(height)
	serviceNodeMinimumStake, err := db.GetStringParam(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
}

func TestGetSetByteArrayParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newOwner, _ := hex.DecodeString("da034209758b78eaea06dd99c07909ab54c99b44")

	err = persistence.SetParam(db, types.ServiceNodeUnstakingBlocksOwner, newOwner)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	//owner, err := db.GetServiceNodeUnstakingBlocksOwner(height)
	owner, err := db.GetBytesParam(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
}
