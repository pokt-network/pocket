package persistence

import (
	"log"

	"github.com/pokt-network/pocket/shared/types"
)

func (m *persistenceModule) HandleDebugMessage(debugMessage *types.DebugMessage) error {
	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case types.DebugMessageAction_DEBUG_CLEAR_STATE:
		m.clearState(debugMessage)
		// TODO_IN_THIS_COMMIT: Figure this out
		m.populateGenesisState(m.GetBus().GetGenesis())
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *types.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1 // -1 because we want the latest committed height
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := types.GetCodec()
	block := &types.Block{}
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d with %d transactions: %+v \n", height, len(block.Transactions), block)
}

func (m *persistenceModule) clearState(_ *types.DebugMessage) {
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
