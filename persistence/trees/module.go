package trees

import (
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.TreeStoreModule = &TreeStore{}

func (*TreeStore) Create(bus modules.Bus, options ...modules.TreeStoreOption) (modules.TreeStoreModule, error) {
	m := &TreeStore{}

	bus.RegisterModule(m)

	for _, option := range options {
		option(m)
	}

	if err := m.setupTrees(); err != nil {
		return nil, err
	}

	return m, nil
}

func Create(bus modules.Bus, options ...modules.TreeStoreOption) (modules.TreeStoreModule, error) {
	return new(TreeStore).Create(bus, options...)
}

// WithLogger assigns a logger for the tree store
func WithLogger(logger *modules.Logger) modules.TreeStoreOption {
	return func(m modules.TreeStoreModule) {
		if mod, ok := m.(*TreeStore); ok {
			mod.logger = logger
		}
	}
}

// WithTreeStoreDirectory assigns the path where the tree store
// saves its data.
func WithTreeStoreDirectory(path string) modules.TreeStoreOption {
	return func(m modules.TreeStoreModule) {
		mod, ok := m.(*TreeStore)
		if ok {
			mod.treeStoreDir = path
		}
	}
}

func (t *TreeStore) GetModuleName() string { return modules.TreeStoreSubmoduleName }

func (t *TreeStore) setupTrees() error {
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

func (t *TreeStore) setupInMemory() error {
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
