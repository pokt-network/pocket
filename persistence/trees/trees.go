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

var _ modules.TreeStoreModule = &TreeStore{}

// TreeStore stores a set of merkle trees that
// it manages. It fulfills the modules.TreeStore interface.
// * It is responsible for atomic commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type TreeStore struct {
	base_modules.IntegratableModule

	bus modules.Bus
	txi indexer.TxIndexer

	treeStoreDir string
	rootTree     *stateTree
	merkleTrees  map[string]*stateTree
}

// GetTree returns the name, root hash, and nodeStore for the matching tree tree
// stored in the TreeStore. This enables the caller to import the smt and not
// change the one stored
func (t *TreeStore) GetTree(name string) ([]byte, kvstore.KVStore) {
	if name == RootTreeName {
		return t.rootTree.tree.Root(), t.rootTree.nodeStore
	}
	if tree, ok := t.merkleTrees[name]; ok {
		return tree.tree.Root(), tree.nodeStore
	}
	return nil, nil
}

// Worldstate holds a ser/deserializable view of the entire tree state.
type Worldstate struct {
	TreeStoreDir string
	RootTree     *stateTree
	MerkleTrees  map[string]*stateTree
}

// Update takes a pgx transaction and a height and updates all of the trees in the TreeStore for that height.
func (t *TreeStore) Update(pgtx pgx.Tx, height uint64) (string, error) {
	previous, err := t.save()
	if err != nil {
		return "", fmt.Errorf("failed to create valid rollback point: %w", err)
	}

	fmt.Printf("previous: %v\n", previous)

	hash, err := t.updateMerkleTrees(pgtx, t.txi, height)
	if err != nil {
		// TODO t.load(previous) and take
		return "", fmt.Errorf("err not handled")
	}
	if err := t.Commit(); err != nil {
		// TODO t.load(previous) state
		return "", fmt.Errorf("failed to commit tree state: %w", err)
	}

	return hash, nil
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
// * it returns the new state hash capturing the state of all the trees or an error if one occurred
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

	if err := t.Commit(); err != nil {
		return "", err
	}

	return t.getStateHash(), nil
}

func (t *TreeStore) Prepare(tx modules.Tx) error {
	return fmt.Errorf("not impl")
}

func (t *TreeStore) save() (*Worldstate, error) {
	w := &Worldstate{
		MerkleTrees: map[string]*stateTree{},
	}

	fmt.Printf("w: %v\n", w)

	return w, nil
}

func (t *TreeStore) Commit() error {
	for treeName, stateTree := range t.merkleTrees {
		if err := stateTree.tree.Commit(); err != nil {
			return fmt.Errorf("failed to commit %s: %w", treeName, err)
		}
	}
	return nil
}

func (t *TreeStore) Rollback() {
	panic("treestore not impl")
}

func (t *TreeStore) getStateHash() string {
	for _, stateTree := range t.merkleTrees {
		if err := t.rootTree.tree.Update([]byte(stateTree.name), stateTree.tree.Root()); err != nil {
			log.Fatalf("failed to update root tree with %s tree's hash: %v", stateTree.name, err)
		}
	}
	return hex.EncodeToString(t.rootTree.tree.Root())
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
