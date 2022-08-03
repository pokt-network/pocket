package test

import (
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

func TestGetSetParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height:     0,
		PostgresDB: testPostgresDB,
	}

	err := db.InitParams()
	require.NoError(t, err)

	newMaxChains := 42

	err = db.SetParam(types.AppMaxChainsParamName, newMaxChains)
	require.NoError(t, err)

	maxChains, err := db.GetMaxAppChains()
	require.NoError(t, err)

	require.Equal(t, newMaxChains, maxChains)
}
