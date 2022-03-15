package consensus

import (
	"fmt"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

const ByzantineThreshold float64 = float64(2) / float64(3)

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

	prepareQC, err := m.getQuorumCertificate(m.Height, Prepare, m.Round)
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
		m.nodeLogError("Could not create a propose message", err)
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

	preCommitQC, err := m.getQuorumCertificate(m.Height, PreCommit, m.Round)
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
		m.nodeLogError("Could not create a propose message", err)
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

	commitQC, err := m.getQuorumCertificate(m.Height, Commit, m.Round)
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

func (m *consensusModule) didReceiveEnoughMessageForStep(step types_consensus.HotstuffStep) (bool, string) {
	return m.isOptimisticThresholdMet(len(m.MessagePool[step]))
}

func (m *consensusModule) isOptimisticThresholdMet(n int) (bool, string) {
	valMap := types.GetTestState(nil).ValidatorMap
	return float64(n) > ByzantineThreshold*float64(len(valMap)), fmt.Sprintf("byzantine safety check: (%d > %.2f?)", n, ByzantineThreshold*float64(len(valMap)))
}

func (m *consensusModule) getQuorumCertificate(height uint64, step types_consensus.HotstuffStep, round uint64) (*types_consensus.QuorumCertificate, error) {
	var pss []*types_consensus.PartialSignature
	for _, msg := range m.MessagePool[step] {
		// TODO(olshansky): Add tests for this
		if msg.GetPartialSignature() == nil {
			m.nodeLog(fmt.Sprintf("[WARN] No partial signature found for step %s which should not happen...", StepToString[step]))
			continue
		}
		// TODO(olshansky): Add tests for this
		if msg.Height != height || msg.Round != round || msg.Step != step {
			m.nodeLog(fmt.Sprintf("[WARN] Message in pool does not match (height, step, round) of QC being generated; %d, %s, %d", height, StepToString[step], round))
			continue
		}
		ps := msg.GetPartialSignature()

		if ps.Signature == nil || len(ps.Address) == 0 {
			m.nodeLog(fmt.Sprintf("[WARN] Partial signature is incomplete for step %s which should not happen...", StepToString[step]))
			continue
		}
		pss = append(pss, msg.GetPartialSignature())
	}

	if ok, reason := m.isOptimisticThresholdMet(len(pss)); !ok {
		return nil, fmt.Errorf("did not receive enough partial signature; %s", reason)
	}

	thresholdSig, err := getThresholdSignature(pss)
	if err != nil {
		return nil, err
	}

	return &types_consensus.QuorumCertificate{
		Height:             m.Height,
		Step:               step,
		Round:              m.Round,
		Block:              m.Block,
		ThresholdSignature: thresholdSig,
	}, nil
}

func (m *consensusModule) findHighQC(step types_consensus.HotstuffStep) (qc *types_consensus.QuorumCertificate) {
	for _, m := range m.MessagePool[step] {
		if m.GetQuorumCertificate() == nil {
			continue
		}
		if qc == nil || m.GetQuorumCertificate().Height > qc.Height {
			qc = m.GetQuorumCertificate()
		}
	}
	return
}
