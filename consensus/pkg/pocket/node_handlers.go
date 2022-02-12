package pocket

import (
	"context"
	"log"

	pcontext "pocket/shared/context"
	"pocket/shared/events"
)

func (node *PocketNode) handleEvent(event *events.PocketEvent) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pocketContext := &pcontext.PocketContext{
		Ctx: ctx,
		Handler: func(...interface{}) (interface{}, error) {
			log.Println("[DEBUG]Handling event: ")
			return nil, nil
		},
	}

	switch event.PocketTopic {

	//case events.CONSENSUS_TELEMETRY_MESSAGE:
	//	node.ConsensusMod.HandleTelemetryMessage(pocketContext, event.NetworkConnection)

	case string(events.CONSENSUS):
		node.GetPocketBusMod().GetConsensusModule().HandleMessage(pocketContext, event.MessageData)

	case string(events.UTILITY_TX_MESSAGE):
		node.GetPocketBusMod().GetConsensusModule().HandleTransaction(pocketContext, event.MessageData)
		// node.GetPocketBusMod().GetConsensusModule().HandleTransaction(pocketContext, event.MessageData)

	// case string(events.UTILITY_EVIDENCE_MESSAGE):
	// 	node.GetPocketBusMod().GetConsensusModule().HandleEvidence(pocketContext, event.MessageData)

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
