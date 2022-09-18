package consensus

import (
	"encoding/base64"
	"log"

	"github.com/pokt-network/pocket/shared/debug"

	"google.golang.org/protobuf/proto"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

// These constants and variables are wrappers around the autogenerated protobuf types and were
// added to simply make the code in the `consensus` module more readable.

const (
	NewRound  = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	Prepare   = typesCons.HotstuffStep_HOTSTUFF_STEP_PREPARE
	PreCommit = typesCons.HotstuffStep_HOTSTUFF_STEP_PRECOMMIT
	Commit    = typesCons.HotstuffStep_HOTSTUFF_STEP_COMMIT
	Decide    = typesCons.HotstuffStep_HOTSTUFF_STEP_DECIDE

	ByzantineThreshold = float64(2) / float64(3)
	HotstuffMessage    = "consensus.HotstuffMessage"
	UtilityMessage     = "consensus.UtilityMessage"
	Propose            = typesCons.HotstuffMessageType_HOTSTUFF_MESAGE_PROPOSE
	Vote               = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_VOTE
)

var (
	HotstuffSteps = [...]typesCons.HotstuffStep{NewRound, Prepare, PreCommit, Commit, Decide}

	maxTxBytes        = 90000             // TODO(olshansky): Move this to config.json.
	lastByzValidators = make([][]byte, 0) // TODO(olshansky): Retrieve this from persistence
)

// ** Hotstuff Helpers ** //

func (m *ConsensusModule) getQuorumCertificate(height uint64, step typesCons.HotstuffStep, round uint64) (*typesCons.QuorumCertificate, error) {
	var pss []*typesCons.PartialSignature
	for _, msg := range m.MessagePool[step] {
		// TODO(olshansky): Add tests for this
		if msg.GetPartialSignature() == nil {
			m.nodeLog(typesCons.WarnMissingPartialSig(msg))
			continue
		}
		// TODO(olshansky): Add tests for this
		if msg.Height != height || msg.Step != step || msg.Round != round {
			m.nodeLog(typesCons.WarnUnexpectedMessageInPool(msg, height, step, round))
			continue
		}
		ps := msg.GetPartialSignature()

		if ps.Signature == nil || len(ps.Address) == 0 {
			m.nodeLog(typesCons.WarnIncompletePartialSig(ps, msg))
			continue
		}
		pss = append(pss, msg.GetPartialSignature())
	}

	if err := m.isOptimisticThresholdMet(len(pss)); err != nil {
		return nil, err
	}

	thresholdSig, err := getThresholdSignature(pss)
	if err != nil {
		return nil, err
	}

	return &typesCons.QuorumCertificate{
		Height:             m.Height,
		Step:               step,
		Round:              m.Round,
		Block:              m.Block,
		ThresholdSignature: thresholdSig,
	}, nil
}

func (m *ConsensusModule) findHighQC(step typesCons.HotstuffStep) (qc *typesCons.QuorumCertificate) {
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

func getThresholdSignature(
	partialSigs []*typesCons.PartialSignature) (*typesCons.ThresholdSignature, error) {
	thresholdSig := new(typesCons.ThresholdSignature)
	thresholdSig.Signatures = make([]*typesCons.PartialSignature, len(partialSigs))
	copy(thresholdSig.Signatures, partialSigs)
	return thresholdSig, nil
}

func isSignatureValid(m *typesCons.HotstuffMessage, pubKeyString string, signature []byte) bool {
	pubKey, err := cryptoPocket.NewPublicKey(pubKeyString)
	if err != nil {
		log.Println("[WARN] Error getting PublicKey from bytes:", err)
		return false
	}
	bytesToVerify, err := getSignableBytes(m)
	if err != nil {
		log.Println("[WARN] Error getting bytes to verify:", err)
		return false
	}
	return pubKey.Verify(bytesToVerify, signature)
}

func (m *ConsensusModule) didReceiveEnoughMessageForStep(step typesCons.HotstuffStep) error {
	return m.isOptimisticThresholdMet(len(m.MessagePool[step]))
}

func (m *ConsensusModule) isOptimisticThresholdMet(n int) error {
	numValidators := len(m.validatorMap)
	if !(float64(n) > ByzantineThreshold*float64(numValidators)) {
		return typesCons.ErrByzantineThresholdCheck(n, ByzantineThreshold*float64(numValidators))
	}
	return nil
}

func (m *ConsensusModule) resetForNewHeight() {
	m.m.Lock()
	defer m.m.Unlock()
	m.Round = 0
	m.Block = nil
	m.HighPrepareQC = nil
	m.LockedQC = nil
}

func protoHash(m proto.Message) string {
	b, err := proto.Marshal(m)
	if err != nil {
		log.Fatalf("Could not marshal proto message: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

/*** P2P Helpers ***/

func (m *ConsensusModule) sendToNode(msg *typesCons.HotstuffMessage) {
	// TODO(olshansky): This can happen due to a race condition with the pacemaker.
	if m.LeaderId == nil {
		m.nodeLogError(typesCons.ErrNilLeaderId.Error(), nil)
		return
	}

	m.nodeLog(typesCons.SendingMessage(msg, *m.LeaderId))
	anyConsensusMessage, err := anypb.New(msg)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateConsensusMessage.Error(), err)
		return
	}

	if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(m.IdToValAddrMap[*m.LeaderId]), anyConsensusMessage, debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC); err != nil {
		m.nodeLogError(typesCons.ErrSendMessage.Error(), err)
		return
	}
}

func (m *ConsensusModule) broadcastToNodes(msg *typesCons.HotstuffMessage) {
	m.nodeLog(typesCons.BroadcastingMessage(msg))
	anyConsensusMessage, err := anypb.New(msg)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateConsensusMessage.Error(), err)
		return
	}

	if err := m.GetBus().GetP2PModule().Broadcast(anyConsensusMessage, debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC); err != nil {
		m.nodeLogError(typesCons.ErrBroadcastMessage.Error(), err)
		return
	}
}

