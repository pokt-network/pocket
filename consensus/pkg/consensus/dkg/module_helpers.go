package dkg

import (
	"log"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	"pocket/shared/events"
	"pocket/shared/messages"

	"google.golang.org/protobuf/types/known/anypb"
)

func (module *dkgModule) broadcastToNodes(message *DKGMessage) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS_MODULE,
		PocketTopic:  string(events.P2P_BROADCAST_MESSAGE),
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) sendToNode(message *DKGMessage, destNode *types.NodeId) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS_MODULE,
		PocketTopic:  string(events.P2P_SEND_MESSAGE),
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

	consensusProtoMsg := &messages.ConsensusMessage{
		Data: data,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	networkProtoMsg := &messages.NetworkMessage{
		Topic: messages.PocketTopic_CONSENSUS.String(),
		Data:  anyProto,
	}

	module.GetPocketBusMod().GetNetworkModule().BroadcastMessage(networkProtoMsg)

	//networkMsg := &p2p_types.NetworkMessage{
	//	Topic: events.CONSENSUS,
	//	Data:  data,
	//}
	//
	//networkMsgEncoded, err := p2p.EncodeNetworkMessage(networkMsg)
	//if err != nil {
	//	log.Println("Error encoding network message: " + err.Error())
	//	return
	//}
	//
	//event.MessageData = networkMsgEncoded
	//module.GetPocketBusMod().PublishEventToBus(event)
}
