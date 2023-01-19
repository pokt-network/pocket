package persistence

import (
	"crypto/sha256"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

// A list of functions to clear data from the DB not associated with protocol actors
var nonActorClearFunctions = []func() string{
	types.ClearAllAccounts,
	types.ClearAllPools,
	types.ClearAllGovParamsQuery,
	types.ClearAllGovFlagsQuery,
	types.ClearAllBlocksQuery,
}

func (m *persistenceModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	// Clears all the state (SQL DB, KV Stores, Trees, etc) to nothing
	case messaging.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE:
		if err := m.clearAllState(debugMessage); err != nil {
			return err
		}
	// Resets all the state (SQL DB, KV Stores, Trees, etc) to the tate specified in the genesis file provided
	case messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS:
		if err := m.clearAllState(debugMessage); err != nil {
			return err
		}
		g := m.genesisState
		m.populateGenesisState(g) // fatal if there's an error
	default:
		m.logger.Debug().Str("message", debugMessage.Message.String()).Msg("Debug message not handled by persistence module")
	}
	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *messaging.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		m.logger.Error().Err(err).Uint64("height", height).Msg("Error getting block from block store")
		return
	}

	block := &coreTypes.Block{}
	if err := codec.GetCodec().Unmarshal(blockBytes, block); err != nil {
		m.logger.Error().Err(err).Uint64("height", height).Msg("Error decoding block from block store")
		return
	}

	m.logger.Info().Uint64("height", height).Str("block", block.String()).Msg("Block from block store")
}

// TECHDEBT: Make sure this is atomic
func (m *persistenceModule) clearAllState(_ *messaging.DebugMessage) error {
	ctx, err := m.NewRWContext(-1)
	if err != nil {
		return err
	}
	postgresCtx := ctx.(*PostgresContext)

	// Clear the SQL DB
	if err := postgresCtx.clearAllSQLState(); err != nil {
		return err
	}

	// Release the SQL context
	if err := m.ReleaseWriteContext(); err != nil {
		return err
	}

	// Clear the BlockStore
	if err := m.blockStore.ClearAll(); err != nil {
		return err
	}

	// Clear all the Trees
	if err := postgresCtx.clearAllTreeState(); err != nil {
		return err
	}

	m.logger.Info().Msg("Cleared all the state")
	return nil
}

func (p *PostgresContext) clearAllSQLState() error {
	ctx, clearTx, err := p.getCtxAndTx()
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

func (p *PostgresContext) clearAllTreeState() error {
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		valueStore := p.stateTrees.valueStores[treeType]
		nodeStore := p.stateTrees.nodeStores[treeType]

		if err := valueStore.ClearAll(); err != nil {
			return err
		}
		if err := nodeStore.ClearAll(); err != nil {
			return err
		}

		// Needed in order to make sure the root is re-set correctly after clearing
		p.stateTrees.merkleTrees[treeType] = smt.NewSparseMerkleTree(valueStore, nodeStore, sha256.New())
	}

	return nil
}
