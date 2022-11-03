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
	// defines the index of the the root hash each independent as they are concatenated together
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
	blocksMerkleTree
	paramsMerkleTree
	flagsMerkleTree
	// txMerkleTree ??

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

	blocksMerkleTree: "blocks",
	paramsMerkleTree: "params",
	flagsMerkleTree:  "flags",
}

var actorTypeToMerkleTreeName map[types.ActorType]merkleTree = map[types.ActorType]merkleTree{
	types.ActorType_App:  appMerkleTree,
	types.ActorType_Val:  valMerkleTree,
	types.ActorType_Fish: fishMerkleTree,
	types.ActorType_Node: serviceNodeMerkleTree,
}

var actorTypeToSchemaName map[types.ActorType]types.ProtocolActorSchema = map[types.ActorType]types.ProtocolActorSchema{
	types.ActorType_App:  types.ApplicationActor,
	types.ActorType_Val:  types.ValidatorActor,
	types.ActorType_Fish: types.FishermanActor,
	types.ActorType_Node: types.ServiceNodeActor,
}

var merkleTreeToActorTypeName map[merkleTree]types.ActorType = map[merkleTree]types.ActorType{
	appMerkleTree:         types.ActorType_App,
	valMerkleTree:         types.ActorType_Val,
	fishMerkleTree:        types.ActorType_Fish,
	serviceNodeMerkleTree: types.ActorType_Node,
}

var merkleTreeToProtoSchema = map[merkleTree]func() proto.Message{
	appMerkleTree:         func() proto.Message { return &types.Actor{} },
	valMerkleTree:         func() proto.Message { return &types.Actor{} },
	fishMerkleTree:        func() proto.Message { return &types.Actor{} },
	serviceNodeMerkleTree: func() proto.Message { return &types.Actor{} },

	accountMerkleTree: func() proto.Message { return &types.Account{} },
	poolMerkleTree:    func() proto.Message { return &types.Account{} },

	blocksMerkleTree: func() proto.Message { return &types.Block{} },
	// paramsMerkleTree: func() proto.Message { return &types.Params{} },
	// flagsMerkleTree:  func() proto.Message { return &types.Flags{} },
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

func (p *PostgresContext) updateStateHash() error {
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
				return fmt.Errorf("no actor type found for merkle tree: %v\n", treeType)
			}
			if err := p.updateActorsTree(actorType, p.Height); err != nil {
				return err
			}

		// Account Merkle Trees
		case accountMerkleTree:
			if err := p.updateAccountTrees(p.Height); err != nil {
				return err
			}
		case poolMerkleTree:
			if err := p.updatePoolTrees(p.Height); err != nil {
				return err
			}

		// Data Merkle Trees
		case blocksMerkleTree:
			p.updateBlockTree(p.Height)
		case paramsMerkleTree:
			// log.Printf("TODO: merkle tree not implemented yet. Merkle tree #{%v}\n", treeType)
		case flagsMerkleTree:
			// log.Printf("TODO: merkle tree not implemented yet. Merkle tree #{%v}\n", treeType)

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

	p.currentStateHash = stateHash[:]
	return nil
}

// Actor Tree Helpers

func (p *PostgresContext) updateActorsTree(actorType types.ActorType, height int64) error {
	actors, err := p.getActorsUpdatedAtHeight(actorType, height)
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

func (p *PostgresContext) updateAccountTrees(height int64) error {
	accounts, err := p.getAccountsUpdated(height)
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

func (p *PostgresContext) updatePoolTrees(height int64) error {
	pools, err := p.getPoolsUpdated(height)
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

func (p *PostgresContext) updateBlockTree(height int64) {

}
