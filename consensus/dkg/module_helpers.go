//go:build test
// +build test

package dkg

import (
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/types/known/anypb"
)

func (module *dkgModule) broadcastToNodes(message *DKGMessage) {
	event := types.Event{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.CONSENSUS),
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) sendToNode(message *DKGMessage, destNode *types_consensus.NodeId) {
	event := types.Event{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.CONSENSUS),
		Destination:  *destNode,
	}
	module.publishEvent(message, &event)
}

func (module *dkgModule) publishEvent(message *DKGMessage, event *types.Event) {
	consensusMessage := &types_consensus.ConsensusMessage{
		Message: message,
		Sender:  module.NodeId,
	}

	data, err := types_consensus.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	consensusProtoMsg := &types_consensus.Message{
		Data: data,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		log.Println("[ERROR] Error encoding message: " + err.Error())
		return
	}

	//networkProtoMsg := &types2.Message{
	//	Topic: types2.PocketTopic_CONSENSUS.String(),
	//	Data:  anyProto,
	//}

	if err := module.GetBus().GetNetworkModule().BroadcastMessage(anyProto, event.PocketTopic); err != nil {
		// TODO handle
		return
	}

	//networkMsg := &p2p_types.Message{
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
