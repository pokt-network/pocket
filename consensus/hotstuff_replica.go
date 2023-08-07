package consensus

import (
	"fmt"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
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

	if err := handler.isMessageValidBasic(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	// Clear the previous utility unitOfWork, if it exists, and create a new one
	if err := m.refreshUtilityUnitOfWork(); err != nil {
		m.logger.Error().Err(err).Msg("Could not refresh utility unitOfWork")
		return
	}

	m.step = Prepare
}

/*** Prepare Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.isMessageValidBasic(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	if err := m.validateProposal(msg); err != nil {
		m.logger.Error().Err(err).Str("message", Prepare.String()).Msg("Invalid proposal")
		m.paceMaker.InterruptRound("invalid proposal")
		return
	}

	block := msg.GetBlock()
	if err := m.applyBlock(block); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrApplyBlock.Error())
		m.paceMaker.InterruptRound("failed to apply block")
		return
	}
	m.block = block
	m.step = PreCommit

	prepareVoteMessage, err := CreateVoteMessage(m.height, m.round, Prepare, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(Prepare).Error())
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToLeader(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.isMessageValidBasic(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(PreCommit).Error())
		m.paceMaker.InterruptRound("invalid quorum certificate")
		return
	}

	m.step = Commit
	m.prepareQC = quorumCert // INVESTIGATE: Why are we never using this for validation?

	preCommitVoteMessage, err := CreateVoteMessage(m.height, m.round, PreCommit, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(PreCommit).Error())
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToLeader(preCommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.isMessageValidBasic(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(Commit).Error())
		m.paceMaker.InterruptRound("invalid quorum certificate")
		return
	}

	m.step = Decide
	m.lockedQC = quorumCert // DISCUSS: How does the replica recover if it's locked? Replica `formally` agrees on the QC while the rest of the network `verbally` agrees on the QC.

	commitVoteMessage, err := CreateVoteMessage(m.height, m.round, Commit, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(Commit).Error())
		return // Not interrupting the round because liveness could continue with one failed vote
	}
	m.sendToLeader(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffReplicaMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.isMessageValidBasic(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	quorumCert := msg.GetQuorumCertificate()
	if err := m.validateQuorumCertificate(quorumCert); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(Decide).Error())
		m.paceMaker.InterruptRound("invalid quorum certificate")
		return
	}

	quorumCertBytes, err := codec.GetCodec().Marshal(quorumCert)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to convert the quorum certificate to bytes")
		return
	}

	m.block.BlockHeader.QuorumCertificate = quorumCertBytes

	if err := m.commitBlock(m.block); err != nil {
		m.logger.Error().Err(err).Msg("Could not commit block")
		m.paceMaker.InterruptRound("failed to commit block")
	return
	}

	m.paceMaker.NewHeight()
}

// isMessageValidBasic is the handler called on every replica message before specific handler
func (handler *HotstuffReplicaMessageHandler) isMessageValidBasic(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	// Basic block metadata validation
	if valid, err := m.isBlockInMessageValidBasic(msg); !valid {
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
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT,
			m.CurrentHeight(),
			typesCons.StepToString[msg.GetStep()],
			m.CurrentRound(),
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
		m.logger.Info().Msg(typesCons.NotLockedOnQC)
		return nil
	}

	// Safety: check the hash of the locked QC
	// The equivalent of `lockedQC.Block.ExtendsFrom(justifyQC.Block)` in the hotstuff whitepaper is done in `applyBlock` below.
	if protoHash(lockedQC.GetBlock()) == protoHash(justifyQC.Block) {
		m.logger.Info().Msg(typesCons.ProposalBlockExtends)
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
func (m *consensusModule) applyBlock(block *coreTypes.Block) error {
	blockHeader := block.BlockHeader
	utilityUnitOfWork := m.utilityUnitOfWork
	if utilityUnitOfWork == nil {
		return fmt.Errorf("utility unit of work is nil")
	}

	// Set the proposal block in the persistence context
	if err := utilityUnitOfWork.SetProposalBlock(blockHeader.StateHash, blockHeader.ProposerAddress, block.Transactions); err != nil {
		return err
	}

	// Apply all the transactions in the block
	if err := utilityUnitOfWork.ApplyBlock(); err != nil {
		return err
	}

	stateHash := utilityUnitOfWork.GetStateHash()

	if blockHeader.StateHash != stateHash {
		return typesCons.ErrInvalidStateHash(blockHeader.StateHash, stateHash)
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

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return err
	}

	actorMapper := typesCons.NewActorMapper(validators)
	validatorMap := actorMapper.GetValidatorMap()
	valAddrToIdMap := actorMapper.GetValAddrToIdMap()

	// TODO(#109): Aggregate signatures once BLS or DKG is implemented
	for _, partialSig := range qc.ThresholdSignature.Signatures {
		validator, ok := validatorMap[partialSig.Address]
		if !ok {
			m.logger.Error().Msgf(typesCons.ErrMissingValidator(partialSig.Address, valAddrToIdMap[partialSig.Address]).Error())
			continue
		}
		// TODO(olshansky): Every call to `IsSignatureValid` does a serialization and should be optimized. We can
		// just serialize `Message` once and verify each signature without re-serializing every time.
		if !isSignatureValid(msgToJustify, validator.GetPublicKey(), partialSig.Signature) {
			m.logger.Warn().Fields(map[string]any{
				"address": partialSig.Address,
				"nodeId":  valAddrToIdMap[partialSig.Address],
			}).Msg("QC contains an invalid partial signature")
			continue
		}
		numValid++
	}
	if err := m.validateOptimisticThresholdMet(numValid, validators); err != nil {
		return err
	}

	return nil
}

func isNodeLockedOnPastQC(justifyQC, lockedQC *typesCons.QuorumCertificate) (bool, error) {
	if isLockedOnPastHeight(justifyQC, lockedQC) {
		return true, typesCons.ErrNodeLockedPastHeight
	} else if isLockedOnCurrHeightAndPastRound(justifyQC, lockedQC) {
		return true, typesCons.ErrNodeLockedPastHeight
	}
	return false, nil
}

func isLockedOnPastHeight(justifyQC, lockedQC *typesCons.QuorumCertificate) bool {
	return justifyQC.Height > lockedQC.Height
}

func isLockedOnCurrHeightAndPastRound(justifyQC, lockedQC *typesCons.QuorumCertificate) bool {
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
