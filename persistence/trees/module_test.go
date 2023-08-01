package trees_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/providers/peerstore_provider"
	"github.com/pokt-network/pocket/persistence/trees"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// DISCUSS: This is duplicated from inside trees package. Is it worth exporting or is it better as duplicate code?
var stateTreeNames = []string{
	"root",
	"app",
	"val",
	"fish",
	"servicer",
	"account",
	"pool",
	"transactions",
	"params",
	"flags",
	"ibc",
}

func TestTreeStore_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
	mockBus := createMockBus(t, mockRuntimeMgr)

	treemod, err := trees.Create(mockBus, trees.WithTreeStoreDirectory(":memory:"))
	assert.NoError(t, err)
	require.NoError(t, treemod.Start())

	// Create should setup a value for each tree
	for _, v := range stateTreeNames {
		root, ns := treemod.GetTree(v)
		require.NotEmpty(t, root)
		require.NotEmpty(t, ns)
	}

	got := treemod.GetBus()
	require.Equal(t, got, mockBus)

	// root hash should be empty for empty tree
	root, ns := treemod.GetTree(trees.TransactionsTreeName)
	require.Equal(t, root, make([]byte, 32))

	// nodestore should have no values in it
	keys, vals, err := ns.GetAll(nil, false)
	require.NoError(t, err)
	require.Empty(t, keys, vals)
}

func TestTreeStore_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRuntimeMgr := mockModules.NewMockRuntimeMgr(ctrl)
	mockBus := createMockBus(t, mockRuntimeMgr)

	// Create returns a started TreeStoreSubmodule
	treemod, err := trees.Create(
		mockBus,
		trees.WithTreeStoreDirectory(":memory:"),
		trees.WithLogger(&zerolog.Logger{}))
	require.NoError(t, err)
	require.NoError(t, treemod.Start())

	// Should stop without error
	require.NoError(t, treemod.Stop())

	// Should error if node store is called after Stop
	for _, treeName := range stateTreeNames {
		_, nodestore := treemod.GetTree(treeName)
		_, _, err = nodestore.GetAll([]byte(""), false)
		require.Error(t, err, "%s tree failed to return an error when expected", treeName)
	}
}

func TestTreeStore_DebugClearAll(t *testing.T) {
	// TODO: Write test case for the DebugClearAll method
	t.Skip("TODO: Write test case for DebugClearAll method")
}

// createMockBus returns a mock bus with stubbed out functions for bus registration
func createMockBus(t *testing.T, runtimeMgr modules.RuntimeMgr) *mockModules.MockBus {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockBus := mockModules.NewMockBus(ctrl)
	mockModulesRegistry := mockModules.NewMockModulesRegistry(ctrl)

	mockBus.EXPECT().
		GetRuntimeMgr().
		Return(runtimeMgr).
		AnyTimes()
	mockBus.EXPECT().
		RegisterModule(gomock.Any()).
		DoAndReturn(func(m modules.Submodule) {
			m.SetBus(mockBus)
		}).
		AnyTimes()
	mockModulesRegistry.EXPECT().
		GetModule(peerstore_provider.PeerstoreProviderSubmoduleName).
		Return(nil, runtime.ErrModuleNotRegistered(peerstore_provider.PeerstoreProviderSubmoduleName)).
		AnyTimes()
	mockModulesRegistry.EXPECT().
		GetModule(modules.CurrentHeightProviderSubmoduleName).
		Return(nil, runtime.ErrModuleNotRegistered(modules.CurrentHeightProviderSubmoduleName)).
		AnyTimes()
	mockBus.EXPECT().
		GetModulesRegistry().
		Return(mockModulesRegistry).
		AnyTimes()
	mockBus.EXPECT().
		PublishEventToBus(gomock.Any()).
		AnyTimes()

	return mockBus
}
