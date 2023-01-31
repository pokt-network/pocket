package pacemaker

import (
	"context"
	"fmt"
	"log"
	"time"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	defaultLogPrefix    = "NODE"
	pacemakerModuleName = "pacemaker"

	// A buffer around the pacemaker timeout to avoid race condition; 30ms was arbitrarily chosen
	timeoutBuffer = 30 * time.Millisecond

	newRound = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	propose  = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
)

var (
	_ modules.Module = &pacemaker{}
	_ PacemakerDebug = &pacemaker{}
	_ Pacemaker      = &pacemaker{}
)

type Pacemaker interface {
	modules.Module
	PacemakerDebug

	ShouldHandleMessage(message *typesCons.HotstuffMessage) (bool, error)
	SetLogPrefix(string)

	RestartTimer()
	NewHeight()
	InterruptRound(reason string)
}

type pacemaker struct {
	bus            modules.Bus
	pacemakerCfg   *configs.PacemakerConfig
	stepCancelFunc context.CancelFunc

	// Only used for development and debugging.
	debug pacemakerDebug

	// REFACTOR: this should be removed, when we build a shared and proper logger
	logPrefix string
}

func CreatePacemaker(bus modules.Bus) (modules.Module, error) {
	var m pacemaker
	return m.Create(bus)
}

func (*pacemaker) Create(bus modules.Bus) (modules.Module, error) {
	m := &pacemaker{
		logPrefix: defaultLogPrefix,
	}

	if err := bus.RegisterModule(m); err != nil {
		return nil, err
	}

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()

	m.pacemakerCfg = cfg.Consensus.PacemakerConfig
	m.debug = pacemakerDebug{
		manualMode:                m.pacemakerCfg.GetManual(),
		debugTimeBetweenStepsMsec: m.pacemakerCfg.GetDebugTimeBetweenStepsMsec(),
		quorumCertificate:         nil,
	}

	return m, nil
}

func (m *pacemaker) Start() error {
	m.RestartTimer()
	return nil
}
func (*pacemaker) Stop() error {
	return nil
}

func (*pacemaker) GetModuleName() string {
	return pacemakerModuleName
}

func (m *pacemaker) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *pacemaker) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *pacemaker) SetLogPrefix(logPrefix string) {
	m.logPrefix = logPrefix
}

func (m *pacemaker) ShouldHandleMessage(msg *typesCons.HotstuffMessage) (bool, error) {
	consensusMod := m.GetBus().GetConsensusModule()

	currentHeight := consensusMod.CurrentHeight()
	currentRound := consensusMod.CurrentRound()
	currentStep := typesCons.HotstuffStep(consensusMod.CurrentStep())

	// Consensus message is from the past
	if msg.Height < currentHeight {
		m.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at height %d > message height %d", currentHeight, msg.Height))
		return false, nil
	}

	// TODO: Need to restart state sync or be in state sync mode right now
	// Current node is out of sync
	if msg.Height > currentHeight {
		m.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at height %d < message at height %d", currentHeight, msg.Height))
		return false, nil
	}

	// TODO(olshansky): This code branch is a result of the optimization in the leader
	//                  handlers. Since the leader also acts as a replica but doesn't use the replica's
	//                  handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
	// Do not handle messages if it is a self proposal
	if consensusMod.IsLeader() && msg.Type == propose && msg.Step != newRound {
		return false, nil
	}

	// Message is from the past
	if msg.Round < currentRound || (msg.Round == currentRound && msg.Step < currentStep) {
		m.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at (height, step, round) (%d, %d, %d) > message at (%d, %d, %d)", currentHeight, currentStep, currentRound, msg.Height, msg.Step, msg.Round))
		return false, nil
	}

	// Everything checks out!
	if msg.Height == currentHeight && msg.Step == currentStep && msg.Round == currentRound {
		return true, nil
	}

	// pacemaker catch up! Node is synched to the right height, but on a previous step/round so we just jump to the latest state.
	if msg.Round > currentRound || (msg.Round == currentRound && msg.Step > currentStep) {
		m.nodeLog(typesCons.PacemakerCatchup(currentHeight, uint64(currentStep), currentRound, msg.Height, uint64(msg.Step), msg.Round))
		consensusMod.SetStep(uint8(msg.Step))
		consensusMod.SetRound(msg.Round)

		// TODO: Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if currentRound != msg.Round || !consensusMod.IsLeaderSet() {
			anyProto, err := anypb.New(msg)
			if err != nil {
				log.Println("[WARN] NewHeight: Failed to convert pacemaker message to proto: ", err)
				return false, err
			}
			// TODO: Add new custom error
			if err := consensusMod.NewLeader(anyProto); err != nil {
				return false, err
			}
		}

		return true, nil
	}

	return false, typesCons.ErrUnexpectedPacemakerCase
}

