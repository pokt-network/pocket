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
	// IMPROVE: Add source of message here
	loggingFields := hotstuffMsgToLoggingFields(msg)

	m.logger.Debug().Fields(loggingFields).Msg("Received hotstuff msg...")

	// Pacemaker - Liveness & safety checks
	if shouldHandle, err := m.paceMaker.ShouldHandleMessage(msg); !shouldHandle {
		m.logger.Debug().Fields(loggingFields).Msg("Not handling hotstuff msg...")
		return err
	}

	// IMPROVE: Add source of message here
	m.logger.Debug().Fields(loggingFields).Msg("About to start handling hotstuff msg...")

	// Elect a leader for the current round if needed
	if m.shouldElectNextLeader() {
		if err := m.electNextLeader(msg); err != nil {
			return err
		}
	}

	if m.IsLeader() {
		// Hotstuff - Handle message as a leader;
		// NB: Leader also acts as a replica, but this logic is implemented in the underlying code
		leaderHandlers[msg.GetStep()](m, msg)
	} else {
		// Hotstuff - Handle message as a replica
		replicaHandlers[msg.GetStep()](m, msg)
	}

	return nil
}

func (m *consensusModule) shouldElectNextLeader() bool {
	// Execute leader election if there is no leader and we are in a NewRound
	return m.step == NewRound && m.leaderId == nil
}
