package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) isLeader() bool {
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}

func (m *consensusModule) electNextLeader(message *types_consensus.HotstuffMessage) {
	leaderId := m.electNextLeaderDeterministic(message)

	if leaderId == 0 {
		m.nodeLogError(fmt.Sprintf("Leader election failed. Validator cannot take part in consensus at height %d round %d", message.Height, message.Round), nil)
		m.clearLeader()
		return
	}

	m.setLeader(&leaderId)
}

func (m *consensusModule) setLeader(leaderId *types_consensus.NodeId) {
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

func (m *consensusModule) electNextLeaderDeterministic(message *types_consensus.HotstuffMessage) types_consensus.NodeId {
	valMap := types.GetTestState(nil).ValidatorMap
	value := int64(message.Height) + int64(message.Round) + int64(message.Step) - 1
	return types_consensus.NodeId(value%int64(len(valMap)) + 1)
}
