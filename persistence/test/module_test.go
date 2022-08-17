package test

import (
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestContextAndCommitComplexFlow(t *testing.T) {
	// variables for testing
	poolName := "fake"
	poolAddress := []byte("address")
	originalAmount := "15"
	modifiedAmount := "10"

	// setup a context, insert a pool and commit it
	context, err := testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)
	require.NoError(t, context.InsertPool(poolName, poolAddress, originalAmount))
	require.NoError(t, context.Commit())

	// verify the insert in the previously committed context worked
	contextA, err := testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)

	contextAOriginalAmount, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, originalAmount, contextAOriginalAmount)

	// modify contextA but do not commit it
	require.NoError(t, contextA.SetPoolAmount(poolName, modifiedAmount))

	contextAModifiedAmount, err := contextA.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.Equal(t, modifiedAmount, contextAModifiedAmount)

	// setup a second context - independent of the previous modified but uncommitted context
	contextB, err := testPersistenceModule.NewReadContext(0)
	require.NoError(t, err)

	// verify context b is unchanged
	contextBOriginalAmount, err := contextB.GetPoolAmount(poolName, 0)
	require.NoError(t, err)
	require.NotEqual(t, modifiedAmount, contextBOriginalAmount)
	require.Equal(t, contextBOriginalAmount, contextAOriginalAmount)
}

// TODO(pocket/issues/149): Need to add support for this sort of test. The call to
// `contextB.SetAccountAmount` currently hangs because we have multiple writes contexts at the
// same height. Some potential solutions may include:
// - keeping a set of all the write contexts and panicking
// - Adding timeouts to the write contexts
// - ???
func TestTwoWriteContextsSameHeight(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	// set amount in write contextA to 10
	contextA, err := testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)
	contextA.SetAccountAmount(addr, "10")

	// set amount in write contextB to 20
	_, err = testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)
	// contextB.SetAccountAmount(addr, "20")

	// // Check that a new read contextC returns a default value
	// contextC, err := testPersistenceModule.NewReadContext(0)
	// require.NoError(t, err)
	// amount, err := contextC.GetAccountAmount(addr, 0)
	// require.NoError(t, err)
	// require.Equal(t, "0", amount) // default amount

	// // contextA still returns 10
	// amount, err = contextA.GetAccountAmount(addr, 0)
	// require.NoError(t, err)
	// require.Equal(t, "10", amount)

	// // contextB still returns 20
	// amount, err = contextB.GetAccountAmount(addr, 0)
	// require.NoError(t, err)
	// require.Equal(t, "20", amount)
}

func TestTwoWriteContextsDifferentHeight(t *testing.T) {
	addr, err := crypto.GenerateAddress()
	require.NoError(t, err)

	// set amount in write contextA to 10 at height 0
	contextZeroWrite, err := testPersistenceModule.NewRWContext(0)
	require.NoError(t, err)
	contextZeroWrite.SetAccountAmount(addr, "10")

	// set amount in write contextB to 20 at height 1
	contextOneWrite, err := testPersistenceModule.NewRWContext(1)
	require.NoError(t, err)
	contextOneWrite.SetAccountAmount(addr, "20")

	// Check that a new read contextZeroRead returns a default value
	contextZeroRead, err := testPersistenceModule.NewReadContext(0)
	require.NoError(t, err)
	amount, err := contextZeroRead.GetAccountAmount(addr, 0)
	require.NoError(t, err)
	require.Equal(t, "0", amount) // default amount

	// Check that a new read contextOneRead returns a default value
	contextOneRead, err := testPersistenceModule.NewReadContext(1)
	require.NoError(t, err)
	amount, err = contextOneRead.GetAccountAmount(addr, 1)
	require.NoError(t, err)
	require.Equal(t, "0", amount) // default amount

	// contextZeroWrite still returns 10
	amount, err = contextZeroWrite.GetAccountAmount(addr, 0)
	require.NoError(t, err)
	require.Equal(t, "10", amount)

	// contextOneWrite still returns 20
	amount, err = contextOneWrite.GetAccountAmount(addr, 1)
	require.NoError(t, err)
	require.Equal(t, "20", amount)

	// contextOneWrite still returns 0 at height 0 (sanity check)
	amount, err = contextOneWrite.GetAccountAmount(addr, 0)
	require.NoError(t, err)
	require.Equal(t, "0", amount)
}
