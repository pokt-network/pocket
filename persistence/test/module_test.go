package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersistenceContextParallelReadWrite(t *testing.T) {
	// Cleanup previous contexts
	testPersistenceMod.ResetContext()
	t.Cleanup(func() {
		testPersistenceMod.ResetContext()
	})

	// variables for testing
	poolName := "fake"
	poolAddress := []byte("address")
	originalAmount := "15"
	modifiedAmount := "10"
	quorumCert := []byte("quorumCert")

	// setup a write context, insert a pool and commit it
	context, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	require.NoError(t, context.InsertPool(poolName, poolAddress, originalAmount))
	require.NoError(t, context.Commit(quorumCert))

	// verify the insert in the previously committed context worked
	contextA, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)

	contextAOriginalAmount, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, originalAmount, contextAOriginalAmount)

	// modify write contextA but do not commit it
	require.NoError(t, contextA.SetPoolAmount(poolName, modifiedAmount))

	contextAModifiedAmount, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, modifiedAmount, contextAModifiedAmount)

	// setup a read context - independent of the previous modified but uncommitted context
	contextB, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)

	// verify context b is unchanged
	contextBOriginalAmount, err := contextB.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.NotEqual(t, modifiedAmount, contextBOriginalAmount)
	require.Equal(t, contextBOriginalAmount, contextAOriginalAmount)
}

func TestPersistenceContextTwoWritesErrors(t *testing.T) {
	// Cleanup previous contexts
	testPersistenceMod.ResetContext()
	t.Cleanup(func() {
		testPersistenceMod.ResetContext()
	})

	// Opening up first write context succeeds
	_, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)

	// Opening up second write context at the same height fails
	_, err = testPersistenceMod.NewRWContext(0)
	require.Error(t, err)

	// Opening up second write context at a different height fails
	_, err = testPersistenceMod.NewRWContext(1)
	require.Error(t, err)
}

func TestPersistenceContextSequentialWrites(t *testing.T) {
	// Opening up first write context succeeds
	writeContext1, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)

	// Close the write context
	require.NoError(t, writeContext1.Release())

	// Opening up second write context at the same height succeeds
	writeContext2, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)

	// Close the write context
	require.NoError(t, writeContext2.Release())

	// Opening up third write context at a different height succeeds
	writeContext3, err := testPersistenceMod.NewRWContext(1)
	require.NoError(t, err)

	// Close the write context
	require.NoError(t, writeContext3.Release())
}

func TestPersistenceContextMultipleParallelReads(t *testing.T) {
	// Opening up first read context succeeds
	readContext1, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)

	// Opening up second read context at the same height succeeds
	readContext2, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)

	// Opening up third read context at a different height succeeds
	readContext3, err := testPersistenceMod.NewReadContext(1)
	require.NoError(t, err)

	require.NoError(t, readContext1.Close())
	require.NoError(t, readContext2.Close())
	require.NoError(t, readContext3.Close())
}
