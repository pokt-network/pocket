package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

type HotstuffReplicaMessageHandler struct{}

var (
	ReplicaMessageHandler HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}
	replicaHandlers                              = map[typesCons.HotstuffStep]func(*consensusModule, *typesCons.HotstuffMessage){
		NewRound:  ReplicaMessageHandler.HandleNewRoundMessage,
		Prepare:   ReplicaMessageHandler.HandlePrepareMessage,
		PreCommit: ReplicaMessageHandler.HandlePrecommitMessage,
		Commit:    ReplicaMessageHandler.HandleCommitMessage,
		Decide:    ReplicaMessageHandler.HandleDecideMessage,
	}
)

/*** NewRound Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleNewRoundMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	m.paceMaker.RestartTimer()
	m.Step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.validateProposal(msg); err != nil {
		m.nodeLogError("Invalid proposal in PREPARE message", err)
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
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.validateQuorumCertificate(msg.GetQuorumCertificate()); err != nil {
		m.nodeLogError("QC is invalid in the PRECOMMIT step", err)
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
	m.sendToNode(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.validateQuorumCertificate(msg.GetQuorumCertificate()); err != nil {
		m.nodeLogError("QC is invalid in the COMMIT step", err)
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
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.validateQuorumCertificate(msg.GetQuorumCertificate()); err != nil {
		m.nodeLogError("QC is invalid in the DECIDE step", err)
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

// anteHandle is the handler called on every replica message before specific handler
func (handler *HotstuffReplicaMessageHandler) anteHandle(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	return nil
}

func (m *consensusModule) validateProposal(msg *typesCons.HotstuffMessage) error {
	if !(msg.Type == Propose && msg.Step == Prepare) {
		return typesCons.ErrProposalNotValidInPrepare
	}

	if err := m.validateBlock(msg.Block); err != nil {
		return err
	}

	// TODO(discuss): A nil QC implies a successfull CommitQC or TimeoutQC, which have been omitted intentionally since
	// they are not needed for consensus validity. However, if a QC is specified, it must be valid.
	if msg.GetQuorumCertificate() != nil {
		if err := m.validateQuorumCertificate(msg.GetQuorumCertificate()); err != nil {
			return err
		}
	}

	lockedQC := m.LockedQC
	justifyQC := msg.GetQuorumCertificate()

	// Safety: not locked
	if lockedQC == nil {
		m.nodeLog(typesCons.NotLockedOnQC)
		return nil
	}

	// Safety: check the hash of the locked QC
	// TODO(olshansky): Extend implementation to adopt `ExtendsFrom` as described in the Hotstuff whitepaper.
	if protoHash(lockedQC.Block) == protoHash(justifyQC.Block) { // && lockedQC.Block.ExtendsFrom(justifyQC.Block)
		m.nodeLog(typesCons.ProposalBlockExtends)
		return nil
	}

	// Liveness: node is locked on a QC from the past. [TODO]: Do we want to set `m.LockedQC = nil` here or something else?
	if justifyQC.Height > lockedQC.Height || (justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round) {
		return typesCons.ErrNodeIsLockedOnPastQC
	}

	return typesCons.ErrUnhandledProposalCase
}

func (m *consensusModule) validateQuorumCertificate(qc *typesCons.QuorumCertificate) error {
	if qc == nil {
		return typesCons.ErrNilQC
	}

	if qc.Block == nil {
		return typesCons.ErrNilBlockInQC
	}

	if qc.ThresholdSignature == nil || len(qc.ThresholdSignature.Signatures) == 0 {
		return typesCons.ErrNilThresholdSigInQC
	}

	msgToJustify := qcToHotstuffMessage(qc)
	valMap := types.GetTestState(nil).ValidatorMap
	numValid := 0
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := valMap[partialSig.Address]
		if !ok {
			m.nodeLog(fmt.Sprintf("[WARN] Validator %d not found in the ValMap but a partial sig was signed by them.", m.ValAddrToIdMap[partialSig.Address]))
			continue
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without re-serializing every time.
		if !isSignatureValid(msgToJustify, validator.PublicKey, partialSig.Signature) {
			m.nodeLog(fmt.Sprintf("[WARN] QC invalid because partial signature from the following node is invalid: %d", m.ValAddrToIdMap[partialSig.Address]))
			continue
		}
		numValid++
	}

	if err := m.isOptimisticThresholdMet(numValid); err != nil {
		return err
	}

	return nil
}

func qcToHotstuffMessage(qc *typesCons.QuorumCertificate) *typesCons.HotstuffMessage {
	return &typesCons.HotstuffMessage{
		Height: qc.Height,
		Step:   qc.Step,
		Round:  qc.Round,
		Block:  qc.Block,
		Justification: &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		},
	}
}
