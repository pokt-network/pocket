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
	priStore1, err := stores.NewTestPrivateStore("priTest1")
	require.NoError(t, err)
	require.Equal(t, priStore1.GetStoreKey(), "priTest1")
	priStore2, err := stores.NewTestPrivateStore("priTest2")
	require.NoError(t, err)
	require.Equal(t, priStore2.GetStoreKey(), "priTest2")
	priStore3, err := stores.NewTestPrivateStore("priTest3")
	require.NoError(t, err)
	require.Equal(t, priStore3.GetStoreKey(), "priTest3")
	priStore4, err := stores.NewTestPrivateStore("priTest4")
	require.NoError(t, err)
	require.Equal(t, priStore4.GetStoreKey(), "priTest4")

	provStore1, err := stores.NewTestProvableStore("provTest1", nil)
	require.NoError(t, err)
	require.Equal(t, provStore1.GetStoreKey(), "provTest1")
	provStore2, err := stores.NewTestProvableStore("provTest2", nil)
	require.NoError(t, err)
	require.Equal(t, provStore2.GetStoreKey(), "provTest2")
	provStore3, err := stores.NewTestProvableStore("provTest3", nil)
	require.NoError(t, err)
	require.Equal(t, provStore3.GetStoreKey(), "provTest3")
	provStore4, err := stores.NewTestProvableStore("provTest4", nil)
	require.NoError(t, err)
	require.Equal(t, provStore4.GetStoreKey(), "provTest4")

	initialPriStores := []modules.PrivateStore{priStore1, priStore2, priStore3}
	initialProvStores := []modules.ProvableStore{provStore1, provStore2, provStore3}

	sm := stores.NewStoreManager()
	for i := 0; i < 3; i++ {
		err := sm.AddPrivateStore(initialPriStores[i])
		require.NoError(t, err)
		err = sm.AddProvableStore(initialProvStores[i])
		require.NoError(t, err)
	}

	testCases := []struct {
		privateStore  modules.PrivateStore
		provableStore modules.ProvableStore
		op            string
		fail          bool
		expected      error
	}{
		{ // Fails to add store that is already in the store manager
			privateStore:  priStore1,
			provableStore: nil,
			op:            "add",
			fail:          true,
			expected:      coreTypes.ErrStoreAlreadyExists("priTest1"),
		},
		{ // Successfully returns store with matching store key when present
			privateStore:  priStore2,
			provableStore: nil,
			op:            "get",
			fail:          false,
			expected:      nil,
		},
		{ // Successfully deletes store with matching store key when present
			privateStore:  priStore3,
			provableStore: nil,
			op:            "remove",
			fail:          false,
			expected:      nil,
		},
		{ // Fails to delete store with matching store key when not present
			privateStore:  priStore4,
			provableStore: nil,
			op:            "remove",
			fail:          true,
			expected:      coreTypes.ErrStoreNotFound("priTest4"),
		},
		{ // Successfully adds a store to the store manager when not already present
			privateStore:  priStore4,
			provableStore: nil,
			op:            "add",
			fail:          false,
			expected:      nil,
		},
		{ // Fail to get store with matching store key when not present
			privateStore:  priStore3,
			provableStore: nil,
			op:            "get",
			fail:          true,
			expected:      coreTypes.ErrStoreNotFound("priTest3"),
		},
		{ // Fails to add store that is already in the store manager
			provableStore: provStore1,
			privateStore:  nil,
			op:            "add",
			fail:          true,
			expected:      coreTypes.ErrStoreAlreadyExists("provTest1"),
		},
		{ // Successfully returns store with matching store key when present
			provableStore: provStore2,
			privateStore:  nil,
			op:            "get",
			fail:          false,
			expected:      nil,
		},
		{ // Successfully deletes store with matching store key when present
			provableStore: provStore3,
			privateStore:  nil,
			op:            "remove",
			fail:          false,
			expected:      nil,
		},
		{ // Fails to delete store with matching store key when not present
			provableStore: provStore4,
			privateStore:  nil,
			op:            "remove",
			fail:          true,
			expected:      coreTypes.ErrStoreNotFound("provTest4"),
		},
		{ // Successfully adds a store to the store manager when not already present
			provableStore: provStore4,
			privateStore:  nil,
			op:            "add",
			fail:          false,
			expected:      nil,
		},
		{ // Fail to get store with matching store key when not present
			provableStore: provStore3,
			privateStore:  nil,
			op:            "get",
			fail:          true,
			expected:      coreTypes.ErrStoreNotFound("provTest3"),
		},
	}

	for _, tc := range testCases {
		switch tc.op {
		case "add":
			var err error
			if tc.privateStore != nil {
				err = sm.AddPrivateStore(tc.privateStore)
			} else {
				err = sm.AddProvableStore(tc.provableStore)
			}
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		case "get":
			if tc.privateStore != nil {
				store, err := sm.GetPrivateStore(tc.privateStore.GetStoreKey())
				if tc.fail {
					require.Error(t, err)
					require.Equal(t, tc.expected, err)
				} else {
					require.Equal(t, store.GetStoreKey(), tc.privateStore.GetStoreKey())
					require.NoError(t, err)
				}
				continue
			}
			store, err := sm.GetProvableStore(tc.provableStore.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.Equal(t, store.GetStoreKey(), tc.provableStore.GetStoreKey())
				require.NoError(t, err)
			}
		case "remove":
			var err error
			if tc.privateStore != nil {
				err = sm.RemoveStore(tc.privateStore.GetStoreKey())
			} else {
				err = sm.RemoveStore(tc.provableStore.GetStoreKey())
			}
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		}
	}

	err = priStore1.Stop()
	require.NoError(t, err)
	err = priStore2.Stop()
	require.NoError(t, err)
	err = priStore3.Stop()
	require.NoError(t, err)
	err = priStore4.Stop()
	require.NoError(t, err)
	err = provStore1.Stop()
	require.NoError(t, err)
	err = provStore2.Stop()
	require.NoError(t, err)
	err = provStore3.Stop()
	require.NoError(t, err)
	err = provStore4.Stop()
	require.NoError(t, err)
}

func TestPrivateStore_StoreOperations(t *testing.T) {
	store, err := stores.NewTestPrivateStore("test1")
	require.NoError(t, err)
	require.Equal(t, store.GetStoreKey(), "test1")

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
