package persistence

import (
	"bytes"
	"crypto/sha256"
	"log"
	"sort"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
)

type MerkleTree float64

// A work-in-progress list of all the trees we need to update to maintain the overall state
const (
	AppMerkleTree MerkleTree = iota
	ValMerkleTree
	FishMerkleTree
	ServiceNodeMerkleTree
	AccountMerkleTree
	PoolMerkleTree
	BlocksMerkleTree
	ParamsMerkleTree
	FlagsMerkleTree
	lastMerkleTree // Used for iteration purposes only - see https://stackoverflow.com/a/64178235/768439
)

func initializeTrees() (map[MerkleTree]*smt.SparseMerkleTree, error) {
	// We need a separate Merkle tree for each type of actor or storage
	trees := make(map[MerkleTree]*smt.SparseMerkleTree, int(lastMerkleTree))

	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		// Initialize two new key-value store to store the nodes and values of the tree
		nodeStore := smt.NewSimpleMap()
		valueStore := smt.NewSimpleMap()

		trees[treeType] = smt.NewSparseMerkleTree(nodeStore, valueStore, sha256.New())
	}
	return trees, nil
}

func loadTrees(map[MerkleTree]*smt.SparseMerkleTree, error) {

}

func (p *PostgresContext) updateStateCommitment() ([]byte, error) {
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		switch treeType {
		case AppMerkleTree:
			apps, err := p.getAppsUpdated(p.Height)
			if err != nil {
				return nil, types.NewError(types.Code(42), "Couldn't figure out apps updated") // TODO_IN_THIS_COMMIT
			}
			for _, app := range apps {
				// OPTIMIZE: Do we want to store the serialized bytes or a hash of it in the KV store?
				appBytes, err := proto.Marshal(app)
				if err != nil {
					return nil, err
				}
				if _, err := p.MerkleTrees[treeType].Update(app.Address, appBytes); err != nil {
					return nil, err
				}
			}
		default:
			log.Fatalln("Not handeled uet in state commitment update")
		}
	}

	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		roots = append(roots, p.MerkleTrees[treeType].Root())
	}

	// Sort the merkle roots lexicographically
	sort.Slice(roots, func(r1, r2 int) bool {
		return bytes.Compare(roots[r1], roots[r2]) < 0
	})

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	return stateHash[:], nil
}

// computeStateHash(root)
// context := p.
// Update the Merkle Tree associated with each actor
// for _, actorType := range typesUtil.ActorTypes {
// 	// Need to get all the actors updated at this height
// 	switch actorType {
// 	case typesUtil.ActorType_App:
// 		apps, err := u.Context.GetAppsUpdated(u.LatestHeight) // shouldn't need to pass in a height here
// 		if err != nil {
// 			return types.NewError(types.Code(42), "Couldn't figure out apps updated")
// 		}
// 		if err := u.Context.UpdateAppTree(apps); err != nil {
// 			return nil
// 		}
// 	case typesUtil.ActorType_Val:
// 		fallthrough
// 	case typesUtil.ActorType_Fish:
// 		fallthrough
// 	case typesUtil.ActorType_Node:
// 		fallthrough
// 	default:
// 		log.Fatalf("Actor type not supported: %s", actorType)
// 	}
// }

// TODO: Update Merkle Tree for Accounts

// TODO: Update Merkle Tree for Pools

// TODO:Update Merkle Tree for Blocks

// TODO: Update Merkle Tree for Params

// TODO: Update Merkle Tree for Flags

// return nil
