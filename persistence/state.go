package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"google.golang.org/protobuf/proto"
)

type merkleTree float64

type stateTrees struct {
	merkleTrees map[merkleTree]*smt.SparseMerkleTree

	// nodeStores & valueStore are part of the SMT, but references are kept below for convenience
	// and debugging purposes
	nodeStores  map[merkleTree]kvstore.KVStore
	valueStores map[merkleTree]kvstore.KVStore
}

// A list of Merkle Trees used to maintain the state hash.
const (
	// VERY IMPORTANT: The order in which these trees are defined is important and strict. It implicitly
	// defines the index of the root hash each independent as they are concatenated together
	// to generate the state hash.

	// Actor Merkle Trees
	appMerkleTree merkleTree = iota
	valMerkleTree
	fishMerkleTree
	serviceNodeMerkleTree

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

var merkleTreeToString = map[merkleTree]string{
	appMerkleTree:         "app",
	valMerkleTree:         "val",
	fishMerkleTree:        "fish",
	serviceNodeMerkleTree: "serviceNode",

	accountMerkleTree: "account",
	poolMerkleTree:    "pool",

	transactionsMerkleTree: "transactions",
	paramsMerkleTree:       "params",
	flagsMerkleTree:        "flags",
}

var actorTypeToMerkleTreeName = map[types.ActorType]merkleTree{
	types.ActorType_App:  appMerkleTree,
	types.ActorType_Val:  valMerkleTree,
	types.ActorType_Fish: fishMerkleTree,
	types.ActorType_Node: serviceNodeMerkleTree,
}

var actorTypeToSchemaName = map[types.ActorType]types.ProtocolActorSchema{
	types.ActorType_App:  types.ApplicationActor,
	types.ActorType_Val:  types.ValidatorActor,
	types.ActorType_Fish: types.FishermanActor,
	types.ActorType_Node: types.ServiceNodeActor,
}

var merkleTreeToActorTypeName = map[merkleTree]types.ActorType{
	appMerkleTree:         types.ActorType_App,
	valMerkleTree:         types.ActorType_Val,
	fishMerkleTree:        types.ActorType_Fish,
	serviceNodeMerkleTree: types.ActorType_Node,
}

func newStateTrees(treesStoreDir string) (*stateTrees, error) {
	if treesStoreDir == "" {
		return newMemStateTrees()
	}

	stateTrees := &stateTrees{
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
		stateTrees.nodeStores[tree] = nodeStore
		stateTrees.valueStores[tree] = valueStore
		stateTrees.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return stateTrees, nil
}

func newMemStateTrees() (*stateTrees, error) {
	stateTrees := &stateTrees{
		merkleTrees: make(map[merkleTree]*smt.SparseMerkleTree, int(numMerkleTrees)),
		nodeStores:  make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
		valueStores: make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
	}
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		nodeStore := kvstore.NewMemKVStore() // For testing, `smt.NewSimpleMap()` can be used as well
		valueStore := kvstore.NewMemKVStore()
		stateTrees.nodeStores[tree] = nodeStore
		stateTrees.valueStores[tree] = valueStore
		stateTrees.merkleTrees[tree] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return stateTrees, nil
}

func (p *PostgresContext) updateMerkleTrees() ([]byte, error) {
	// Update all the merkle trees
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		switch treeType {
		// Actor Merkle Trees
		case appMerkleTree:
			fallthrough
		case valMerkleTree:
			fallthrough
		case fishMerkleTree:
			fallthrough
		case serviceNodeMerkleTree:
			actorType, ok := merkleTreeToActorTypeName[treeType]
			if !ok {
				return nil, fmt.Errorf("no actor type found for merkle tree: %v\n", treeType)
			}
			if err := p.updateActorsTree(actorType); err != nil {
				return nil, err
			}

		// Account Merkle Trees
		case accountMerkleTree:
			if err := p.updateAccountTrees(); err != nil {
				return nil, err
			}
		case poolMerkleTree:
			if err := p.updatePoolTrees(); err != nil {
				return nil, err
			}

		// Data Merkle Trees
		case transactionsMerkleTree:
			if err := p.updateTransactionsTree(); err != nil {
				return nil, err
			}
		case paramsMerkleTree:
			if err := p.updateParamsTree(); err != nil {
				return nil, err
			}
		case flagsMerkleTree:
			if err := p.updateFlagsTree(); err != nil {
				return nil, err
			}

		// Default
		default:
			log.Fatalf("Not handled yet in state commitment update. Merkle tree #{%v}\n", treeType)
		}
	}

	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		roots = append(roots, p.stateTrees.merkleTrees[tree].Root())
	}

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	return stateHash[:], nil
}

// Actor Tree Helpers

func (p *PostgresContext) updateActorsTree(actorType types.ActorType) error {
	actors, err := p.getActorsUpdatedAtHeight(actorType, p.Height)
	if err != nil {
		return err
	}

	for _, actor := range actors {
		bzAddr, err := hex.DecodeString(actor.GetAddress())
		if err != nil {
			return err
		}

		actorBz, err := proto.Marshal(actor)
		if err != nil {
			return err
		}

		merkleTreeName, ok := actorTypeToMerkleTreeName[actorType]
		if !ok {
			return fmt.Errorf("no merkle tree found for actor type: %s", actorType)
		}
		if _, err := p.stateTrees.merkleTrees[merkleTreeName].Update(bzAddr, actorBz); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) getActorsUpdatedAtHeight(actorType types.ActorType, height int64) (actors []*types.Actor, err error) {
	actorSchema, ok := actorTypeToSchemaName[actorType]
	if !ok {
		return nil, fmt.Errorf("no schema found for actor type: %s", actorType)
	}

	schemaActors, err := p.GetActorsUpdated(actorSchema, height)
	if err != nil {
		return nil, err
	}

	actors = make([]*types.Actor, len(schemaActors))
	for i, schemaActor := range schemaActors {
		actor := &types.Actor{
			ActorType:       actorType,
			Address:         schemaActor.Address,
			PublicKey:       schemaActor.PublicKey,
			Chains:          schemaActor.Chains,
			GenericParam:    schemaActor.GenericParam,
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

func (p *PostgresContext) updateAccountTrees() error {
	accounts, err := p.GetAccountsUpdated(p.Height)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		bzAddr, err := hex.DecodeString(account.GetAddress())
		if err != nil {
			return err
		}

		accBz, err := proto.Marshal(account)
		if err != nil {
			return err
		}

		if _, err := p.stateTrees.merkleTrees[accountMerkleTree].Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) updatePoolTrees() error {
	pools, err := p.GetPoolsUpdated(p.Height)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		bzAddr := []byte(pool.GetAddress())
		accBz, err := proto.Marshal(pool)
		if err != nil {
			return err
		}

		if _, err := p.stateTrees.merkleTrees[poolMerkleTree].Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

// Data Tree Helpers

func (p *PostgresContext) updateTransactionsTree() error {
	txResults, err := p.txIndexer.GetByHeight(p.Height, false)
	if err != nil {
		return err
	}

	for _, txResult := range txResults {
		txHash, err := txResult.Hash()
		if err != nil {
			return err
		}
		if _, err := p.stateTrees.merkleTrees[transactionsMerkleTree].Update(txHash, txResult.GetTx()); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) updateParamsTree() error {
	// TODO_IN_NEXT_COMMIT(olshansky): Implement me
	return nil
}

func (p *PostgresContext) updateFlagsTree() error {
	// TODO_IN_NEXT_COMMIT(olshansky): Implement me
	return nil
}
