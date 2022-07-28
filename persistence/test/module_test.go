package test

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContextAndCommit(t *testing.T) {
	// variables for testing
	poolName := "fake"
	poolAddress := []byte("address")
	originalAmount := "15"
	modifiedAmount := "10"
	// setup two separate contexts
	contextA, err := PersistenceModule.NewRWContext(0)
	require.NoError(t, contextA.InsertPool(poolName, poolAddress, originalAmount))
	require.NoError(t, contextA.Commit())
	// verify the insert worked
	contextA, err = PersistenceModule.NewRWContext(0)
	contextAOriginal, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, originalAmount, contextAOriginal)
	require.NoError(t, err)
	contextB, err := PersistenceModule.NewRWContext(0)
	require.NoError(t, err)
	// modify only in context a and check that modification worked
	require.NoError(t, contextA.SetPoolAmount(poolName, modifiedAmount))
	contextAAfter, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, modifiedAmount, contextAAfter)
	// ensure context b is unchanged
	contextBOriginal, err := contextB.GetPoolAmount(poolName, 0)
	require.NotEqual(t, modifiedAmount, contextBOriginal)
	require.Equal(t, contextBOriginal, contextAOriginal)
}
