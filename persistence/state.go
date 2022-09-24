package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sort"

	"github.com/celestiaorg/smt"
	typesUtil "github.com/pokt-network/pocket/utility/types"
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
	accountMerkleTree
	poolMerkleTree

	// Data / State Merkle Trees
	blocksMerkleTree
	paramsMerkleTree
	flagsMerkleTree

	// Used for iteration purposes only - see https://stackoverflow.com/a/64178235/768439
	lastMerkleTree
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

// Question: Is this the right approach?
func (p *PostgresContext) updateStateHash() error {
	// Update all the merkle trees
	for treeType := MerkleTree(0); treeType < lastMerkleTree; treeType++ {
		switch treeType {
		case appMerkleTree:
			apps, err := p.getApplicationsUpdatedAtHeight(p.Height)
			if err != nil {
				// TODO_IN_THIS_COMMIT: Update this error
				return typesUtil.NewError(typesUtil.Code(42), "Couldn't figure out apps updated")
			}
			for _, app := range apps {
				appBz, err := proto.Marshal(app)
				if err != nil {
					return err
				}
				// An update results in a create/update that is idempotent
				addrBz, err := hex.DecodeString(app.Address)
				if err != nil {
					return err
				}
				if _, err := p.MerkleTrees[treeType].Update(addrBz, appBz); err != nil {
					return err
				}
			}
		// TODO_IN_THIS_COMMIT: re
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

	p.stateHash = stateHash[:]
	return nil
}
