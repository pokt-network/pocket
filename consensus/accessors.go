package consensus

import typesCons "github.com/pokt-network/pocket/consensus/types"

/*** General Purpose Getters / Setters ***/

func (m *ConsensusModule) incrementRound() {
	m.m.Lock()
	defer m.m.Unlock()
	m.Round++
}

func (m *ConsensusModule) incrementHeight() {
	m.m.Lock()
	defer m.m.Unlock()
	m.Height++
}

func (m *ConsensusModule) getHeight() uint64 {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.Height
}

func (m *ConsensusModule) getRound() uint64 {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.Round
}

func (m *ConsensusModule) getStep() typesCons.HotstuffStep {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.Step
}

func (m *ConsensusModule) setStep(step typesCons.HotstuffStep) {
	m.m.Lock()
	defer m.m.Unlock()
	m.Step = step
}

func (m *ConsensusModule) setRound(round uint64) {
	m.m.Lock()
	defer m.m.Unlock()
	m.Round = round
}

func (m *ConsensusModule) setHighPrepareQC(qc *typesCons.QuorumCertificate) {
	m.m.Lock()
	defer m.m.Unlock()
	m.HighPrepareQC = qc
}

func (m *ConsensusModule) setMessagePoolForStep(step typesCons.HotstuffStep, msgs []*typesCons.HotstuffMessage) {
	m.m.Lock()
	defer m.m.Unlock()
	m.MessagePool[step] = msgs
}

func (m *ConsensusModule) setLeaderId(leaderId typesCons.NodeId) {
	m.m.Lock()
	defer m.m.Unlock()
	m.LeaderId = &leaderId
}

func (m *ConsensusModule) isCurrentNodeLeader() bool {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}
