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

// A work-in-progress list of all the trees we need to update to maintain the overall state
const (
	// Actor  Merkle Trees
	appMerkleTree MerkleTree = iota
	valMerkleTree
	fishMerkleTree
	serviceNodeMerkleTree

	// Account Merkle Trees
	accountMerkleTree
	poolMerkleTree

	// Data / State Merkle Trees
	blocksMerkleTree
	paramsMerkleTree
	flagsMerkleTree

	// Used for iteration purposes only - see https://stackoverflow.com/a/64178235/768439
	lastMerkleTree
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
	// We need a separate Merkle tree for each type of actor or storage
	trees := make(map[MerkleTree]*smt.SparseMerkleTree, int(lastMerkleTree))

	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		// TODO_IN_THIS_COMMIT: Rather than using `NewSimpleMap`, use a disk based key-value store
		nodeStore := smt.NewSimpleMap()
		valueStore := smt.NewSimpleMap()

		trees[treeType] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return trees, nil
}

func loadMerkleTrees(map[MerkleTree]*smt.SparseMerkleTree, error) {
	log.Fatalf("loadMerkleTrees not implemented yet")
}

func (p *PostgresContext) updateStateHash() error {
	// Update all the merkle trees
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
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
			log.Println("TODO: merkle tree not implemented", treeType)
		default:
			log.Fatalln("Not handled yet in state commitment update", treeType)
		}
	}

	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		roots = append(roots, p.merkleTrees[treeType].Root())
	}

	// Currently using the order of the merkle tree definition (i.e. defined in code)
	// Sort the merkle roots lexicographically
	// sort.Slice(roots, func(r1, r2 int) bool {
	// 	return bytes.Compare(roots[r1], roots[r2]) < 0
	// })

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	p.currentStateHash = stateHash[:] //[]byte(fmt.Sprintf("%x", ))
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
			GenericParam:    actor.ActorSpecificParam,
			StakedAmount:    actor.StakedTokens,
			PausedHeight:    actor.PausedHeight,
			UnstakingHeight: actor.UnstakingHeight,
			Output:          actor.OutputAddress,
		}
		actors = append(actors, actor)
	}
	return
}
