package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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
		// if the message height is higher than node's current height, node needs to start active state sync
		if msg.Height > m.height {
			// if the message is a decide message (which is the final consensus cycle, and it means proposed block is persisted), set active sync height to message height
			// else, set active sync height to message height - 1.
			if msg.Step == Decide {
				m.stateSync.SetActiveSyncHeight(msg.Height)
			} else {
				m.stateSync.SetActiveSyncHeight(msg.Height - 1)
			}
			err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynced)
			return err
		}
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
