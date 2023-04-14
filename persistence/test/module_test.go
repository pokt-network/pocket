package test

import (
	"encoding/hex"
	"testing"

	"github.com/pokt-network/pocket/runtime/test_artifacts/keygen"
	"github.com/stretchr/testify/require"
)

func TestPersistenceContextParallelReadWrite(t *testing.T) {
	prepareAndCleanContext(t)

	// variables for testing
	_, _, poolAddr := keygen.GetInstance().Next()
	addrBz, err := hex.DecodeString(poolAddr)
	require.NoError(t, err)
	originalAmount := "15"
	modifiedAmount := "10"
	proposerAddr := []byte("proposerAddr")
	quorumCert := []byte("quorumCert")

	// setup a write context, insert a pool and commit it
	context, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer context.Release()

	require.NoError(t, context.InsertPool(addrBz, originalAmount))
	require.NoError(t, context.Commit(proposerAddr, quorumCert))

	// verify the insert in the previously committed context worked
	contextA, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer contextA.Release()

	contextAOriginalAmount, err := contextA.GetPoolAmount(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, originalAmount, contextAOriginalAmount)

	// modify write contextA but do not commit it
	require.NoError(t, contextA.SetPoolAmount(addrBz, modifiedAmount))

	contextAModifiedAmount, err := contextA.GetPoolAmount(addrBz, 0)
	require.NoError(t, err)
	require.Equal(t, modifiedAmount, contextAModifiedAmount)

	// setup a read context - independent of the previous modified but uncommitted context
	contextB, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)
	defer contextB.Release()

	// verify context b is unchanged
	contextBOriginalAmount, err := contextB.GetPoolAmount(addrBz, 0)
	require.NoError(t, err)
	require.NotEqual(t, modifiedAmount, contextBOriginalAmount)
	require.Equal(t, contextAOriginalAmount, contextBOriginalAmount)
}

func TestPersistenceContextTwoWritesErrors(t *testing.T) {
	prepareAndCleanContext(t)

	// Opening up first write context succeeds
	rwCtx1, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	defer rwCtx1.Release()

	// Opening up second write context at the same height fails
	_, err = testPersistenceMod.NewRWContext(0)
	require.Error(t, err)

	// Opening up a third second write context at a different height fails
	_, err = testPersistenceMod.NewRWContext(1)
	require.Error(t, err)
}

func TestPersistenceContextSequentialWrites(t *testing.T) {
	prepareAndCleanContext(t)

	// Opening up first write context succeeds
	writeContext1, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	writeContext1.Release()

	// Opening up second write context at the same height succeeds
	writeContext2, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	writeContext2.Release()

	// Opening up third write context at a different height succeeds
	writeContext3, err := testPersistenceMod.NewRWContext(1)
	require.NoError(t, err)
	writeContext3.Release()
}

func TestPersistenceContextMultipleParallelReads(t *testing.T) {
	prepareAndCleanContext(t)

	// Opening up first read context succeeds
	readContext1, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)

	// Opening up second read context at the same height succeeds
	readContext2, err := testPersistenceMod.NewReadContext(0)
	require.NoError(t, err)

	// Opening up third read context at a different height succeeds
	readContext3, err := testPersistenceMod.NewReadContext(1)
	require.NoError(t, err)

	readContext1.Release()
	readContext2.Release()
	readContext3.Release()
}

func prepareAndCleanContext(t *testing.T) {
	// Cleanup context after the test
	t.Cleanup(clearAllState)

	clearAllState()
}
