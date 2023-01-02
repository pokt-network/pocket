package consensus

// TODO: Split this file into multiple helpers (e.g. signatures.go, hotstuff_helpers.go, etc...)
import (
	"encoding/base64"
	"log"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/proto"
)

// These constants and variables are wrappers around the autogenerated protobuf types and were
// added to simply make the code in the `consensus` module more readable.
const (
	NewRound  = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	Prepare   = typesCons.HotstuffStep_HOTSTUFF_STEP_PREPARE
	PreCommit = typesCons.HotstuffStep_HOTSTUFF_STEP_PRECOMMIT
	Commit    = typesCons.HotstuffStep_HOTSTUFF_STEP_COMMIT
	Decide    = typesCons.HotstuffStep_HOTSTUFF_STEP_DECIDE

	Propose = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
	Vote    = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_VOTE

	ByzantineThreshold = float64(2) / float64(3)

	HotstuffMessageContentType = "consensus.HotstuffMessage"
)

var (
	HotstuffSteps = [...]typesCons.HotstuffStep{NewRound, Prepare, PreCommit, Commit, Decide}
)

// ** Hotstuff Helpers ** //

// IMPROVE: Avoid having the `ConsensusModule` be a receiver of this; making it more functional.
// TODO: Add unit tests for all quorumCert creation & validation logic...
func (m *consensusModule) getQuorumCertificate(height uint64, step typesCons.HotstuffStep, round uint64) (*typesCons.QuorumCertificate, error) {
	var pss []*typesCons.PartialSignature
	for _, msg := range m.messagePool[step] {
		if msg.GetPartialSignature() == nil {
			m.nodeLog(typesCons.WarnMissingPartialSig(msg))
			continue
		}
		if msg.GetHeight() != height || msg.GetStep() != step || msg.GetRound() != round {
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
		Height:             height,
		Step:               step,
		Round:              round,
		Block:              m.block,
		ThresholdSignature: thresholdSig,
	}, nil
}

func (m *consensusModule) findHighQC(msgs []*typesCons.HotstuffMessage) (qc *typesCons.QuorumCertificate) {
	for _, m := range msgs {
		if m.GetQuorumCertificate() == nil {
			continue
		}
		// TODO: Make sure to validate the "highest QC" first and add tests
		if qc == nil || m.GetQuorumCertificate().Height > qc.Height {
			qc = m.GetQuorumCertificate()
		}
	}
	return
}

func getThresholdSignature(partialSigs []*typesCons.PartialSignature) (*typesCons.ThresholdSignature, error) {
	thresholdSig := new(typesCons.ThresholdSignature)
	thresholdSig.Signatures = make([]*typesCons.PartialSignature, len(partialSigs))
	copy(thresholdSig.Signatures, partialSigs)
	return thresholdSig, nil
}

func isSignatureValid(msg *typesCons.HotstuffMessage, pubKeyString string, signature []byte) bool {
	pubKey, err := cryptoPocket.NewPublicKey(pubKeyString)
	if err != nil {
		log.Println("[WARN] Error getting PublicKey from bytes:", err)
		return false
	}
	bytesToVerify, err := getSignableBytes(msg)
	if err != nil {
		log.Println("[WARN] Error getting bytes to verify:", err)
		return false
	}
	return pubKey.Verify(bytesToVerify, signature)
}

func (m *consensusModule) didReceiveEnoughMessageForStep(step typesCons.HotstuffStep) error {
	return m.isOptimisticThresholdMet(len(m.messagePool[step]))
}

func (m *consensusModule) isOptimisticThresholdMet(n int) error {
	numValidators := len(m.validatorMap)
	if !(float64(n) > ByzantineThreshold*float64(numValidators)) {
		return typesCons.ErrByzantineThresholdCheck(n, ByzantineThreshold*float64(numValidators))
	}
	return nil
}

func (m *consensusModule) resetForNewHeight() {
	m.round = 0
	m.block = nil
	m.highPrepareQC = nil
	m.lockedQC = nil
}

func protoHash(m proto.Message) string {
	b, err := codec.GetCodec().Marshal(m)
	if err != nil {
		log.Fatalf("Could not marshal proto message: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

/*** P2P Helpers ***/

func (m *consensusModule) sendToLeader(msg *typesCons.HotstuffMessage) {
	m.nodeLog(typesCons.SendingMessage(msg, *m.leaderId))

	// TODO: This can happen due to a race condition with the pacemaker.
	if m.leaderId == nil {
		m.nodeLogError(typesCons.ErrNilLeaderId.Error(), nil)
		return
	}

	anyConsensusMessage, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateConsensusMessage.Error(), err)
		return
	}

	if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(m.idToValAddrMap[*m.leaderId]), anyConsensusMessage); err != nil {
		m.nodeLogError(typesCons.ErrSendMessage.Error(), err)
		return
	}
}

// Star-like (O(n)) broadcast - send to all nodes directly
// INVESTIGATE: Re-evaluate if we should be using our structured broadcast (RainTree O(log3(n))) algorithm instead
func (m *consensusModule) broadcastToValidators(msg *typesCons.HotstuffMessage) {
	m.nodeLog(typesCons.BroadcastingMessage(msg))

	anyConsensusMessage, err := codec.GetCodec().ToAny(msg)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateConsensusMessage.Error(), err)
		return
	}

	for _, val := range m.validatorMap {
		if err := m.GetBus().GetP2PModule().Send(cryptoPocket.AddressFromString(val.GetAddress()), anyConsensusMessage); err != nil {
			m.nodeLogError(typesCons.ErrBroadcastMessage.Error(), err)
		}
	}
}

