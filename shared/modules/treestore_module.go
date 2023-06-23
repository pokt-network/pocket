package modules

import (
	"github.com/jackc/pgx/v5"
)

const (
	TreeStoreModuleName = "tree_store"
)

type TreeStoreOption func(TreeStoreModule)

type TreeStoreFactory = FactoryWithOptions[TreeStoreModule, TreeStoreOption]

// TreeStoreModules defines the interface for atomic updates and rollbacks to the internal
// merkle trees that compose the state hash of pocket.
type TreeStoreModule interface {
	IntegratableModule

	// Update returns the new state hash for a given height.
	// * Height is passed through to the Update function and is used to query the TxIndexer for transactions
	// to update into the merkle tree set
	// * Passing a higher height will cause a change but repeatedly calling the same or a lower height will
	// not incur a change.
	// * By nature of it taking a pgx transaction at runtime, Update inherits the pgx transaction's read view of the
	// database.
	Update(pgtx pgx.Tx, height uint64) (string, error)
	// DebugClearAll completely clears the state of the trees. For debugging purposes only.
	DebugClearAll() error
}
