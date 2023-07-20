package trees

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/smt"
)

var _ modules.TreeStoreModule = &treeStore{}

// Create returns a TreeStoreSubmodule that has been setup with the provided TreeStoreOptions, started,
// and then registered to the bus.
func (*treeStore) Create(bus modules.Bus, options ...modules.TreeStoreOption) (modules.TreeStoreModule, error) {
	m := &treeStore{}

	for _, option := range options {
		option(m)
	}

	if err := m.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s: %w", modules.TreeStoreSubmoduleName, err)
	}

	bus.RegisterModule(m)

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
		mod, ok := m.(*treeStore)
		if ok {
			mod.treeStoreDir = path
		}
	}
}

// Start loads up the trees from the configured tree store directory.
func (t *treeStore) Start() error {
	return t.setupTrees()
}

// Stop shuts down the database connection to the nodestore for the root tree and then for each merkle tree.
// If Commit has not been called before Stop is called, data will be lost.
func (t *treeStore) Stop() error {
	t.logger.Debug().Msgf("🛑 tree store stop initiated at %s 🛑", hex.EncodeToString(t.rootTree.tree.Root()))
	errs := []error{}
	if err := t.rootTree.nodeStore.Stop(); err != nil {
		errs = append(errs, err)
	}
	for _, st := range t.merkleTrees {
		if err := st.nodeStore.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *treeStore) GetModuleName() string { return modules.TreeStoreSubmoduleName }

// setupTrees is called by Start and it loads the treestore at the given directory
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
