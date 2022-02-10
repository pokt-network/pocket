package consensus

import (
	"fmt"
	"log"
	"math/rand"

	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/types"
)

func (m *consensusModule) electNextLeader(message *HotstuffMessage) {
	leaderId := m.electNextLeaderDeterministic(message)

	if leaderId == 0 {
		m.nodeLogError(fmt.Sprintf("Leader election failed. Validator cannot take part in consensus at height %d round %d", message.Height, message.Round))
		m.clearLeader()
		return
	}

	m.setLeader(&leaderId)
}

func (m *consensusModule) setLeader(leaderId *types.NodeId) {
	m.LeaderId = leaderId

	if m.LeaderId != nil && *m.LeaderId == m.NodeId {
		m.logPrefix = "LEADER"
		m.nodeLog(fmt.Sprintf("👑👑👑👑👑   %d   👑👑👑👑👑", m.NodeId))
	} else {
		m.logPrefix = "REPLICA"
		m.nodeLog(fmt.Sprintf("Elected %d as 👑.", *m.LeaderId))
	}
}

func (m *consensusModule) clearLeader() {
	m.logPrefix = DefaultLogPrefix
	m.LeaderId = nil
}

func (m *consensusModule) electNextLeaderDeterministic(message *HotstuffMessage) types.NodeId {
	valMap := shared.GetPocketState().ValidatorMap
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1
	return types.NodeId(value%int64(len(valMap)) + 1)
}

func (m *consensusModule) electNextLeaderPseudoRandom(message *HotstuffMessage) types.NodeId {
	valMap := shared.GetPocketState().ValidatorMap
	value := int64(message.Height) + int64(message.Round) + int64(message.Step)
	rand.Seed(value)
	return types.NodeId(rand.Intn(len(valMap)) + 1)
}

func (m *consensusModule) electNextLeaderRoundRobin(message *HotstuffMessage) types.NodeId {
	log.Fatalf("Not supported right now")
	return types.NodeId(0)

	// leaderNum := uint32(0)
	// valMap := shared.GetPocketState().ValidatorMap
	// if m.PreviousLeader != nil {
	// 	leaderNum = uint32(*m.PreviousLeader)
	// }
	// if leaderNum >= uint32(len(valMap)) {
	// 	return types.NodeId(leaderNum)
	// }
	// return types.NodeId(leaderNum + 1)
}