/*** Persistence Helpers ***/

func (m *ConsensusModule) clearMessagesPool() {
	m.m.Lock()
	defer m.m.Unlock()
	for _, step := range HotstuffSteps {
		m.MessagePool[step] = make([]*typesCons.HotstuffMessage, 0)
	}
}

/*** Leader Election Helpers ***/

func (m *ConsensusModule) isLeaderUnknown() bool {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.LeaderId == nil
}

func (m *ConsensusModule) isLeader() bool {
	m.m.RLock()
	defer m.m.RUnlock()
	return m.LeaderId != nil && *m.LeaderId == m.NodeId
}

func (m *ConsensusModule) isReplica() bool {
	return !m.isLeader()
}

func (m *ConsensusModule) clearLeader() {
	m.m.Lock()
	defer m.m.Unlock()
	m.logPrefix = DefaultLogPrefix
	m.LeaderId = nil
}

func (m *ConsensusModule) electNextLeader(message *typesCons.HotstuffMessage) {
	leaderId, err := m.leaderElectionMod.ElectNextLeader(message)
	if err != nil || leaderId == 0 {
		m.nodeLogError(typesCons.ErrLeaderElection(message).Error(), err)
		m.clearLeader()
		return
	}

	m.setLeaderId(leaderId)

	if m.isCurrentNodeLeader() {
		m.setLogPrefix("LEADER")
		m.nodeLog(typesCons.ElectedSelfAsNewLeader(m.IdToValAddrMap[*m.LeaderId], *m.LeaderId, m.Height, m.Round))
	} else {
		m.setLogPrefix("REPLICA")
		m.nodeLog(typesCons.ElectedNewLeader(m.IdToValAddrMap[*m.LeaderId], *m.LeaderId, m.Height, m.Round))
	}

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

/*** General Infrastructure Helpers ***/

func (m *ConsensusModule) nodeLog(s string) {
	m.m.RLock()
	defer m.m.RUnlock()
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.NodeId, s)
}

func (m *ConsensusModule) nodeLogError(s string, err error) {
	m.m.RLock()
	defer m.m.RUnlock()
	log.Printf("[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.NodeId, s, err)
}

func (m *ConsensusModule) setLogPrefix(logPrefix string) {
	m.m.Lock()
	defer m.m.Unlock()
	m.logPrefix = logPrefix
}
