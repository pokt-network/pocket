package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	stores "github.com/pokt-network/pocket/ibc/stores"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

var defaultValue []byte = nil

func TestStoreManager_StoreManagerOperations(t *testing.T) {
	store1 := stores.NewTestPrivateStore("test1")
	require.Equal(t, store1.GetStoreKey(), "test1")
	require.False(t, store1.IsProvable())
	store2 := stores.NewTestPrivateStore("test2")
	require.Equal(t, store2.GetStoreKey(), "test2")
	require.False(t, store2.IsProvable())
	store3 := stores.NewTestPrivateStore("test3")
	require.Equal(t, store3.GetStoreKey(), "test3")
	require.False(t, store3.IsProvable())
	store4 := stores.NewTestProvableStore("test4", nil)
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
			err := sm.AddStore(tc.store)
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
			err := sm.RemoveStore(tc.store.GetStoreKey())
			if tc.fail {
				require.Error(t, err)
				require.Equal(t, tc.expected, err)
			} else {
				require.NoError(t, err)
			}
		}
	}

	err := store1.Stop()
	require.NoError(t, err)
	err = store2.Stop()
	require.NoError(t, err)
	err = store3.Stop()
	require.NoError(t, err)
	err = store4.Stop()
	require.NoError(t, err)
}

