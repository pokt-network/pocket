package consensus

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
	sender := 4 // message.Sender
	m.nodeLog(fmt.Sprintf("[DEBUG] (%d->%d) - Height: %d; Type: %s; Round: %d.", sender, m.NodeId, message.Height, StepToString[message.Step], message.Round))

	// TODO(olshansky): Basics metadata & byte checks.

	// Liveness & safety checks.
	shouldHandle, reason := m.paceMaker.ShouldHandleMessage(message)
	if !shouldHandle {
		m.nodeLog(fmt.Sprintf("[WARN] Discarding hotstuff message because: %s", reason))
		return
	}
	m.nodeLog(reason)

	// Discard messages with invalid partial signatures.
	if !m.isMessagePartialSigValid(message) {
		return
	}

	// TODO(olshansky): Move this over into the persistence m.
	m.MessagePool[message.Step] = append(m.MessagePool[message.Step], *message)

	if m.LeaderId == nil && message.Step == NewRound {
		m.electNextLeader(message)
	}

	m.nodeLog(fmt.Sprintf("About to process %s message.", StepToString[m.Step]))
	if m.isLeader() {
		leaderMessageMapper[message.Step](m, message)
	} else {
		replicaMessageMapper[message.Step](m, message)
	}
}
