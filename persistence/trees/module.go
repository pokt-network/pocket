package trees

import (
	"fmt"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.Module = &TreeStore{}

func (*TreeStore) Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	m := &TreeStore{}

	bus.RegisterModule(m)

	for _, option := range options {
		option(m)
	}

	if m.txi == nil {
		m.txi = bus.GetPersistenceModule().GetTxIndexer()
	}

	if err := m.setupTrees(); err != nil {
		return nil, err
	}

	return m, nil
}

func Create(bus modules.Bus, options ...modules.ModuleOption) (modules.Module, error) {
	return new(TreeStore).Create(bus, options...)
}

// WithTreeStoreDirectory assigns the path where the tree store
// saves its data.
func WithTreeStoreDirectory(path string) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*TreeStore)
		if ok {
			mod.treeStoreDir = path
		}
	}
}

// WithTxIndexer assigns a TxIndexer for use during operation.
func WithTxIndexer(txi indexer.TxIndexer) modules.ModuleOption {
	return func(m modules.InitializableModule) {
		mod, ok := m.(*TreeStore)
		if ok {
			mod.txi = txi
		}
	}
}

func (t *TreeStore) GetModuleName() string  { return modules.TreeStoreModuleName }
func (t *TreeStore) Start() error           { return nil }
func (t *TreeStore) Stop() error            { return nil }
func (t *TreeStore) GetBus() modules.Bus    { return t.bus }
func (t *TreeStore) SetBus(bus modules.Bus) { t.bus = bus }

func (t *TreeStore) setupTrees() error {
	if t.treeStoreDir == ":memory:" {
		return t.setupInMemory()
	}

	t.merkleTrees = make(map[merkleTree]*smt.SMT, int(numMerkleTrees))
	t.nodeStores = make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees))

	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", t.treeStoreDir, merkleTreeToString[tree]))
		if err != nil {
			return err
		}
		t.nodeStores[tree] = nodeStore
		t.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, smtTreeHasher)
	}

	return nil
}

func (t *TreeStore) setupInMemory() error {
	t.merkleTrees = make(map[merkleTree]*smt.SMT, int(numMerkleTrees))
	t.nodeStores = make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees))

	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		t.nodeStores[tree] = nodeStore
		t.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, smtTreeHasher)
	}

	return nil
}
