package test

import (
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
)

func TestInitParams(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	err := db.InitParams()
	require.NoError(t, err)
}

func TestGetSetParam(t *testing.T) {
	db := persistence.PostgresContext{
		Height: 0,
		DB:     *PostgresDB,
	}
	err := db.InitParams()
	require.NoError(t, err)
	err = db.SetParam(types.AppMaxChainsParamName, ParamToUpdate)
	require.NoError(t, err)
	maxChains, err := db.GetMaxAppChains()
	require.NoError(t, err)
	if maxChains != ParamToUpdate {
		t.Fatal("unexpected param value")
	}
}
