package consensus

import (
	"fmt"
)

type HotstuffMessageHandler interface {
	HandleNewRoundMessage(*consensusModule, *HotstuffMessage)
	HandlePrepareMessage(*consensusModule, *HotstuffMessage)
	HandlePrecommitMessage(*consensusModule, *HotstuffMessage)
	HandleCommitMessage(*consensusModule, *HotstuffMessage)
	HandleDecideMessage(*consensusModule, *HotstuffMessage)
}

/*
TODO: Think/discuss: The reason we do not assign both the leader and the replica handlers
to the leader (which should also act as a replica) is because it can create a weird
inconsistent state (e.g. the replica handler restarts the PaceMaker timeout). This requires
additional "replica-like" logic in the leader handler which has both pros and cons:
Pros:
 * The leader can short-circuit and optimize replica messages.
 * Allows for micr-Micro-optimizationAllows optimization on both
Cons:
 * The leader's replica code utilizes a different code-path.
*/
var (
	// TODO: Should we just make these singletons?
	LeaderMessageHandler  HotstuffMessageHandler = &HotstuffLeaderMessageHandler{}
	ReplicaMessageHandler HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}
)

var replicaMessageMapper map[Step]func(*consensusModule, *HotstuffMessage) = map[Step]func(*consensusModule, *HotstuffMessage){
	NewRound:  ReplicaMessageHandler.HandleNewRoundMessage,
	Prepare:   ReplicaMessageHandler.HandlePrepareMessage,
	PreCommit: ReplicaMessageHandler.HandlePrecommitMessage,
	Commit:    ReplicaMessageHandler.HandleCommitMessage,
	Decide:    ReplicaMessageHandler.HandleDecideMessage,
}

var leaderMessageMapper map[Step]func(*consensusModule, *HotstuffMessage) = map[Step]func(*consensusModule, *HotstuffMessage){
	NewRound:  LeaderMessageHandler.HandleNewRoundMessage,
	Prepare:   LeaderMessageHandler.HandlePrepareMessage,
	PreCommit: LeaderMessageHandler.HandlePrecommitMessage,
	Commit:    LeaderMessageHandler.HandleCommitMessage,
	Decide:    LeaderMessageHandler.HandleDecideMessage,
}

func (m *consensusModule) handleHotstuffMessage(message *HotstuffMessage) {
	m.nodeLog(fmt.Sprintf("[DEBUG] (%d->%d) - Height: %d; Type: %s; Round: %d.", message.Sender, m.NodeId, message.Height, StepToString[message.Step], message.Round))

	// TODO: Basics metadata & byte checks.

	// Liveness & safety checks & updates.
	if !m.paceMaker.ShouldHandleMessage(message) {
		return
	}

	// Discard messages with invalid partial signatures.
	if !m.isMessagePartialSigValid(message) {
		return
	}

	// TODO: Move this over into the persistance m.
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
