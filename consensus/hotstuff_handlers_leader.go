package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
)

var _ HotstuffMessageHandler = &HotstuffLeaderMessageHandler{}

type HotstuffLeaderMessageHandler struct{}

/*** Prepare Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleNewRoundMessage(m *consensusModule, _ *types_consensus.HotstuffMessage) {
	if ok, reason := m.didReceiveEnoughMessageForStep(NewRound); !ok {
		m.nodeLog(fmt.Sprintf("Still waiting for more NEWROUND messages; %s", reason))
		return
	}

	// TODO(olshansky): Do we need to pause for `MinBlockFreqMSec` here to let more transactions come in?
	m.nodeLog("Received enough NEWROUND messages!")

	// Likely to be `nil` if blockchain is progressing well.
	highPrepareQC := m.findHighQC(NewRound)

	// TODO(olshansky): Add more unit tests for these checks...
	if highPrepareQC == nil || highPrepareQC.Height < m.Height || highPrepareQC.Round < m.Round {
		block, err := m.prepareBlock()
		if err != nil {
			m.nodeLogError("Could not prepare a block for proposal", err)
			m.paceMaker.InterruptRound()
			return
		}
		m.Block = block
	} else {
		// TODO(discuss): Do we need to validate highPrepareQC here?
		m.Block = highPrepareQC.Block
	}

	m.Step = Prepare
	m.MessagePool[NewRound] = nil
	m.paceMaker.RestartTimer()

	prepareProposeMessage, err := CreateProposeMessage(m, Prepare, highPrepareQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(prepareProposeMessage)

	// Leader also acts like a replica
	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, m.Block)
	if err != nil {
		m.nodeLogError("Leader could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *consensusModule, _ *types_consensus.HotstuffMessage) {
	if ok, reason := m.didReceiveEnoughMessageForStep(Prepare); !ok {
		m.nodeLog(fmt.Sprintf("Still waiting for more PREPARE messages; %s", reason))
		return
	}
	m.nodeLog("Received enough PREPARE votes!")

	prepareQC, err := m.getQuorumCertificateForStep(Prepare)
	if err != nil {
		m.nodeLogError("Could not get QC for PREPARE step", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = PreCommit
	m.HighPrepareQC = prepareQC
	m.MessagePool[Prepare] = nil
	m.paceMaker.RestartTimer()

	precommitProposeMessages, err := CreateProposeMessage(m, PreCommit, prepareQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message.", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(precommitProposeMessages)

	// Leader also acts like a replica
	precommitVoteMessage, err := CreateVoteMessage(m, PreCommit, m.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *consensusModule, _ *types_consensus.HotstuffMessage) {
	if ok, reason := m.didReceiveEnoughMessageForStep(PreCommit); !ok {
		m.nodeLog(fmt.Sprintf("Still waiting for more PRECOMMIT messages; %s", reason))
		return
	}
	m.nodeLog("received enough PRECOMMIT votes!")

	preCommitQC, err := m.getQuorumCertificateForStep(PreCommit)
	if err != nil {
		m.nodeLogError("Could not get QC for PRECOMMIT step", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Commit
	m.LockedQC = preCommitQC
	m.MessagePool[PreCommit] = nil
	m.paceMaker.RestartTimer()

	commitProposeMessage, err := CreateProposeMessage(m, Commit, preCommitQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message.", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(commitProposeMessage)

	// Leader also acts like a replica
	commitVoteMessage, err := CreateVoteMessage(m, Commit, m.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.hotstuffNodeSend(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *consensusModule, _ *types_consensus.HotstuffMessage) {
	if ok, reason := m.didReceiveEnoughMessageForStep(Commit); !ok {
		m.nodeLog(fmt.Sprintf("Still waiting for more COMMIT messages; %s", reason))
		return
	}
	m.nodeLog("Received enough COMMIT votes!")

	commitQC, err := m.getQuorumCertificateForStep(Commit)
	if err != nil {
		m.nodeLogError("Could not get QC for COMMIT step.", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Decide
	m.MessagePool[Commit] = nil
	m.paceMaker.RestartTimer()

	decideProposeMessage, err := CreateProposeMessage(m, Decide, commitQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(decideProposeMessage)

	if err := m.commitBlock(m.Block); err != nil {
		m.nodeLogError("Leader could not commit block during DECIDE step", err)
		m.paceMaker.InterruptRound()
		return
	}

	// There is no "replica behaviour" to immitate here

	m.paceMaker.NewHeight()
}

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *consensusModule, _ *types_consensus.HotstuffMessage) {
	m.nodeLog("[NOOP] Leader does nothing on DECIDE message.")
}

/*** Helpers ***/

func (m *consensusModule) hotstuffLeaderBroadcast(msg *types_consensus.HotstuffMessage) {
	m.nodeLog(fmt.Sprintf("Broadcasting %s message.", StepToString[msg.Step]))
	m.broadcastToNodes(msg, HotstuffMessage)
}
