package consensus

import (
	"encoding/hex"
	"unsafe"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
)

type HotstuffLeaderMessageHandler struct{}

var (
	LeaderMessageHandler HotstuffMessageHandler = &HotstuffLeaderMessageHandler{}
	leaderHandlers                              = map[typesCons.HotstuffStep]func(*ConsensusModule, *typesCons.HotstuffMessage){
		NewRound:  LeaderMessageHandler.HandleNewRoundMessage,
		Prepare:   LeaderMessageHandler.HandlePrepareMessage,
		PreCommit: LeaderMessageHandler.HandlePrecommitMessage,
		Commit:    LeaderMessageHandler.HandleCommitMessage,
		Decide:    LeaderMessageHandler.HandleDecideMessage,
	}
)

/*** Prepare Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleNewRoundMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	// DISCUSS: Do we need to pause for `MinBlockFreqMSec` here to let more transactions or should we stick with optimistic responsiveness?

	if err := m.didReceiveEnoughMessageForStep(NewRound); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(NewRound, err.Error()))
		return
	}
	m.nodeLog(typesCons.OptimisticVoteCountPassed(NewRound))

	// Clear the previous utility context, if it exists, and create a new one
	if err := m.refreshUtilityContext(); err != nil {
		return
	}

	// Likely to be `nil` if blockchain is progressing well.
	highPrepareQC := m.findHighQC(NewRound) // TECHDEBT: How do we validate `highPrepareQC` here?

	// TODO: Add more unit tests for these checks...
	if highPrepareQC == nil || highPrepareQC.Height < m.Height || highPrepareQC.Round < m.Round {
		// Leader prepares a new block if `highPrepareQC` is not applicable
		block, err := m.prepareAndApplyBlock()
		if err != nil {
			m.nodeLogError(typesCons.ErrPrepareBlock.Error(), err)
			m.paceMaker.InterruptRound()
			return
		}
		m.Block = block
	} else {
		// Leader acts like a replica if `highPrepareQC` is not `nil`
		if err := m.applyBlock(highPrepareQC.Block); err != nil {
			m.nodeLogError(typesCons.ErrApplyBlock.Error(), err)
			m.paceMaker.InterruptRound()
			return
		}
		m.Block = highPrepareQC.Block
	}

	m.Step = Prepare
	m.MessagePool[NewRound] = nil

	prepareProposeMessage, err := CreateProposeMessage(m.Height, m.Round, Prepare, m.Block, highPrepareQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(Prepare).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(prepareProposeMessage)

	// Leader also acts like a replica
	prepareVoteMessage, err := CreateVoteMessage(m.Height, m.Round, Prepare, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Prepare).Error(), err)
		return
	}
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

	if err := m.didReceiveEnoughMessageForStep(Prepare); err != nil {
		m.nodeLog(typesCons.OptimisticVoteCountWaiting(Prepare, err.Error()))
		return
	}
	m.nodeLog(typesCons.OptimisticVoteCountPassed(Prepare))

	// DISCUSS: What prevents leader from swapping out the block here?
	prepareQC, err := m.getQuorumCertificate(m.Height, Prepare, m.Round)
	if err != nil {
		m.nodeLogError(typesCons.ErrQCInvalid(Prepare).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = PreCommit
	m.HighPrepareQC = prepareQC
	m.MessagePool[Prepare] = nil

	precommitProposeMessages, err := CreateProposeMessage(m.Height, m.Round, PreCommit, m.Block, prepareQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(PreCommit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(precommitProposeMessages)

	// Leader also acts like a replica
	precommitVoteMessage, err := CreateVoteMessage(m.Height, m.Round, PreCommit, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(PreCommit).Error(), err)
		return
	}
	m.sendToNode(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

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

	commitProposeMessage, err := CreateProposeMessage(m.Height, m.Round, Commit, m.Block, preCommitQC)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateProposeMessage(Commit).Error(), err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(commitProposeMessage)

	// Leader also acts like a replica
	commitVoteMessage, err := CreateVoteMessage(m.Height, m.Round, Commit, m.Block, m.privateKey)
	if err != nil {
		m.nodeLogError(typesCons.ErrCreateVoteMessage(Commit).Error(), err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}

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

	decideProposeMessage, err := CreateProposeMessage(m.Height, m.Round, Decide, m.Block, commitQC)
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

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	handler.emitTelemetryEvent(m, msg)
	defer m.paceMaker.RestartTimer()

	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError(typesCons.ErrHotstuffValidation.Error(), err)
		return
	}
}

// anteHandle is the general handler called for every before every specific HotstuffLeaderMessageHandler handler
func (handler *HotstuffLeaderMessageHandler) anteHandle(m *ConsensusModule, msg *typesCons.HotstuffMessage) error {
	if err := handler.validateBasic(m, msg); err != nil {
		return err
	}
	m.aggregateMessage(msg)
	return nil
}

func (handler *HotstuffLeaderMessageHandler) emitTelemetryEvent(m *ConsensusModule, msg *typesCons.HotstuffMessage) {
	m.GetBus().
		GetTelemetryModule().
		GetEventMetricsAgent().
		EmitEvent(
			consensusTelemetry.CONSENSUS_EVENT_METRICS_NAMESPACE,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_NAME,
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_HEIGHT, m.CurrentHeight(),
			typesCons.StepToString[msg.GetStep()],
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)
}

// ValidateBasic general validation checks that apply to every HotstuffLeaderMessage
func (handler *HotstuffLeaderMessageHandler) validateBasic(m *ConsensusModule, msg *typesCons.HotstuffMessage) error {
	// Discard messages with invalid partial signatures before storing it in the leader's consensus mempool
	if err := m.validatePartialSignature(msg); err != nil {
		return err
	}
	return nil
}

func (m *ConsensusModule) validatePartialSignature(msg *typesCons.HotstuffMessage) error {
	if msg.GetStep() == NewRound {
		m.nodeLog(typesCons.ErrUnnecessaryPartialSigForNewRound.Error())
		return nil
	}

	if msg.GetType() == Propose {
		m.nodeLog(typesCons.ErrUnnecessaryPartialSigForLeaderProposal.Error())
		return nil
	}

	if msg.GetPartialSignature() == nil {
		return typesCons.ErrNilPartialSig
	}
	partialSig := msg.GetPartialSignature()

	if partialSig.Signature == nil || len(partialSig.GetAddress()) == 0 {
		return typesCons.ErrNilPartialSigOrSourceNotSpecified
	}

	address := partialSig.GetAddress()
	validator, ok := m.validatorMap[address]
	if !ok {
		return typesCons.ErrMissingValidator(address, m.ValAddrToIdMap[address])
	}
	pubKey := validator.GetPublicKey()
	if isSignatureValid(msg, pubKey, partialSig.GetSignature()) {
		return nil
	}

	return typesCons.ErrValidatingPartialSig(
		address, m.ValAddrToIdMap[address], msg, pubKey)
}

func (m *ConsensusModule) aggregateMessage(msg *typesCons.HotstuffMessage) {
	// TODO(olshansky): Add proper tests for this when we figure out where the mempool should live.
	// NOTE: This is just a placeholder at the moment. It doesn't actually work because SizeOf returns
	// the size of the map pointer, and does not recursively determine the size of all the underlying elements.
	if m.consCfg.GetMaxMempoolBytes() < uint64(unsafe.Sizeof(m.MessagePool)) {
		m.nodeLogError(typesCons.DisregardHotstuffMessage, typesCons.ErrConsensusMempoolFull)
		return
	}

	// Only the leader needs to aggregate consensus related messages.
	m.MessagePool[msg.Step] = append(m.MessagePool[msg.Step], msg)
}

// This is a helper function intended to be called by a leader/validator during a view change
// to prepare a new block that is applied to the new underlying context.
func (m *ConsensusModule) prepareAndApplyBlock() (*typesCons.Block, error) {
	if m.isReplica() {
		return nil, typesCons.ErrReplicaPrepareBlock
	}

	// TECHDEBT: Retrieve this from consensus consensus config
	maxTxBytes := 90000

	// TECHDEBT: Retrieve this from persistence
	lastByzValidators := make([][]byte, 0)

	// Reap the mempool for transactions to be applied in this block
	txs, err := m.utilityContext.GetProposalTransactions(m.privateKey.Address(), maxTxBytes, lastByzValidators)
	if err != nil {
		return nil, err
	}

	// Apply all the transactions in the block
	appHash, err := m.utilityContext.ApplyBlock(int64(m.Height), m.privateKey.Address(), txs, lastByzValidators)
	if err != nil {
		return nil, err
	}

	// Construct the block
	blockHeader := &typesCons.BlockHeader{
		Height:            int64(m.Height),
		Hash:              hex.EncodeToString(appHash),
		NumTxs:            uint32(len(txs)),
		LastBlockHash:     m.lastAppHash,
		ProposerAddress:   m.privateKey.Address().Bytes(),
		QuorumCertificate: []byte("HACK: Temporary placeholder"),
	}
	block := &typesCons.Block{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	return block, nil
}
