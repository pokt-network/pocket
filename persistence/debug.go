package persistence

import (
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
)

func (m *persistenceModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CLEAR_STATE:
		m.clearState(debugMessage)
		g := m.genesisState.(*types.PersistenceGenesisState)
		m.populateGenesisState(g)
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

// TODO(olshansky): Create a shared interface `Block` to avoid the use of typesCons here.
func (m *persistenceModule) showLatestBlockInStore(_ *messaging.DebugMessage) {
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

func (m *persistenceModule) clearState(_ *messaging.DebugMessage) {
	context, err := m.NewRWContext(-1)
	if err != nil {
		log.Printf("Error creating new context: %s \n", err)
		return
	}
	defer context.Commit()

	if err := context.(*PostgresContext).DebugClearAll(); err != nil {
		log.Printf("Error clearing state: %s \n", err)
		return
	}
	if err := m.blockStore.ClearAll(); err != nil {
		log.Printf("Error clearing block store: %s \n", err)
		return
	}
}
