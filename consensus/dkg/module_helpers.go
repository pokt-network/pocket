package dkg

import (
	"log"
	consensus_types "pocket/consensus/types"
	"pocket/shared/types"

	"google.golang.org/protobuf/types/known/anypb"
)

func (module *dkgModule) broadcastToNodes(message *DKGMessage) {
	event := types.PocketEvent{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.P2P_BROADCAST_MESSAGE),
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) sendToNode(message *DKGMessage, destNode *consensus_types.NodeId) {
	event := types.PocketEvent{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.P2P_SEND_MESSAGE),
		Destination:  *destNode,
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) publishEvent(message *DKGMessage, event *types.PocketEvent) {
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  module.NodeId,
	}

	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	consensusProtoMsg := &types.ConsensusMessage{
		Data: data,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	networkProtoMsg := &types.NetworkMessage{
		Topic: types.PocketTopic_CONSENSUS.String(),
		Data:  anyProto,
	}

	module.GetBus().GetNetworkModule().BroadcastMessage(networkProtoMsg)

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
	//module.GetBus().PublishEventToBus(event)
}
