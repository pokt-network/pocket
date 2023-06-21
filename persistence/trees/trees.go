// package trees maintains a set of sparse merkle trees
// each backed by the KVStore interface. It offers an atomic
// commit and rollback mechanism for interacting with
// that core resource map of merkle trees.
package trees

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

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

var smtTreeHasher hash.Hash = sha256.New()

var merkleTreeToString = map[merkleTree]string{
	appMerkleTree:      "app",
	valMerkleTree:      "val",
	fishMerkleTree:     "fish",
	servicerMerkleTree: "servicer",

	accountMerkleTree: "account",
	poolMerkleTree:    "pool",

	transactionsMerkleTree: "transactions",
	paramsMerkleTree:       "params",
	flagsMerkleTree:        "flags",
}

var actorTypeToMerkleTreeName = map[coreTypes.ActorType]merkleTree{
	coreTypes.ActorType_ACTOR_TYPE_APP:      appMerkleTree,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      valMerkleTree,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     fishMerkleTree,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: servicerMerkleTree,
}

var merkleTreeToActorTypeName = map[merkleTree]coreTypes.ActorType{
	appMerkleTree:      coreTypes.ActorType_ACTOR_TYPE_APP,
	valMerkleTree:      coreTypes.ActorType_ACTOR_TYPE_VAL,
	fishMerkleTree:     coreTypes.ActorType_ACTOR_TYPE_FISH,
	servicerMerkleTree: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
}

type merkleTree float64

// A list of Merkle Trees used to maintain the state hash.
const (
	// IMPORTANT: The order in which these trees are defined is important and strict. It implicitly // defines the index of the root hash each independent as they are concatenated together
	// to generate the state hash.

	// TECHDEBT(#834): Remove the need for enforced ordering

	// Actor Merkle Trees
	appMerkleTree merkleTree = iota
	valMerkleTree
	fishMerkleTree
	servicerMerkleTree

	// Account Merkle Trees
	accountMerkleTree
	poolMerkleTree

	// Data Merkle Trees
	transactionsMerkleTree
	paramsMerkleTree
	flagsMerkleTree

	// Used for iteration purposes only; see https://stackoverflow.com/a/64178235/768439 as a reference
	numMerkleTrees
)

// Ensure treeStore implements TreeStore
var _ modules.TreeStoreModule = &treeStore{}

// treeStore stores a set of merkle trees that
// it manages. It fulfills the modules.TreeStore interface.
// * It is responsible for atomic commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type treeStore struct {
	base_modules.IntegratableModule

	treeStoreDir string
	merkleTrees  map[merkleTree]*smt.SMT
	nodeStores   map[merkleTree]kvstore.KVStore
}

// Update takes a transaction and a height and updates
// all of the trees in the treeStore for that height.
func (t *treeStore) Update(pgtx pgx.Tx, height uint64) (string, error) {
	txi := t.GetBus().GetPersistenceModule().GetTxIndexer()
	return t.updateMerkleTrees(pgtx, txi, height)
}

// DebugClearAll is used by the debug cli to completely reset all merkle trees.
// This should only be called by the debug CLI.
// TECHDEBT: Move this into a separate file with a debug build flag to avoid accidental usage in prod
func (t *treeStore) DebugClearAll() error {
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		nodeStore := t.nodeStores[treeType]
		if err := nodeStore.ClearAll(); err != nil {
			return fmt.Errorf("failed to clear %s node store: %w", merkleTreeToString[treeType], err)
		}
		t.merkleTrees[treeType] = smt.NewSparseMerkleTree(nodeStore, smtTreeHasher)
	}
	return nil
}

// updateMerkleTrees updates all of the merkle trees in order defined by `numMerkleTrees`
// * it returns the new state hash capturing the state of all the trees or an error if one occured
func (t *treeStore) updateMerkleTrees(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		switch treeType {
		// Actor Merkle Trees
		case appMerkleTree, valMerkleTree, fishMerkleTree, servicerMerkleTree:
			actorType, ok := merkleTreeToActorTypeName[treeType]
			if !ok {
				return "", fmt.Errorf("no actor type found for merkle tree: %v", treeType)
			}

			actors, err := sql.GetActors(pgtx, actorType, height)
			if err != nil {
				return "", fmt.Errorf("failed to get actors at height: %w", err)
			}

			if err := t.updateActorsTree(actorType, actors); err != nil {
				return "", fmt.Errorf("failed to update actors tree for treeType: %v, actorType: %v - %w", treeType, actorType, err)
			}

		// Account Merkle Trees
		case accountMerkleTree:
			accounts, err := sql.GetAccounts(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get accounts: %w", err)
			}
			if err := t.updateAccountTrees(accounts); err != nil {
				return "", fmt.Errorf("failed to update account trees: %w", err)
			}
		case poolMerkleTree:
			pools, err := sql.GetPools(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updatePoolTrees(pools); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case transactionsMerkleTree:
			indexedTxs, err := sql.GetTransactions(txi, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updateTransactionsTree(indexedTxs); err != nil {
				return "", fmt.Errorf("failed to update transactions: %w", err)
			}
		case paramsMerkleTree:
			params, err := sql.GetParams(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get params: %w", err)
			}
			if err := t.updateParamsTree(params); err != nil {
				return "", fmt.Errorf("failed to update params tree: %w", err)
			}
		case flagsMerkleTree:
			flags, err := sql.GetFlags(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get flags from transaction: %w", err)
			}
			if err := t.updateFlagsTree(flags); err != nil {
				return "", fmt.Errorf("failed to update flags tree - %w", err)
			}
		// Default
		default:
			panic(fmt.Sprintf("not handled in state commitment update. Merkle tree #{%v}", treeType))
		}
	}

	if err := t.commit(); err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}
	return t.getStateHash(), nil
}

func (t *treeStore) commit() error {
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		if err := t.merkleTrees[tree].Commit(); err != nil {
			return fmt.Errorf("failed to commit %s: %w", merkleTreeToString[tree], err)
		}
	}
	return nil
}

func (t *treeStore) getStateHash() string {
	// create an order-matters list of roots
	roots := make([][]byte, 0)
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		roots = append(roots, t.merkleTrees[tree].Root())
	}

	// combine them and hash the result
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	// Convert the array to a slice and return it
	// REF: https://stackoverflow.com/questions/28886616/convert-array-to-slice-in-go
	return hex.EncodeToString(stateHash[:])
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
		if err := t.merkleTrees[merkleTreeName].Update(bzAddr, actorBz); err != nil {
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

		if err := t.merkleTrees[accountMerkleTree].Update(bzAddr, accBz); err != nil {
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

		if err := t.merkleTrees[poolMerkleTree].Update(bzAddr, accBz); err != nil {
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
		if err := t.merkleTrees[transactionsMerkleTree].Update(txHash, txBz); err != nil {
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
		if err := t.merkleTrees[paramsMerkleTree].Update(paramKey, paramBz); err != nil {
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
		if err := t.merkleTrees[flagsMerkleTree].Update(flagKey, flagBz); err != nil {
			return err
		}
	}

	return nil
}
