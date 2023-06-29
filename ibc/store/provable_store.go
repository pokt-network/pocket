package store

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/ibc/host"
	"github.com/pokt-network/pocket/persistence/kvstore"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
)

var _ modules.ProvableStore = &provableStore{}

// CachedEntry represents a local change made to the IBC store prior to it being
// committed to the state tree. These should be written to disk in the to prevent a
// loss of data and pruned when included in the state tree
// written to disk as follows:
// "{height}/{prefixedKey}" => value
type cachedEntry struct {
	storeName   string
	height      uint64
	prefixedKey []byte
	value       []byte
}

// prepare returns the key and value to be written to disk
func (c *cachedEntry) prepare() (key, value []byte) {
	return []byte(fmt.Sprintf("%s/%d/%s", c.storeName, c.height, string(c.prefixedKey))), c.value
}

// provableStore is a struct that interfaces with the PostgresDB instance
// obtained from Persistence. It is used to Get/Set/Delete the keys in the
// IBC state tree, in doing so it will trigger the creation of
type provableStore struct {
	bus    modules.Bus                // used to interact with persistence (passed from IBCHost)
	name   string                     // store name in storeManager
	prefix coreTypes.CommitmentPrefix // []byte(name)
	cache  []*cachedEntry             // in-memory cache of local changes to be written to disk
}

// newProvableStore returns a new instance of provableStore with the bus and prefix provided
func newProvableStore(bus modules.Bus, prefix coreTypes.CommitmentPrefix) *provableStore {
	return &provableStore{
		bus:    bus,
		name:   string(prefix),
		prefix: prefix,
		cache:  make([]*cachedEntry, 0),
	}
}

// Get queries the persistence layer for the latest value stored in the IBC state tree
// keys are automatically prefixed with the CommitmentPrefix if not present
func (p *provableStore) Get(key []byte) ([]byte, error) {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	rCtx, err := p.bus.GetPersistenceModule().NewReadContext(currHeight)
	if err != nil {
		return nil, err
	}
	defer rCtx.Release()
	return rCtx.GetIBCStoreEntry(prefixed, currHeight) // returns latest value stored
}

// Get queries the persistence layer for the latest value stored in the IBC state tree
// it then generates a proof by importing the IBC state tree from the TreeStore
// keys are automatically prefixed with the CommitmentPrefix if not present
func (p *provableStore) GetAndProve(key []byte, membership bool) ([]byte, *ics23.CommitmentProof, error) {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	rCtx, err := p.bus.GetPersistenceModule().NewReadContext(currHeight)
	if err != nil {
		return nil, nil, err
	}
	value, err := rCtx.GetIBCStoreEntry(prefixed, currHeight) // returns latest value stored
	if err != nil {
		return nil, nil, err
	}
	defer rCtx.Release()
	var proof *ics23.CommitmentProof
	if membership {
		proof, err = p.CreateMembershipProof(key, value)
	} else {
		proof, err = p.CreateNonMembershipProof(key)
	}
	if err != nil {
		return nil, nil, err
	}
	return value, proof, nil
}

// CreateMembershipProof creates a membership proof for the key-value pair with the key
// prefixed with the CommitmentPrefix, by importing the state tree from the TreeStore
func (p *provableStore) CreateMembershipProof(key, value []byte) (*ics23.CommitmentProof, error) {
	// import IBC state tree
	// TODO(#854): Implement tree retrieval
	/**
	prefixed := applyPrefix(p.prefix, key)
	root, nodeStore := p.bus.GetTreeStore().GetTree(trees.IBCStateTree)
	lazy := smt.ImportSparseMerkleTree(nodeStore, root, sha256.New())
	return createMembershipProof(lazy, prefixed, value)
	**/
	return nil, nil
}

// CreateNonMembershipProof creates a non-membership proof for the key prefixed with the
// CommitmentPrefix, by importing the state tree from the TreeStore
func (p *provableStore) CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	// import IBC state tree
	// TODO(#854): Implement tree retrieval
	/**
	prefixed := applyPrefix(p.prefix, key)
	root, nodeStore := p.bus.GetTreeStore().GetTree(trees.IBCStateTree)
	lazy := smt.ImportSparseMerkleTree(nodeStore, root, sha256.New())
	return createNonMembershipProof(lazy, prefixed)
	**/
	return nil, nil
}

