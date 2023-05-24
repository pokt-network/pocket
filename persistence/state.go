package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/smt"
)

type merkleTree float64

type stateTrees struct {
	merkleTrees map[merkleTree]*smt.SMT

	// nodeStores are part of the SMT, but references are kept below for convenience
	// and debugging purposes
	nodeStores map[merkleTree]kvstore.KVStore
}

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

var actorTypeToSchemaName = map[coreTypes.ActorType]types.ProtocolActorSchema{
	coreTypes.ActorType_ACTOR_TYPE_APP:      types.ApplicationActor,
	coreTypes.ActorType_ACTOR_TYPE_VAL:      types.ValidatorActor,
	coreTypes.ActorType_ACTOR_TYPE_FISH:     types.FishermanActor,
	coreTypes.ActorType_ACTOR_TYPE_SERVICER: types.ServicerActor,
}

var merkleTreeToActorTypeName = map[merkleTree]coreTypes.ActorType{
	appMerkleTree:      coreTypes.ActorType_ACTOR_TYPE_APP,
	valMerkleTree:      coreTypes.ActorType_ACTOR_TYPE_VAL,
	fishMerkleTree:     coreTypes.ActorType_ACTOR_TYPE_FISH,
	servicerMerkleTree: coreTypes.ActorType_ACTOR_TYPE_SERVICER,
}

func newStateTrees(treesStoreDir string) (*stateTrees, error) {
	if treesStoreDir == "" {
		return newMemStateTrees()
	}

	stateTrees := &stateTrees{
		merkleTrees: make(map[merkleTree]*smt.SMT, int(numMerkleTrees)),
		nodeStores:  make(map[merkleTree]kvstore.KVStore, int(numMerkleTrees)),
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

func newMemStateTrees() (*stateTrees, error) {
	stateTrees := &stateTrees{
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

func (p *PostgresContext) updateMerkleTrees() (string, error) {
	// Update all the merkle trees
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		switch treeType {
		// Actor Merkle Trees
		case appMerkleTree, valMerkleTree, fishMerkleTree, servicerMerkleTree:
			actorType, ok := merkleTreeToActorTypeName[treeType]
			if !ok {
				return "", fmt.Errorf("no actor type found for merkle tree: %v", treeType)
			}
			if err := p.updateActorsTree(actorType); err != nil {
				return "", fmt.Errorf("failed to update actors tree for treeType: %v, actorType: %v - %w", treeType, actorType, err)
			}

		// Account Merkle Trees
		case accountMerkleTree:
			if err := p.updateAccountTrees(); err != nil {
				return "", fmt.Errorf("failed to update account trees - %w", err)
			}
		case poolMerkleTree:
			if err := p.updatePoolTrees(); err != nil {
				return "", fmt.Errorf("failed to update pool trees - %w", err)
			}

		// Data Merkle Trees
		case transactionsMerkleTree:
			if err := p.updateTransactionsTree(); err != nil {
				return "", fmt.Errorf("failed to update transactions tree - %w", err)
			}
		case paramsMerkleTree:
			if err := p.updateParamsTree(); err != nil {
				return "", fmt.Errorf("failed to update params tree - %w", err)
			}
		case flagsMerkleTree:
			if err := p.updateFlagsTree(); err != nil {
				return "", fmt.Errorf("failed to update flags tree - %w", err)
			}

		// Default
		default:
			p.logger.Fatal().Msgf("Not handled yet in state commitment update. Merkle tree #{%v}", treeType)
		}
	}

	return p.getStateHash(), nil
}

func (p *PostgresContext) getStateHash() string {
	// Get the root of each Merkle Tree
	roots := make([][]byte, 0)
	for tree := merkleTree(0); tree < numMerkleTrees; tree++ {
		roots = append(roots, p.stateTrees.merkleTrees[tree].Root())
	}

	// Get the state hash
	rootsConcat := bytes.Join(roots, []byte{})
	stateHash := sha256.Sum256(rootsConcat)

	// Convert the array to a slice and return it
	return hex.EncodeToString(stateHash[:])
}

// Actor Tree Helpers

func (p *PostgresContext) updateActorsTree(actorType coreTypes.ActorType) error {
	actors, err := p.getActorsUpdatedAtHeight(actorType, p.Height)
	if err != nil {
		return err
	}

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
		if err := p.stateTrees.merkleTrees[merkleTreeName].Update(bzAddr, actorBz); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) getActorsUpdatedAtHeight(actorType coreTypes.ActorType, height int64) (actors []*coreTypes.Actor, err error) {
	actorSchema, ok := actorTypeToSchemaName[actorType]
	if !ok {
		return nil, fmt.Errorf("no schema found for actor type: %s", actorType)
	}

	schemaActors, err := p.GetActorsUpdated(actorSchema, height)
	if err != nil {
		return nil, err
	}

	actors = make([]*coreTypes.Actor, len(schemaActors))
	for i, schemaActor := range schemaActors {
		actor := &coreTypes.Actor{
			ActorType:       actorType,
			Address:         schemaActor.Address,
			PublicKey:       schemaActor.PublicKey,
			Chains:          schemaActor.Chains,
			ServiceUrl:      schemaActor.ServiceUrl,
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

		accBz, err := codec.GetCodec().Marshal(account)
		if err != nil {
			return err
		}

		if err := p.stateTrees.merkleTrees[accountMerkleTree].Update(bzAddr, accBz); err != nil {
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
		bzAddr, err := hex.DecodeString(pool.GetAddress())
		if err != nil {
			return err
		}

		accBz, err := codec.GetCodec().Marshal(pool)
		if err != nil {
			return err
		}

		if err := p.stateTrees.merkleTrees[poolMerkleTree].Update(bzAddr, accBz); err != nil {
			return err
		}
	}

	return nil
}

// Data Tree Helpers

func (p *PostgresContext) updateTransactionsTree() error {
	indexedTxs, err := p.txIndexer.GetByHeight(p.Height, false)
	if err != nil {
		return err
	}

	for _, idxTx := range indexedTxs {
		txBz := idxTx.GetTx()
		txHash := crypto.SHA3Hash(txBz)
		if err := p.stateTrees.merkleTrees[transactionsMerkleTree].Update(txHash, txBz); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) updateParamsTree() error {
	params, err := p.getParamsUpdated(p.Height)
	if err != nil {
		return err
	}

	for _, param := range params {
		paramBz, err := codec.GetCodec().Marshal(param)
		paramKey := crypto.SHA3Hash([]byte(param.Name))
		if err != nil {
			return err
		}
		if err := p.stateTrees.merkleTrees[paramsMerkleTree].Update(paramKey, paramBz); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresContext) updateFlagsTree() error {
	flags, err := p.getFlagsUpdated(p.Height)
	if err != nil {
		return err
	}

	for _, flag := range flags {
		flagBz, err := codec.GetCodec().Marshal(flag)
		flagKey := crypto.SHA3Hash([]byte(flag.Name))
		if err != nil {
			return err
		}
		if err := p.stateTrees.merkleTrees[flagsMerkleTree].Update(flagKey, flagBz); err != nil {
			return err
		}
	}

	return nil
}
