package consensus

import (
	"context"
	"fmt"
	"log"
	"time"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	pacemakerModuleName = "pacemaker"
	timeoutBuffer       = 30 * time.Millisecond // A buffer around the pacemaker timeout to avoid race condition; 30ms was arbitrarily chosen
)

type Pacemaker interface {
	modules.Module
	PacemakerDebug

	// TODO(olshansky): Rather than exposing the underlying `ConsensusModule` struct,
	// we could create a `ConsensusModuleDebug` interface that'll expose setters/getters
	// for the height/round/step/etc, and interface with the module that way.
	SetConsensusModule(module *consensusModule)

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

	// TODO(olshansky): The reason `pacemaker_*` files are not a sub-package under consensus
	// due to it's dependency on the underlying implementation of `ConsensusModule`. Think
	// through a way to decouple these. This could be fixed with reflection but that's not
	// a great idea in production code.
	consensusMod *consensusModule

	pacemakerCfg *configs.PacemakerConfig

	stepCancelFunc context.CancelFunc

	// Only used for development and debugging.
	paceMakerDebug
}

func CreatePacemaker(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	var m paceMaker
	return m.Create(runtimeMgr)
}

func (m *paceMaker) Create(runtimeMgr modules.RuntimeMgr) (modules.Module, error) {
	cfg := runtimeMgr.GetConfig()

	pacemakerCfg := cfg.Consensus.PacemakerConfig

	return &paceMaker{
		bus:          nil,
		consensusMod: nil,

		pacemakerCfg: pacemakerCfg,

		stepCancelFunc: nil, // Only set on restarts

		paceMakerDebug: paceMakerDebug{
			manualMode:                pacemakerCfg.Manual,
			debugTimeBetweenStepsMsec: pacemakerCfg.DebugTimeBetweenStepsMsec,
			quorumCertificate:         nil,
		},
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
	return pacemakerModuleName
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

func (m *paceMaker) SetConsensusModule(c *consensusModule) {
	m.consensusMod = c
}

func (p *paceMaker) ShouldHandleMessage(msg *typesCons.HotstuffMessage) (bool, error) {
	currentHeight := p.consensusMod.height
	currentRound := p.consensusMod.round
	currentStep := p.consensusMod.step

	// Consensus message is from the past
	if msg.Height < currentHeight {
		p.consensusMod.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at height %d > message height %d", currentHeight, msg.Height))
		return false, nil
	}

	// TODO: Need to restart state sync or be in state sync mode right now
	// Current node is out of sync
	if msg.Height > currentHeight {
		p.consensusMod.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at height %d < message at height %d", currentHeight, msg.Height))
		return false, nil
	}

	// TODO(olshansky): This code branch is a result of the optimization in the leader
	//                  handlers. Since the leader also acts as a replica but doesn't use the replica's
	//                  handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
	// Do not handle messages if it is a self proposal
	if p.consensusMod.isLeader() && msg.Type == Propose && msg.Step != NewRound {
		// TODO: Noisy log - make it a DEBUG
		// p.consensusMod.nodeLog(typesCons.ErrSelfProposal.Error())
		return false, nil
	}

	// Message is from the past
	if msg.Round < currentRound || (msg.Round == currentRound && msg.Step < currentStep) {
		p.consensusMod.nodeLog(fmt.Sprintf("⚠️ [WARN][DISCARDING] ⚠️ Node at (height, step, round) (%d, %d, %d) > message at (%d, %d, %d)", currentHeight, currentStep, currentRound, msg.Height, msg.Step, msg.Round))
		return false, nil
	}

	// Everything checks out!
	if msg.Height == currentHeight && msg.Step == currentStep && msg.Round == currentRound {
		return true, nil
	}

	// Pacemaker catch up! Node is synched to the right height, but on a previous step/round so we just jump to the latest state.
	if msg.Round > currentRound || (msg.Round == currentRound && msg.Step > currentStep) {
		p.consensusMod.nodeLog(typesCons.PacemakerCatchup(currentHeight, uint64(currentStep), currentRound, msg.Height, uint64(msg.Step), msg.Round))
		p.consensusMod.step = msg.Step
		p.consensusMod.round = msg.Round

		// TODO(olshansky): Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if currentRound != msg.Round || p.consensusMod.leaderId == nil {
			p.consensusMod.electNextLeader(msg)
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

	stepTimeout := p.getStepTimeout(p.consensusMod.round)

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

func (p *paceMaker) InterruptRound(reason string) {
	p.consensusMod.nodeLog(typesCons.PacemakerInterrupt(reason, p.consensusMod.CurrentHeight(), p.consensusMod.step, p.consensusMod.round))

	p.consensusMod.round++
	p.startNextView(p.consensusMod.highPrepareQC, false)
}

func (p *paceMaker) NewHeight() {
	p.consensusMod.nodeLog(typesCons.PacemakerNewHeight(p.consensusMod.CurrentHeight() + 1))

	p.consensusMod.height++
	p.consensusMod.resetForNewHeight()

	p.startNextView(nil, false) // TODO(design): We are omitting CommitQC and TimeoutQC here.

	p.consensusMod.
		GetBus().
		GetTelemetryModule().
		GetTimeSeriesAgent().
		CounterIncrement(
			consensusTelemetry.CONSENSUS_BLOCKCHAIN_HEIGHT_COUNTER_NAME,
		)
}

func (p *paceMaker) startNextView(qc *typesCons.QuorumCertificate, forceNextView bool) {
	// DISCUSS: Should we lock the consensus module here?

	p.consensusMod.step = NewRound
	p.consensusMod.clearLeader()
	p.consensusMod.clearMessagesPool()
	// TECHDEBT: This should be avoidable altogether
	if p.consensusMod.utilityContext != nil {
		if err := p.consensusMod.utilityContext.Release(); err != nil {
			log.Println("[WARN] Failed to release utility context: ", err)
		}
		p.consensusMod.utilityContext = nil
	}

	// TECHDEBT: This if structure for debug purposes only; think of a way to externalize it from the main consensus flow...
	if p.manualMode && !forceNextView {
		p.quorumCertificate = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        p.consensusMod.height,
		Step:          NewRound,
		Round:         p.consensusMod.round,
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	p.RestartTimer()
	p.consensusMod.broadcastToValidators(hotstuffMessage)
}

// TODO(olshansky): Increase timeout using exponential backoff.
func (p *paceMaker) getStepTimeout(round uint64) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(p.pacemakerCfg.TimeoutMsec))
	return baseTimeout
}
