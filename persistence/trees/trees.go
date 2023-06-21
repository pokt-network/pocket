// package trees maintains a set of sparse merkle trees
// each backed by the KVStore interface. It offers an atomic
// commit and rollback mechanism for interacting with
// that core resource map of merkle trees.
package trees

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/sql"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/smt"
)

var smtTreeHasher hash.Hash = sha256.New()

const (
	RootTreeName         = "root"
	AppTreeName          = "app"
	ValTreeName          = "val"
	FishTreeName         = "fish"
	ServicerTreeName     = "servicer"
	AccountTreeName      = "account"
	PoolTreeName         = "pool"
	TransactionsTreeName = "transactions"
	ParamsTreeName       = "params"
	FlagsTreeName        = "flags"
)

var actorTypeToMerkleTreeName = map[coreTypes.ActorType]string{
	coreTypes.ActorType_ACTOR_TYPE_APP:      AppTreeName,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      ValTreeName,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     FishTreeName,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: ServicerTreeName,
}

var merkleTreeNameToActorTypeName = map[string]coreTypes.ActorType{
	AppTreeName:      coreTypes.ActorType_ACTOR_TYPE_APP,
	ValTreeName:      coreTypes.ActorType_ACTOR_TYPE_VAL,
	FishTreeName:     coreTypes.ActorType_ACTOR_TYPE_FISH,
	ServicerTreeName: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
}

var stateTreeNames = []string{
	// Actor Trees
	AppTreeName, ValTreeName, FishTreeName, ServicerTreeName,
	// Account Trees
	AccountTreeName, PoolTreeName,
	// Data Trees
	TransactionsTreeName, ParamsTreeName, FlagsTreeName,
}

// stateTree is a wrapper around the SMT that contains an identifying
// key alongside the tree and nodeStore that backs the tree
type stateTree struct {
	name      string
	tree      *smt.SMT
	nodeStore kvstore.KVStore
}

// treeStore stores a set of merkle trees that
// it manages. It fulfills the modules.TreeStore interface.
// * It is responsible for atomic commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type treeStore struct {
	treeStoreDir string
	rootTree     *stateTree
	merkleTrees  map[string]*stateTree
}

// GetTree returns the name, root hash, and nodeStore for the matching tree tree
// stored in the TreeStore. This enables the caller to import the smt
func (t *treeStore) GetTree(name string) *smt.SMT {
	if name == RootTreeName {
		return t.rootTree.tree
	}
	if tree, ok := t.merkleTrees[name]; ok {
		return tree.tree
	}
	return nil
}

func (t *treeStore) Update(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
	return t.updateMerkleTrees(pgtx, txi, height)
}

// NewStateTrees is the constructor object for a treeStore and initializes and configures a new
// tree for the appropriate type of store, i.e. in-memory vs file system storage.
func NewStateTrees(treesStoreDir string) (*treeStore, error) {
	if treesStoreDir == ":memory:" {
		return newMemStateTrees()
	}

	nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", treesStoreDir, "root"))
	if err != nil {
		return nil, err
	}
	rootTree := &stateTree{
		name:      RootTreeName,
		tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
		nodeStore: nodeStore,
	}

	stateTrees := &treeStore{
		treeStoreDir: treesStoreDir,
		rootTree:     rootTree,
		merkleTrees:  make(map[string]*stateTree, len(stateTreeNames)),
	}

	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", treesStoreDir, stateTreeNames[i]))
		if err != nil {
			return nil, err
		}
		tree := &stateTree{
			name:      stateTreeNames[i],
			tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
			nodeStore: nodeStore,
		}
		stateTrees.merkleTrees[stateTreeNames[i]] = tree
	}
	return stateTrees, nil
}

// DebugClearAll is used by the debug cli to completely reset all merkle trees.
// This should only be called by the debug CLI.
// TECHDEBT: Move this into a separate file with a debug build flag to avoid accidental usage in prod
func (t *treeStore) DebugClearAll() error {
	if err := t.rootTree.nodeStore.ClearAll(); err != nil {
		return fmt.Errorf("failed to clear root node store: %w", err)
	}
	t.rootTree.tree = smt.NewSparseMerkleTree(t.rootTree.nodeStore, smtTreeHasher)
	for treeName, stateTree := range t.merkleTrees {
		nodeStore := stateTree.nodeStore
		if err := nodeStore.ClearAll(); err != nil {
			return fmt.Errorf("failed to clear %s node store: %w", treeName, err)
		}
		stateTree.tree = smt.NewSparseMerkleTree(nodeStore, smtTreeHasher)
	}
	return nil
}

// newMemstateTrees creates a new in-memory state tree
func newMemStateTrees() (*treeStore, error) {
	nodeStore := kvstore.NewMemKVStore()
	rootTree := &stateTree{
		name:      RootTreeName,
		tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
		nodeStore: nodeStore,
	}
	stateTrees := &treeStore{
		rootTree:    rootTree,
		merkleTrees: make(map[string]*stateTree, len(stateTreeNames)),
	}
	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		tree := &stateTree{
			name:      stateTreeNames[i],
			tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
			nodeStore: nodeStore,
		}
		stateTrees.merkleTrees[stateTreeNames[i]] = tree
	}
	return stateTrees, nil
}

