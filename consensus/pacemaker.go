package consensus

import (
	"context"
	"log"
	timePkg "time"

	consensusTelemetry "github.com/pokt-network/pocket/consensus/telemetry"
	typesCons "github.com/pokt-network/pocket/consensus/types"

	"github.com/pokt-network/pocket/shared/modules"
)

const (
	PacemakerModuleName = "pacemaker"
)

type Pacemaker interface {
	modules.Module
	PacemakerDebug

	// TODO(olshansky): Rather than exposing the underlying `ConsensusModule` struct,
	// we could create a `ConsensusModuleDebug` interface that'll expose setters/getters
	// for the height/round/step/etc, and interface with the module that way.
	SetConsensusModule(module *ConsensusModule)

	ValidateMessage(message *typesCons.HotstuffMessage) error
	RestartTimer()
	NewHeight()
	InterruptRound()
}

var _ modules.Module = &paceMaker{}
var _ PacemakerDebug = &paceMaker{}

type paceMaker struct {
	bus modules.Bus

	// TODO(olshansky): The reason `pacemaker_*` files are not a sub-package under consensus
	// due to it's dependency on the underlying implementation of `ConsensusModule`. Think
	// through a way to decouple these. This could be fixed with reflection but that's not
	// a great idea in production code.
	consensusMod *ConsensusModule

	pacemakerConfigs modules.PacemakerConfig

	stepCancelFunc context.CancelFunc

	// Only used for development and debugging.
	paceMakerDebug
}

func (p *paceMaker) InitConfig(pathToConfigJSON string) (config modules.IConfig, err error) {
	return // No-op
}

func (p *paceMaker) InitGenesis(pathToGenesisJSON string) (genesis modules.IGenesis, err error) {
	return // No-op
}

func CreatePacemaker(cfg *typesCons.ConsensusConfig) (m *paceMaker, err error) {
	return &paceMaker{
		bus:          nil,
		consensusMod: nil,

		pacemakerConfigs: cfg.GetPaceMakerConfig(),

		stepCancelFunc: nil, // Only set on restarts

		paceMakerDebug: paceMakerDebug{
			manualMode:                cfg.GetPaceMakerConfig().GetManual(),
			debugTimeBetweenStepsMsec: cfg.GetPaceMakerConfig().GetDebugTimeBetweenStepsMsec(),
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
	return PacemakerModuleName
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

func (m *paceMaker) SetConsensusModule(c *ConsensusModule) {
	m.consensusMod = c
}

func (p *paceMaker) ValidateMessage(m *typesCons.HotstuffMessage) error {
	currentHeight := p.consensusMod.Height
	currentRound := p.consensusMod.Round
	// Consensus message is from the past
	if m.Height < currentHeight {
		return typesCons.ErrPacemakerUnexpectedMessageHeight(typesCons.ErrOlderMessage, currentHeight, m.Height)
	}

	// Current node is out of sync
	if m.Height > currentHeight {
		// TODO(design): Need to restart state sync
		return typesCons.ErrPacemakerUnexpectedMessageHeight(typesCons.ErrFutureMessage, currentHeight, m.Height)
	}

	// Do not handle messages if it is a self proposal
	if p.consensusMod.isLeader() && m.Type == Propose && m.Step != NewRound {
		// TODO(olshansky): This code branch is a result of the optimization in the leader
		// handlers. Since the leader also acts as a replica but doesn't use the replica's
		// handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
		return typesCons.ErrSelfProposal
	}

	// Message is from the past
	if m.Round < currentRound || (m.Round == currentRound && m.Step < p.consensusMod.Step) {
		return typesCons.ErrPacemakerUnexpectedMessageStepRound(typesCons.ErrOlderStepRound, p.consensusMod.Step, currentRound, m)
	}

	// Everything checks out!
	if m.Height == currentHeight && m.Step == p.consensusMod.Step && m.Round == currentRound {
		return nil
	}

	// Pacemaker catch up! Node is synched to the right height, but on a previous step/round so we just jump to the latest state.
	if m.Round > currentRound || (m.Round == currentRound && m.Step > p.consensusMod.Step) {
		p.consensusMod.nodeLog(typesCons.PacemakerCatchup(currentHeight, uint64(p.consensusMod.Step), currentRound, m.Height, uint64(m.Step), m.Round))
		p.consensusMod.Step = m.Step
		p.consensusMod.Round = m.Round

		// TODO(olshansky): Add tests for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if currentRound != m.Round || p.consensusMod.LeaderId == nil {
			p.consensusMod.electNextLeader(m)
		}

		return nil
	}

	return typesCons.ErrUnexpectedPacemakerCase
}

func (p *paceMaker) RestartTimer() {
	if p.stepCancelFunc != nil {
		p.stepCancelFunc()
	}

	// NOTE: Not defering a cancel call because this function is asynchronous.

	stepTimeout := p.getStepTimeout(p.consensusMod.Round)

	clock := p.bus.GetClock()

	ctx, cancel := clock.WithTimeout(context.TODO(), stepTimeout)
	p.stepCancelFunc = cancel

	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				p.consensusMod.nodeLog(typesCons.PacemakerTimeout(p.consensusMod.CurrentHeight(), p.consensusMod.Step, p.consensusMod.Round))
				p.InterruptRound()
			}
		case <-clock.After(stepTimeout + 30*timePkg.Millisecond): // Adding 30ms to the context timeout to avoid race condition.
			return
		}
	}()
}

func (p *paceMaker) InterruptRound() {
	p.consensusMod.nodeLog(typesCons.PacemakerInterrupt(p.consensusMod.CurrentHeight(), p.consensusMod.Step, p.consensusMod.Round))

	p.consensusMod.Round++
	p.startNextView(p.consensusMod.HighPrepareQC, false)
}

func (p *paceMaker) NewHeight() {
	p.consensusMod.nodeLog(typesCons.PacemakerNewHeight(p.consensusMod.CurrentHeight() + 1))

	p.consensusMod.Height++
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
	p.consensusMod.Step = NewRound
	p.consensusMod.clearLeader()
	p.consensusMod.clearMessagesPool()

	// TODO(olshansky): This if structure for debug purposes only; think of a way to externalize it...
	if p.manualMode && !forceNextView {
		p.quorumCertificate = qc
		return
	}

	hotstuffMessage := &typesCons.HotstuffMessage{
		Type:          Propose,
		Height:        p.consensusMod.Height,
		Step:          NewRound,
		Round:         p.consensusMod.Round,
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &typesCons.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	p.RestartTimer()
	p.consensusMod.broadcastToNodes(hotstuffMessage)
}

// TODO(olshansky): Increase timeout using exponential backoff.
func (p *paceMaker) getStepTimeout(round uint64) timePkg.Duration {
	baseTimeout := timePkg.Duration(int64(timePkg.Millisecond) * int64(p.pacemakerConfigs.GetTimeoutMsec()))
	return baseTimeout
}
