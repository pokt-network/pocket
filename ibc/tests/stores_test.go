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
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

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

	testCases := []struct {
		key           []byte
		value         []byte
		nonmembership bool
	}{
		{ // Successfully generates and verifies a membership proof for a key stored
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
		},
		{ // Successfully generates and verifies a non-membership proof for a key not stored
			key:           []byte("baz"),
			value:         nil,
			nonmembership: true,
		},
		{ // Successfully generates and verifies a non-membership proof for an unset nil key
			key:           nil,
			value:         nil,
			nonmembership: true,
		},
	}

	for _, tc := range testCases {
		var proof *coreTypes.CommitmentProof
		if tc.nonmembership {
			proof, err = store.CreateNonMembershipProof(tc.key)
			require.NoError(t, err)
			require.NotNil(t, proof)
			require.NotNil(t, proof.GetNonMembershipLeafData())
		} else {
			proof, err = store.CreateMembershipProof(tc.key, tc.value)
			require.NoError(t, err)
			require.NotNil(t, proof)
			require.Nil(t, proof.GetNonMembershipLeafData())
		}
		require.Equal(t, tc.key, proof.GetKey())
		require.Equal(t, tc.value, proof.GetValue())
	}

	err = store.Stop()
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

	memProof, err := store.CreateMembershipProof([]byte("foo"), []byte("bar"))
	require.NoError(t, err)
	require.NotNil(t, memProof)
	require.Nil(t, memProof.GetNonMembershipLeafData())

	nonMemProof, err := store.CreateNonMembershipProof([]byte("baz"))
	require.NoError(t, err)
	require.NotNil(t, nonMemProof)
	require.NotNil(t, nonMemProof.GetNonMembershipLeafData())

	testCases := []struct {
		modify        func(proof *coreTypes.CommitmentProof)
		key           []byte
		value         []byte
		nonmembership bool
		valid         bool
		fail          bool
		expected      error
	}{
		{ // Successfully verifies a membership proof for a key-value stored pair
			modify:        nil,
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         true,
			fail:          false,
			expected:      nil,
		},
		{ // Successfully verifies a non-membership proof for a key-value pair not stored
			modify:        nil,
			key:           []byte("baz"),
			value:         nil,
			nonmembership: true,
			valid:         true,
			fail:          false,
			expected:      nil,
		},
		{ // Fails to verify proof for an invalid value
			modify:        nil,
			key:           []byte("foo"),
			value:         []byte("baz"),
			nonmembership: false,
			valid:         false,
			fail:          true,
			expected:      coreTypes.ErrInvalidProof("provided value does not match proof value"),
		},
		{ // Fails to verify proof for an invalid key
			modify:        nil,
			key:           []byte("baz"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         false,
			fail:          true,
			expected:      coreTypes.ErrInvalidProof("provided key does not match proof key"),
		},
		{ // Fails to verify membership proof with non nil non membership leaf data
			modify: func(proof *coreTypes.CommitmentProof) {
				proof.NonMembershipLeafData = []byte("invalid non membership data")
			},
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         false,
			fail:          true,
			expected:      coreTypes.ErrInvalidProof("non membership leaf data must be nil"),
		},
		{ // Fails to verify non-membership proof with nil non membership leaf data
			modify: func(proof *coreTypes.CommitmentProof) {
				proof.NonMembershipLeafData = nil
			},
			key:           []byte("baz"),
			value:         nil,
			nonmembership: true,
			valid:         false,
			fail:          true,
			expected:      coreTypes.ErrInvalidProof("non membership leaf data must not be nil"),
		},
		{ // Fails to verify membership proof when computed root is different from provided root
			modify: func(proof *coreTypes.CommitmentProof) {
				proof.Value = []byte("new value")
			},
			key:           []byte("foo"),
			value:         []byte("new value"),
			nonmembership: false,
			valid:         false,
			fail:          false,
			expected:      nil,
		},
		{ // Fails to verify non-membership proof when key exists
			modify: func(proof *coreTypes.CommitmentProof) { //nolint:unused // param not used in modify function
				err := store.Set([]byte("baz"), []byte("bar")) // Set a value at the missing key
				require.NoError(t, err)
				root = store.Root() // Get new root
			},
			key:           []byte("baz"),
			value:         nil,
			nonmembership: true,
			valid:         false,
			fail:          false,
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		var proof *coreTypes.CommitmentProof
		if tc.nonmembership {
			proof = codec.GetCodec().Clone(nonMemProof).(*coreTypes.CommitmentProof)
		} else {
			proof = codec.GetCodec().Clone(memProof).(*coreTypes.CommitmentProof)
		}

		if tc.modify != nil {
			tc.modify(proof)
		}

		var valid bool
		var err error
		if tc.nonmembership {
			valid, err = ibc.VerifyNonMembership(store.TreeSpec(), proof, root, tc.key)
		} else {
			valid, err = ibc.VerifyMembership(store.TreeSpec(), proof, root, tc.key, tc.value)
		}

		if tc.fail {
			require.EqualError(t, err, tc.expected.Error())
		} else {
			require.NoError(t, err)
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
