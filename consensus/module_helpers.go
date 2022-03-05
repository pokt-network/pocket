package consensus

import (
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) broadcastToNodes(message proto.Message, messageType types_consensus.ConsensusMessageType) {
	anyConsensusMessage, err := anypb.New(message)
	if err != nil {
		m.nodeLogError("Failed to create inner consensus message", err)
		return
	}

	consensusMessage := &types_consensus.ConsensusMessage{
		Type:    messageType,
		Message: anyConsensusMessage,
	}

	any, err := anypb.New(consensusMessage)
	if err != nil {
		m.nodeLogError("Failed to create consensus message", err)
		return
	}

	if err := m.GetBus().GetP2PModule().Broadcast(any, types.PocketTopic_CONSENSUS_MESSAGE_TOPIC); err != nil {
		m.nodeLogError("Error broadcasting message:", err)
		return
	}
}

func (m *consensusModule) sendToNode(message proto.Message, messageType types_consensus.ConsensusMessageType, destNode types_consensus.NodeId) {
	anyConsensusMessage, err := anypb.New(message)
	if err != nil {
		m.nodeLogError("Failed to create inner consensus message", err)
		return
	}

	consensusMessage := &types_consensus.ConsensusMessage{
		Type:    messageType,
		Message: anyConsensusMessage,
	}

	any, err := anypb.New(consensusMessage)
	if err != nil {
		m.nodeLogError("Failed to create consensus message", err)
		return
	}

	if err := m.GetBus().GetP2PModule().Send([]byte(m.IdToValMap[destNode]), any, types.PocketTopic_CONSENSUS_MESSAGE_TOPIC); err != nil {
		m.nodeLogError("Error broadcasting message:", err)
		return
	}
}

// TODO(olshansky): Move this into persistence.
func (m *consensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.MessagePool[step] = make([]types_consensus.HotstuffMessage, 0)
	}
}

func (m *consensusModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *consensusModule) nodeLogError(s string, err error) {
	log.Printf("[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.NodeId, s, err)
}