/*** Persistence Helpers ***/

// TECHDEBT(#388): Integrate this with the `persistence` module or a real mempool.
func (m *consensusModule) clearMessagesPool() {
	for _, step := range HotstuffSteps {
		m.messagePool[step] = make([]*typesCons.HotstuffMessage, 0)
	}
}

/*** Leader Election Helpers ***/

func (m *consensusModule) isLeaderUnknown() bool {
	return m.leaderId == nil
}

func (m *consensusModule) isLeader() bool {
	return m.leaderId != nil && *m.leaderId == m.nodeId
}

func (m *consensusModule) isReplica() bool {
	return !m.isLeader()
}

func (m *consensusModule) clearLeader() {
	m.logPrefix = DefaultLogPrefix
	m.leaderId = nil
}

func (m *consensusModule) electNextLeader(message *typesCons.HotstuffMessage) error {
	leaderId, err := m.leaderElectionMod.ElectNextLeader(message)
	if err != nil || leaderId == 0 {
		m.nodeLogError(typesCons.ErrLeaderElection(message).Error(), err)
		m.clearLeader()
		return err
	}
	m.leaderId = &leaderId

	if m.isLeader() {
		m.setLogPrefix("LEADER")
		m.nodeLog(typesCons.ElectedSelfAsNewLeader(m.idToValAddrMap[*m.leaderId], *m.leaderId, m.height, m.round))
	} else {
		m.setLogPrefix("REPLICA")
		m.nodeLog(typesCons.ElectedNewLeader(m.idToValAddrMap[*m.leaderId], *m.leaderId, m.height, m.round))
	}

	return nil
}

/*** General Infrastructure Helpers ***/

// TODO(#164): Remove this once we have a proper logging system.
func (m *consensusModule) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.nodeId, s)
}

// TODO(#164): Remove this once we have a proper logging system.
func (m *consensusModule) nodeLogError(s string, err error) {
	log.Printf("🐞[ERROR][%s][%d] %s: %v\n", m.logPrefix, m.nodeId, s, err)
}

func (m *consensusModule) setLogPrefix(logPrefix string) {
	m.logPrefix = logPrefix
}
