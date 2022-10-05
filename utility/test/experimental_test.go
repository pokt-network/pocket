package test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/require"
)

// This test just tries to add amounts to an account and check the app hash via the real implementation
func TestExperimentalAppHashWithoutPersistenceMock(t *testing.T) {
	ctx := NewTestingUtilityContext(t, 0)

	// Prepare a new account
	pubKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	addr := []byte(pubKey.Address())

	// Add some amount to the account
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)

	// Verify the account's amount
	amount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(3), amount)

	// Check the hash the first time it was called - hardcoded value in the persistence module
	appHash, err := ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte("A real app hash, I am not"), appHash)

	// Commit the context - noop with respect to app hash at the moment
	require.NoError(t, ctx.Store().Commit())

	// Check the hash the second time it was called - same thing because it's hardcoded
	appHash, err = ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte("A real app hash, I am not"), appHash)
}

// This test just tries to add amounts to an account and check the app hash via a mocked implementation
// of the app hash even though the amount addition triggers the real codepath
func TestExperimentalAppHashWithPersistenceMock(t *testing.T) {
	height := int64(0)
	ctx := newTestingUtilityContextWithPersistenceContext(t, height, newMockablePersistenceContextForAppHashTest(t, height))

	// Prepare a new account
	pubKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	addr := []byte(pubKey.Address())

	// Add some amount to the account
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)

	// Verify the account's amount
	amount, err := ctx.GetAccountAmount(addr)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(3), amount)

	// Check the hash the first time it was called - the first mocked value
	appHash, err := ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte("first commit"), appHash)

	// Commit the context - noop with respect to app hash at the moment
	require.NoError(t, ctx.Store().Commit())

	// Check the hash the second time it was called - the second mocked value
	appHash, err = ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte("second commit"), appHash)

	// Calling this a third time won't work because we limited the number of mock calls to 2
	// appHash, err = ctx.GetAppHash()
}

func newMockablePersistenceContextForAppHashTest(t *testing.T, height int64) modules.PersistenceRWContext {
	// Create one instance of the real version of the persistence context
	persistenceContext, err := testPersistenceMod.NewRWContext(0)
	require.NoError(t, err)
	t.Cleanup(func() {
		testPersistenceMod.ResetContext()
	})

	// Create one version of a mock of the persistence context
	ctrl := gomock.NewController(t)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)

	// --- Passthrough implementation ---
	// The functions below simply call the real implementation of the persistence context so it does
	// not need to be re-implemented.

	persistenceContextMock.EXPECT().
		GetHeight().
		DoAndReturn(persistenceContext.GetHeight).
		AnyTimes()

	persistenceContextMock.EXPECT().
		AddAccountAmount(gomock.Any(), gomock.Any()).
		DoAndReturn(persistenceContext.AddAccountAmount).
		AnyTimes()

	persistenceContextMock.EXPECT().
		GetAccountAmount(gomock.Any(), gomock.Any()).
		DoAndReturn(persistenceContext.GetAccountAmount).
		AnyTimes()

	persistenceContextMock.EXPECT().
		Commit().
		DoAndReturn(persistenceContext.Commit).
		AnyTimes()

	// --- Mocked implementation ---
	// The functions below allow AppHash (not implemented yet) to be called twice and return
	// different values each time.

	persistenceContextMock.EXPECT().
		AppHash().
		Return([]byte("first commit"), nil).
		Times(1)

	persistenceContextMock.EXPECT().
		AppHash().
		Return([]byte("second commit"), nil).
		Times(1)

	return persistenceContextMock
}

// This mocks both the account addition amount and the app hash implementation and doesn't use
// any parts of the real module implementation. It also shows how a mock can be used to be
// more stateful in between different calls.
func TestExperimentalAppHashWithStatefulPersistenceMock(t *testing.T) {
	height := int64(0)
	ctx := newTestingUtilityContextWithPersistenceContext(t, height, newMockablePersistenceContextForAppHashTestWithMoreState(t, height))

	// Prepare a new account
	pubKey, err := crypto.GeneratePublicKey()
	require.NoError(t, err)

	addr := []byte(pubKey.Address())

	// Add some amount to the account - doesn't actually call real logic
	err = ctx.AddAccountAmount(addr, big.NewInt(1))
	require.NoError(t, err)

	// Check the app hash
	appHash, err := ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte(" 1  getHash "), appHash)

	// Add some amount to the account - doesn't actually call real logic
	err = ctx.AddAccountAmount(addr, big.NewInt(42))
	require.NoError(t, err)

	// Check the app hash
	appHash, err = ctx.GetAppHash()
	require.NoError(t, err)
	require.Equal(t, []byte(" 1  getHash  42  getHash "), appHash)
}

func newMockablePersistenceContextForAppHashTestWithMoreState(t *testing.T, height int64) modules.PersistenceRWContext {
	// Create one version of a mock of the persistence context
	ctrl := gomock.NewController(t)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)

	// --- Stateful Mocked Implementation ---
	// Here we can inline specific functions that are dependant on an external state (just a local
	// variable for simplicity) that changes every time we call it.

	appHash := []byte("")

	persistenceContextMock.EXPECT().
		AddAccountAmount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(address []byte, amount string) error {
			appHash = append(appHash, fmt.Sprintf(" %s ", amount)...)
			return nil
		}).
		AnyTimes()

	persistenceContextMock.EXPECT().
		AppHash().
		DoAndReturn(func() ([]byte, error) {
			appHash = append(appHash, " getHash "...)
			return appHash, nil
		}).
		Times(2)

	return persistenceContextMock
}
