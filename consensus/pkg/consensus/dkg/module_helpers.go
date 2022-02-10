package dkg

import (
	"log"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/p2p"
	"pocket/consensus/pkg/p2p/p2p_types"
	"pocket/consensus/pkg/shared/events"
	"pocket/consensus/pkg/types"
)

func (module *dkgModule) broadcastToNodes(message *DKGMessage) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS,
		PocketTopic:  events.P2P_BROADCAST_MESSAGE,
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) sendToNode(message *DKGMessage, destNode *types.NodeId) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS,
		PocketTopic:  events.P2P_SEND_MESSAGE,
		Destination:  *destNode,
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) publishEvent(message *DKGMessage, event *events.PocketEvent) {
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  module.NodeId,
	}
	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	networkMsg := &p2p_types.NetworkMessage{
		Topic: events.CONSENSUS_MESSAGE,
		Data:  data,
	}

	networkMsgEncoded, err := p2p.EncodeNetworkMessage(networkMsg)
	if err != nil {
		log.Println("Error encoding network message: " + err.Error())
		return
	}

	event.MessageData = networkMsgEncoded
	module.GetPocketBusMod().PublishEventToBus(event)
}
