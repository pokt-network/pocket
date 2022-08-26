package persistence

import (
	"encoding/json"
	types2 "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/debug"
	"log"
)

func (m *persistenceModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case debug.DebugMessageAction_DEBUG_CLEAR_STATE:
		m.clearState(debugMessage)

		var persistenceGenesisState *types.PersistenceGenesisState
		genBz := m.GetBus().GetGenesis()["Persistence"]
		if err := json.Unmarshal(genBz, persistenceGenesisState); err != nil {
			return err
		}
		m.populateGenesisState(persistenceGenesisState)
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *debug.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1 // -1 because we want the latest committed height
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := codec.GetCodec()
	block := &types2.Block{} // TODO in_this_commit PREVENT THIS IMPORT
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d with %d transactions: %+v \n", height, len(block.Transactions), block)
}

func (m *persistenceModule) clearState(_ *debug.DebugMessage) {
	context, err := m.NewRWContext(-1)
	defer context.Commit()
	if err != nil {
		log.Printf("Error creating new context: %s \n", err)
		return
	}
	if err := context.(PostgresContext).DebugClearAll(); err != nil {
		log.Printf("Error clearing state: %s \n", err)
		return
	}
	if err := m.blockStore.ClearAll(); err != nil {
		log.Printf("Error clearing block store: %s \n", err)
		return
	}
}
