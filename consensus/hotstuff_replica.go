package consensus

import (
	"encoding/hex"
	"fmt"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	"github.com/pokt-network/pocket/consensus/types"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
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
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	// Clear the previous utility context, if it exists, and create a new one
	if err := m.refreshUtilityContext(); err != nil {
		m.nodeLogError("Could not refresh utility context", err)
		return
	}

	m.Step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

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

	prepareVoteMessage, err := CreateVoteMessage(m.Height, m.Round, Prepare, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Prepare).Error(), err)
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

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
	m.highPrepareQC = quorumCert // INVESTIGATE: Why are we never using this for validation?

	preCommitVoteMessage, err := CreateVoteMessage(m.Height, m.Round, PreCommit, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(PreCommit).Error(), err)
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToNode(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

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
	m.lockedQC = quorumCert // DISCUSS: How does the replica recover if it's locked? Replica `formally` agrees on the QC while the rest of the network `verbally` agrees on the QC.

	commitVoteMessage, err := CreateVoteMessage(m.Height, m.Round, Commit, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Commit).Error(), err)
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

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

	if err := m.commitBlock(m.Block); err != nil {
		m.nodeLogError("Could not commit block", err)
		m.paceMaker.InterruptRound()
		return
	}

	m.paceMaker.NewHeight()
}

// anteHandle is the handler called on every replica message before specific handler
func (handler *HotstuffReplicaMessageHandler) anteHandle(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	// Basic block metadata validation
	if err := m.validateBlockBasic(msg.GetBlock()); err != nil {
		return err
	}

	return nil
}

func (handler *HotstuffReplicaMessageHandler) emitTelemetryEvent(m *consensusModule, msg *typesCons.HotstuffMessage) {
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

func (m *consensusModule) validateProposal(msg *typesCons.HotstuffMessage) error {
	// Check if node should be accepting proposals
	if !(msg.GetType() == Propose && msg.GetStep() == Prepare) {
		return typesCons.ErrProposalNotValidInPrepare
	}

	quorumCert := msg.GetQuorumCertificate()
	// A nil QC implies a successful CommitQC or TimeoutQC, which have been omitted intentionally
	// since they are not needed for consensus validity. However, if a QC is specified, it must be valid.
	if quorumCert != nil {
		if err := m.validateQuorumCertificate(quorumCert); err != nil {
			return err
		}
	}

	lockedQC := m.lockedQC
	justifyQC := quorumCert

	// Safety: not locked
	if lockedQC == nil {
		m.nodeLog(typesCons.NotLockedOnQC)
		return nil
	}

	// Safety: check the hash of the locked QC
	// The equivalent of `lockedQC.Block.ExtendsFrom(justifyQC.Block)` in the hotstuff whitepaper is done in `applyBlock` below.
	if protoHash(lockedQC.GetBlock()) == protoHash(justifyQC.Block) {
		m.nodeLog(typesCons.ProposalBlockExtends)
		return nil
	}

	// Liveness: is node locked on a QC from the past?
	// DISCUSS: Where should additional logic be added to unlock the node?
	if isLocked, err := isNodeLockedOnPastQC(justifyQC, lockedQC); isLocked {
		return err
	}

	return typesCons.ErrUnhandledProposalCase
}

// This helper applies the block metadata to the utility & persistence layers
func (m *consensusModule) applyBlock(block *typesCons.Block) error {
	blockProtoBz, err := codec.GetCodec().Marshal(block)
	if err != nil {
		return err
	}
	persistenceContext := m.utilityContext.GetPersistenceContext()
	// Set the proposal block in the persistence context
	if err = persistenceContext.SetProposalBlock(block.BlockHeader.Hash, blockProtoBz, block.BlockHeader.ProposerAddress, block.Transactions); err != nil {
		return err
	}

	// Apply all the transactions in the block and get the appHash
	appHash, err := m.utilityContext.ApplyBlock()
	if err != nil {
		return err
	}

	// CONSOLIDATE: Terminology of `appHash` and `stateHash`
	if block.BlockHeader.Hash != hex.EncodeToString(appHash) {
		return typesCons.ErrInvalidAppHash(block.BlockHeader.Hash, hex.EncodeToString(appHash))
	}

	return nil
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
	numValid := 0

	// TODO(#109): Aggregate signatures once BLS or DKG is implemented
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := m.validatorMap[partialSig.Address]
		if !ok {
			m.nodeLogError(typesCons.ErrMissingValidator(partialSig.Address, m.valAddrToIdMap[partialSig.Address]).Error(), nil)
			continue
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without re-serializing every time.
		if !isSignatureValid(msgToJustify, validator.GetPublicKey(), partialSig.Signature) {
			m.nodeLog(typesCons.WarnInvalidPartialSigInQC(partialSig.Address, m.valAddrToIdMap[partialSig.Address]))
			continue
		}
		numValid++
	}
	if err := m.isOptimisticThresholdMet(numValid); err != nil {
		return err
	}

	return nil
}

func isNodeLockedOnPastQC(justifyQC, lockedQC *types.QuorumCertificate) (bool, error) {
	if isLockedOnPastHeight(justifyQC, lockedQC) {
		return true, types.ErrNodeLockedPastHeight
	} else if isLockedOnCurrHeightAndPastRound(justifyQC, lockedQC) {
		return true, types.ErrNodeLockedPastHeight
	}
	return false, nil
}

func isLockedOnPastHeight(justifyQC, lockedQC *types.QuorumCertificate) bool {
	return justifyQC.Height > lockedQC.Height
}

func isLockedOnCurrHeightAndPastRound(justifyQC, lockedQC *types.QuorumCertificate) bool {
	return justifyQC.Height == lockedQC.Height && justifyQC.Round > lockedQC.Round
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
