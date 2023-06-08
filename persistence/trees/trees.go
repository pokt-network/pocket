// package trees maintains a set of sparse merkle trees
// each backed by the KVStore interface. It offers an atomic
// commit and rollback mechanism for interacting with
// that core resource map of merkle trees.
package trees

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/kvstore"
	ptypes "github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/smt"
)

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

var actorTypeToSchemaName = map[coreTypes.ActorType]ptypes.ProtocolActorSchema{
	coreTypes.ActorType_ACTOR_TYPE_APP:      ptypes.ApplicationActor,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      ptypes.ValidatorActor,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     ptypes.FishermanActor,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: ptypes.ServicerActor,
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
// it manages. It fulfills the modules.TreeStore interface.
// * It is responsible for commit or rollback behavior
// of the underlying trees by utilizing the lazy loading
// functionality provided by the underlying smt library.
type treeStore struct {
	treeStoreDir string
	merkleTrees  map[merkleTree]*smt.SMT
	nodeStores   map[merkleTree]kvstore.KVStore
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

	stateTrees := &treeStore{
		treeStoreDir: treesStoreDir,
		merkleTrees:  make(map[merkleTree]*smt.SMT, int(numMerkleTrees)),
		nodeStores:   make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
	}

	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore, err := kvstore.NewKVStore(fmt.Sprintf("%s/%s_nodes", treesStoreDir, merkleTreeToString[tree]))
		if err != nil {
			return nil, err
		}
		stateTrees.nodeStores[tree] = nodeStore
		stateTrees.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, sha256.New())
	}
	return stateTrees, nil
}

// DebugClearAll is used by the debug cli to completely reset all merkle trees.
// This should only be called by the debug CLI.
func (t *treeStore) DebugClearAll() error {
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		nodeStore := t.nodeStores[treeType]
		if err := nodeStore.ClearAll(); err != nil {
			return fmt.Errorf("failed to clear %s node store: %w", merkleTreeToString[treeType], err)
		}
		t.merkleTrees[treeType] = smt.NewSparseMerkleTree(nodeStore, sha256.New())
	}

	return nil
}

// newMemStateTrees creates a new in-memory state tree
func newMemStateTrees() (*treeStore, error) {
	stateTrees := &treeStore{
		merkleTrees: make(map[merkleTree]*smt.SMT, int(numMerkleTrees)),
		nodeStores:  make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
	}
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		stateTrees.nodeStores[tree] = nodeStore
		stateTrees.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, sha256.New())
	}
	return stateTrees, nil
}

// updateMerkleTrees updates all of the merkle trees that TreeStore manages.
// * it returns an hash of the output or an error.
func (t *treeStore) updateMerkleTrees(pgtx pgx.Tx, txi indexer.TxIndexer, height uint64) (string, error) {
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
			accounts, err := t.getAccounts(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get accounts: %w", err)
			}
			if err := t.updateAccountTrees(accounts); err != nil {
				return "", fmt.Errorf("failed to update account trees: %w", err)
			}
		case poolMerkleTree:
			pools, err := t.getPools(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updatePoolTrees(pools); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case transactionsMerkleTree:
			indexedTxs, err := t.getTransactions(txi, height)
			if err != nil {
				return "", fmt.Errorf("failed to get transactions: %w", err)
			}
			if err := t.updateTransactionsTree(indexedTxs); err != nil {
				return "", fmt.Errorf("failed to update transactions: %w", err)
			}
		case paramsMerkleTree:
			params, err := t.getParams(pgtx, height)
			if err != nil {
				return "", fmt.Errorf("failed to get params: %w", err)
			}
			if err := t.updateParamsTree(params); err != nil {
				return "", fmt.Errorf("failed to update params tree: %w", err)
			}
		case flagsMerkleTree:
			flags, err := t.getFlags(pgtx, height)
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

// getActorsUpdated is responsible for fetching the actors that have been updated at a given height.
func (t *treeStore) getActorsUpdated(
	pgtx pgx.Tx,
	actorType coreTypes.ActorType,
	height uint64,
) ([]*coreTypes.Actor, error) {
	actorSchema, ok := actorTypeToSchemaName[actorType]
	if !ok {
		return nil, fmt.Errorf("no schema found for actor type: %s", actorType)
	}

	// TECHDEBT (ISSUE #813): Avoid this cast to int64
	// https://github.com/pokt-network/pocket/issues/813
	query := actorSchema.GetUpdatedAtHeightQuery(int64(height))
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addrs := make([][]byte, 0)
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addrBz, err := hex.DecodeString(addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addrBz)
	}

	actors := make([]*coreTypes.Actor, len(addrs))
	for i, addr := range addrs {
		// TECHDEBT #XXX: Avoid this cast to int64
		actor, err := t.getActor(pgtx, actorSchema, addr, int64(height))
		if err != nil {
			return nil, err
		}
		actors[i] = actor
	}
	rows.Close()

	return actors, nil
}

func (t *treeStore) getAccountsUpdated(
	pgtx pgx.Tx,
	acctType ptypes.ProtocolAccountSchema,
	height uint64,
) ([]*coreTypes.Account, error) {
	accounts := []*coreTypes.Account{}

	// TECHDEBT (ISSUE #813): Avoid this cast to int64
	// https://github.com/pokt-network/pocket/issues/813
	query := acctType.GetAccountsUpdatedAtHeightQuery(int64(height))
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		acc := new(coreTypes.Account)
		if err := rows.Scan(&acc.Address, &acc.Amount); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (t *treeStore) getTransactions(txi indexer.TxIndexer, height uint64) ([]*coreTypes.IndexedTransaction, error) {
	// TECHDEBT (ISSUE #813): Avoid this cast to int64
	// https://github.com/pokt-network/pocket/issues/813
	indexedTxs, err := txi.GetByHeight(int64(height), false)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by height: %w", err)
	}
	return indexedTxs, nil
}

// getPools returns the pools updated at the given height
func (t *treeStore) getPools(pgtx pgx.Tx, height uint64) ([]*coreTypes.Account, error) {
	pools, err := t.getAccountsUpdated(pgtx, ptypes.Pool, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %w", err)
	}
	return pools, nil
}

// getAccounts returns the list of accounts updated at the provided height
func (t *treeStore) getAccounts(pgtx pgx.Tx, height uint64) ([]*coreTypes.Account, error) {
	accounts, err := t.getAccountsUpdated(pgtx, ptypes.Account, height)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return accounts, nil
}

func (t *treeStore) getFlags(pgtx pgx.Tx, height uint64) ([]*coreTypes.Flag, error) {
	fields := "name,value,enabled"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY name ASC", fields, ptypes.FlagsTableName, height)
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags: %w", err)
	}
	defer rows.Close()

	flagSlice := []*coreTypes.Flag{}
	for rows.Next() {
		flag := new(coreTypes.Flag)
		if err := rows.Scan(&flag.Name, &flag.Value, &flag.Enabled); err != nil {
			return nil, err
		}
		flag.Height = int64(height)
		flagSlice = append(flagSlice, flag)
	}

	return flagSlice, nil
}

