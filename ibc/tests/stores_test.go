package tests

import (
	"testing"

	stores "github.com/pokt-network/pocket/ibc/stores"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestStoreManager(t *testing.T) {
	store1, err := stores.NewTestStore("test1", false)
	require.NoError(t, err)
	store2, err := stores.NewTestStore("test2", false)
	require.NoError(t, err)
	store3, err := stores.NewTestStore("test3", false)
	require.NoError(t, err)
	store4, err := stores.NewTestStore("test4", false)
	require.NoError(t, err)
	initialStores := []modules.Store{store1, store2, store3}

	sm := stores.NewStoreManager()
	for i := range initialStores {
		err := sm.AddStore(initialStores[i])
		require.NoError(t, err)
	}

	testCases := []struct {
		store     modules.Store
		operation string
		fail      bool
		expected  error
	}{
		{ // Fails to add store that is already in the store manager
			store:     store1,
			operation: "add",
			fail:      true,
			expected:  coreTypes.ErrStoreAlreadyExists("test1"),
		},
		{ // Successfully returns store with matching store key when present
			store:     store2,
			operation: "get",
			fail:      false,
			expected:  nil,
		},
		{ // Successfully deletes store with matching store key when present
			store:     store3,
			operation: "remove",
			fail:      false,
			expected:  nil,
		},
		{ // Fails to delete store with matching store key when not present
			store:     store4,
			operation: "remove",
			fail:      true,
			expected:  coreTypes.ErrStoreNotFound("test4"),
		},
		{ // Successfully adds a store to the store manager when not already present
			store:     store4,
			operation: "add",
			fail:      false,
			expected:  nil,
		},
		{ // Fail to get store with matching store key when not present
			store:     store3,
			operation: "get",
			fail:      true,
			expected:  coreTypes.ErrStoreNotFound("test3"),
		},
	}

	for _, tc := range testCases {
		switch tc.operation {
		case "add":
			err := sm.AddStore(tc.store)
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		case "get":
			_, err := sm.GetStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		case "remove":
			err := sm.RemoveStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		}
	}
}