// Set updates the persistence layer with the new key-value pair at the latest height and
// emits an UpdateIBCStore event to the bus for propagation throughout the network, to be
// included in each node's mempool and thus the next block
func (p *provableStore) Set(key, value []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	rwCtx, err := p.bus.GetPersistenceModule().NewRWContext(currHeight)
	if err != nil {
		return err
	}
	defer rwCtx.Release()
	if err := rwCtx.SetIBCStoreEntry(prefixed, value); err != nil {
		return err
	}
	p.cache = append(p.cache, &cachedEntry{
		storeName:   p.name,
		height:      uint64(currHeight),
		prefixedKey: prefixed,
		value:       value,
	})
	// TODO(#854): Implement emit functions
	// return emitUpdateStoreEvent(p.prefix, key, value)
	return nil
}

// Delete updates the persistence layer with the key and nil value pair at the latest height
// and emits an PruneIBCStore event to the bus for propagation throughout the network, to be
// included in each node's mempool and thus the next block
func (p *provableStore) Delete(key []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	rwCtx, err := p.bus.GetPersistenceModule().NewRWContext(currHeight)
	if err != nil {
		return err
	}
	defer rwCtx.Release()
	if err := rwCtx.SetIBCStoreEntry(prefixed, nil); err != nil {
		return err
	}
	p.cache = append(p.cache, &cachedEntry{
		storeName:   p.name,
		height:      uint64(currHeight),
		prefixedKey: prefixed,
		value:       nil,
	})
	// TODO(#854): Implement emit functions
	// return emitPruneStoreEvent(p.prefix, key)
	return nil
}

// GetCommitmentPrefix returns the CommitmentPrefix for the store
func (p *provableStore) GetCommitmentPrefix() coreTypes.CommitmentPrefix { return p.prefix }

// Root returns the current root of the IBC state tree
func (p *provableStore) Root() ics23.CommitmentRoot {
	// TODO(#854): Implement tree retrieval
	/**
	root, _ := p.bus.GetTreeStore().GetTree(trees.IBCStateTree)
	return root
	**/
	return nil
}

// CacheEntries writes all local changes to disk and clears the in-memory cache
func (p *provableStore) CacheEntries(store kvstore.KVStore) error {
	for _, entry := range p.cache {
		key, value := entry.prepare()
		if err := store.Set(key, value); err != nil {
			return err
		}
	}
	p.cache = make([]*cachedEntry, 0)
	return nil
}

// PruneCache removes all entries from the cache at the given height
func (p *provableStore) PruneCache(store kvstore.KVStore, height uint64) error {
	keys, _, err := store.GetAll([]byte(fmt.Sprintf("%s/%d", p.name, height)), false)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

// RestoreCache loads all entries from disk into the cache
func (p *provableStore) RestoreCache(store kvstore.KVStore) error {
	keys, values, err := store.GetAll(p.prefix, false)
	if err != nil {
		return err
	}
	for i := 0; i < len(keys); i++ {
		parts := strings.SplitN(string(keys[i]), "/", 2) // name, heightStr, prefixedKeyStr
		height, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		}
		value := values[i]
		p.cache = append(p.cache, &cachedEntry{
			storeName:   parts[0],
			height:      height,
			prefixedKey: []byte(parts[2]),
			value:       value,
		})
	}
	return nil
}

// applyPrefix will apply the CommitmentPrefix to the key provided if not already applied
func applyPrefix(prefix coreTypes.CommitmentPrefix, key []byte) coreTypes.CommitmentPath {
	slashed := make([]byte, 0, len(key)+1)
	slashed = append(slashed, key...)
	slashed = append(slashed, []byte("/")...)
	if bytes.Equal(prefix[:len(slashed)], slashed) {
		return key
	}
	return host.ApplyPrefix(prefix, string(key))
}
