package consensus

import typesCons "github.com/pokt-network/pocket/consensus/types"

// TODO(discuss): Low priority design: think of a way to make `hotstuff_*` files be a sub-package under consensus.
// This is currently not possible because functions tied to the `ConsensusModule`
// struct (implementing the ConsensusModule module), which spans multiple files.
/*
TODO(discuss): The reason we do not assign both the leader and the replica handlers
to the leader (which should also act as a replica when it is a leader) is because it
can create a weird inconsistent state (e.g. both the replica and leader try to restart
the Pacemaker timeout). This requires additional "replica-like" logic in the leader
handler which has both pros and cons:
	Pros:
		* The leader can short-circuit and optimize replica related logic
		* Avoids additional code flowing through the P2P pipeline
		* Allows for micro-optimizations
	Cons:
		* The leader's "replica related logic" requires an additional code path
		* Code is less "generalizable" and therefore potentially more error prone
*/

type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrepareMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandlePrecommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleCommitMessage(*ConsensusModule, *typesCons.HotstuffMessage)
	HandleDecideMessage(*ConsensusModule, *typesCons.HotstuffMessage)
}

func (m *ConsensusModule) handleHotstuffMessage(msg *typesCons.HotstuffMessage) {
	m.nodeLog(typesCons.DebugHandlingHotstuffMessage(msg))

	step := msg.GetStep()
	// Liveness & safety checks
	if err := m.paceMaker.ValidateMessage(msg); err != nil {
		// If a replica is not a leader for this round, but has already determined a leader,
		// and continues to receive NewRound messages, we avoid logging the "message discard"
		// because it creates unnecessary spam.
		if !(m.LeaderId != nil && !m.isLeader() && step == NewRound) {
			m.nodeLog(typesCons.WarnDiscardHotstuffMessage(msg, err.Error()))
		}
		return
	}

	// Need to execute leader election if there is no leader and we are in a new round.
	if m.Step == NewRound && m.LeaderId == nil {
		m.electNextLeader(msg)
	}

	if m.isReplica() {
		replicaHandlers[step](m, msg)
		return
	}

	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderHandlers[step](m, msg)
}
