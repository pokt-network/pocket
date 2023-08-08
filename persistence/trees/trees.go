// package trees maintains a set of sparse merkle trees
// each backed by the `KVStore` interface. It offers an atomic
// commit and rollback mechanism for interacting with
// its core resource - a set of merkle trees.
// - `Update` is called, which will fetch and apply the contextual changes to the respective trees.
// - `Savepoint` is first called to create a new anchor in time that can be rolled back to
// - `Commit` must be called after any `Update` calls to persist changes applied to disk.
// - If `Rollback` is called at any point before committing, it rolls the TreeStore state back to the
//    earlier savepoint. This means that the caller is responsible for correctly managing atomic updates
//     of the TreeStore.
// In most contexts, this is from the perspective of the `utility/unit_of_work` package.

package trees

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"os"
	"path/filepath"

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

var _ modules.TreeStoreModule = &treeStore{}

// ErrFailedRollback is thrown when a rollback fails to reset the TreeStore to a known good state
var ErrFailedRollback = fmt.Errorf("failed to rollback")

// treeStore stores a set of merkle trees that it manages.
// It fulfills the modules.treeStore interface
// * It is responsible for atomic commit or rollback behavior of the underlying
// trees by utilizing the lazy loading functionality of the smt library.
// TECHDEBT(#880): treeStore is exported for testing purposes to avoid import cycle errors.
// Make it private and export a custom struct with a test build tag when necessary.
type treeStore struct {
	base_modules.IntegrableModule

	logger *modules.Logger

	treeStoreDir string
	rootTree     *stateTree
	merkleTrees  map[string]*stateTree

	// prevState holds a previous view of the worldState.
	// The tree store rolls back to this view if errors are encountered during block application.
	prevState *worldState
}

// worldState holds a (de)serializable view of the entire tree state.
// TECHDEBT(#566) - Hook this up to node CLI subcommands
type worldState struct {
	treeStoreDir string
	rootTree     *stateTree
	rootHash     []byte
	merkleTrees  map[string]*stateTree
	merkleRoots  map[string][]byte
}

// worldStateJson holds exported members for proper JSON marshaling and unmarshaling.
// It contains the root hash of the merkle roots as a byte slice and a map of the MerkleRoots
// where each key is the name of the file in the same directory that corresponds to the baderDB
// backup file for that tree. That tree's hash is the value of that object for checking the integrity
// of each file and tree.
type worldStateJson struct {
	RootHash    []byte
	MerkleRoots map[string][]byte //
}

// GetTree returns the root hash and nodeStore for the matching tree stored in the TreeStore.
// This enables the caller to import the SMT without changing the one stored unless they call
// `Commit()` to write to the nodestore.
func (t *treeStore) GetTree(name string) ([]byte, kvstore.KVStore) {
	if name == RootTreeName && t.rootTree.tree != nil {
		return t.rootTree.tree.Root(), t.rootTree.nodeStore
	}
	if tree, ok := t.merkleTrees[name]; ok && tree != nil {
		return tree.tree.Root(), tree.nodeStore
	}
	return nil, nil
}

// Prove generates and verifies a proof against the tree name stored in the TreeStore
// using the given key-value pair. If value == nil this will be an exclusion proof,
// otherwise it will be an inclusion proof.
func (t *treeStore) Prove(name string, key, value []byte) (bool, error) {
	st, ok := t.merkleTrees[name]
	if !ok {
		return false, fmt.Errorf("tree not found: %s", name)
	}
	proof, err := st.tree.Prove(key)
	if err != nil {
		return false, fmt.Errorf("error generating proof (%s): %w", name, err)
	}
	if valid := smt.VerifyProof(proof, st.tree.Root(), key, value, st.tree.Spec()); !valid {
		return false, nil
	}
	return true, nil
}

// GetTreeHashes returns a map of tree names to their root hashes for all
// the trees tracked by the treestore, excluding the root tree
func (t *treeStore) GetTreeHashes() map[string]string {
	hashes := make(map[string]string, len(t.merkleTrees))
	for treeName, stateTree := range t.merkleTrees {
		hashes[treeName] = hex.EncodeToString(stateTree.tree.Root())
	}
	return hashes
}

// Update takes a transaction and a height and updates
// all of the trees in the treeStore for that height.
func (t *treeStore) Update(pgtx pgx.Tx, height uint64) (string, error) {
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

// updateMerkleTrees updates all of the merkle trees in order defined by `numMerkleTrees`
// * It returns the new state hash capturing the state of all the trees or an error if one occurred.
// * This function does not commit state to disk. The caller must manually invoke `Commit` to persist
// changes to disk.
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

	return t.getStateHash(), nil
}

// Commit commits changes in the sub-trees to the root tree and then commits updates for each sub-tree.
func (t *treeStore) Commit() error {
	if err := t.rootTree.tree.Commit(); err != nil {
		t.logger.Err(err).Msg("TECHDEBT: failed to commit root tree: changes to sub-trees will not be committed - this should be investigated")
		return fmt.Errorf("failed to commit root tree: %w", err)
	}

	for name, treeStore := range t.merkleTrees {
		if err := treeStore.tree.Commit(); err != nil {
			t.logger.Err(err).Msgf("TECHDEBT: failed to commit to %s tree: changes will not be saved - this should be investigated", name)
			return fmt.Errorf("failed to commit %s: %w", name, err)
		}
	}

	return nil
}

