package persistence

import (
	"log"

	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/debug"
)

func (m *persistenceModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case debug.DebugMessageAction_DEBUG_CLEAR_STATE:
		if err := m.ClearState(debugMessage); err != nil {
			log.Fatalf("Error clearing state: %s \n", err)
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
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1 // -1 because we want the latest committed height
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := codec.GetCodec()
	block := &types.Block{}
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d with %d transactions: %+v \n", height, len(block.Transactions), block)
}

func (m *persistenceModule) ClearState(_ *debug.DebugMessage) error {
	context, err := m.NewRWContext(-1)
	if err != nil {
		return err
	}

	if err := context.(*PostgresContext).DebugClearAll(); err != nil {
		return err
	}

	if err := m.blockStore.ClearAll(); err != nil {
		return err
	}

	if err := m.ReleaseWriteContext(); err != nil {
		return err
	}

	return nil
}
