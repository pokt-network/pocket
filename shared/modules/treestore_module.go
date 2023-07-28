package modules

import (
	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/kvstore"
)

//go:generate mockgen -destination=./mocks/treestore_module_mock.go github.com/pokt-network/pocket/shared/modules TreeStoreModule

const (
	TreeStoreSubmoduleName = "tree_store"
)

type TreeStoreOption func(TreeStoreModule)

type treeStoreFactory = FactoryWithOptions[TreeStoreModule, TreeStoreOption]

// TreeStoreModules defines the interface for atomic updates and rollbacks to the internal
// merkle trees that compose the state hash of pocket.
type TreeStoreModule interface {
	Submodule
	InterruptableModule
	treeStoreFactory

	AtomicStore
	Archivable

	// Update returns the computed state hash for a given height.
	// * Height is passed through to the Update function and is used to query the TxIndexer for transactions
	// to update into the merkle tree set
	// * Passing a higher height will cause a change but repeatedly calling the same or a lower height will
	// not incur a change.
	// * By nature of it taking a pgx transaction at runtime, Update inherits the pgx transaction's read view of the
	// database.
	// * Commit must be called after Update to persist any changes it made to disk.
	Update(pgtx pgx.Tx, height uint64) (string, error)
	// DebugClearAll completely clears the state of the trees. For debugging purposes only.
	DebugClearAll() error
	// Prove generates and verifies a proof against the tree with the matching name using the given
	// key and value. If value == nil, it will verify non-membership of the key, otherwise membership.
	Prove(treeName string, key, value []byte) (bool, error)
	// GetTree returns the specified tree's root and nodeStore in order to be imported elsewhere
	GetTree(name string) ([]byte, kvstore.KVStore)
	// GetTreeHashes returns a map of tree names to their root hashes
	GetTreeHashes() map[string]string
}

type Archivable interface {
	Backup(dir string) error
	Load(dir string) error
}
