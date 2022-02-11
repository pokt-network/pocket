package consensus

import (
	"log"

	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/types"
	"pocket/shared/events"
)

func (m *consensusModule) broadcastToNodes(message consensus_types.GenericConsensusMessage) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS,
		PocketTopic:  events.P2P_BROADCAST_MESSAGE,
	}
	m.publishConsensusEvent(message, &event)
}

func (m *consensusModule) sendToNode(message consensus_types.GenericConsensusMessage, destNode *types.NodeId) {
	event := events.PocketEvent{
		SourceModule: events.CONSENSUS,
		PocketTopic:  events.P2P_SEND_MESSAGE,
		Destination:  *destNode,
	}
	m.publishConsensusEvent(message, &event)
}

func (m *consensusModule) publishConsensusEvent(message consensus_types.GenericConsensusMessage, event *events.PocketEvent) {
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  m.NodeId,
	}
	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	if err != nil {
		m.nodeLogError("Error encoding message: " + err.Error())
		return
	}

	//fmt.Println("Can't publish yet.", data)
	if err := m.GetPocketBusMod().GetNetworkModule().ConsensusBroadcast(data); err != nil {
		m.nodeLogError("Error broadcasting message: " + err.Error())
		return
	}
}

// TODO: Move this into persistence.
func (m *consensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.MessagePool[step] = make([]HotstuffMessage, 0)
	}
}

func (m *consensusModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *consensusModule) nodeLogError(s string) {
	log.Printf("[ERROR][%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *consensusModule) isLeader() bool {
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}
