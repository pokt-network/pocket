package trees

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/smt"
)

type TreeStore interface {
	Update(pgx.Tx) error
}

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

var actorTypeToSchemaName = map[coreTypes.ActorType]types.ProtocolActorSchema{
	coreTypes.ActorType_ACTOR_TYPE_APP:      types.ApplicationActor,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      types.ValidatorActor,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     types.FishermanActor,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: types.ServicerActor,
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
	// IMPORTANT: The order in which these trees are defined is important and strict. It implicitly
	// defines the index of the root hash each independent as they are concatenated together
	// to generate the state hash.

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

// treeStore stores a set of merkle trees that
// it manages.
// * It is responsible for commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type treeStore struct {
	merkleTrees map[merkleTree]*smt.SparseMerkleTree

	// nodeStores & valueStore are part of the SMT, but references are kept below for convenience
	// and debugging purposes
	nodeStores  map[merkleTree]kvstore.KVStore
	valueStores map[merkleTree]kvstore.KVStore
}

// var _ TreeStore = &treeStore{} // TODO

func NewTreeStore() *treeStore {
	panic("not impl")
}

// Update takes a transaction and a height and updates
// all of the trees in the treeStore for that height.
func (t *treeStore) Update(pgtx pgx.Tx, height uint64) error {
	hash, err := t.updateMerkleTrees(pgtx, height)
	fmt.Printf("hash updated: %s", hash)
	return err
}

func NewtreeStore(treesStoreDir string) (*treeStore, error) {
	if treesStoreDir == "" {
		return newMemtreeStore()
	}

	treeStore := &treeStore{
		merkleTrees: make(map[merkleTree]*smt.SparseMerkleTree, int(numMerkleTrees)),
		nodeStores:  make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
		valueStores: make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
	}

	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", treesStoreDir, merkleTreeToString[tree]))
		if err != nil {
			return nil, err
		}
		valueStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_values", treesStoreDir, merkleTreeToString[tree]))
		if err != nil {
			return nil, err
		}
		treeStore.nodeStores[tree] = nodeStore
		treeStore.valueStores[tree] = valueStore
		treeStore.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return treeStore, nil
}

func newMemtreeStore() (*treeStore, error) {
	treeStore := &treeStore{
		merkleTrees: make(map[merkleTree]*smt.SparseMerkleTree, int(numMerkleTrees)),
		nodeStores:  make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
		valueStores: make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
	}
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		valueStore := kvstore.NewMemKVStore()
		treeStore.nodeStores[tree] = nodeStore
		treeStore.valueStores[tree] = valueStore
		treeStore.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return treeStore, nil
}

// updateMerkleTrees updates all of the merkle trees that TreeStore manages.
// * it returns an hash of the output or an error.
func (t *treeStore) updateMerkleTrees(pgtx pgx.Tx, height uint64) (string, error) {
	// Update all the merkle trees
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		switch treeType {
		// Actor Merkle Trees
		case appMerkleTree, valMerkleTree, fishMerkleTree, servicerMerkleTree:
			actorType, ok := merkleTreeToActorTypeName[treeType]
			if !ok {
				return "", fmt.Errorf("no actor type found for merkle tree: %v", treeType)
			}

			actors, err := t.getActorsUpdated(pgtx, actorType, height)
			if err != nil {
				return "", fmt.Errorf("failed to get actors at height: %w", err)
			}

			if err := t.updateActorsTree(actorType, actors); err != nil {
				return "", fmt.Errorf("failed to update actors tree for treeType: %v, actorType: %v - %w", treeType, actorType, err)
			}

		// Account Merkle Trees
		case accountMerkleTree:
			accounts, err := t.getAccounts(pgtx)
			if err != nil {
				return "", fmt.Errorf("failed to get accounts: %w", err)
			}
			if err := t.updateAccountTrees(accounts); err != nil {
				return "", fmt.Errorf("failed to update account trees: %w", err)
			}
		case poolMerkleTree:
			pools, err := t.getPools(pgtx)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updatePoolTrees(pools); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case transactionsMerkleTree:
			indexedTxs, err := t.getTransactions(pgtx)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updateTransactionsTree(indexedTxs); err != nil {
				return "", fmt.Errorf("failed to update transactions: %w", err)
			}
		case paramsMerkleTree:
			params, err := t.getParams(pgtx)
			if err != nil {
				return "", fmt.Errorf("failed to get params: %w", err)
			}
			if err := t.updateParamsTree(params); err != nil {
				return "", fmt.Errorf("failed to update params tree: %w", err)
			}
		case flagsMerkleTree:
			flags, err := t.getFlags(pgtx)
			if err != nil {
				return "", fmt.Errorf("failed to get flags from transaction: %w", err)
			}
			if err := t.updateFlagsTree(flags); err != nil {
				return "", fmt.Errorf("failed to update flags tree - %w", err)
			}
		// Default
		default:
			// t.logger.Fatal().Msgf("Not handled yet in state commitment update. Merkle tree #{%v}", treeType)
		}
	}

	return t.getStateHash(), nil
}

