package consensus

import (
	"fmt"
	"unsafe"

	"github.com/pokt-network/pocket/shared/types"

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
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(NewRound); err != nil {
		m.nodeLog(fmt.Sprintf("Still waiting for more NEWROUND messages; %s", err.Error()))
		return
	}

	// TODO(olshansky): Do we need to pause for `MinBlockFreqMSec` here to let more transactions come in?
	m.nodeLog("Received enough NEWROUND messages!")

	// Likely to be `nil` if blockchain is progressing well.
	highPrepareQC := m.findHighQC(NewRound)

	// TODO(olshansky): Add more unit tests for these checks...
	if highPrepareQC == nil || highPrepareQC.Height < m.Height || highPrepareQC.Round < m.Round {
		block, err := m.prepareBlock()
		if err != nil {
			m.nodeLogError("Could not prepare a block for proposal", err)
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
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(prepareProposeMessage)

	// Leader also acts like a replica
	prepareVoteMessage, err := CreateVoteMessage(m, Prepare, m.Block)
	if err != nil {
		m.nodeLogError("Leader could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(prepareVoteMessage)
}

/*** PreCommit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrepareMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(Prepare); err != nil {
		m.nodeLog(fmt.Sprintf("Still waiting for more PREPARE messages; %s", err.Error()))
		return
	}
	m.nodeLog("Received enough PREPARE votes!")

	prepareQC, err := m.getQuorumCertificate(m.Height, Prepare, m.Round)
	if err != nil {
		m.nodeLogError("Could not get QC for PREPARE step", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = PreCommit
	m.HighPrepareQC = prepareQC
	m.MessagePool[Prepare] = nil
	m.paceMaker.RestartTimer()

	precommitProposeMessages, err := CreateProposeMessage(m, PreCommit, prepareQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(precommitProposeMessages)

	// Leader also acts like a replica
	precommitVoteMessage, err := CreateVoteMessage(m, PreCommit, m.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(precommitVoteMessage)
}

/*** Commit Step ***/

func (handler *HotstuffLeaderMessageHandler) HandlePrecommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(PreCommit); err != nil {
		m.nodeLog(fmt.Sprintf("Still waiting for more PRECOMMIT messages; %s", err.Error()))
		return
	}
	m.nodeLog("received enough PRECOMMIT votes!")

	preCommitQC, err := m.getQuorumCertificate(m.Height, PreCommit, m.Round)
	if err != nil {
		m.nodeLogError("Could not get QC for PRECOMMIT step", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Commit
	m.LockedQC = preCommitQC
	m.MessagePool[PreCommit] = nil
	m.paceMaker.RestartTimer()

	commitProposeMessage, err := CreateProposeMessage(m, Commit, preCommitQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(commitProposeMessage)

	// Leader also acts like a replica
	commitVoteMessage, err := CreateVoteMessage(m, Commit, m.Block)
	if err != nil {
		m.nodeLogError("Could not create a vote message", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}
	m.sendToNode(commitVoteMessage)
}

/*** Decide Step ***/

func (handler *HotstuffLeaderMessageHandler) HandleCommitMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	// TODO(olshansky): add step specific validation
	if err := m.didReceiveEnoughMessageForStep(Commit); err != nil {
		m.nodeLog(fmt.Sprintf("Still waiting for more COMMIT messages; %s", err.Error()))
		return
	}
	m.nodeLog("Received enough COMMIT votes!")

	commitQC, err := m.getQuorumCertificate(m.Height, Commit, m.Round)
	if err != nil {
		m.nodeLogError("Could not get QC for COMMIT step.", err)
		return // TODO(olshansky): Should we interrupt the round here?
	}

	m.Step = Decide
	m.MessagePool[Commit] = nil
	m.paceMaker.RestartTimer()

	decideProposeMessage, err := CreateProposeMessage(m, Decide, commitQC)
	if err != nil {
		m.nodeLogError("Could not create a propose message", err)
		m.paceMaker.InterruptRound()
		return
	}
	m.broadcastToNodes(decideProposeMessage)

	if err := m.commitBlock(m.Block); err != nil {
		m.nodeLogError("Leader could not commit block during DECIDE step", err)
		m.paceMaker.InterruptRound()
		return
	}

	// There is no "replica behaviour" to immitate here

	m.paceMaker.NewHeight()
}

func (handler *HotstuffLeaderMessageHandler) HandleDecideMessage(m *consensusModule, msg *typesCons.HotstuffMessage) {
	if err := handler.anteHandle(m, msg); err != nil {
		m.nodeLogError("Discarding hotstuff message because ante validation failed", err)
		return
	}
	m.nodeLog("[NOOP] Leader does nothing on DECIDE message.")
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

	valMap := types.GetTestState(nil).ValidatorMap
	address := msg.GetPartialSignature().Address
	validator, ok := valMap[address]
	if !ok {
		return typesCons.ErrMissingValidator(address, uint64(m.ValAddrToIdMap[address]))
	}
	pubKey := validator.PublicKey
	if isSignatureValid(msg, pubKey, msg.GetPartialSignature().Signature) {
		return nil
	}

	return typesCons.ErrValidatingPartialSig(
		address, uint64(m.ValAddrToIdMap[address]), uint64(msg.Height), uint64(msg.Round), StepToString[msg.Step],
		string(msg.GetPartialSignature().Signature), protoHash(msg.Block), pubKey.String())
}

func (m *consensusModule) aggregateMessage(msg *typesCons.HotstuffMessage) {
	// TODO(olshansky): Add proper tests for this when we figure out where the mempool should live.
	// NOTE: This is just a placeholder at the moment. It doesn't actually work because SizeOf returns
	// the size of the map pointer, and does not recursively determine the size of all the underlying elements.
	if m.consCfg.MaxMempoolBytes < uint64(unsafe.Sizeof(m.MessagePool)) {
		m.nodeLogError("Discarding hotstuff message because the mempool is full", fmt.Errorf("mempool is full"))
		return
	}

	// Only the leader needs to aggregate consensus related messages.
	m.MessagePool[msg.Step] = append(m.MessagePool[msg.Step], msg)
}
