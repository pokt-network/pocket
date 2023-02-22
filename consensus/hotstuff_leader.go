package consensus

import (
	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// CONSOLIDATE: Last/Prev & AppHash/StateHash/BlockHash

type HotstuffLeaderMessageHandler struct{}

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

/*** Prepare Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleNewRoundMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.logger.Error().Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	// DISCUSS: Do we need to pause for `MinBlockFreqMSec` here to let more transactions or should we stick with optimistic responsiveness?

	if err := m.didReceiveEnoughMessageForStep(NewRound); err != nil {
		m.logger.Info().Msg(typesCons.OptimisticVoteCountWaiting(NewRound, err.Error()))
		return
	}

	m.logger.Info().Fields(
		map[string]any{
			"height": m.height,
			"round":  m.round,
			"step":   NewRound,
		},
	).Msg("📬 Received enough 📬 votes")

	// Clear the previous utility context, if it exists, and create a new one
	if err := m.refreshUtilityContext(); err != nil {
		m.logger.Error().Err(err).Msg("Could not refresh utility context")
		return
	}

	// Likely to be `nil` if blockchain is progressing well.
	// TECHDEBT: How do we properly validate `prepareQC` here?
	// CONSIDERATION(M5): could this be improved by incrementally keeping track of highQC when we add/remove messages to the mempool? Probably premature optimization for now but something to keep in mind.
	highPrepareQC := m.findHighQC(m.hotstuffMempool[NewRound].GetAll())

	// TODO: Add test to make sure same block is not applied twice if round is interrupted after being 'Applied'.
	// TODO: Add more unit tests for these checks...
	if m.shouldPrepareNewBlock(highPrepareQC) {
		block, err := m.prepareAndApplyBlock(highPrepareQC)
		if err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrPrepareBlock.Error())
			m.paceMaker.InterruptRound("failed to prepare & apply block")
			return
		}
		m.block = block
	} else {
		// Leader acts like a replica if `prepareQC` is not `nil`
		// TODO: Do we need to call `validateProposal` here similar to how replicas does it
		if err := m.applyBlock(highPrepareQC.Block); err != nil {
			m.logger.Error().Err(err).Msg(typesCons.ErrApplyBlock.Error())
			m.paceMaker.InterruptRound("failed to apply block")
			return
		}
		m.block = highPrepareQC.Block
	}

	m.step = Prepare
	m.hotstuffMempool[NewRound].Clear()

	prepareProposeMessage, err := CreateProposeMessage(m.height, m.round, Prepare, m.block, highPrepareQC)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateProposeMessage(Prepare).Error())
		m.paceMaker.InterruptRound("failed to create propose message")
		return
	}
	m.broadcastToValidators(prepareProposeMessage)

	// Leader also acts like a replica
	prepareVoteMessage, err := CreateVoteMessage(m.height, m.round, Prepare, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(Prepare).Error())
		return
	}
	m.sendToLeader(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.logger.Error().Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	if err := m.didReceiveEnoughMessageForStep(Prepare); err != nil {
		m.logger.Info().Msg(typesCons.OptimisticVoteCountWaiting(Prepare, err.Error()))
		return
	}

	m.logger.Info().Fields(
		map[string]any{
			"height": m.height,
			"round":  m.round,
			"step":   Prepare,
		},
	).Msg("📬 Received enough 📬 votes")

	prepareQC, err := m.getQuorumCertificate(m.height, Prepare, m.round)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(Prepare).Error())
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.step = PreCommit
	m.prepareQC = prepareQC
	m.hotstuffMempool[Prepare].Clear()

	preCommitProposeMessage, err := CreateProposeMessage(m.height, m.round, PreCommit, m.block, prepareQC)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateProposeMessage(PreCommit).Error())
		m.paceMaker.InterruptRound("failed to create propose message")
		return
	}
	m.broadcastToValidators(preCommitProposeMessage)

	// Leader also acts like a replica
	precommitVoteMessage, err := CreateVoteMessage(m.height, m.round, PreCommit, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(PreCommit).Error())
		return
	}
	m.sendToLeader(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	if err := m.didReceiveEnoughMessageForStep(PreCommit); err != nil {
		m.logger.Info().Msg(typesCons.OptimisticVoteCountWaiting(PreCommit, err.Error()))
		return
	}

	m.logger.Info().Fields(
		map[string]any{
			"height": m.height,
			"round":  m.round,
			"step":   PreCommit,
		},
	).Msg("📬 Received enough 📬 votes")

	preCommitQC, err := m.getQuorumCertificate(m.height, PreCommit, m.round)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(PreCommit).Error())
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.step = Commit
	m.lockedQC = preCommitQC
	m.hotstuffMempool[PreCommit].Clear()

	commitProposeMessage, err := CreateProposeMessage(m.height, m.round, Commit, m.block, preCommitQC)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateProposeMessage(Commit).Error())
		m.paceMaker.InterruptRound("failed to create propose message")
		return
	}
	m.broadcastToValidators(commitProposeMessage)

	// Leader also acts like a replica
	commitVoteMessage, err := CreateVoteMessage(m.height, m.round, Commit, m.block, m.privateKey)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateVoteMessage(Commit).Error())
		return
	}
	m.sendToLeader(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}

	if err := m.didReceiveEnoughMessageForStep(Commit); err != nil {
		m.logger.Info().Msg(typesCons.OptimisticVoteCountWaiting(Commit, err.Error()))
		return
	}

	m.logger.Info().Fields(
		map[string]any{
			"height": m.height,
			"round":  m.round,
			"step":   Commit,
		},
	).Msg("📬 Received enough 📬 votes")

	commitQC, err := m.getQuorumCertificate(m.height, Commit, m.round)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrQCInvalid(Commit).Error())
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.step = Decide
	m.hotstuffMempool[Commit].Clear()

	decideProposeMessage, err := CreateProposeMessage(m.height, m.round, Decide, m.block, commitQC)
	if err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCreateProposeMessage(Decide).Error())
		m.paceMaker.InterruptRound("failed to create propose message")
		return
	}
	m.broadcastToValidators(decideProposeMessage)

	if err := m.commitBlock(m.block); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrCommitBlock.Error())
		m.paceMaker.InterruptRound("failed to commit block")
		return
	}

	// There is no "replica behavior" to imitate here because the leader already committed the block proposal.

	m.paceMaker.NewHeight()
	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	defer m.paceMaker.RestartTimer()
	handler.emitTelemetryEvent(m, msg)

	if err := handler.anteHandle(m, msg); err != nil {
		m.logger.Error().Err(err).Msg(typesCons.ErrHotstuffValidation.Error())
		return
	}
}

// anteHandle is the general handler called for every before every specific HotstuffLeaderMessageHandler handler
func (handler *HotstuffLeaderMessageHandler) anteHandle(m *consensusModule, msg *typesCons.HotstuffMessage) error {
	// Basic block metadata validation

	if valid, err := m.isValidMessageBlock(msg); !valid {
		return err
	}

	// Discard messages with invalid partial signatures before storing it in the leader's consensus mempool
	if err := m.validateMessageSignature(msg); err != nil {
		return err
	}

	// Index the hotstuff message in the consensus mempool
	if err := m.indexHotstuffMessage(msg); err != nil {
		return err
	}

	return nil
}

func (handler *HotstuffLeaderMessageHandler) emitTelemetryEvent(m *consensusModule, msg *typesCons.HotstuffMessage) {
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
			consensusTelemetry.HOTPOKT_MESSAGE_EVENT_METRIC_LABEL_VALIDATOR_TYPE_LEADER,
		)
}

func (m *consensusModule) validateMessageSignature(msg *typesCons.HotstuffMessage) error {
	partialSig := msg.GetPartialSignature()

	if msg.GetStep() == NewRound {
		if partialSig != nil {
			m.logger.Error().Err(typesCons.ErrUnnecessaryPartialSigForNewRound)
		}
		return nil
	}

	if msg.GetType() == Propose {
		if partialSig != nil {
			m.logger.Error().Err(typesCons.ErrUnnecessaryPartialSigForLeaderProposal)
		}
		return nil
	}

	if partialSig == nil {
		return typesCons.ErrNilPartialSig
	}

	if partialSig.Signature == nil || partialSig.GetAddress() == "" {
		return typesCons.ErrNilPartialSigOrSourceNotSpecified
	}

	address := partialSig.GetAddress()

	validators, err := m.getValidatorsAtHeight(m.CurrentHeight())
	if err != nil {
		return err
	}

	actorMapper := typesCons.NewActorMapper(validators)
	validatorMap := actorMapper.GetValidatorMap()
	valAddrToIdMap := actorMapper.GetValAddrToIdMap()

	validator, ok := validatorMap[address]
	if !ok {
		return typesCons.ErrMissingValidator(address, valAddrToIdMap[address])
	}
	pubKey := validator.GetPublicKey()
	if isSignatureValid(msg, pubKey, partialSig.GetSignature()) {
		return nil
	}

	return typesCons.ErrValidatingPartialSig(
		address, valAddrToIdMap[address], msg, pubKey)
}

// TODO(#388): Utilize the shared mempool implementation for consensus messages.
//
//	It doesn't actually work because SizeOf returns the size of the map pointer,
//	and does not recursively determine the size of all the underlying elements
//	Add proper tests and implementation once the mempool is implemented.
func (m *consensusModule) indexHotstuffMessage(msg *typesCons.HotstuffMessage) error {
	if m.consCfg.MaxMempoolBytes < uint64(m.hotstuffMempool[typesCons.HotstuffStep(msg.Type)].TotalMsgBytes()) {
		m.logger.Error().Err(typesCons.ErrConsensusMempoolFull).Msg(typesCons.DisregardHotstuffMessage)
		return typesCons.ErrConsensusMempoolFull
	}

	// Only the leader needs to aggregate consensus related messages.
	step := msg.GetStep()
	if err := m.hotstuffMempool[step].Push(msg); err != nil {
		return err
	}

	return nil
}

// This is a helper function intended to be called by a leader/validator during a view change
// to prepare a new block that is applied to the new underlying context.
// TODO: Split this into atomic & functional `prepareBlock` and `applyBlock` methods
func (m *consensusModule) prepareAndApplyBlock(qc *typesCons.QuorumCertificate) (*coreTypes.Block, error) {
	if m.isReplica() {
		return nil, typesCons.ErrReplicaPrepareBlock
	}

	// TECHDEBT: Retrieve this from consensus consensus config
	maxTxBytes := 90000

	// Reap the mempool for transactions to be applied in this block
	stateHash, txs, err := m.utilityContext.CreateAndApplyProposalBlock(m.privateKey.Address(), maxTxBytes)
	if err != nil {
		return nil, err
	}

	prevHeight := int64(m.height) - 1
	readCtx, err := m.GetBus().GetPersistenceModule().NewReadContext(prevHeight)
	if err != nil {
		return nil, err
	}

	// IMPROVE: This data can be read via an ephemeral read context - no need to use the utility's persistence context
	prevBlockHash, err := readCtx.GetBlockHash(prevHeight)
	if err != nil {
		return nil, err
	}

	qcBytes, err := codec.GetCodec().Marshal(qc)
	if err != nil {
		return nil, err
	}

	// Construct the block
	blockHeader := &coreTypes.BlockHeader{
		Height:            m.height,
		StateHash:         stateHash,
		PrevStateHash:     prevBlockHash,
		NumTxs:            uint32(len(txs)),
		ProposerAddress:   m.privateKey.Address().Bytes(),
		QuorumCertificate: qcBytes,
	}
	block := &coreTypes.Block{
		BlockHeader:  blockHeader,
		Transactions: txs,
	}

	// Set the proposal block in the persistence context
	if err := m.utilityContext.SetProposalBlock(blockHeader.StateHash, blockHeader.ProposerAddress, block.Transactions); err != nil {
		return nil, err
	}

	return block, nil
}

// Return true if this node, the leader, should prepare a new block.
// ADDTEST: Add more tests for all the different scenarios here
func (m *consensusModule) shouldPrepareNewBlock(highPrepareQC *typesCons.QuorumCertificate) bool {
	if highPrepareQC == nil {
		m.logger.Info().Msg("Preparing a new block - no prepareQC found")
		return true
	} else if m.isPrepareQCFromPast(highPrepareQC) {
		m.logger.Info().Msg("Preparing a new block - prepareQC is from the past")
		return true
	} else if highPrepareQC.Block == nil {
		m.logger.Warn().Msg("Preparing a new block - prepareQC SHOULD be used but block is nil")
		return true
	}
	return false
}

// The `prepareQC` is from the past so we can safely ignore it
func (m *consensusModule) isPrepareQCFromPast(highPrepareQC *typesCons.QuorumCertificate) bool {
	return highPrepareQC.Height < m.height || highPrepareQC.Round < m.round
}
