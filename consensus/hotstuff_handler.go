package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

// DISCUSS: Should these functions return an error?
type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandlePrepareMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandlePrecommitMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandleCommitMessage(*consensusModule, *typesCons.HotstuffMessage)
	HandleDecideMessage(*consensusModule, *typesCons.HotstuffMessage)
}

func (m *consensusModule) handleHotstuffMessage(msg *typesCons.HotstuffMessage) error {
	m.nodeLog(typesCons.DebugHandlingHotstuffMessage(msg))

	step := msg.GetStep()

	// Pacemaker - Liveness & safety checks
	if err := m.paceMaker.ValidateMessage(msg); err != nil {
		if m.shouldHandleHotstuffMessage(step) {
			m.nodeLog(typesCons.WarnDiscardHotstuffMessage(msg, err.Error()))
			return err
		}
		return nil
	}

	if m.shouldElectNextLeader() {
		if err := m.electNextLeader(msg); err != nil {
			return err
		}
	}

	// Hotstuff - Handle message
	if m.isReplica() {
		replicaHandlers[step](m, msg)
	}
	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderHandlers[step](m, msg)

	return nil
}

func (m *consensusModule) shouldElectNextLeader() bool {
	// Execute leader election if there is no leader and we are in a new round
	return m.Step == NewRound && m.LeaderId == nil
}

func (m *consensusModule) shouldHandleHotstuffMessage(step typesCons.HotstuffStep) bool {
	// If a replica is not a leader for this round, but has already determined a leader,
	// and continues to receive NewRound messages, we avoid logging the "message discard"
	// because it creates unnecessary spam.
	return !(m.LeaderId != nil && !m.isLeader() && step == NewRound)
}
