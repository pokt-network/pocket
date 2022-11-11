package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBlockHash(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	// Cannot get prev hash at height 0
	appHash, err := db.GetBlockHash(0)
	require.NoError(t, err)
	require.NotEmpty(t, appHash)

	// Cannot get a hash at height 1 since it doesn't exist
	appHash, err = db.GetBlockHash(1)
	require.Error(t, err)

	// Cannot get a hash at height 10 since it doesn't exist
	appHash, err = db.GetBlockHash(10)
	require.Error(t, err)

}
