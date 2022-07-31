package persistence

import (
	"log"

	"github.com/pokt-network/pocket/shared/types"
)

func (p *persistenceModule) HandleDebugMessage(debugMessage *types.DebugMessage) error {
	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		p.showLatestBlockInStore(debugMessage)
	case types.DebugMessageAction_DEBUG_CLEAR_STATE:
		p.clearState(debugMessage)
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

func (p *persistenceModule) showLatestBlockInStore(_ *types.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := p.GetBus().GetConsensusModule().CurrentHeight() - 1 // -1 because we want the latest committed height
	blockBytes, err := p.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := types.GetCodec()
	block := &types.Block{}
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d with %d transactions: %+v \n", height, len(block.Transactions), block)
}

func (p *persistenceModule) clearState(_ *types.DebugMessage) {
	context, err := p.NewContext(-1)
	if err != nil {
		log.Printf("Error creating new context: %s \n", err)
		return
	}
	if err := context.(PostgresContext).ClearAllDebug(); err != nil {
		log.Printf("Error clearing state: %s \n", err)
		return
	}
	if err := p.blockStore.ClearAll(); err != nil {
		log.Printf("Error clearing block store: %s \n", err)
		return
	}
}