func TestPrivateStore_StoreOperations(t *testing.T) {
	store := stores.NewTestPrivateStore("test1")
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

	err := store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_StoreOperations(t *testing.T) {
	store := stores.NewTestProvableStore("test1", nil)
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

	err := store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_UpdatesPersist(t *testing.T) {
	nodeStore := kvstore.NewMemKVStore()
	store := stores.NewTestProvableStore("test1", nodeStore)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.True(t, store.IsProvable())

	// Set a value in the store
	err := store.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)

	// Calculate path and leaf digest of stored value
	hasher := sha256.New()
	hasher.Write([]byte("foo"))
	sum := hasher.Sum(nil)
	hasher.Reset()
	preDigest := []byte{0}
	preDigest = append(preDigest, sum...)
	preDigest = append(preDigest, []byte("bar")...)
	hasher.Write(preDigest)
	digest := hasher.Sum(nil)

	// Check that the value stored in  the nodeStore is the pre-hashed digest of the leaf
	val, err := nodeStore.Get(digest)
	require.NoError(t, err)
	require.True(t, bytes.Equal(val, preDigest))

	// Delete the value from the store
	err = store.Delete([]byte("foo"))
	require.NoError(t, err)

	// Check the nodeStore no longer contains the pre-hashed digest of the leaf
	_, err = nodeStore.Get(digest)
	require.EqualError(t, err, badger.ErrKeyNotFound.Error())

	err = store.Stop()
	require.NoError(t, err)
}

// Proof generation cannot fail but verification can
func TestProvableStore_GenerateCommitmentProofs(t *testing.T) {
	store1 := stores.NewTestProvableStore("test1", nil)
	require.Equal(t, store1.GetStoreKey(), "test1")
	require.True(t, store1.IsProvable())

	// Set a value in the store
	err := store1.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	err = store1.Set([]byte("foo2"), []byte("bar2"))
	require.NoError(t, err)
	err = store1.Set([]byte("bar"), []byte("foo"))
	require.NoError(t, err)
	err = store1.Set([]byte("bar2"), []byte("foo2"))
	require.NoError(t, err)

	testCases := []struct {
		store         modules.ProvableStore
		key           []byte
		value         []byte
		nonmembership bool
		fails         bool
		expected      error
	}{
		{ // Successfully generates and verifies a membership proof for a key stored
			store:         store1,
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			fails:         false,
			expected:      nil,
		},
		{ // Successfully generates and verifies a non-membership proof for a key not stored
			store: store1,
			key:   []byte("baz"),
			// unrelated leaf data
			value: []byte{
				0x0, 0x84, 0x4b, 0x7f, 0xde, 0xcd, 0xa0, 0x76, 0x33,
				0x76, 0x96, 0xdb, 0xd, 0xca, 0xc0, 0xce, 0x9d, 0x83, 0x83, 0x19, 0xac, 0x0, 0xb2, 0x43, 0x80, 0xd6,
				0xc6, 0x2d, 0x44, 0x2d, 0x6, 0x80, 0xaf, 0x66, 0x6f, 0x6f, 0x32,
			},
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
		{ // Successfully generates and verifies a non-membership proof for an unset nil key
			store: store1,
			key:   nil,
			// unrelated leaf data
			value: []byte{
				0x0, 0xfc, 0xde, 0x2b, 0x2e, 0xdb, 0xa5, 0x6b, 0xf4,
				0x8, 0x60, 0x1f, 0xb7, 0x21, 0xfe, 0x9b, 0x5c, 0x33, 0x8d, 0x10, 0xee, 0x42, 0x9e, 0xa0, 0x4f, 0xae,
				0x55, 0x11, 0xb6, 0x8f, 0xbf, 0x8f, 0xb9, 0x66, 0x6f, 0x6f,
			},
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		var proof *ics23.CommitmentProof
		if tc.nonmembership {
			proof, err = tc.store.CreateNonMembershipProof(tc.key)
		} else {
			proof, err = tc.store.CreateMembershipProof(tc.key, tc.value)
		}
		if tc.fails {
			require.EqualError(t, err, tc.expected.Error())
			require.Nil(t, proof)
			continue
		}
		require.NoError(t, err)
		require.NotNil(t, proof)
		require.Equal(t, tc.key, proof.GetExist().GetKey())
		require.Equal(t, tc.value, proof.GetExist().GetValue())
		require.NotNil(t, proof.GetExist().GetLeaf())
		require.NotNil(t, proof.GetExist().GetPath())
	}

	err = store1.Stop()
	require.NoError(t, err)
}

func TestProvableStore_VerifyCommitmentProofs(t *testing.T) {
	store := stores.NewTestProvableStore("test1", nil)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.True(t, store.IsProvable())

	// Set a value in the store
	err := store.Set([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	err = store.Set([]byte("foo2"), []byte("bar2"))
	require.NoError(t, err)
	err = store.Set([]byte("bar"), []byte("foo"))
	require.NoError(t, err)
	err = store.Set([]byte("bar2"), []byte("foo2"))
	require.NoError(t, err)

	root := store.Root()
	require.NotNil(t, root)

	testCases := []struct {
		modify        func(proof *ics23.CommitmentProof)
		key           []byte
		value         []byte
		nonmembership bool
		valid         bool
	}{
		{ // Successfully verifies a membership proof for a key-value stored pair
			modify:        nil,
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         true,
		},
		{ // Successfully verifies a non-membership proof for a key-value pair not stored
			modify:        nil,
			key:           []byte("foo"),
			value:         nil,
			nonmembership: true,
			valid:         true,
		},
		// TODO: Add more test cases
	}

	for _, tc := range testCases {
		proof := new(ics23.CommitmentProof)
		var err error

		if tc.nonmembership {
			proof, err = store.CreateNonMembershipProof(tc.key)
		} else {
			proof, err = store.CreateMembershipProof(tc.key, tc.value)
		}
		require.NoError(t, err)
		require.NotNil(t, proof)
		require.NotNil(t, proof.GetExist())

		if tc.modify != nil {
			tc.modify(proof)
		}

		var valid bool
		if tc.nonmembership {
			valid = stores.VerifyNonMembership(root, proof, tc.key)
		} else {
			valid = stores.VerifyMembership(root, proof, tc.key, tc.value)
		}

		if tc.valid {
			require.True(t, valid)
		} else {
			require.False(t, valid)
		}
	}

	err = store.Stop()
	require.NoError(t, err)
}
