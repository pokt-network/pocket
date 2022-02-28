package shared

import (
	"log"
	"pocket/shared/types"
)

func (node *Node) handleEvent(event *types.PocketEvent) error {
	switch event.Topic {
	case types.PocketTopic_CONSENSUS_MESSAGE_TOPIC:
		node.GetBus().GetConsensusModule().HandleMessage(event.Data)
	default:
		log.Printf("Unsupported event: %s \n", event.Topic)

	}
	return nil
}
