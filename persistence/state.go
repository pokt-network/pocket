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
	// Actors
	AppMerkleTree MerkleTree = iota
	ValMerkleTree
	FishMerkleTree
	ServiceNodeMerkleTree
	AccountMerkleTree
	PoolMerkleTree
	// Data / state
	BlocksMerkleTree
	ParamsMerkleTree
	FlagsMerkleTree
	lastMerkleTree // Used for iteration purposes only - see https://stackoverflow.com/a/64178235/768439
)

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

// DISCUSS_IN_THIS_COMMIT(drewskey): Thoughts on this approach?
// 1. Retrieves all of the actors / data types updated at the current height
// 2. Updates the Merkle Tree associated with each actor / data type
//    - This operation is idempotent so you can call `updateStateHash` as often as you want
// 3. Update the context's "cached" state hash
// 4. Returns the state hash
func (p *PostgresContext) updateStateHash() ([]byte, error) {
	// Update all the merkle trees
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		switch treeType {
		case AppMerkleTree:
			apps, err := p.getAppsUpdated(p.Height)
			if err != nil {
				return nil, types.NewError(types.Code(42), "Couldn't figure out apps updated") // TODO_IN_THIS_COMMIT
			}
			for _, app := range apps {
				appBytes, err := proto.Marshal(app)
				if err != nil {
					return nil, err
				}
				// An update results in a create/update that is idempotent
				if _, err := p.MerkleTrees[treeType].Update(app.Address, appBytes); err != nil {
					return nil, err
				}
				// TODO_IN_THIS_COMMIT: Add support for `Delete` operations to remove it from the tree
			}
		default:
			log.Fatalln("Not handled yet in state commitment update", treeType)
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

	p.StateHash = stateHash[:]
	return p.StateHash, nil
}
