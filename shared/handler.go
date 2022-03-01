package shared

import (
	"log"

	"github.com/pokt-network/pocket/shared/types"
)

func (node *Node) handleEvent(event *types.Event) error {
	switch event.PocketTopic {
	case types.ConsensusMessage:
		node.GetBus().GetConsensusModule().HandleMessage(event.MessageData)
	default:
		// TODO(discuss): Should we panic here?
		log.Printf("[WARN] Unsupported event: %s \n", event.PocketTopic)
	}
	return nil
}
