package shared

import (
	"fmt"
	"log"
	"pocket/shared/types"
)

func (node *Node) handleEvent(event *types.Event) error {
	switch event.PocketTopic {

	//case events.CONSENSUS_TELEMETRY_MESSAGE:
	//	node.ConsensusMod.HandleTelemetryMessage(pocketContext, event.NetworkConnection)

	case string(types.CONSENSUS):
		fmt.Println(event.MessageData)
		node.GetBus().GetConsensusModule().HandleMessage(event.MessageData)

	// TODO(Andrew): This is where the broadcasted utility message will be forwarded to the consensus module.
	case string(types.UTILITY_TX_MESSAGE):
		node.GetBus().GetConsensusModule().HandleTransaction(event.MessageData)
		// node.GetBus().GetConsensusModule().HandleTransaction(pocketContext, event.MessageData)

	// case string(events.UTILITY_EVIDENCE_MESSAGE):
	// 	node.GetBus().GetConsensusModule().HandleEvidence(pocketContext, event.MessageData)

	//case events.P2P_BROADCAST_MESSAGE:
	//	message, err := p2p.DecodeNetworkMessage(event.MessageData)
	//	if err != nil {
	//		return err
	//	}
	//	node.NetworkMod.Broadcast(pocketContext, message)
	//
	//case events.P2P_SEND_MESSAGE:
	//	message, err := p2p.DecodeNetworkMessage(event.MessageData)
	//	if err != nil {
	//		return err
	//	}
	//	node.NetworkMod.Send(pocketContext, message, event.Destination)

	default:
		log.Printf("Unsupported event: %s \n", event.PocketTopic)

	}
	return nil
}