func (t *treeStore) getStateHash() string {
	for _, stateTree := range t.merkleTrees {
		key := []byte(stateTree.name)
		val := stateTree.tree.Root()
		if err := t.rootTree.tree.Update(key, val); err != nil {
			log.Fatalf("failed to update root tree with %s tree's hash: %v", stateTree.name, err)
		}
	}
	// Convert the array to a slice and return it
	// REF: https://stackoverflow.com/questions/28886616/convert-array-to-slice-in-go
	root := t.rootTree.tree.Root()
	hexHash := hex.EncodeToString(root)
	t.logger.Info().Msgf("#️⃣ calculated state hash: %s", hexHash)
	return hexHash
}

////////////////////////////////
// AtomicStore Implementation //
////////////////////////////////

// Savepoint generates a new savepoint (i.e. a worldState) for the tree store and saves it internally.
func (t *treeStore) Savepoint() error {
	w, err := t.save()
	if err != nil {
		return err
	}
	t.prevState = w
	return nil
}

// Rollback returns the treeStore to the last saved worldState maintained by the treeStore.
// If no worldState has been saved, it returns ErrFailedRollback
func (t *treeStore) Rollback() error {
	if t.prevState != nil {
		t.merkleTrees = t.prevState.merkleTrees
		t.rootTree = t.prevState.rootTree
		return nil
	}
	t.logger.Err(ErrFailedRollback)
	return ErrFailedRollback
}

// Load sets the TreeStore merkle and root trees to the values provided in the worldstate
func (t *treeStore) Load(dir string) error {
	// look for a worldstate.json file to hydrate
	data, err := readFile(filepath.Join(dir, "worldstate.json"))
	if err != nil {
		return err
	}

	// assign tree store directory to dir if a valid worldstate.json exists
	t.treeStoreDir = dir

	// hydrate a worldstate from the json object
	var w *worldStateJson
	err = json.Unmarshal(data, &w)
	if err != nil {
		return err
	}

	t.logger.Info().Msgf("🌏 worldstate detected, beginning import at %s", dir)

	// create a new root tree and node store
	nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", t.treeStoreDir, RootTreeName))
	if err != nil {
		return err
	}
	t.rootTree = &stateTree{
		name:      RootTreeName,
		tree:      smt.NewSparseMerkleTree(nodeStore, smtTreeHasher),
		nodeStore: nodeStore,
	}

	// import merkle trees with the proper hash
	t.merkleTrees = make(map[string]*stateTree)
	for treeName, treeRootHash := range w.MerkleRoots {
		treePath := fmt.Sprintf("%s/%s_nodes", dir, treeName)
		nodeStore, err := kvstore.NewKVStore(treePath)
		if err != nil {
			return err
		}
		t.merkleTrees[treeName] = &stateTree{
			name:      treeName,
			tree:      smt.ImportSparseMerkleTree(nodeStore, smtTreeHasher, treeRootHash),
			nodeStore: nodeStore,
		}
		t.logger.Info().Msgf("🌳 %s initialized at %s", treeName, hex.EncodeToString(w.MerkleRoots[treeName]))
	}

	return nil
}

// save commits any pending changes to the trees and creates a copy of the current state of the
// tree store then saves that copy as a rollback point for later use if errors are encountered.
// OPTIMIZE: Consider saving only the root hash of each tree and the tree directory here and then
// load the trees up in Rollback instead of setting them up here.
func (t *treeStore) save() (*worldState, error) {
	if err := t.Commit(); err != nil {
		return nil, err
	}

	w := &worldState{
		treeStoreDir: t.treeStoreDir,
		merkleRoots:  make(map[string][]byte),
		merkleTrees:  make(map[string]*stateTree),
		rootHash:     t.rootTree.tree.Root(),
		rootTree:     t.rootTree,
	}

	for treeName := range t.merkleTrees {
		root, nodeStore := t.GetTree(treeName)
		tree := smt.ImportSparseMerkleTree(nodeStore, smtTreeHasher, root)
		w.merkleTrees[treeName] = &stateTree{
			name:      treeName,
			tree:      tree,
			nodeStore: nodeStore,
		}
	}

	root, nodeStore := t.GetTree(RootTreeName)
	tree := smt.ImportSparseMerkleTree(nodeStore, smtTreeHasher, root)
	w.rootTree = &stateTree{
		name:      RootTreeName,
		tree:      tree,
		nodeStore: nodeStore,
	}

	return w, nil
}

// Backup creates a new backup of each tree in the tree store to the provided directory.
// Each tree is backed up in an eponymous file in the provided backupDir.
func (t *treeStore) Backup(backupDir string) error {
	// save all current branches
	if err := t.Commit(); err != nil {
		return err
	}

	w := &worldStateJson{
		RootHash:    []byte(t.getStateHash()),
		MerkleRoots: make(map[string][]byte),
	}

	for _, st := range t.merkleTrees {
		treePath := fmt.Sprintf("%s/%s_nodes.bak", backupDir, st.name)
		if err := st.nodeStore.Backup(treePath); err != nil {
			t.logger.Err(err).Msgf("failed to backup %s tree: %+v", st.name, err)
			return err
		}
		w.MerkleRoots[st.name] = st.tree.Root()
	}

	worldstatePath := filepath.Join(backupDir, "worldstate.json")
	err := writeFile(worldstatePath, w)
	if err != nil {
		return err
	}

	t.logger.Info().Msgf("💾 backup created at %s", backupDir)

	return nil
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

func (t *treeStore) updateIBCTree(keys, values [][]byte) error {
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

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Use os.Stat to get file size and read the content into a byte slice
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func writeFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use the json.MarshalIndent function to encode data into JSON format with indentation
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