// updateMerkleTrees updates all of the merkle trees in order defined by `numMerkleTrees`
// * it returns the new state hash capturing the state of all the trees or an error if one occurred
func (t *treeStore) updateMerkleTrees(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
	for treeName := range t.merkleTrees {
		switch treeName {
		// Actor Merkle Trees
		case AppTreeName, ValTreeName, FishTreeName, ServicerTreeName:
			actorType, ok := merkleTreeNameToActorTypeName[treeName]
			if !ok {
				return "", fmt.Errorf("no actor type found for merkle tree: %s", treeName)
			}

			actors, err := sql.GetActors(pgtx, actorType, height)
			if err != nil {
				return "", fmt.Errorf("failed to get actors at height: %w", err)
			}

			if err := t.updateActorsTree(actorType, actors); err != nil {
				return "", fmt.Errorf("failed to update actors tree for treeType: %s, actorType: %v - %w", treeName, actorType, err)
			}

		// Account Merkle Trees
		case AccountTreeName:
			accounts, err := sql.GetAccounts(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get accounts: %w", err)
			}
			if err := t.updateAccountTrees(accounts); err != nil {
				return "", fmt.Errorf("failed to update account trees: %w", err)
			}
		case PoolTreeName:
			pools, err := sql.GetPools(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updatePoolTrees(pools); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case TransactionsTreeName:
			indexedTxs, err := sql.GetTransactions(txi, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updateTransactionsTree(indexedTxs); err != nil {
				return "", fmt.Errorf("failed to update transactions: %w", err)
			}
		case ParamsTreeName:
			params, err := sql.GetParams(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get params: %w", err)
			}
			if err := t.updateParamsTree(params); err != nil {
				return "", fmt.Errorf("failed to update params tree: %w", err)
			}
		case FlagsTreeName:
			flags, err := sql.GetFlags(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get flags from transaction: %w", err)
			}
			if err := t.updateFlagsTree(flags); err != nil {
				return "", fmt.Errorf("failed to update flags tree - %w", err)
			}
		// Default
		default:
			log.Fatalf("not handled in state commitment update. Merkle tree: %s", treeName)
		}
	}

	if err := t.commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return t.getStateHash(), nil
}

func (t *treeStore) commit() error {
	for treeName, stateTree := range t.merkleTrees {
		if err := stateTree.tree.Commit(); err != nil {
			return fmt.Errorf("failed to commit %s: %w", treeName, err)
		}
	}
	return nil
}

func (t *treeStore) getStateHash() string {
	for _, stateTree := range t.merkleTrees {
		if err := t.rootTree.tree.Update([]byte(stateTree.name), stateTree.tree.Root()); err != nil {
			logger.Global.Err(err).Str("treeStore", "failed to update root tree: %s")
		}
	}
	return hex.EncodeToString(t.rootTree.tree.Root())
}

////////////////////////
// Actor Tree Helpers //
////////////////////////

// NB: I think this needs to be done manually for all 4 types.
func (t *treeStore) updateActorsTree(actorType coreTypes.ActorType, actors []*coreTypes.Actor) error {
	for _, actor := range actors {
		bzAddr, err := hex.DecodeString(actor.GetAddress())
		if err != nil {
			return err
		}

		actorBz, err := codec.GetCodec().Marshal(actor)
		if err != nil {
			return err
		}

		merkleTreeName, ok := actorTypeToMerkleTreeName[actorType]
		if !ok {
			return fmt.Errorf("no merkle tree found for actor type: %s", actorType)
		}
		if err := t.merkleTrees[merkleTreeName].tree.Update(bzAddr, actorBz); err != nil {
			return err
		}
	}

	return nil
}

//////////////////////////
// Account Tree Helpers //
//////////////////////////

func (t *treeStore) updateAccountTrees(accounts []*coreTypes.Account) error {
	for _, account := range accounts {
		bzAddr, err := hex.DecodeString(account.GetAddress())
		if err != nil {
			return err
		}

		accBz, err := codec.GetCodec().Marshal(account)
		if err != nil {
			return err
		}

		if err := t.merkleTrees[AccountTreeName].tree.Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

func (t *treeStore) updatePoolTrees(pools []*coreTypes.Account) error {
	for _, pool := range pools {
		bzAddr, err := hex.DecodeString(pool.GetAddress())
		if err != nil {
			return err
		}

		accBz, err := codec.GetCodec().Marshal(pool)
		if err != nil {
			return err
		}

		if err := t.merkleTrees[PoolTreeName].tree.Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

///////////////////////
// Data Tree Helpers //
///////////////////////

func (t *treeStore) updateTransactionsTree(indexedTxs []*coreTypes.IndexedTransaction) error {
	for _, idxTx := range indexedTxs {
		txBz := idxTx.GetTx()
		txHash := crypto.SHA3Hash(txBz)
		if err := t.merkleTrees[TransactionsTreeName].tree.Update(txHash, txBz); err != nil {
			return err
		}
	}
	return nil
}

func (t *treeStore) updateParamsTree(params []*coreTypes.Param) error {
	for _, param := range params {
		paramBz, err := codec.GetCodec().Marshal(param)
		paramKey := crypto.SHA3Hash([]byte(param.Name))
		if err != nil {
			return err
		}
		if err := t.merkleTrees[ParamsTreeName].tree.Update(paramKey, paramBz); err != nil {
			return err
		}
	}

	return nil
}

func (t *treeStore) updateFlagsTree(flags []*coreTypes.Flag) error {
	for _, flag := range flags {
		flagBz, err := codec.GetCodec().Marshal(flag)
		flagKey := crypto.SHA3Hash([]byte(flag.Name))
		if err != nil {
			return err
		}
		if err := t.merkleTrees[FlagsTreeName].tree.Update(flagKey, flagBz); err != nil {
			return err
		}
	}

	return nil
}
