package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"github.com/pokt-network/pocket/ibc"
	stores "github.com/pokt-network/pocket/ibc/stores"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

func TestStoreManager_StoreManagerOperations(t *testing.T) {
	store1, err := stores.NewTestPrivateStore("test1")
	require.NoError(t, err)
	require.Equal(t, store1.GetStoreKey(), "test1")
	require.False(t, store1.IsProvable())
	store2, err := stores.NewTestPrivateStore("test2")
	require.NoError(t, err)
	require.Equal(t, store2.GetStoreKey(), "test2")
	require.False(t, store2.IsProvable())
	store3, err := stores.NewTestPrivateStore("test3")
	require.NoError(t, err)
	require.Equal(t, store3.GetStoreKey(), "test3")
	require.False(t, store3.IsProvable())
	store4, err := stores.NewTestProvableStore("test4", nil)
	require.NoError(t, err)
	require.Equal(t, store4.GetStoreKey(), "test4")
	require.True(t, store4.IsProvable())

	initialStores := []modules.Store{store1, store2, store3}

	sm := stores.NewStoreManager()
	for i := 0; i < 3; i++ {
		err := sm.AddStore(initialStores[i])
		require.NoError(t, err)
	}

	testCases := []struct {
		store    modules.Store
		op       string
		fail     bool
		expected error
	}{
		{ // Fails to add store that is already in the store manager
			store:    store1,
			op:       "add",
			fail:     true,
			expected: coreTypes.ErrStoreAlreadyExists("test1"),
		},
		{ // Successfully returns store with matching store key when present
			store:    store2,
			op:       "get",
			fail:     false,
			expected: nil,
		},
		{ // Successfully deletes store with matching store key when present
			store:    store3,
			op:       "remove",
			fail:     false,
			expected: nil,
		},
		{ // Fails to delete store with matching store key when not present
			store:    store4,
			op:       "remove",
			fail:     true,
			expected: coreTypes.ErrStoreNotFound("test4"),
		},
		{ // Successfully adds a store to the store manager when not already present
			store:    store4,
			op:       "add",
			fail:     false,
			expected: nil,
		},
		{ // Fail to get store with matching store key when not present
			store:    store3,
			op:       "get",
			fail:     true,
			expected: coreTypes.ErrStoreNotFound("test3"),
		},
		{ // Successfully returns provable store instance for provable stores
			store:    store4,
			op:       "getprovable",
			fail:     false,
			expected: nil,
		},
		{ // Fails to return a provable store when the store is private
			store:    store1,
			op:       "getprovable",
			fail:     true,
			expected: coreTypes.ErrStoreNotProvable("test1"),
		},
	}

	for _, tc := range testCases {
		switch tc.op {
		case "add":
			err = sm.AddStore(tc.store)
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		case "get":
			store, err := sm.GetStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.Equal(t, store.GetStoreKey(), tc.store.GetStoreKey())
				require.NoError(t, err)
			}
		case "getprovable":
			store, err := sm.GetProvableStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.Equal(t, store.GetStoreKey(), tc.store.GetStoreKey())
				require.True(t, store.IsProvable())
				require.NoError(t, err)
			}
		case "remove":
			err = sm.RemoveStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		}
	}

	err = store1.Stop()
	require.NoError(t, err)
	err = store2.Stop()
	require.NoError(t, err)
	err = store3.Stop()
	require.NoError(t, err)
	err = store4.Stop()
	require.NoError(t, err)
}

