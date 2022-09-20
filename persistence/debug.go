package persistence

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/debug"
)

func (m *PersistenceModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case debug.DebugMessageAction_DEBUG_CLEAR_STATE:
		m.clearState(debugMessage)
		g, err := runtime.ParseGenesisJSON(m.genesisPath)
		if err != nil {
			return err
		}
		m.populateGenesisState(g.PersistenceGenesisState)
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

// TODO(olshansky): Create a shared interface `Block` to avoid the use of typesCons here.
func (m *PersistenceModule) showLatestBlockInStore(_ *debug.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1 // -1 because we want the latest committed height
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := codec.GetCodec()
	block := &typesCons.Block{}
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d with %d transactions: %+v \n", height, len(block.Transactions), block)
}

func (m *PersistenceModule) clearState(_ *debug.DebugMessage) {
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
