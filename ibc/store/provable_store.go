package store

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/pokt-network/pocket/ibc/path"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/trees"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.ProvableStore = &provableStore{}

// cachedEntry represents a local change made to the IBC store prior to it being
// committed to the state tree. These are written to disk in the to prevent a
// loss of data and pruned when included in the state tree
type cachedEntry struct {
	storeName   string
	height      uint64
	prefixedKey []byte
	value       []byte
}

// prepare returns the key and value to be written to disk
// "{height}/{prefixedKey}" => value
func (c *cachedEntry) prepare() (key, value []byte) {
	return []byte(fmt.Sprintf("%s/%d/%s", c.storeName, c.height, string(c.prefixedKey))), c.value
}

// provableStore is a struct that interfaces with the persistence layer to
// Get/Set/Delete the keys in the IBC state tree, in doing so it will trigger
// the creation of IBC messages that are broadcasted through the network and
// included in the mempool/next block to change the state of the IBC tree
type provableStore struct {
	m          sync.Mutex
	bus        modules.Bus                // used to interact with persistence (passed from IBCHost)
	name       string                     // store name in storeManager
	prefix     coreTypes.CommitmentPrefix // []byte(name)
	cache      map[string]*cachedEntry    // in-memory cache of local changes to be written to disk
	privateKey string
}

// newProvableStore returns a new instance of provableStore with the bus and prefix provided
func newProvableStore(bus modules.Bus, prefix coreTypes.CommitmentPrefix, privateKey string) *provableStore {
	return &provableStore{
		m:          sync.Mutex{},
		bus:        bus,
		name:       string(prefix),
		prefix:     prefix,
		cache:      make(map[string]*cachedEntry, 0),
		privateKey: privateKey,
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
	value, err := rCtx.GetIBCStoreEntry(prefixed, currHeight) // returns latest value stored
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetAndProve queries the persistence layer for the latest value stored in the IBC state
// tree it then generates a proof by importing the IBC state tree from the TreeStore
// keys are automatically prefixed with the CommitmentPrefix if not present
func (p *provableStore) GetAndProve(key []byte) ([]byte, *ics23.CommitmentProof, error) {
	found := true
	value, err := p.Get(key)
	if errors.Is(err, coreTypes.ErrIBCKeyDoesNotExist(string(key))) {
		found = false // key not found create non-membership proof
	} else if err != nil {
		return nil, nil, err
	}
	var proof *ics23.CommitmentProof
	if found {
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
	prefixed := applyPrefix(p.prefix, key)
	root, nodeStore := p.bus.GetTreeStore().GetTree(trees.IBCTreeName)
	lazy := smt.ImportSparseMerkleTree(nodeStore, sha256.New(), root)
	return createMembershipProof(lazy, prefixed, value)
}

// CreateNonMembershipProof creates a non-membership proof for the key prefixed with the
// CommitmentPrefix, by importing the state tree from the TreeStore
func (p *provableStore) CreateNonMembershipProof(key []byte) (*ics23.CommitmentProof, error) {
	// import IBC state tree
	prefixed := applyPrefix(p.prefix, key)
	root, nodeStore := p.bus.GetTreeStore().GetTree(trees.IBCTreeName)
	lazy := smt.ImportSparseMerkleTree(nodeStore, sha256.New(), root)
	return createNonMembershipProof(lazy, prefixed)
}

// Set updates the persistence layer with the new key-value pair at the latest height and
// emits an UpdateIBCStore event to the bus for propagation throughout the network, to be
// included in each node's mempool and thus the next block
func (p *provableStore) Set(key, value []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	p.m.Lock()
	defer p.m.Unlock()
	p.cache[string(prefixed)] = &cachedEntry{
		storeName:   p.name,
		height:      uint64(currHeight),
		prefixedKey: prefixed,
		value:       value,
	}
	return emitUpdateStoreEvent(p.bus, p.privateKey, key, value)
}

// Delete updates the persistence layer with the key and nil value pair at the latest height
// and emits an PruneIBCStore event to the bus for propagation throughout the network, to be
// included in each node's mempool and thus the next block
func (p *provableStore) Delete(key []byte) error {
	prefixed := applyPrefix(p.prefix, key)
	currHeight := int64(p.bus.GetConsensusModule().CurrentHeight())
	p.m.Lock()
	defer p.m.Unlock()
	p.cache[string(prefixed)] = &cachedEntry{
		storeName:   p.name,
		height:      uint64(currHeight),
		prefixedKey: prefixed,
		value:       nil,
	}
	return emitPruneStoreEvent(p.bus, p.privateKey, key)
}

// GetCommitmentPrefix returns the CommitmentPrefix for the store
func (p *provableStore) GetCommitmentPrefix() coreTypes.CommitmentPrefix { return p.prefix }

// Root returns the current root of the IBC state tree
func (p *provableStore) Root() ics23.CommitmentRoot {
	root, _ := p.bus.GetTreeStore().GetTree(trees.IBCTreeName)
	return root
}

// FlushEntries writes all local changes to disk and clears the in-memory cache
func (p *provableStore) FlushEntries(store kvstore.KVStore) error {
	p.m.Lock()
	defer p.m.Unlock()
	for _, entry := range p.cache {
		key, value := entry.prepare()
		if err := store.Set(key, value); err != nil {
			return err
		}
		delete(p.cache, string(entry.prefixedKey))
	}
	return nil
}

// PruneCache removes all entries from the cache at the given height
func (p *provableStore) PruneCache(store kvstore.KVStore, height uint64) error {
	p.m.Lock()
	defer p.m.Unlock()
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
	p.m.Lock()
	defer p.m.Unlock()
	keys, values, err := store.GetAll([]byte(fmt.Sprintf("%s/", p.name)), false)
	if err != nil {
		return err
	}
	for i, key := range keys {
		parts := strings.SplitN(string(key), "/", 3) // name, heightStr, prefixedKeyStr
		height, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return err
		}
		value := values[i]
		p.cache[parts[1]] = &cachedEntry{
			storeName:   parts[0],
			height:      height,
			prefixedKey: []byte(parts[2]),
			value:       value,
		}
	}
	return nil
}

// applyPrefix will apply the CommitmentPrefix to the key provided if not already applied
func applyPrefix(prefix coreTypes.CommitmentPrefix, key []byte) coreTypes.CommitmentPath {
	delim := []byte("/")
	slashed := make([]byte, 0, len(key)+len(delim))
	slashed = append(slashed, key...)
	slashed = append(slashed, delim...)
	if len(prefix) > len(slashed) && bytes.Equal(prefix[:len(slashed)], slashed) {
		return key
	}
	return path.ApplyPrefix(prefix, string(key))
}
