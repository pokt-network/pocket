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

type MerkleTree float64

// A list of Merkle Trees used to maintain the state hash.
const (
	// VERY IMPORTANT: The order in which these trees are defined is important and strict. It implicitly
	// defines the index of the the root hash each independent as they are concatenated together
	// to generate the state hash.

	// Actor Merkle Trees
	appMerkleTree MerkleTree = iota
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

var actorTypeToMerkleTreeName map[types.ActorType]MerkleTree = map[types.ActorType]MerkleTree{
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

var merkleTreeToActorTypeName map[MerkleTree]types.ActorType = map[MerkleTree]types.ActorType{
	appMerkleTree:         types.ActorType_App,
	valMerkleTree:         types.ActorType_Val,
	fishMerkleTree:        types.ActorType_Fish,
	serviceNodeMerkleTree: types.ActorType_Node,
}

func newMerkleTrees() (map[MerkleTree]*smt.SparseMerkleTree, error) {
	trees := make(map[MerkleTree]*smt.SparseMerkleTree, int(numMerkleTrees))

	for treeType := MerkleTree(0); treeType < numMerkleTrees; treeType++ {
		// TODO_IN_THIS_COMMIT: Rather than using `NewSimpleMap`, use a disk based key-value store
		// nodeStore := smt.NewSimpleMap()
		// valueStore := smt.NewSimpleMap()
		nodeStore := kvstore.NewMemKVStore()
		valueStore := kvstore.NewMemKVStore()

		trees[treeType] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return trees, nil
}

func loadMerkleTrees(map[MerkleTree]*smt.SparseMerkleTree, error) {
	log.Fatalf("TODO: loadMerkleTrees not implemented yet")
}

func (p *PostgresContext) updateStateHash() error {
	// Update all the merkle trees
	for treeType := MerkleTree(0); treeType < numMerkleTrees; treeType++ {
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
			if err := p.updateAccountTrees(false, p.Height); err != nil {
				return err
			}
		case poolMerkleTree:
			if err := p.updateAccountTrees(true, p.Height); err != nil {
				return err
			}

		// Data Merkle Trees
		case blocksMerkleTree:
			p.updateBlockTree(p.Height)
		case paramsMerkleTree:
			log.Printf("TODO: merkle tree not implemented yet. Merkle tree #{%v}\n", treeType)
		case flagsMerkleTree:
			log.Printf("TODO: merkle tree not implemented yet. Merkle tree #{%v}\n", treeType)

		// Default
		default:
			log.Fatalf("Not handled yet in state commitment update. Merkle tree #{%v}\n", treeType)
		}
	}

	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for treeType := MerkleTree(0); treeType < numMerkleTrees; treeType++ {
		roots = append(roots, p.merkleTrees[treeType].Root())
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

		if _, err := p.merkleTrees[merkleTreeName].Update(bzAddr, actorBz); err != nil {
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
	for _, actor := range schemaActors {
		actor := &types.Actor{
			ActorType:       actorType,
			Address:         actor.Address,
			PublicKey:       actor.PublicKey,
			Chains:          actor.Chains,
			GenericParam:    actor.GenericParam,
			StakedAmount:    actor.StakedAmount,
			PausedHeight:    actor.PausedHeight,
			UnstakingHeight: actor.UnstakingHeight,
			Output:          actor.Output,
		}
		actors = append(actors, actor)
	}
	return
}

// Account Tree Helpers

// Helper to update both `Pool` and `Account` Merkle Trees. The use of `isPool` is a bit hacky, but
// but simplifies the code since Pools are just specialized versions of accounts.
func (p *PostgresContext) updateAccountTrees(isPool bool, height int64) error {
	var merkleTreeName MerkleTree
	var accounts []*types.Account
	var err error

	if isPool {
		merkleTreeName = poolMerkleTree
		accounts, err = p.getPoolsUpdated(height)
	} else {
		merkleTreeName = accountMerkleTree
		accounts, err = p.getAccountsUpdated(height)
	}
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

		if _, err := p.merkleTrees[merkleTreeName].Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

// Data Tree Helpers

func (p *PostgresContext) updateBlockTree(height int64) {

}
