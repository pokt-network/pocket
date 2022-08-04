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

func TestGetSetToggleIntFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitFlags()
	require.NoError(t, err)

	newMaxChains := 42

	// insert with false
	err = persistence.SetFlag(db, types.AppMaxChainsParamName, newMaxChains, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	maxChains, enabled, err := db.GetIntFlag(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)

	require.Equal(t, false, enabled)

	// toggle to true
	err = persistence.SetFlag(db, types.AppMaxChainsParamName, newMaxChains, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	maxChains, enabled, err = db.GetIntFlag(types.AppMaxChainsParamName, height)
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)

	require.Equal(t, true, enabled)
}

func TestGetSetToggleStringFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newServiceNodeMinimumStake := "99999999"

	// insert with false
	err = persistence.SetFlag(db, types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	serviceNodeMinimumStake, enabled, err := db.GetStringFlag(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
	require.Equal(t, false, enabled)

	//toggle to true
	err = persistence.SetFlag(db, types.ServiceNodeMinimumStakeParamName, newServiceNodeMinimumStake, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	serviceNodeMinimumStake, enabled, err = db.GetStringFlag(types.ServiceNodeMinimumStakeParamName, height)
	require.NoError(t, err)

	require.Equal(t, newServiceNodeMinimumStake, serviceNodeMinimumStake)
	require.Equal(t, true, enabled)

}

func TestGetSetToggleByteArrayFlag(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newOwner, _ := hex.DecodeString("da034209758b78eaea06dd99c07909ab54c99b44")

	// insert with false
	err = persistence.SetFlag(db, types.ServiceNodeUnstakingBlocksOwner, newOwner, false)
	require.NoError(t, err)

	height, err := db.GetHeight()
	require.NoError(t, err)

	owner, enabled, err := db.GetBytesFlag(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
	require.Equal(t, false, enabled)

	//toggle to true
	err = persistence.SetFlag(db, types.ServiceNodeUnstakingBlocksOwner, newOwner, true)
	require.NoError(t, err)

	height, err = db.GetHeight()
	require.NoError(t, err)

	owner, enabled, err = db.GetBytesFlag(types.ServiceNodeUnstakingBlocksOwner, height)
	require.NoError(t, err)

	require.Equal(t, newOwner, owner)
	require.Equal(t, true, enabled)

}
