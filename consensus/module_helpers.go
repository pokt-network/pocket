package consensus

import (
	"fmt"
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

/*** P2P Helpers ***/

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

	if err := m.GetBus().GetP2PModule().Send(crypto.AddressFromString(m.IdToValAddrMap[destNode]), any, types.PocketTopic_CONSENSUS_MESSAGE_TOPIC); err != nil {
		m.nodeLogError("Error broadcasting message:", err)
		return
	}
}

/*** Persistence Helpers ***/

func (m *consensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.MessagePool[step] = make([]*types_consensus.HotstuffMessage, 0)
	}
}

/*** Leader Election Helpers ***/

func (m *consensusModule) isLeader() bool {
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}

func (m *consensusModule) clearLeader() {
	m.logPrefix = DefaultLogPrefix
	m.LeaderId = nil
}

func (m *consensusModule) electNextLeader(message *types_consensus.HotstuffMessage) {
	leaderId := m.electNextLeaderDeterministic(message)

	if leaderId == 0 {
		m.nodeLogError(fmt.Sprintf("Leader election failed. Validator cannot take part in consensus at height %d round %d", message.Height, message.Round), nil)
		m.clearLeader()
		return
	}

	m.LeaderId = &leaderId

	if m.LeaderId != nil && *m.LeaderId == m.NodeId {
		m.logPrefix = "LEADER"
		m.nodeLog(fmt.Sprintf("👑👑👑👑👑   %d   👑👑👑👑👑", m.NodeId))
	} else {
		m.logPrefix = "REPLICA"
		m.nodeLog(fmt.Sprintf("Elected %d as 👑.", *m.LeaderId))
	}
}

func (m *consensusModule) electNextLeaderDeterministic(message *types_consensus.HotstuffMessage) types_consensus.NodeId {
	valMap := types.GetTestState(nil).ValidatorMap
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1
	return types_consensus.NodeId(value%int64(len(valMap)) + 1)
}

/*** General Infrastructure Helpers ***/

func (m *consensusModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *consensusModule) nodeLogError(s string, err error) {
	log.Printf("[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.NodeId, s, err)
}
