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
	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/sql"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/base_modules"
	"github.com/pokt-network/smt"
)

// smtTreeHasher sets the hasher used by the tree SMT trees
// as a package level variable for visibility and internal use.
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
	IBCTreeName          = "ibc"
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
	TransactionsTreeName, ParamsTreeName, FlagsTreeName, IBCTreeName,
}

// stateTree is a wrapper around the SMT that contains an identifying
// key alongside the tree and nodeStore that backs the tree
type stateTree struct {
	name      string
	tree      *smt.SMT
	nodeStore kvstore.KVStore
}

var _ modules.TreeStoreModule = &TreeStore{}

// TreeStore stores a set of merkle trees that
// it manages. It fulfills the modules.TreeStore interface.
// * It is responsible for atomic commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type TreeStore struct {
	base_modules.IntegrableModule

	logger *modules.Logger

	TreeStoreDir string
	rootTree     *stateTree
	merkleTrees  map[string]*stateTree
}

// GetTree returns the root hash and nodeStore for the matching tree
// stored in the TreeStore. This enables the caller to import the smt
// and not change the one stored.
func (t *TreeStore) GetTree(name string) ([]byte, kvstore.KVStore) {
	if name == RootTreeName {
		return t.rootTree.tree.Root(), t.rootTree.nodeStore
	}
	if tree, ok := t.merkleTrees[name]; ok {
		return tree.tree.Root(), tree.nodeStore
	}
	return nil, nil
}

// Update takes a pgx transaction and a height and updates all of the trees in the TreeStore for that height.
// It is atomic and handles its own savepoint and rollback creation.
func (t *TreeStore) Update(pgtx pgx.Tx, height uint64) (string, error) {
	t.logger.Info().Msgf("🌴 updating state trees at height %d", height)
	txi := t.GetBus().GetPersistenceModule().GetTxIndexer()
	stateHash, err := t.updateMerkleTrees(pgtx, txi, height)
	if err != nil {
		return "", fmt.Errorf("failed to update merkle trees: %w", err)
	}
	return stateHash, nil
}

// DebugClearAll is used by the debug cli to completely reset all merkle trees.
// This should only be called by the debug CLI.
// TECHDEBT: Move this into a separate file with a debug build flag to avoid accidental usage in prod
func (t *TreeStore) DebugClearAll() error {
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

// updateMerkleTrees updates all of the merkle trees in order defined by `numMerkleTrees`
// * it returns the new state hash capturing the state of all the trees or an error if one occurred.
// * This function does not commit state to disk. The caller must manually invoke `Commit` to persist
// changes to disk.
func (t *TreeStore) updateMerkleTrees(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
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
				return "", fmt.Errorf("failed to get actors at height %d: %w", height, err)
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
			indexedTxs, err := getTransactions(txi, height)
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
		case IBCTreeName:
			keys, values, err := sql.GetIBCStoreUpdates(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get IBC store updates: %w", err)
			}
			if err := t.updateIBCTree(keys, values); err != nil {
				return "", fmt.Errorf("failed to update IBC tree: %w", err)
			}
		// Default
		default:
			t.logger.Panic().Msgf("unhandled merkle tree type: %s", treeName)
		}
	}

	if err := t.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return t.getStateHash(), nil
}

func (t *TreeStore) Commit() error {
	if err := t.rootTree.tree.Commit(); err != nil {
		return fmt.Errorf("failed to commit root tree: %w", err)
	}

	for name, treeStore := range t.merkleTrees {
		if err := treeStore.tree.Commit(); err != nil {
			return fmt.Errorf("failed to commit %s: %w", name, err)
		}
	}
	return nil
}

func (t *TreeStore) getStateHash() string {
	for _, stateTree := range t.merkleTrees {
		key := []byte(stateTree.name)
		val := stateTree.tree.Root()
		if err := t.rootTree.tree.Update(key, val); err != nil {
			log.Fatalf("failed to update root tree with %s tree's hash: %v", stateTree.name, err)
		}
	}
	// Convert the array to a slice and return it
	// REF: https://stackoverflow.com/questions/28886616/convert-array-to-slice-in-go
	hexHash := hex.EncodeToString(t.rootTree.tree.Root())
	t.logger.Info().Msgf("#️⃣ calculated state hash: %s", hexHash)
	return hexHash
}

////////////////////////
// Actor Tree Helpers //
////////////////////////

// NB: I think this needs to be done manually for all 4 types.
func (t *TreeStore) updateActorsTree(actorType coreTypes.ActorType, actors []*coreTypes.Actor) error {
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

func (t *TreeStore) updateAccountTrees(accounts []*coreTypes.Account) error {
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

func (t *TreeStore) updatePoolTrees(pools []*coreTypes.Account) error {
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

func (t *TreeStore) updateTransactionsTree(indexedTxs []*coreTypes.IndexedTransaction) error {
	for _, idxTx := range indexedTxs {
		txBz := idxTx.GetTx()
		txHash := crypto.SHA3Hash(txBz)
		if err := t.merkleTrees[TransactionsTreeName].tree.Update(txHash, txBz); err != nil {
			return err
		}
	}
	return nil
}

func (t *TreeStore) updateParamsTree(params []*coreTypes.Param) error {
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

func (t *TreeStore) updateFlagsTree(flags []*coreTypes.Flag) error {
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

func (t *TreeStore) updateIBCTree(keys, values [][]byte) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys and values must be the same length")
	}
	for i, key := range keys {
		value := values[i]
		if value == nil {
			if err := t.merkleTrees[IBCTreeName].tree.Delete(key); err != nil {
				return err
			}
			continue
		}
		if err := t.merkleTrees[IBCTreeName].tree.Update(key, value); err != nil {
			return err
		}
	}
	return nil
}

// getTransactions takes a transaction indexer and returns the transactions for the current height
func getTransactions(txi indexer.TxIndexer, height uint64) ([]*coreTypes.IndexedTransaction, error) {
	// TECHDEBT(#813): Avoid this cast to int64
	indexedTxs, err := txi.GetByHeight(int64(height), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by height: %w", err)
	}
	return indexedTxs, nil
}
