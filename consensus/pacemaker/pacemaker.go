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
	//DefaultLogPrefix    = "NODE"
	pacemakerModuleName = "pacemaker"
	//NewRound            = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	//Propose             = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
	timeoutBuffer = 30 * time.Millisecond // A buffer around the pacemaker timeout to avoid race condition; 30ms was arbitrarily chosen

	NewRound = typesCons.HotstuffStep_HOTSTUFF_STEP_NEWROUND
	Propose  = typesCons.HotstuffMessageType_HOTSTUFF_MESSAGE_PROPOSE
)

type Pacemaker interface {
	modules.Module
	PacemakerDebug

	ShouldHandleMessage(message *typesCons.HotstuffMessage) (bool, error)

	RestartTimer()
	NewHeight()
	InterruptRound(reason string)
}

var (
	_ modules.Module = &paceMaker{}
	_ PacemakerDebug = &paceMaker{}
)

type paceMaker struct {
	bus modules.Bus

	pacemakerCfg *configs.PacemakerConfig

	stepCancelFunc context.CancelFunc

	// Only used for development and debugging.
	paceMakerDebug

	//REFACTOR: this should be removed, when we build a shared and proper logger
	logPrefix string
}

func CreatePacemaker(bus modules.Bus) (modules.Module, error) {
	var m paceMaker
	return m.Create(bus)
}

func (m *paceMaker) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtimeMgr.GetConfig()
	if err := m.ValidateConfig(cfg); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	pacemakerCfg := cfg.GetConsensusConfig().(HasPacemakerConfig).GetPacemakerConfig()

	return &paceMaker{
		bus: nil,

		pacemakerCfg: pacemakerCfg,

		stepCancelFunc: nil, // Only set on restarts

		paceMakerDebug: paceMakerDebug{
			manualMode:                pacemakerCfg.GetManual(),
			debugTimeBetweenStepsMsec: pacemakerCfg.GetDebugTimeBetweenStepsMsec(),
			previousRoundQC:           nil,
		},

		logPrefix: DefaultLogPrefix,
	}, nil
}

func (p *paceMaker) Start() error {
	p.RestartTimer()
	return nil
}
func (p *paceMaker) Stop() error {
	return nil
}

func (p *paceMaker) GetModuleName() string {
	return modules.PacemakerModuleName
}

func (m *paceMaker) SetBus(pocketBus modules.Bus) {
	m.bus = pocketBus
}

func (m *paceMaker) GetBus() modules.Bus {
	if m.bus == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.bus
}

func (m *paceMaker) ShouldHandleMessage(msg *typesCons.HotstuffMessage) (bool, error) {
	consensusMod := m.GetBus().GetConsensusModule()

	currentHeight := m.GetBus().GetConsensusModule().CurrentHeight()
	currentRound := m.GetBus().GetConsensusModule().CurrentRound()
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

	if m.GetBus().GetConsensusModule().IsLeader() && msg.Type == Propose && msg.Step != NewRound 
		// TODO(olshansky): This code branch is a result of the optimization in the leader
		// handlers. Since the leader also acts as a replica but doesn't use the replica's
		// handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
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

	// Pacemaker catch up! Node is synched to the right height, but on a previous step/round so we just jump to the latest state.
	if m.Round > currentRound || (m.Round == currentRound && m.Step > currentStep) {
		//TODO: Add node log (typesCons.PacemakerCatchup(currentHeight, uint64(currentStep), currentRound, m.Height, uint64(m.Step), m.Round))

		consensusMod.SetStep(uint64(m.Step))
		consensusMod.SetRound(m.Round)

		// TODO(olshansky): Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if currentRound != m.Round || !consensusMod.IsLeaderSet() {
			anyProto, err := anypb.New(m)
			if err != nil {
				log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
				return false, err
			}
			m.GetBus().GetConsensusModule().NewLeader(anyProto)


		}

		return true, nil
	}

	return false, typesCons.ErrUnexpectedPacemakerCase
}

func (p *paceMaker) RestartTimer() {
	// NOTE: Not deferring a cancel call because this function is asynchronous.
	if p.stepCancelFunc != nil {
		p.stepCancelFunc()
	}

	// NOTE: Not defering a cancel call because this function is asynchronous.

	stepTimeout := p.getStepTimeout(p.GetBus().GetConsensusModule().CurrentRound())

	clock := p.GetBus().GetRuntimeMgr().GetClock()

	ctx, cancel := clock.WithTimeout(context.TODO(), stepTimeout)
	p.stepCancelFunc = cancel

	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				p.InterruptRound("timeout")
			}
		case <-clock.After(stepTimeout + timeoutBuffer):
			return
		}
	}()
}

func (p *paceMaker) InterruptRound() {
	//TODO: Add node log: (typesCons.PacemakerInterrupt(p.GetBus().GetConsensusModule().CurrentHeight(), typesCons.HotstuffStep(p.GetBus().GetConsensusModule().CurrentStep()), p.GetBus().GetConsensusModule().CurrentRound()))

	consensusMod := p.GetBus().GetConsensusModule()
	currentRound := consensusMod.CurrentRound()
	currentRound++
	consensusMod.SetRound(currentRound)

	msg, err := codec.GetCodec().FromAny(consensusMod.GetPrepareQC())

	if err != nil {
		return
	}
	quorumCertificate, ok := msg.(*typesCons.QuorumCertificate)
	if !ok {
		return
	}
	m.startNextView(quorumCertificate, false)
}

func (p *paceMaker) NewHeight() {
	//TODO: Add node log: (typesCons.PacemakerNewHeight(p.consensusMod.CurrentHeight() + 1))

	consensusMod := p.GetBus().GetConsensusModule()
	currentHeight := consensusMod.CurrentHeight()
	currentHeight++
	consensusMod.SetHeight(currentHeight)
	consensusMod.ResetForNewHeight()

	p.startNextView(nil, false) // TODO(design): We are omitting CommitQC and TimeoutQC here.

	p.
		GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (p *paceMaker) startNextView(qc *typesCons.QuorumCertificate, forceNextView bool) {
	// DISCUSS: Should we lock the consensus module here?

	consensusMod := p.GetBus().GetConsensusModule()
	consensusMod.SetStep(uint64(NewRound))
	consensusMod.ClearLeaderMessagesPool()
	consensusMod.ReleaseUtilityContext()

	// TECHDEBT: This if structure for debug purposes only; think of a way to externalize it from the main consensus flow...
	if p.manualMode && !forceNextView {
		p.previousRoundQC = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        consensusMod.CurrentHeight(),
		Step:          NewRound,
		Round:         consensusMod.CurrentRound(),
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	p.RestartTimer()

	anyProto, err := anypb.New(hotstuffMessage)
	if err != nil {
		log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
		return
	}
	consensusMod.BroadcastMessageToNodes(anyProto)
}

// TODO(olshansky): Increase timeout using exponential backoff.
func (m *paceMaker) getStepTimeout(round uint64) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(m.pacemakerCfg.TimeoutMsec))
	return baseTimeout
}

// TODO: Remove once we have a proper logging system.
func (m *paceMaker) nodeLog(s string) {
	log.Printf("[%s][%d] %s\n", m.logPrefix, m.GetBus().GetConsensusModule().GetNodeId(), s)
}

// HasPacemakerConfig is used to determine if a ConsensusConfig includes a PacemakerConfig without having to cast to the struct
// (which would break mocks and/or pollute the codebase with mock types casts and checks)
type HasPacemakerConfig interface {
	GetPacemakerConfig() *typesCons.PacemakerConfig
}
