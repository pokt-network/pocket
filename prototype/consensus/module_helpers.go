package consensus

import (
	"log"
	consensus_types "pocket/consensus/types"
	"pocket/shared/types"
	"strconv"

	"google.golang.org/protobuf/types/known/anypb"
)

func (m *ConsensusModule) broadcastToNodes(message consensus_types.GenericConsensusMessage) {
	event := types.Event{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.CONSENSUS),
	}
	bz, err := consensus_types.EncodeConsensusMessage(&consensus_types.ConsensusMessage{
		Message: message,
		Sender:  0,
	})
	if err != nil {
		panic(err)
	}
	any, err := anypb.New(&consensus_types.Message{Data: bz})
	if err != nil {
		panic(err)
	}
	if err := m.GetBus().GetNetworkModule().BroadcastMessage(any, event.PocketTopic); err != nil {
		// TODO handle
	}
}

func (m *ConsensusModule) sendToNode(message consensus_types.GenericConsensusMessage, destNode *consensus_types.NodeId) {
	event := types.Event{
		SourceModule: types.CONSENSUS_MODULE,
		PocketTopic:  string(types.CONSENSUS),
		Destination:  *destNode,
	}
	// PreP2P hack
	addr := strconv.Itoa(int(*destNode))
	bz, err := consensus_types.EncodeConsensusMessage(&consensus_types.ConsensusMessage{
		Message: message,
		Sender:  0,
	})
	any, err := anypb.New(&consensus_types.Message{Data: bz})
	if err != nil {
		panic(err)
	}
	if err := m.GetBus().GetNetworkModule().Send(addr, any, event.PocketTopic); err != nil {
		// TODO handle
	}
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
