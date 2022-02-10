package consensus

import (
	"fmt"
)

type HotstuffLeaderMessageHandler struct{}

/*** Prepare Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleNewRoundMessage(m *consensusModule, message *HotstuffMessage) {
	if !m.didReceiveEnoughMessageForStep(NewRound) {
		m.nodeLog("still waiting for more NEWROUND messages.")
		return
	}

	// TODO: Do we need to pause for MinBlockFreqMSec here to let more transactions come in?
	m.nodeLog("received enough NEWROUND messages.")

	// Likely to be `nil` if blockchain is progressing well.
	highPrepareQC := m.findHighQC(NewRound)

	// TODO: Is this check sufficient?
	if highPrepareQC == nil || highPrepareQC.Height < m.Height || highPrepareQC.Round < m.Round {
		block, err := m.prepareBlock()
		if err != nil {
			m.nodeLogError(fmt.Sprintf("Could not prepare a block for proposal: %v", err))
			m.paceMaker.InterruptRound()
			return
		}
		m.Block = block
		m.deliverTxToUtility(block)
	} else {
		// TODO: Do we need to validate highPrepareQC here?
		m.Block = highPrepareQC.Block
	}

	m.Step = Prepare
	m.MessagePool[NewRound] = nil
	m.paceMaker.RestartTimer()

	prepareProposeMessage, err := CreateProposeMessage(m, Prepare, highPrepareQC)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a propose message: %v", err))
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(prepareProposeMessage)

	// Leader also acts like a replica.
	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, m.Block)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a vote message: %v", err))
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *consensusModule, message *HotstuffMessage) {
	if !m.didReceiveEnoughMessageForStep(Prepare) {
		m.nodeLog("Still waiting for more PREPARE messages...")
		return
	}
	m.nodeLog("received enough PREPARE messages.")

	prepareQC, err := m.getQCForStep(Prepare)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not get QC for PREPARE step: %v", err))
		return
	}

	m.Step = PreCommit
	m.HighPrepareQC = prepareQC
	m.MessagePool[Prepare] = nil
	m.paceMaker.RestartTimer()

	precommitProposeMessages, err := CreateProposeMessage(m, PreCommit, prepareQC)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a propose message: %v", err))
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(precommitProposeMessages)

	// Leader also acts like a replica.
	precommitVoteMessage, err := CreateVoteMessage(m, PreCommit, m.Block)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a vote message: %v", err))
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *consensusModule, message *HotstuffMessage) {
	if !m.didReceiveEnoughMessageForStep(PreCommit) {
		m.nodeLog("still waiting for more PRECOMMIT votes.")
		return
	}
	m.nodeLog("received enough PRECOMMIT votes.")

	preCommitQC, err := m.getQCForStep(PreCommit)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not get QC for PRECOMMIT step: %v", err))
		return
	}

	m.Step = Commit
	m.LockedQC = preCommitQC
	m.MessagePool[PreCommit] = nil
	m.paceMaker.RestartTimer()

	commitProposeMessage, err := CreateProposeMessage(m, Commit, preCommitQC)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a propose message: %v", err))
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(commitProposeMessage)

	// Leader also acts like a replica.
	commitVoteMessage, err := CreateVoteMessage(m, Commit, m.Block)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a vote message: %v", err))
		return // TODO: Should we interrupt the round here?
	}
	m.hotstuffNodeSend(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *consensusModule, message *HotstuffMessage) {
	if !m.didReceiveEnoughMessageForStep(Commit) {
		m.nodeLog("still waiting for more COMMIT votes.")
		return
	}
	m.nodeLog("received enough COMMIT votes.")

	commitQC, err := m.getQCForStep(Commit)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not get QC for COMMIT step: %v", err))
		return
	}

	m.Step = Decide
	m.MessagePool[Commit] = nil
	m.paceMaker.RestartTimer()

	decideProposeMessage, err := CreateProposeMessage(m, Decide, commitQC)
	if err != nil {
		m.nodeLogError(fmt.Sprintf("Could not create a propose message: %v", err))
		m.paceMaker.InterruptRound()
		return
	}
	m.hotstuffLeaderBroadcast(decideProposeMessage)

	if err := m.commitBlock(m.Block); err != nil {
		m.nodeLogError(fmt.Sprintf("Could not commit block: %v", err))
		m.paceMaker.InterruptRound()
		return
	}

	m.paceMaker.NewHeight()
}

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *consensusModule, message *HotstuffMessage) {
	m.nodeLog("[NOOP] Leader does nothing on DECIDE message.")
}

// Helpers

func (m *consensusModule) hotstuffLeaderBroadcast(message *HotstuffMessage) {
	m.nodeLog(fmt.Sprintf("Broadcasting %s message.", StepToString[message.Step]))
	m.broadcastToNodes(message)
}
