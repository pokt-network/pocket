package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// IMPROVE: Need to expand contexts to test contexts as the height changes.
func TestContextAndCommit(t *testing.T) {
	// variables for testing
	poolName := "fake"
	poolAddress := []byte("address")
	originalAmount := "15"
	// modifiedAmount := "10"

	// setup a context, make a change and commit it
	contextA, err := testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)

	require.NoError(t, contextA.InsertPool(poolName, poolAddress, originalAmount))
	require.NoError(t, contextA.Commit())

	// verify the insert in the previous context worked
	contextA, err = testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)

	contextAOriginal, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)

	require.Equal(t, originalAmount, contextAOriginal)

	contextA.Release()
	// setup a second context
	_, err = testPersistenceModule.NewReadContext(0)
	// testPersistenceModule.NewReadContext(0)
	require.NoError(t, err) // Cannot open a second write context with the same height when it's already open

	// // modify only in context a and check that modification worked
	// require.NoError(t, contextA.SetPoolAmount(poolName, modifiedAmount))
	// contextAAfter, err := contextA.GetPoolAmount(poolName, 0)
	// require.NoError(t, err)
	// require.Equal(t, modifiedAmount, contextAAfter)

	// // ensure context b is unchanged
	// contextBOriginal, err := contextB.GetPoolAmount(poolName, 0)
	// require.NoError(t, err)
	// require.NotEqual(t, modifiedAmount, contextBOriginal)
	// require.Equal(t, contextBOriginal, contextAOriginal)

	// contextA.Release()
	// contextB.Release()
}
