package shared

import (
	"log"
	"pocket/shared/types"
)

func (node *Node) handleEvent(event *types.Event) error {
	// TODO: Only supporting a subset of topics because not all are used.
	switch event.PocketTopic {

	case string(types.CONSENSUS):
		node.GetBus().GetConsensusModule().HandleMessage(event.MessageData)

	case string(types.UTILITY_TX_MESSAGE):
		node.GetBus().GetConsensusModule().HandleTransaction(event.MessageData)

	default:
		log.Printf("Unsupported event: %s \n", event.PocketTopic)

	}
	return nil
}