func TestPrivateStore_StoreOperations(t *testing.T) {
	store, err := stores.NewTestPrivateStore("test1")
	require.NoError(t, err)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.False(t, store.IsProvable())

	invalidKey := [65001]byte{}
	testCases := []struct {
		op       string
		key      []byte
		value    []byte
		fail     bool
		expected error
	}{
		{ // Successfully sets a value in the store
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("baz"),
			fail:     false,
			expected: nil,
		},
		{ // Successfully updates a value in the store
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{ // Fails to set value to nil key
			op:       "set",
			key:      nil,
			value:    []byte("bar"),
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{ // Fails to set a value to a key that is too large
			op:       "set",
			key:      invalidKey[:],
			value:    []byte("bar"),
			fail:     true,
			expected: errors.Errorf("Key with size 65001 exceeded 65000 limit. Key:\n%s", hex.Dump(invalidKey[:1<<10])),
		},
		{ // Successfully manages to retrieve a value from the store
			op:       "get",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{ // Fails to get a value that is not stored
			op:       "get",
			key:      []byte("bar"),
			value:    nil,
			fail:     true,
			expected: badger.ErrKeyNotFound,
		},
		{ // Fails when the key is empty
			op:       "get",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{ // Successfully deletes a value in the store
			op:       "delete",
			key:      []byte("foo"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Fails to delete a value not in the store
			op:       "delete",
			key:      []byte("bar"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Fails to set value to nil key
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{ // Fails to set a value to a key that is too large
			op:       "delete",
			key:      invalidKey[:],
			value:    nil,
			fail:     true,
			expected: errors.Errorf("Key with size 65001 exceeded 65000 limit. Key:\n%s", hex.Dump(invalidKey[:1<<10])),
		},
	}

	for _, tc := range testCases {
		switch tc.op {
		case "set":
			err := store.Set(tc.key, tc.value)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				got, err := store.Get(tc.key)
				require.NoError(t, err)
				require.Equal(t, tc.value, got)
			}
		case "get":
			got, err := store.Get(tc.key)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.value, got)
			}
		case "delete":
			err := store.Delete(tc.key)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				_, err := store.Get(tc.key)
				require.EqualError(t, err, badger.ErrKeyNotFound.Error())
			}
		}
	}

	err = store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_StoreOperations(t *testing.T) {
	store, err := stores.NewTestProvableStore("test1", nil)
	require.NoError(t, err)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.True(t, store.IsProvable())

	testCases := []struct {
		op       string
		key      []byte
		value    []byte
		fail     bool
		expected error
	}{
		{ // Successfully sets a value in the store
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("baz"),
			fail:     false,
			expected: nil,
		},
		{ // Successfully updates a value in the store
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{ // Successfully sets a nil key to value
			op:       "set",
			key:      nil,
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{ // Successfully deletes value stored at nil key
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Successfully manages to retrieve a value from the store
			op:       "get",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{ // Successfully returns default value for a key that is not stored
			op:       "get",
			key:      []byte("bar"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Successfully returns defaultValue for a nil path
			op:       "get",
			key:      nil,
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Successfully deletes a value in the store
			op:       "delete",
			key:      []byte("foo"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{ // Fails to delete a value not in the store
			op:       "delete",
			key:      []byte("bar"),
			value:    nil,
			fail:     true,
			expected: coreTypes.ErrStoreUpdate(smt.ErrKeyNotPresent),
		},
		{ // Fails to delete a nil key
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: coreTypes.ErrStoreUpdate(smt.ErrKeyNotPresent),
		},
	}

	for _, tc := range testCases {
		switch tc.op {
		case "set":
			err := store.Set(tc.key, tc.value)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				got, err := store.Get(tc.key)
				require.NoError(t, err)
				require.Equal(t, tc.value, got)
			}
		case "get":
			got, err := store.Get(tc.key)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.value, got)
			}
		case "delete":
			err := store.Delete(tc.key)
			if tc.fail {
				require.Error(t, err)
				require.EqualError(t, tc.expected, err.Error())
			} else {
				require.NoError(t, err)
				got, err := store.Get(tc.key)
				require.NoError(t, err)
				require.True(t, bytes.Equal(got, []byte{}))
			}
		}
	}

	err = store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_UpdatesPersist(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	store, err := stores.NewTestProvableStore("test1", nodeStore)
	require.NoError(t, err)

	// Set a value in the store
	err = store.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)

	// Calculate path and leaf digest of stored value
	hasher := sha256.New()
	hasher.Write([]byte("foo"))
	sum := hasher.Sum(nil)
	hasher.Reset()
	preDigest := []byte{0}
	preDigest = append(preDigest, sum[:]...)
	preDigest = append(preDigest, []byte("bar")...)
	hasher.Write(preDigest)
	digest := hasher.Sum(nil)

	// Check that the value stored in  the nodeStore is the pre-hashed digest of the leaf
	val, err := nodeStore.Get(digest[:])
	require.NoError(t, err)
	require.True(t, bytes.Equal(val, preDigest))

	// Delete the value from the store
	err = store.Delete([]byte("foo"))
	require.NoError(t, err)

	// Check the nodeStore no longer contains the pre-hashed digest of the leaf
	_, err = nodeStore.Get(digest[:])
	require.EqualError(t, err, badger.ErrKeyNotFound.Error())

	err = store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_GenerateCommitmentProofs(t *testing.T) {
	store, err := stores.NewTestProvableStore("test1", nil)
	require.NoError(t, err)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.True(t, store.IsProvable())

	// Set a value in the store
	err = store.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	// err = store.Set([]byte("foo2"), []byte("bar2"))
	// require.NoError(t, err)

	root := store.Root()

	// Create membership proof
	proof, err := store.CreateMembershipProof([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Basic validation of the proof
	require.Equal(t, []byte("foo"), proof.GetKey())
	require.Equal(t, []byte("bar"), proof.GetValue())
	require.Nil(t, proof.GetNonMembershipLeafData())

	// Verify the proof
	valid, err := ibc.VerifyMembership(store.TreeSpec(), proof, root, []byte("foo"), []byte("bar"))
	require.NoError(t, err)
	require.True(t, valid)

	err = store.Stop()
	require.NoError(t, err)
}
