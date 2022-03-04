package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

var _ HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}

type HotstuffReplicaMessageHandler struct{}

func (handler *HotstuffReplicaMessageHandler) HandleNewRoundMessage(m *consensusModule, message *types_consensus.HotstuffMessage) {
	m.paceMaker.RestartTimer()
	m.Step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *consensusModule, message *types_consensus.HotstuffMessage) {
	if !m.isValidProposal(message) {
		m.nodeLogError("Invalid proposal in PREPARE message.", nil)
		m.paceMaker.InterruptRound()
		return
	}

	if !m.isValidBlock(message.Block) {
		m.nodeLogError("Invalid block in PREPARE message.", nil)
		m.paceMaker.InterruptRound()
		return
	}
	m.deliverTxToUtility(message.Block)

	m.Step = PreCommit
	m.paceMaker.RestartTimer()

	prepareVoteMessages, err := CreateVoteMessage(m, Prepare, message.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message: %v", err)
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(prepareVoteMessages)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *consensusModule, message *types_consensus.HotstuffMessage) {
	if !m.isQCValid(message.GetQuorumCertificate()) {
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Commit
	m.HighPrepareQC = message.GetQuorumCertificate() // TODO: DISCUSS why are we never using this for validation?
	m.paceMaker.RestartTimer()

	preCommitVoteMessage, err := CreateVoteMessage(m, PreCommit, message.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message: %v", err)
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *consensusModule, message *types_consensus.HotstuffMessage) {
	if !m.isQCValid(message.GetQuorumCertificate()) {
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Decide
	m.LockedQC = message.GetQuorumCertificate() // TODO: Discuss how the replica recover if it's locked? Replica `formally` agrees on the QC while the rest of the network `verbally` agrees on the QC.
	m.paceMaker.RestartTimer()

	commitVoteMessage, err := CreateVoteMessage(m, Commit, message.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message: %v", err)
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *consensusModule, message *types_consensus.HotstuffMessage) {
	if !m.isQCValid(message.GetQuorumCertificate()) {
		m.paceMaker.InterruptRound()
		return
	}

	if err := m.commitBlock(message.Block); err != nil {
		m.nodeLogError("Could not commit block: %v", err)
		m.paceMaker.InterruptRound()
		return
	}

	m.paceMaker.NewHeight()
}

// Helpers

func (m *consensusModule) hotstuffNodeSend(message *types_consensus.HotstuffMessage) {
	// TODO: This can happen due to a race condition with the pacemaker.
	if m.LeaderId == nil {
		m.nodeLogError("[TODO] Why am I trying to send a message to a nil leader?", nil)
		return
	}

	m.nodeLog(fmt.Sprintf("Sending %s vote.", StepToString[message.Step]))
	m.sendToNode(message, HotstuffMessage, m.LeaderId)
}
