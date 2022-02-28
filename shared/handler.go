package shared

import (
	"log"
	"pocket/shared/types"
)

// TODO: Only supporting a subset of topics because not all are used.
func (node *Node) handleEvent(event *types.Event) error {
	switch event.PocketTopic {
	case types.CONSENSUS_MESSAGE:
		node.GetBus().GetConsensusModule().HandleMessage(event.MessageData)
	default:
		log.Printf("Unsupported event: %s \n", event.PocketTopic)

	}
	return nil
}
