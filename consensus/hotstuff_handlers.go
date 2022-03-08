package consensus

// TODO(olshansky): Low priority design: think of a way to make `hotstuff_*` files be a sub-package under consensus.

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
can create a weird inconsistent state (e.g. the replica handler restarts the
PaceMaker timeout). This requires additional "replica-like" logic in the leader handler
which has both pros and cons:
	Pros:
		* The leader can short-circuit and optimize replica related logic
		* Allows for micro-optimizations
		* Avoids code flowing through the P2P pipeline
	Cons:
		* The leader's "replica related logic" utilizes a different code-path
		* Code is less "generalizable" and therefore potentially more error prone.
*/
var (
	// TODO: Should we just make these singletons?
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

func (m *consensusModule) handleHotstuffMessage(message *types_consensus.HotstuffMessage) {
	// TODO(olshansky): Can we add the senderID back in here?
	m.nodeLog(fmt.Sprintf("[DEBUG] (%s->%d) - Height: %d; Type: %s; Round: %d.", "???", m.NodeId, message.Height, StepToString[message.Step], message.Round))

	// TODO(olshansky): Basics metadata & byte checks.

	// Liveness & safety checks.
	shouldHandle, reason := m.paceMaker.ShouldHandleMessage(message)
	if !shouldHandle {
		m.nodeLog(fmt.Sprintf("[WARN] Discarding hotstuff message because: %s", reason))
		return
	}
	m.nodeLog(reason)

	// Discard messages with invalid partial signatures.
	validPartialSig, reason := m.isMessagePartialSigValid(message)
	if !validPartialSig {
		m.nodeLogError("Discarding hotstuff message because the partial signature is invalid.", fmt.Errorf(reason))
		return
	}

	// TODO(olshansky): Move this over into the persistence module.
	m.MessagePool[message.Step] = append(m.MessagePool[message.Step], message)

	// Need to execute leader election if there is no leader and we are in a new round.
	if m.LeaderId == nil && message.Step == NewRound {
		m.electNextLeader(message)
	}

	m.nodeLog(fmt.Sprintf("About to process %s message.", StepToString[m.Step]))
	if m.isLeader() {
		// Note that the leader also acts as a replica, but this logic is implemented in the underlying code.
		leaderMessageMapper[message.Step](m, message)
	} else {
		replicaMessageMapper[message.Step](m, message)
	}
}
