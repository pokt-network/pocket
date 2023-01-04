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

// func (*paceMaker) Create(bus modules.Bus) (modules.Module, error) {
// 	m := &paceMaker{}
// 	bus.RegisterModule(m)

// <<<<<<< HEAD
// 	runtimeMgr := bus.GetRuntimeMgr()
// 	cfg := runtimeMgr.GetConfig()
// =======
// 	return &paceMaker{
// 		bus: nil,
// 		//consensusMod: nil,
// >>>>>>> be3e4368 (refactor (consensus): WIP, seperate interfaces, remove consensusMod field from pacemaker)

// 	pacemakerCfg := cfg.Consensus.PacemakerConfig

// 	m.pacemakerCfg = pacemakerCfg
// 	m.paceMakerDebug = paceMakerDebug{
// 		manualMode:                pacemakerCfg.GetManual(),
// 		debugTimeBetweenStepsMsec: pacemakerCfg.GetDebugTimeBetweenStepsMsec(),
// 		quorumCertificate:         nil,
// 	}

// 	return m, nil
// }

func (*paceMaker) Create(bus modules.Bus) (modules.Module, error) {
	m := &paceMaker{}
	bus.RegisterModule(m)

	runtimeMgr := bus.GetRuntimeMgr()
	cfg := runtimeMgr.GetConfig()

	pacemakerCfg := cfg.Consensus.PacemakerConfig

	m.pacemakerCfg = pacemakerCfg
	m.paceMakerDebug = paceMakerDebug{
		manualMode:                pacemakerCfg.GetManual(),
		debugTimeBetweenStepsMsec: pacemakerCfg.GetDebugTimeBetweenStepsMsec(),
		quorumCertificate:         nil,
	}

	return m, nil
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
	if m.IsLeader() && msg.Type == Propose && msg.Step != NewRound {
		// TODO: Noisy log - make it a DEBUG
		// p.consensusMod.nodeLog(typesCons.ErrSelfProposal.Error())
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
	if msg.Round > currentRound || (msg.Round == currentRound && msg.Step > currentStep) {
		m.nodeLog(typesCons.PacemakerCatchup(currentHeight, uint64(currentStep), currentRound, msg.Height, uint64(msg.Step), msg.Round))
		consensusMod.SetStep(uint8(msg.Step))
		consensusMod.SetRound(msg.Round)

		// TODO(olshansky): Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		// if currentRound != msg.Round || m.consensusMod.leaderId == nil {
		// 	m.electNextLeader(msg)
		// }

		if currentRound != msg.Round || !consensusMod.IsLeaderSet() {
			anyProto, err := anypb.New(msg)
			if err != nil {
				log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
				return false, err
			}
			consensusMod.NewLeader(anyProto)
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

	stepTimeout := p.getStepTimeout(p.GetBus().GetConsensusModule().CurrentRound()) //p.consensusMod.round// p.Get)

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

func (m *paceMaker) InterruptRound(reason string) {
	consensusMod := m.GetBus().GetConsensusModule()
	m.nodeLog(typesCons.PacemakerInterrupt(reason, consensusMod.CurrentHeight(), typesCons.HotstuffStep(consensusMod.CurrentStep()), consensusMod.CurrentRound()))

	consensusMod.SetRound(consensusMod.CurrentRound() + 1)

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

func (p *paceMaker) NewHeight() {
	//TODO: Add log
	//p.consensusMod.nodeLog(typesCons.PacemakerNewHeight(p.consensusMod.CurrentHeight() + 1))

	currentHeight := p.GetBus().GetConsensusModule().CurrentHeight()
	currentHeight++
	p.GetBus().GetConsensusModule().SetHeight(currentHeight)

	p.GetBus().GetConsensusModule().ResetForNewHeight()

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

	//p.consensusMod.step = NewRound
	p.GetBus().GetConsensusModule().SetStep(uint64(NewRound))

	// p.consensusMod.clearLeader()
	// p.consensusMod.clearMessagesPool()
	p.GetBus().GetConsensusModule().ClearLeaderMessagesPool()

	// if p.consensusMod.utilityContext != nil {
	// 	if err := p.consensusMod.utilityContext.Release(); err != nil {
	// 		log.Println("[WARN] Failed to release utility context: ", err)
	// 	}
	// 	p.consensusMod.utilityContext = nil
	// }
	p.GetBus().GetConsensusModule().ReleaseUtilityContext()

	// TECHDEBT: This if structure for debug purposes only; think of a way to externalize it from the main consensus flow...
	if p.manualMode && !forceNextView {
		p.quorumCertificate = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        p.GetBus().GetConsensusModule().CurrentHeight(),
		Step:          NewRound,
		Round:         p.GetBus().GetConsensusModule().CurrentRound(),
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	p.RestartTimer()
	//p.consensusMod.broadcastToNodes(hotstuffMessage)
	anyProto, err := anypb.New(hotstuffMessage)
	if err != nil {
		log.Println("[WARN] NewHeight: Failed to convert paceMaker message to proto: ", err)
		return
	}
	p.GetBus().GetConsensusModule().BroadcastMessageToNodes(anyProto)
}

// TODO(olshansky): Increase timeout using exponential backoff.
func (p *paceMaker) getStepTimeout(round uint64) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(p.pacemakerCfg.TimeoutMsec))
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
