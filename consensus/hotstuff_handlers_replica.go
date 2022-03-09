package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

var _ HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}

type HotstuffReplicaMessageHandler struct{}

func (handler *HotstuffReplicaMessageHandler) HandleNewRoundMessage(m *consensusModule, msg *types_consensus.HotstuffMessage) {
	m.paceMaker.RestartTimer()
	m.Step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *consensusModule, msg *types_consensus.HotstuffMessage) {
	if valid, reason := m.isValidProposal(msg); !valid {
		m.nodeLogError("Invalid proposal in PREPARE message", fmt.Errorf(reason))
		m.paceMaker.InterruptRound()
		return
	}

	if valid, reason := m.isValidBlock(msg.Block); !valid {
		m.nodeLogError("Invalid block in PREPARE message", fmt.Errorf(reason))
		m.paceMaker.InterruptRound()
		return
	}

	if err := m.applyBlock(msg.Block); err != nil {
		m.nodeLogError("Could not apply the block", err)
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = PreCommit
	m.paceMaker.RestartTimer()

	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, msg.Block)
	if err != nil {
		m.nodeLogError("Could not create a PREPARE Vote", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *types_consensus.HotstuffMessage) {
	if valid, reason := m.isQuorumCertificateValid(msg.GetQuorumCertificate()); !valid {
		m.nodeLogError("QC is invalid in the PRECOMMIT step", fmt.Errorf(reason))
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Commit
	m.HighPrepareQC = msg.GetQuorumCertificate() // TODO(discuss): Why are we never using this for validation?
	m.paceMaker.RestartTimer()

	preCommitVoteMessage, err := CreateVoteMessage(m, PreCommit, msg.Block)
	if err != nil {
		m.nodeLogError("Could not create a PRECOMMIT Vote", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *consensusModule, msg *types_consensus.HotstuffMessage) {
	if valid, reason := m.isQuorumCertificateValid(msg.GetQuorumCertificate()); !valid {
		m.nodeLogError("QC is invalid in the COMMIT step", fmt.Errorf(reason))
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Decide
	m.LockedQC = msg.GetQuorumCertificate() // TODO(discuss): How do the replica recover if it's locked? Replica `formally` agrees on the QC while the rest of the network `verbally` agrees on the QC.
	m.paceMaker.RestartTimer()

	commitVoteMessage, err := CreateVoteMessage(m, Commit, msg.Block)
	if err != nil {
		m.nodeLogError("Could not create a COMMIT Vote", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *consensusModule, msg *types_consensus.HotstuffMessage) {
	if valid, reason := m.isQuorumCertificateValid(msg.GetQuorumCertificate()); !valid {
		m.nodeLogError("QC is invalid in the DECIDE step", fmt.Errorf(reason))
		m.paceMaker.InterruptRound()
		return
	}

	if err := m.commitBlock(msg.Block); err != nil {
		m.nodeLogError("Could not commit block: %v", err)
		m.paceMaker.InterruptRound()
		return
	}

	m.paceMaker.NewHeight()
}

/*** Helpers ***/

func (m *consensusModule) hotstuffNodeSend(msg *types_consensus.HotstuffMessage) {
	// TODO(olshansky): This can happen due to a race condition with the pacemaker.
	if m.LeaderId == nil {
		m.nodeLogError("[TODO] How/why am I trying to send a message to a nil leader?", nil)
		return
	}

	m.nodeLog(fmt.Sprintf("Sending %s vote.", StepToString[msg.Step]))
	m.sendToNode(msg, HotstuffMessage, *m.LeaderId)
}
