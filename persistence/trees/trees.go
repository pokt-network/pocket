// package trees maintains a set of sparse merkle trees
// each backed by the KVStore interface. It offers an atomic
// commit and rollback mechanism for interacting with
// that core resource map of merkle trees.
package trees

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/sql"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/smt"
)

const (
	appTreeName          = "app"
	valTreeName          = "val"
	fishTreeName         = "fish"
	servicerTreeName     = "servicer"
	accountTreeName      = "account"
	poolTreeName         = "pool"
	transactionsTreeName = "transactions"
	paramsTreeName       = "params"
	flagsTreeName        = "flags"
)

var actorTypeToMerkleTreeName = map[coreTypes.ActorType]string{
	coreTypes.ActorType_ACTOR_TYPE_APP:      appTreeName,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      valTreeName,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     fishTreeName,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: servicerTreeName,
}

var merkleTreeNameToActorTypeName = map[string]coreTypes.ActorType{
	appTreeName:      coreTypes.ActorType_ACTOR_TYPE_APP,
	valTreeName:      coreTypes.ActorType_ACTOR_TYPE_VAL,
	fishTreeName:     coreTypes.ActorType_ACTOR_TYPE_FISH,
	servicerTreeName: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
}

var stateTreeNames = []string{
	// Actor Trees
	appTreeName, valTreeName, fishTreeName, servicerTreeName,
	// Account Trees
	accountTreeName, poolTreeName,
	// Data Trees
	transactionsTreeName, paramsTreeName, flagsTreeName,
}

// stateTree is a wrapper around the SMT that contains an identifying
// key alongside the tree and nodeStore that backs the tree
type stateTree struct {
	key       []byte
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

// Update takes a transaction and a height and updates
// all of the trees in the treeStore for that height.
func (t *treeStore) Update(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
	return t.updateMerkleTrees(pgtx, txi, height)
}

func NewStateTrees(treesStoreDir string) (*treeStore, error) {
	if treesStoreDir == ":memory:" {
		return newMemStateTrees()
	}

	nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", treesStoreDir, "root"))
	if err != nil {
		return nil, err
	}
	rootTree := &stateTree{
		key:       []byte("root"),
		tree:      smt.NewSparseMerkleTree(nodeStore, sha256.New()),
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
			key:       []byte(stateTreeNames[i]),
			tree:      smt.NewSparseMerkleTree(nodeStore, sha256.New()),
			nodeStore: nodeStore,
		}
		stateTrees.merkleTrees[stateTreeNames[i]] = tree
	}
	return stateTrees, nil
}

// DebugClearAll is used by the debug cli to completely reset all merkle trees.
// This should only be called by the debug CLI.
func (t *treeStore) DebugClearAll() error {
	if err := t.rootTree.nodeStore.ClearAll(); err != nil {
		return fmt.Errorf("failed to clear root node store: %w", err)
	}
	t.rootTree.tree = smt.NewSparseMerkleTree(t.rootTree.nodeStore, sha256.New())
	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore := t.merkleTrees[stateTreeNames[i]].nodeStore
		if err := nodeStore.ClearAll(); err != nil {
			return fmt.Errorf("failed to clear %s node store: %w", string(t.merkleTrees[stateTreeNames[i]].key), err)
		}
		t.merkleTrees[stateTreeNames[i]].tree = smt.NewSparseMerkleTree(nodeStore, sha256.New())
	}

	return nil
}

// newMemStateTrees creates a new in-memory state tree
func newMemStateTrees() (*treeStore, error) {
	nodeStore := kvstore.NewMemKVStore()
	rootTree := &stateTree{
		key:       []byte("root"),
		tree:      smt.NewSparseMerkleTree(nodeStore, sha256.New()),
		nodeStore: nodeStore,
	}
	stateTrees := &treeStore{
		rootTree:    rootTree,
		merkleTrees: make(map[string]*stateTree, len(stateTreeNames)),
	}
	for i := 0; i < len(stateTreeNames); i++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		tree := &stateTree{
			key:       []byte(stateTreeNames[i]),
			tree:      smt.NewSparseMerkleTree(nodeStore, sha256.New()),
			nodeStore: nodeStore,
		}
		stateTrees.merkleTrees[stateTreeNames[i]] = tree
	}
	return stateTrees, nil
}

// updateMerkleTrees updates all of the merkle trees that TreeStore manages.
// * it returns an hash of the output or an error.
func (t *treeStore) updateMerkleTrees(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
	for i := 0; i < len(stateTreeNames); i++ {
		switch key := string(t.merkleTrees[stateTreeNames[i]].key); key {
		// Actor Merkle Trees
		case appTreeName, valTreeName, fishTreeName, servicerTreeName:
			actorType, ok := merkleTreeNameToActorTypeName[key]
			if !ok {
				return "", fmt.Errorf("no actor type found for merkle tree: %s", key)
			}

			actors, err := sql.GetActors(pgtx, actorType, height)
			if err != nil {
				return "", fmt.Errorf("failed to get actors at height: %w", err)
			}

			if err := t.updateActorsTree(actorType, actors); err != nil {
				return "", fmt.Errorf("failed to update actors tree for treeType: %s, actorType: %v - %w", key, actorType, err)
			}

		// Account Merkle Trees
		case accountTreeName:
			accounts, err := sql.GetAccounts(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get accounts: %w", err)
			}
			if err := t.updateAccountTrees(accounts); err != nil {
				return "", fmt.Errorf("failed to update account trees: %w", err)
			}
		case poolTreeName:
			pools, err := sql.GetPools(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updatePoolTrees(pools); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case transactionsTreeName:
			indexedTxs, err := sql.GetTransactions(txi, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updateTransactionsTree(indexedTxs); err != nil {
				return "", fmt.Errorf("failed to update transactions: %w", err)
			}
		case paramsTreeName:
			params, err := sql.GetParams(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get params: %w", err)
			}
			if err := t.updateParamsTree(params); err != nil {
				return "", fmt.Errorf("failed to update params tree: %w", err)
			}
		case flagsTreeName:
			flags, err := sql.GetFlags(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get flags from transaction: %w", err)
			}
			if err := t.updateFlagsTree(flags); err != nil {
				return "", fmt.Errorf("failed to update flags tree - %w", err)
			}
		// Default
		default:
			panic(fmt.Sprintf("not handled in state commitment update. Merkle tree: %s", key))
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
		if err := t.rootTree.tree.Update(stateTree.key, stateTree.tree.Root()); err != nil {
			log.Fatalf("failed to update root tree: %s", err.Error())
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

		if err := t.merkleTrees[accountTreeName].tree.Update(bzAddr, accBz); err != nil {
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

		if err := t.merkleTrees[poolTreeName].tree.Update(bzAddr, accBz); err != nil {
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
		if err := t.merkleTrees[transactionsTreeName].tree.Update(txHash, txBz); err != nil {
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
		if err := t.merkleTrees[paramsTreeName].tree.Update(paramKey, paramBz); err != nil {
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
		if err := t.merkleTrees[flagsTreeName].tree.Update(flagKey, flagBz); err != nil {
			return err
		}
	}

	return nil
}
