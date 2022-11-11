package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPrevAppHash(t *testing.T) {
	db := NewTestPostgresContext(t, 0)

	// Cannot get prev hash at height 0
	_, err := db.GetPrevAppHash()
	require.Error(t, err)

	db.Close()
	db = NewTestPostgresContext(t, 1)

	// Cannot a non empty prev hash at height 1 (i.e. the genesis hash)
	appHash, err := db.GetPrevAppHash()
	require.NoError(t, err)
	require.NotEmpty(t, appHash)

	db.Close()
	db = NewTestPostgresContext(t, 10)

	// This hash does not exist
	appHash, err = db.GetPrevAppHash()
	require.Error(t, err)
}
