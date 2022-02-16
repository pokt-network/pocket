package consensus

import (
	"log"
	consensus_types "pocket/consensus/types"
	"pocket/shared/types"
	"strconv"

	"google.golang.org/protobuf/types/known/anypb"
)

func (m *ConsensusModule) broadcastToNodes(message consensus_types.GenericConsensusMessage) {
	event := types.PocketEvent{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.P2P_BROADCAST_MESSAGE),
	}
	networkProtoMsg := m.getConsensusNetworkMessage(message, &event)
	m.GetBus().GetNetworkModule().BroadcastMessage(networkProtoMsg)
}

func (m *ConsensusModule) sendToNode(message consensus_types.GenericConsensusMessage, destNode *consensus_types.NodeId) {
	event := types.PocketEvent{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.P2P_SEND_MESSAGE),
		Destination:  *destNode,
	}
	// PreP2P hack
	addr := strconv.Itoa(int(*destNode))
	networkProtoMsg := m.getConsensusNetworkMessage(message, &event)
	m.GetBus().GetNetworkModule().Send(addr, networkProtoMsg)
}

func (m *ConsensusModule) getConsensusNetworkMessage(message consensus_types.GenericConsensusMessage, event *types.PocketEvent) *types.NetworkMessage {
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  m.NodeId,
	}

	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		m.nodeLogError("Error encoding message: " + err.Error())
		return nil
	}

	consensusProtoMsg := &types.ConsensusMessage{
		Data: data,
	}

	anyProto, err := anypb.New(consensusProtoMsg)
	if err != nil {
		m.nodeLogError("Error encoding any proto: " + err.Error())
		return nil
	}

	networkProtoMsg := &types.NetworkMessage{
		Topic: types.PocketTopic_CONSENSUS,
		Data:  anyProto,
	}
	return networkProtoMsg
}

// TODO: Move this into persistence.
func (m *ConsensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.MessagePool[step] = make([]HotstuffMessage, 0)
	}
}

func (m *ConsensusModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *ConsensusModule) nodeLogError(s string) {
	log.Printf("[ERROR][%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *ConsensusModule) isLeader() bool {
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}
