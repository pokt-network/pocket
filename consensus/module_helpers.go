package consensus

import (
	"log"
	"strconv"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/types/known/anypb"
)

func (m *consensusModule) broadcastToNodes(message types_consensus.GenericConsensusMessage) {
	event := types.PocketEvent{
		Topic: types.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	}
	bz, err := types_consensus.EncodeConsensusMessage(&types_consensus.ConsensusMessage{
		Message: message,
		// Sender:  0,
	})
	if err != nil {
		panic(err)
	}
	any, err := anypb.New(&types_consensus.Message{Data: bz})
	if err != nil {
		panic(err)
	}
	if err := m.GetBus().GetP2PModule().Broadcast(any, event.Topic); err != nil {
		// TODO handle
	}
}

func (m *consensusModule) sendToNode(message types_consensus.GenericConsensusMessage, destNode *types_consensus.NodeId) {
	event := types.PocketEvent{
		Topic: types.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		// Destination: *destNode,
	}
	// PreP2P hack
	addr := strconv.Itoa(int(*destNode))
	bz, err := types_consensus.EncodeConsensusMessage(&types_consensus.ConsensusMessage{
		Message: message,
		// Sender:  0,
	})
	any, err := anypb.New(&types_consensus.Message{Data: bz})
	if err != nil {
		panic(err)
	}
	actualAddressHack := crypto.Address([]byte(addr))
	if err := m.GetBus().GetP2PModule().Send(actualAddressHack, any, event.Topic); err != nil {
		// TODO handle
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
