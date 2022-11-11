package persistence

import (
	"crypto/sha256"
	"log"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/debug"
)

// A list of functions to clear data from the DB not associated with protocol actors
var nonActorClearFunctions = []func() string{
	types.ClearAllAccounts,
	types.ClearAllPools,
	types.ClearAllGovParamsQuery,
	types.ClearAllGovFlagsQuery,
	types.ClearAllBlocksQuery,
}

func (m *persistenceModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	// Clears all the state (SQL DB, KV Stores, Trees, etc) to nothing
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE:
		if err := m.clearAllState(debugMessage); err != nil {
			return err
		}
	// Clears all the state (SQL DB, KV Stores, Trees, etc) to the tate specified in the genesis file provided
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS:
		if err := m.clearAllState(debugMessage); err != nil {
			return err
		}
		g := m.genesisState.(*types.PersistenceGenesisState)
		m.populateGenesisState(g) // fatal if there's an error
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *debug.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight()
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	block := &types.Block{}
	codec.GetCodec().Unmarshal(blockBytes, block)

	log.Printf("Block at height %d: %+v \n", height, block)
}

// TODO: Make sure this is atomic
func (m *persistenceModule) clearAllState(_ *debug.DebugMessage) error {
	context, err := m.NewRWContext(-1)
	if err != nil {
		return err
	}

	// Clear the SQL DB
	if err := context.(*PostgresContext).clearAllSQLState(); err != nil {
		return err
	}
	if err := m.ReleaseWriteContext(); err != nil {
		return err
	}

	// Clear the KV Stores
	if err := m.blockStore.ClearAll(); err != nil {
		return err
	}
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		valueStore := m.stateTrees.valueStores[treeType]
		nodeStore := m.stateTrees.nodeStores[treeType]

		if err := valueStore.ClearAll(); err != nil {
			return err
		}
		if err := nodeStore.ClearAll(); err != nil {
			return err
		}

		// Needed in order to make sure the root is re-set correctly after clearing
		m.stateTrees.merkleTrees[treeType] = smt.NewSparseMerkleTree(valueStore, nodeStore, sha256.New())
	}

	log.Println("Cleared all the state")
	return nil
}

func (p *PostgresContext) clearAllSQLState() error {
	ctx, clearTx, err := p.GetCtxAndTx()
	if err != nil {
		return err
	}

	for _, actor := range protocolActorSchemas {
		if _, err = clearTx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if actor.GetChainsTableName() != "" {
			if _, err = clearTx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
				return err
			}
		}
	}

	for _, clearFn := range nonActorClearFunctions {
		if _, err := clearTx.Exec(ctx, clearFn()); err != nil {
			return err
		}
	}

	if err = clearTx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
