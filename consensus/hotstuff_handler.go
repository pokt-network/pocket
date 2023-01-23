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
	step := msg.GetStep()

	m.nodeLog(typesCons.DebugReceivedHandlingHotstuffMessage(msg))
	// Pacemaker - Liveness & safety checks
	if shouldHandle, err := m.paceMaker.ShouldHandleMessage(msg); !shouldHandle {
		return err
	}
	m.nodeLog(typesCons.DebugHandlingHotstuffMessage(msg))

	// Elect a leader for the current round if needed
	if m.shouldElectNextLeader() {
		if err := m.electNextLeader(msg); err != nil {
			return err
		}
	}

	// Hotstuff - Handle message as a replica
	if m.isReplica() {
		replicaHandlers[step](m, msg)
	}

	// Hotstuff - Handle message as a leader
	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderHandlers[step](m, msg)

	return nil
}

func (m *consensusModule) shouldElectNextLeader() bool {
	// Execute leader election if there is no leader and we are in a new round
	return m.step == NewRound && m.leaderId == nil
}
