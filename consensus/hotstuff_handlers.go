package consensus

// TODO(olshansky): Low priority design: think of a way to make `hotstuff_*` files be a sub-package under consensus.
// This is currently not possible because functions tied to the `consensusModule` struct (implementing the ConsensusModule module)
// spans multiple files.

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*consensusModule, *types_consensus.HotstuffMessage)
	HandlePrepareMessage(*consensusModule, *types_consensus.HotstuffMessage)
	HandlePrecommitMessage(*consensusModule, *types_consensus.HotstuffMessage)
	HandleCommitMessage(*consensusModule, *types_consensus.HotstuffMessage)
	HandleDecideMessage(*consensusModule, *types_consensus.HotstuffMessage)
}

/*
TODO(design): The reason we do not assign both the leader and the replica handlers
to the leader (which should also act as a replica when it is a leader) is because it
can create a weird inconsistent state (e.g. both the replica and leader try to restart
the PaceMaker timeout). This requires additional "replica-like" logic in the leader
handler which has both pros and cons:
	Pros:
		* The leader can short-circuit and optimize replica related logic
		* Avoids additional code flowing through the P2P pipeline
		* Allows for micro-optimizations
	Cons:
		* The leader's "replica related logic" requires an additional code path
		* Code is less "generalizable" and therefore potentially more error prone
*/

// TODO(olshansky): Should we just make these singletons or embed them directly in the consensusModule?
var (
	LeaderMessageHandler  HotstuffMessageHandler = &HotstuffLeaderMessageHandler{}
	ReplicaMessageHandler HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}
)

var replicaMessageMapper map[types_consensus.HotstuffStep]func(*consensusModule, *types_consensus.HotstuffMessage) = map[types_consensus.HotstuffStep]func(*consensusModule, *types_consensus.HotstuffMessage){
	NewRound:  ReplicaMessageHandler.HandleNewRoundMessage,
	Prepare:   ReplicaMessageHandler.HandlePrepareMessage,
	PreCommit: ReplicaMessageHandler.HandlePrecommitMessage,
	Commit:    ReplicaMessageHandler.HandleCommitMessage,
	Decide:    ReplicaMessageHandler.HandleDecideMessage,
}

var leaderMessageMapper map[types_consensus.HotstuffStep]func(*consensusModule, *types_consensus.HotstuffMessage) = map[types_consensus.HotstuffStep]func(*consensusModule, *types_consensus.HotstuffMessage){
	NewRound:  LeaderMessageHandler.HandleNewRoundMessage,
	Prepare:   LeaderMessageHandler.HandlePrepareMessage,
	PreCommit: LeaderMessageHandler.HandlePrecommitMessage,
	Commit:    LeaderMessageHandler.HandleCommitMessage,
	Decide:    LeaderMessageHandler.HandleDecideMessage,
}

func (m *consensusModule) handleHotstuffMessage(msg *types_consensus.HotstuffMessage) {
	// TODO(olshansky): How can we add the `senderId` back in here for debugging purposes?
	m.nodeLog(fmt.Sprintf("[DEBUG] (%s->%d) - Height: %d; Type: %s; Round: %d.", "???", m.NodeId, msg.Height, StepToString[msg.Step], msg.Round))

	// Basic metadata checks
	if valid, reason := m.isMessageBasicValid(msg); !valid {
		m.nodeLog(fmt.Sprintf("[WARN] Discarding hotstuff message because: %s", reason))
	}

	// Liveness & safety checks
	if shouldHandle, reason := m.paceMaker.ShouldHandleMessage(msg); !shouldHandle {
		m.nodeLog(fmt.Sprintf("[WARN] Discarding hotstuff message because: %s", reason))
		return
	}

	// Need to execute leader election if there is no leader and we are in a new round.
	if msg.Step == NewRound && m.LeaderId == nil {
		m.electNextLeader(msg)
	}

	m.nodeLog(fmt.Sprintf("About to process %s msg.", StepToString[m.Step]))

	if !m.isLeader() {
		replicaMessageMapper[msg.Step](m, msg)
		return
	}

	// Discard messages with invalid partial signatures before storing it in the leader's consensus mempool
	if validPartialSig, reason := m.isValidPartialSignature(msg); !validPartialSig {
		m.nodeLogError("Discarding hotstuff message because the partial signature is invalid", fmt.Errorf(reason))
		return
	}

	// Only the leader needs to aggregate consensus related messages.
	m.MessagePool[msg.Step] = append(m.MessagePool[msg.Step], msg)

	// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
	leaderMessageMapper[msg.Step](m, msg)
}