func (t *treeStore) getParams(pgtx pgx.Tx, height uint64) ([]*coreTypes.Param, error) {
	fields := "name,value"
	query := fmt.Sprintf("SELECT %s FROM %s WHERE height=%d ORDER BY name ASC", fields, ptypes.ParamsTableName, height)
	rows, err := pgtx.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paramSlice []*coreTypes.Param
	for rows.Next() {
		param := new(coreTypes.Param)
		if err := rows.Scan(&param.Name, &param.Value); err != nil {
			return nil, err
		}
		param.Height = int64(height)
		paramSlice = append(paramSlice, param)
	}

	return paramSlice, nil
}

func (t *treeStore) getActor(tx pgx.Tx, actorSchema ptypes.ProtocolActorSchema, address []byte, height int64) (actor *coreTypes.Actor, err error) {
	ctx := context.TODO()
	actor, height, err = t.getActorFromRow(actorSchema.GetActorType(), tx.QueryRow(ctx, actorSchema.GetQuery(hex.EncodeToString(address), height)))
	if err != nil {
		return
	}
	return t.getChainsForActor(ctx, tx, actorSchema, actor, height)
}

func (t *treeStore) getActorFromRow(actorType coreTypes.ActorType, row pgx.Row) (actor *coreTypes.Actor, height int64, err error) {
	actor = &coreTypes.Actor{
		ActorType: actorType,
	}
	err = row.Scan(
		&actor.Address,
		&actor.PublicKey,
		&actor.StakedAmount,
		&actor.ServiceUrl,
		&actor.Output,
		&actor.PausedHeight,
		&actor.UnstakingHeight,
		&height)
	return
}

func (t *treeStore) getChainsForActor(
	ctx context.Context,
	tx pgx.Tx,
	actorSchema ptypes.ProtocolActorSchema,
	actor *coreTypes.Actor,
	height int64,
) (a *coreTypes.Actor, err error) {
	if actorSchema.GetChainsTableName() == "" {
		return actor, nil
	}
	rows, err := tx.Query(ctx, actorSchema.GetChainsQuery(actor.Address, height))
	if err != nil {
		return actor, err
	}
	defer rows.Close()

	var chainAddr string
	var chainID string
	var chainEndHeight int64 // DISCUSS: why is this commented as "unused"?
	for rows.Next() {
		err = rows.Scan(&chainAddr, &chainID, &chainEndHeight)
		if err != nil {
			return
		}
		if chainAddr != actor.Address {
			return actor, fmt.Errorf("unexpected address %s, expected %s when reading chains", chainAddr, actor.Address)
		}
		actor.Chains = append(actor.Chains, chainID)
	}
	return actor, nil
}
