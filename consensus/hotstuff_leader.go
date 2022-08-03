package consensus

import (
	"encoding/hex"
	"unsafe"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

var (
	LeaderMessageHandler HotstuffMessageHandler = &HotstuffLeaderMessageHandler{}
	leaderHandlers                              = map[typesCons.HotstuffStep]func(*consensusModule, *typesCons.HotstuffMessage){
		NewRound:  LeaderMessageHandler.HandleNewRoundMessage,
		Prepare:   LeaderMessageHandler.HandlePrepareMessage,
		PreCommit: LeaderMessageHandler.HandlePrecommitMessage,
		Commit:    LeaderMessageHandler.HandleCommitMessage,
		Decide:    LeaderMessageHandler.HandleDecideMessage,
	}
)

type HotstuffLeaderMessageHandler struct{}

/*** Prepare Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleNewRoundMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.BlockHeight(),
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_NEW_ROUND,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(NewRound); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(NewRound, err.Error()))
		return
	}

	// TODO(olshansky): Do we need to pause for `MinBlockFreqMSec` here to let more transactions come in?
	m.nodeLog(typesCons.OptimisticVoteCountPassed(NewRound))

	// Likely to be `nil` if blockchain is progressing well.
	highPrepareQC := m.findHighQC(NewRound)

	// TODO(olshansky): Add more unit tests for these checks...
	if highPrepareQC == nil || highPrepareQC.Height < m.Height || highPrepareQC.Round < m.Round {
		block, err := m.prepareBlock()
		if err != nil {
			m.nodeLogError(typesCons.ErrPrepareBlock.Error(), err)
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
		m.nodeLogError(typesCons.ErrCreateProposeMessage(Prepare).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(prepareProposeMessage)

	// Leader also acts like a replica
	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, m.Block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Prepare).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.BlockHeight(),
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_PREPARE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(Prepare); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(Prepare, err.Error()))
		return
	}
	m.nodeLog(typesCons.OptimisticVoteCountPassed(Prepare))

	prepareQC, err := m.getQuorumCertificate(m.Height, Prepare, m.Round)
	if err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(Prepare).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = PreCommit
	m.HighPrepareQC = prepareQC
	m.MessagePool[Prepare] = nil
	m.paceMaker.RestartTimer()

	precommitProposeMessages, err := CreateProposeMessage(m, PreCommit, prepareQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(PreCommit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(precommitProposeMessages)

	// Leader also acts like a replica
	precommitVoteMessage, err := CreateVoteMessage(m, PreCommit, m.Block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(PreCommit).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.BlockHeight(),
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_PRECOMMIT,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(PreCommit); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(PreCommit, err.Error()))
		return
	}
	m.nodeLog(typesCons.OptimisticVoteCountPassed(PreCommit))

	preCommitQC, err := m.getQuorumCertificate(m.Height, PreCommit, m.Round)
	if err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(PreCommit).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Commit
	m.LockedQC = preCommitQC
	m.MessagePool[PreCommit] = nil
	m.paceMaker.RestartTimer()

	commitProposeMessage, err := CreateProposeMessage(m, Commit, preCommitQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(Commit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(commitProposeMessage)

	// Leader also acts like a replica
	commitVoteMessage, err := CreateVoteMessage(m, Commit, m.Block)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Commit).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.BlockHeight(),
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_COMMIT,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(Commit); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(Commit, err.Error()))
		return
	}
	m.nodeLog(typesCons.OptimisticVoteCountPassed(Commit))

	commitQC, err := m.getQuorumCertificate(m.Height, Commit, m.Round)
	if err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(Commit).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Decide
	m.MessagePool[Commit] = nil
	m.paceMaker.RestartTimer()

	decideProposeMessage, err := CreateProposeMessage(m, Decide, commitQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(Decide).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(decideProposeMessage)

	if err := m.commitBlock(m.Block); err != nil {
		m.nodeLogError(typesCons.ErrCommitBlock.Error(), err)
		m.paceMaker.InterruptRound()
		return
	}

	// There is no "replica behavior" to imitate here

	m.paceMaker.NewHeight()
	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.BlockHeight(),
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_STEP_DECIDE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
}

// anteHandle is the general handler called for every before every specific HotstuffLeaderMessageHandler handler
func (handler *HotstuffLeaderMessageHandler) anteHandle(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	if err := handler.validateBasic(m, msg); err != nil {
		return err
	}
	m.aggregateMessage(msg)
	return nil
}

// ValidateBasic general validation checks that apply to every HotstuffLeaderMessage
func (handler *HotstuffLeaderMessageHandler) validateBasic(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	// Discard messages with invalid partial signatures before storing it in the leader's consensus mempool
	if err := m.validatePartialSignature(msg); err != nil {
		return err
	}
	return nil
}

func (m *consensusModule) validatePartialSignature(msg *typesCons.HotstuffMessage) error {
	if msg.Step == NewRound {
		m.nodeLog(typesCons.ErrUnnecessaryPartialSigForNewRound.Error())
		return nil
	}

	if msg.Type == Propose {
		m.nodeLog(typesCons.ErrUnnecessaryPartialSigForLeaderProposal.Error())
		return nil
	}

	if msg.GetPartialSignature() == nil {
		return typesCons.ErrNilPartialSig
	}

	if msg.GetPartialSignature().Signature == nil || len(msg.GetPartialSignature().Address) == 0 {
		return typesCons.ErrNilPartialSigOrSourceNotSpecified
	}

	address := msg.GetPartialSignature().Address
	validator, ok := m.validatorMap[address]
	if !ok {
		return typesCons.ErrMissingValidator(address, m.ValAddrToIdMap[address])
	}
	pubKey := validator.PublicKey
	if isSignatureValid(msg, pubKey, msg.GetPartialSignature().Signature) {
		return nil
	}

	return typesCons.ErrValidatingPartialSig(
		address, m.ValAddrToIdMap[address], msg, hex.EncodeToString(pubKey))
}

func (m *consensusModule) aggregateMessage(msg *typesCons.HotstuffMessage) {
	// TODO(olshansky): Add proper tests for this when we figure out where the mempool should live.
	// NOTE: This is just a placeholder at the moment. It doesn't actually work because SizeOf returns
	// the size of the map pointer, and does not recursively determine the size of all the underlying elements.
	if m.consCfg.MaxMempoolBytes < uint64(unsafe.Sizeof(m.MessagePool)) {
		m.nodeLogError(typesCons.DisregardHotstuffMessage, typesCons.ErrConsensusMempoolFull)
		return
	}

	// Only the leader needs to aggregate consensus related messages.
	m.MessagePool[msg.Step] = append(m.MessagePool[msg.Step], msg)
}
