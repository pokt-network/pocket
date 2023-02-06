package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBlockBlockHash(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	// Cannot get prev hash at height 0
	blockHash, err := db.GetBlockHash(0)
	require.NoError(t, err)
	require.NotEmpty(t, blockHash)

	// Cannot get a hash at height 1 since it doesn't exist
	blockHash, err = db.GetBlockHash(1)
	require.Error(t, err)
	require.Equal(t, blockHash, "")

	// Cannot get a hash at height 10 since it doesn't exist
	blockHash, err = db.GetBlockHash(10)
	require.Error(t, err)
	require.Equal(t, blockHash, "")
}
