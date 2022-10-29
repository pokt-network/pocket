package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/types"
	"google.golang.org/protobuf/proto"
)

type MerkleTree float64

// A list of Merkle Trees used to maintain the state hash
const (
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

	// Used for iteration purposes only - see https://stackoverflow.com/a/64178235/768439
	numMerkleTrees
)

var actorTypeToMerkleTreeName map[types.ActorType]MerkleTree = map[types.ActorType]MerkleTree{
	types.ActorType_App:  appMerkleTree,
	types.ActorType_Val:  valMerkleTree,
	types.ActorType_Fish: fishMerkleTree,
	types.ActorType_Node: serviceNodeMerkleTree,
}

var merkleTreeToActorTypeName map[MerkleTree]types.ActorType = map[MerkleTree]types.ActorType{
	appMerkleTree:         types.ActorType_App,
	valMerkleTree:         types.ActorType_Val,
	fishMerkleTree:        types.ActorType_Fish,
	serviceNodeMerkleTree: types.ActorType_Node,
}

var actorTypeToSchemaName map[types.ActorType]types.ProtocolActorSchema = map[types.ActorType]types.ProtocolActorSchema{
	types.ActorType_App:  types.ApplicationActor,
	types.ActorType_Val:  types.ValidatorActor,
	types.ActorType_Fish: types.FishermanActor,
	types.ActorType_Node: types.ServiceNodeActor,
}

func newMerkleTrees() (map[MerkleTree]*smt.SparseMerkleTree, error) {
	trees := make(map[MerkleTree]*smt.SparseMerkleTree, int(numMerkleTrees))

	for treeType := MerkleTree(0); treeType < numMerkleTrees; treeType++ {
		// TODO_IN_THIS_COMMIT: Rather than using `NewSimpleMap`, use a disk based key-value store
		nodeStore := smt.NewSimpleMap()
		valueStore := smt.NewSimpleMap()

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
			actors, err := p.getActorsUpdatedAtHeight(actorType, p.Height)
			if err != nil {
				return err
			}
			if err != p.updateActorsTree(actorType, actors) {
				return err
			}
		case accountMerkleTree:
			fallthrough
		case poolMerkleTree:
			fallthrough
		case blocksMerkleTree:
			fallthrough
		case paramsMerkleTree:
			fallthrough
		case flagsMerkleTree:
			// log.Println("TODO: merkle tree not implemented", treeType)
		default:
			log.Fatalln("Not handled yet in state commitment update", treeType)
		}
	}

	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for treeType := MerkleTree(0); treeType < numMerkleTrees; treeType++ {
		roots = append(roots, p.merkleTrees[treeType].Root())
	}

	// DISCUSS(drewsky): In #152, we discussed the ordering of the roots
	// 	Strict Ordering: sha3(app_tree_root + fish_tree_root + service_node_tree_root + validator_tree_root)
	// 	Value Ordering sha3(app_tree_root <= + fish_tree_root <= + service_node_tree_root <= + validator_tree_root)
	// If we don't do the lexographic ordering below, then it follows the string ordering of
	// the merkle trees declared above. I have a feeling you're not a fan of this solution, but curious
	// to hear your thoughts.

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	p.currentStateHash = stateHash[:]
	return nil
}

func (p PostgresContext) updateActorsTree(actorType types.ActorType, actors []*types.Actor) error {
	for _, actor := range actors {
		bzAddr, err := hex.DecodeString(actor.GetAddress())
		if err != nil {
			return err
		}

		appBz, err := proto.Marshal(actor)
		if err != nil {
			return err
		}

		merkleTreeName, ok := actorTypeToMerkleTreeName[actorType]
		if !ok {
			return fmt.Errorf("no merkle tree found for actor type: %s", actorType)
		}

		if _, err := p.merkleTrees[merkleTreeName].Update(bzAddr, appBz); err != nil {
			return err
		}
	}

	return nil
}

func (p PostgresContext) getActorsUpdatedAtHeight(actorType types.ActorType, height int64) (actors []*types.Actor, err error) {
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
