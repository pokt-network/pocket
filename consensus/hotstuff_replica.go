package consensus

import (
	"encoding/hex"
	"fmt"
	"log"
	"unsafe"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

type HotstuffReplicaMessageHandler struct{}

var (
	ReplicaMessageHandler HotstuffMessageHandler = &HotstuffReplicaMessageHandler{}
	replicaHandlers                              = map[typesCons.HotstuffStep]func(*ConsensusModule, *typesCons.HotstuffMessage){
		NewRound:  ReplicaMessageHandler.HandleNewRoundMessage,
		Prepare:   ReplicaMessageHandler.HandlePrepareMessage,
		PreCommit: ReplicaMessageHandler.HandlePrecommitMessage,
		Commit:    ReplicaMessageHandler.HandleCommitMessage,
		Decide:    ReplicaMessageHandler.HandleDecideMessage,
	}
)

/*** NewRound Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleNewRoundMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	// Clear the previous utility context, if it exists, and create a new one
	if err := m.refreshUtilityContext(); err != nil {
		return
	}

	m.Step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	if err := m.validateProposal(msg); err != nil {
		m.nodeLogError(fmt.Sprintf("Invalid proposal in %s message", Prepare), err)
		m.paceMaker.InterruptRound()
		return
	}

	block := msg.GetBlock()
	if err := m.applyBlock(block); err != nil {
		m.nodeLogError(typesCons.ErrApplyBlock.Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.Block = block

	m.Step = PreCommit

	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Prepare).Error(), err)
		return
	}
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(PreCommit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Commit
	m.HighPrepareQC = quorumCert // INVESTIGATE: Why are we never using this for validation?

	preCommitVoteMessage, err := CreateVoteMessage(m, PreCommit, msg.Block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(PreCommit).Error(), err)
		return
	}
	m.sendToNode(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(Commit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}

	m.Step = Decide
	m.LockedQC = quorumCert // DISCUSS: How does the replica recover if it's locked? Replica `formally` agrees on the QC while the rest of the network `verbally` agrees on the QC.

	commitVoteMessage, err := CreateVoteMessage(m, Commit, msg.Block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Commit).Error(), err)
		return
	}
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(Decide).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}

	if err := m.commitBlock(msg.Block); err != nil {
		m.nodeLogError("Could not commit block", err)
		m.paceMaker.InterruptRound()
		return
	}

	m.paceMaker.NewHeight()
}

// anteHandle is the handler called on every replica message before specific handler
func (handler *HotstuffReplicaMessageHandler) anteHandle(m *ConsensusModule, msg *typesCons.HotstuffMessage) error {
	log.Println("TODO: Hotstuff replica ante handle not implemented yet")
	return nil
}

func (handler *HotstuffReplicaMessageHandler) emitTelemetryEvent(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.CurrentHeight(),
			typesCons.StepToString[msg.GetStep()],
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_REPLICA,
		)
}

func (m *ConsensusModule) validateProposal(msg *typesCons.HotstuffMessage) error {
	if !(msg.GetType() == Propose && msg.GetStep() == Prepare) {
		return typesCons.ErrProposalNotValidInPrepare
	}

	if err := m.validateBlock(msg.Block); err != nil {
		return err
	}

	// TODO(discuss): A nil QC implies a successful CommitQC or TimeoutQC, which have been omitted intentionally since
	// they are not needed for consensus validity. However, if a QC is specified, it must be valid.
	quorumCert :=
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

func (m *ConsensusModule) validateQuorumCertificate(qc *typesCons.QuorumCertificate) error {
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
	numValid := 0
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := m.validatorMap[partialSig.Address]
		if !ok {
			m.nodeLogError(typesCons.ErrMissingValidator(partialSig.Address, m.ValAddrToIdMap[partialSig.Address]).Error(), nil)
			continue
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without re-serializing every time.
		if !isSignatureValid(msgToJustify, validator.GetPublicKey(), partialSig.Signature) {
			m.nodeLog(typesCons.WarnInvalidPartialSigInQC(partialSig.Address, m.ValAddrToIdMap[partialSig.Address]))
			continue
		}
		numValid++
	}

	if err := m.isOptimisticThresholdMet(numValid); err != nil {
		return err
	}

	return nil
}

// This is a helper function intended to be called by a replica/voter during a view change
func (m *ConsensusModule) applyBlock(block *typesCons.Block) error {
	// TODO: Add unit tests to verify this.
	if unsafe.Sizeof(*block) > uintptr(m.MaxBlockBytes) {
		return typesCons.ErrInvalidBlockSize(uint64(unsafe.Sizeof(*block)), m.MaxBlockBytes)
	}

	// TECHDEBT: Retrieve this from persistence
	lastByzValidators := make([][]byte, 0)

	// Apply all the transactions in the block and get the appHash
	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), block.BlockHeader.ProposerAddress, block.Transactions, lastByzValidators)
	if err != nil {
		return err
	}

	// CONSOLIDATE: Terminology of `blockHash`, `appHash` and `stateHash`
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return typesCons.ErrInvalidAppHash(block.BlockHeader.Hash, hex.EncodeToString(appHash))
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