func (m *pacemaker) RestartTimer() {
	// NOTE: Not deferring a cancel call because this function is asynchronous.
	if m.stepCancelFunc != nil {
		m.stepCancelFunc()
	}

	// NOTE: Not defering a cancel call because this function is asynchronous.
	consensusMod := m.GetBus().GetConsensusModule()
	stepTimeout := m.getStepTimeout(consensusMod.CurrentRound())
	clock := m.GetBus().GetRuntimeMgr().GetClock()

	ctx, cancel := clock.WithTimeout(context.TODO(), stepTimeout)
	m.stepCancelFunc = cancel

	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				m.InterruptRound("timeout")
			}
		case <-clock.After(stepTimeout + timeoutBuffer):
			return
		}
	}()
}

func (m *pacemaker) InterruptRound(reason string) {
	consensusMod := m.GetBus().GetConsensusModule()
	m.nodeLog(typesCons.PacemakerInterrupt(reason, consensusMod.CurrentHeight(), typesCons.HotstuffStep(consensusMod.CurrentStep()), consensusMod.CurrentRound()))

	consensusMod.SetRound(consensusMod.CurrentRound() + 1)

	//ADDTEST: check if this is indeed ensured after a successful round
	if m.GetBus().GetConsensusModule().IsPrepareQCNil() {
		m.startNextView(nil, false)
		return
	}

	msgAny, err := consensusMod.GetPrepareQC()
	if err != nil {
		return
	}

	msg, err := codec.GetCodec().FromAny(msgAny)
	if err != nil {
		return
	}

	quorumCertificate, ok := msg.(*typesCons.QuorumCertificate)
	if !ok {
		return
	}

	m.startNextView(quorumCertificate, false)
}

func (m *pacemaker) NewHeight() {
	consensusMod := m.GetBus().GetConsensusModule()

	m.nodeLog(typesCons.PacemakerNewHeight(consensusMod.CurrentHeight() + 1))
	consensusMod.SetHeight(consensusMod.CurrentHeight() + 1)
	consensusMod.ResetForNewHeight()

	m.startNextView(nil, false) // TODO(design): We are omitting CommitQC and TimeoutQC here.

	m.GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (m *pacemaker) startNextView(qc *typesCons.QuorumCertificate, forceNextView bool) {
	// DISCUSS: Should we lock the consensus module here?
	consensusMod := m.GetBus().GetConsensusModule()
	consensusMod.SetStep(uint8(newRound))
	consensusMod.ResetRound()
	if err := consensusMod.ReleaseUtilityContext(); err != nil {
		log.Println("[WARN] NewHeight: Failed to release utility context: ", err)
		return
	}

	// TECHDEBT: This if structure for debug purposes only; think of a way to externalize it from the main consensus flow...
	if m.debug.manualMode && !forceNextView {
		m.debug.quorumCertificate = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          propose,
		Height:        consensusMod.CurrentHeight(),
		Step:          newRound,
		Round:         consensusMod.CurrentRound(),
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	m.RestartTimer()

	anyProto, err := anypb.New(hotstuffMessage)
	if err != nil {
		log.Println("[WARN] NewHeight: Failed to convert pacemaker message to proto: ", err)
		return
	}
	if err := consensusMod.BroadcastMessageToValidators(anyProto); err != nil {
		log.Println("[WARN] NewHeight: Failed to broadcast message to validators: ", err)
		return
	}
}

// TODO: Increase timeout using exponential backoff.
func (m *pacemaker) getStepTimeout(round uint64) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(m.pacemakerCfg.TimeoutMsec))
	return baseTimeout
}

// TODO: Remove once we have a proper logging system.
func (m *pacemaker) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s)
}