func (t *treeStore) getStateHash() string {
	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		roots = append(roots, t.merkleTrees[tree].Root())
	}

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	// Convert the array to a slice and return it
	return hex.EncodeToString(stateHash[:])
}

// Actor Tree Helpers

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
		if _, err := t.merkleTrees[merkleTreeName].Update(bzAddr, actorBz); err != nil {
			return err
		}
	}

	return nil
}

func (t *treeStore) getActorsUpdatedAtHeight(pgtx pgx.Tx, actorType coreTypes.ActorType, height int64) (actors []*coreTypes.Actor, err error) {
	actorSchema, ok := actorTypeToSchemaName[actorType]
	if !ok {
		return nil, fmt.Errorf("no schema found for actor type: %s", actorType)
	}

	schemaActors, err := t.getActorsUpdated(pgtx, actorSchema, uint64(height))
	if err != nil {
		return nil, err
	}

	actors = make([]*coreTypes.Actor, len(schemaActors))
	for i, schemaActor := range schemaActors {
		actor := &coreTypes.Actor{
			ActorType:       actorType,
			Address:         schemaActor.Address,
			PublicKey:       schemaActor.PublicKey,
			Chains:          schemaActor.Chains,
			ServiceUrl:      schemaActor.ServiceUrl,
			StakedAmount:    schemaActor.StakedAmount,
			PausedHeight:    schemaActor.PausedHeight,
			UnstakingHeight: schemaActor.UnstakingHeight,
			Output:          schemaActor.Output,
		}
		actors[i] = actor
	}
	return
}

// Account Tree Helpers

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

		if _, err := t.merkleTrees[accountMerkleTree].Update(bzAddr, accBz); err != nil {
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

		if _, err := t.merkleTrees[poolMerkleTree].Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

// Data Tree Helpers

func (t *treeStore) updateTransactionsTree(indexedTxs []*coreTypes.IndexedTransaction) error {
	for _, idxTx := range indexedTxs {
		txBz := idxTx.GetTx()
		txHash := crypto.SHA3Hash(txBz)
		if _, err := t.merkleTrees[transactionsMerkleTree].Update(txHash, txBz); err != nil {
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
		if _, err := t.merkleTrees[paramsMerkleTree].Update(paramKey, paramBz); err != nil {
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
		if _, err := t.merkleTrees[flagsMerkleTree].Update(flagKey, flagBz); err != nil {
			return err
		}
	}

	return nil
}

func (t *treeStore) getActorsUpdated(pgtx pgx.Tx, actorSchema types.ProtocolActorSchema, height uint64) ([]*coreTypes.Actor, error) {
	return nil, fmt.Errorf("not impl")
}

func (t *treeStore) getTransactions(pgtx pgx.Tx) ([]*coreTypes.IndexedTransaction, error) {
	return nil, fmt.Errorf("not impl")
}

func (t *treeStore) getPools(pgtx pgx.Tx) ([]*coreTypes.Account, error) {
	return nil, fmt.Errorf("not impl")
}

func (t *treeStore) getAccounts(pgtx pgx.Tx) ([]*coreTypes.Account, error) {
	return nil, fmt.Errorf("not impl")
}

func (t *treeStore) getFlags(pgtx pgx.Tx) ([]*coreTypes.Flag, error) {
	return nil, fmt.Errorf("not impl")
}

func (t *treeStore) getParams(pgtx pgx.Tx) ([]*coreTypes.Param, error) {
	return nil, fmt.Errorf("not impl")
}
