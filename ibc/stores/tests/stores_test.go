package tests

import (
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

	sm := stores.NewStoreManager("")
	for i := 0; i < 3; i++ {
		err := sm.AddExistingStore(initialStores[i])
		require.NoError(t, err)
	}

	testCases := []struct {
		name     string
		store    modules.Store
		op       string
		fail     bool
		expected error
	}{
		{
			name:     "Fails to add store that is already in the store manager",
			store:    store1,
			op:       "add",
			fail:     true,
			expected: coreTypes.ErrIBCStoreAlreadyExists("test1"),
		},
		{
			name:     "Successfully returns store with matching store key when present",
			store:    store2,
			op:       "get",
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully deletes store with matching store key when present",
			store:    store3,
			op:       "remove",
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to delete store with matching store key when not present",
			store:    store4,
			op:       "remove",
			fail:     true,
			expected: coreTypes.ErrIBCStoreNotFound("test4"),
		},
		{
			name:     "Successfully adds a store to the store manager when not already present",
			store:    store4,
			op:       "add",
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fail to get store with matching store key when not present",
			store:    store3,
			op:       "get",
			fail:     true,
			expected: coreTypes.ErrIBCStoreNotFound("test3"),
		},
		{
			name:     "Successfully returns provable store instance for provable stores",
			store:    store4,
			op:       "getprovable",
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to return a provable store when the store is private",
			store:    store1,
			op:       "getprovable",
			fail:     true,
			expected: coreTypes.ErrIBCStoreNotProvable("test1"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.op {
			case "add":
				err := sm.AddExistingStore(tc.store)
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
		})
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
		name     string
		op       string
		key      []byte
		value    []byte
		fail     bool
		expected error
	}{
		{
			name:     "Successfully sets a value in the store",
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("baz"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully updates a value in the store",
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to set value to nil key",
			op:       "set",
			key:      nil,
			value:    []byte("bar"),
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{
			name:     "Fails to set a value to a key that is too large",
			op:       "set",
			key:      invalidKey[:],
			value:    []byte("bar"),
			fail:     true,
			expected: errors.Errorf("Key with size 65001 exceeded 65000 limit. Key:\n%s", hex.Dump(invalidKey[:1<<10])),
		},
		{
			name:     "Successfully manages to retrieve a value from the store",
			op:       "get",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to get a value that is not stored",
			op:       "get",
			key:      []byte("bar"),
			value:    nil,
			fail:     true,
			expected: badger.ErrKeyNotFound,
		},
		{
			name:     "Fails when the key is empty",
			op:       "get",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{
			name:     "Successfully deletes a value in the store",
			op:       "delete",
			key:      []byte("foo"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to delete a value not in the store",
			op:       "delete",
			key:      []byte("bar"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to set value to nil key",
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: badger.ErrEmptyKey,
		},
		{
			name:     "Fails to set a value to a key that is too large",
			op:       "delete",
			key:      invalidKey[:],
			value:    nil,
			fail:     true,
			expected: errors.Errorf("Key with size 65001 exceeded 65000 limit. Key:\n%s", hex.Dump(invalidKey[:1<<10])),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}

	err := store.Stop()
	require.NoError(t, err)
}

func TestProvableStore_StoreOperations(t *testing.T) {
	store := stores.NewTestProvableStore("test1", nil)
	require.Equal(t, store.GetStoreKey(), "test1")
	require.True(t, store.IsProvable())

	testCases := []struct {
		name     string
		op       string
		key      []byte
		value    []byte
		fail     bool
		expected error
	}{
		{
			name:     "Successfully sets a value in the store",
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("baz"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully updates a value in the store",
			op:       "set",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully sets a nil key to value",
			op:       "set",
			key:      nil,
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully deletes value stored at nil key",
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully manages to retrieve a value from the store",
			op:       "get",
			key:      []byte("foo"),
			value:    []byte("bar"),
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully returns default value for a key that is not stored",
			op:       "get",
			key:      []byte("bar"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully returns defaultValue for a nil path",
			op:       "get",
			key:      nil,
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Successfully deletes a value in the store",
			op:       "delete",
			key:      []byte("foo"),
			value:    nil,
			fail:     false,
			expected: nil,
		},
		{
			name:     "Fails to delete a value not in the store",
			op:       "delete",
			key:      []byte("bar"),
			value:    nil,
			fail:     true,
			expected: coreTypes.ErrIBCStoreUpdate(smt.ErrKeyNotPresent),
		},
		{
			name:     "Fails to delete a nil key",
			op:       "delete",
			key:      nil,
			value:    nil,
			fail:     true,
			expected: coreTypes.ErrIBCStoreUpdate(smt.ErrKeyNotPresent),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
					require.Equal(t, got, []byte(nil))
				}
			}
		})
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
	require.Equal(t, val, preDigest)

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
		name          string
		store         modules.ProvableStore
		key           []byte
		value         []byte
		nonmembership bool
		fails         bool
		expected      error
	}{
		{
			name:          "Successfully generates a membership proof for a key stored",
			store:         store1,
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			fails:         false,
			expected:      nil,
		},
		{
			name:          "Successfully generates a non-membership proof for a key not stored",
			store:         store1,
			key:           []byte("baz"),
			value:         []byte("foo2"), // unrelated leaf data
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
		{
			name:          "Successfully generates a non-membership proof for an unset nil key",
			store:         store1,
			key:           nil,
			value:         []byte("foo"), // unrelated leaf data
			nonmembership: true,
			fails:         false,
			expected:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var proof *ics23.CommitmentProof
			if tc.nonmembership {
				proof, err = tc.store.CreateNonMembershipProof(tc.key)
			} else {
				proof, err = tc.store.CreateMembershipProof(tc.key, tc.value)
			}
			if tc.fails {
				require.EqualError(t, err, tc.expected.Error())
				require.Nil(t, proof)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, proof)
			require.Equal(t, tc.value, proof.GetExist().GetValue())
			require.NotNil(t, proof.GetExist().GetLeaf())
			require.NotNil(t, proof.GetExist().GetPath())
		})
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
		name          string
		modify        func(proof *ics23.CommitmentProof)
		key           []byte
		value         []byte
		nonmembership bool
		valid         bool
	}{
		{
			name:          "Successfully verifies a membership proof for a key-value stored pair",
			modify:        nil,
			key:           []byte("foo"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         true,
		},
		{
			name:          "Successfully verifies a non-membership proof for a key-value pair not stored",
			modify:        nil,
			key:           []byte("not stored"),
			value:         nil,
			nonmembership: true,
			valid:         true,
		},
		{
			name:          "Fails to verify a membership proof for a key-value pair not stored",
			modify:        nil,
			key:           []byte("baz"),
			value:         []byte("bar"),
			nonmembership: false,
			valid:         false,
		},
		{
			name: "Fails to verify a non-membership proof for a key-value pair stored",
			modify: func(proof *ics23.CommitmentProof) {
				proof.GetExist().Value = []byte("bar")
			},
			key:           []byte("foo"),
			value:         nil,
			nonmembership: true,
			valid:         false,
		},
		{
			name:          "Fails to verify a non-membership proof for a key stored in the tree",
			modify:        nil,
			key:           []byte("foo"),
			value:         nil,
			nonmembership: true,
			valid:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}

	err = store.Stop()
	require.NoError(t, err)
}
