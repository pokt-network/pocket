package trees

import (
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

func (*treeStore) Create(bus modules.Bus, options ...modules.TreeStoreOption) (modules.TreeStoreModule, error) {
	m := &treeStore{}

	for _, option := range options {
		option(m)
	}

	m.SetBus(bus)

	if err := m.setupTrees(); err != nil {
		return nil, err
	}

	return m, nil
}

func Create(bus modules.Bus, options ...modules.TreeStoreOption) (modules.TreeStoreModule, error) {
	return new(treeStore).Create(bus, options...)
}

// WithLogger assigns a logger for the tree store
func WithLogger(logger *modules.Logger) modules.TreeStoreOption {
	return func(m modules.TreeStoreModule) {
		if mod, ok := m.(*treeStore); ok {
			mod.logger = logger
		}
	}
}

// WithTreeStoreDirectory assigns the path where the tree store
// saves its data.
func WithTreeStoreDirectory(path string) modules.TreeStoreOption {
	return func(m modules.TreeStoreModule) {
		if mod, ok := m.(*treeStore); ok {
			mod.treeStoreDir = path
		}
	}
}

func (t *treeStore) setupTrees() error {
	if t.treeStoreDir == ":memory:" {
		return t.setupInMemory()
	}

	nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", t.treeStoreDir, RootTreeName))
	if err != nil {
		return err
	}
	t.rootTree = &stateTree{
		name:      RootTreeName,
		tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
		nodeStore: nodeStore,
	}
	t.merkleTrees = make(map[string]*stateTree, len(stateTreeNames))

	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", t.treeStoreDir, stateTreeNames[i]))
		if err != nil {
			return err
		}
		t.merkleTrees[stateTreeNames[i]] = &stateTree{
			name:      stateTreeNames[i],
			tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
			nodeStore: nodeStore,
		}
	}

	return nil
}

func (t *treeStore) setupInMemory() error {
	nodeStore := kvstore.NewMemKVStore()
	t.rootTree = &stateTree{
		name:      RootTreeName,
		tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
		nodeStore: nodeStore,
	}
	t.merkleTrees = make(map[string]*stateTree, len(stateTreeNames))
	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		t.merkleTrees[stateTreeNames[i]] = &stateTree{
			name:      stateTreeNames[i],
			tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
			nodeStore: nodeStore,
		}
	}
	return nil
}
