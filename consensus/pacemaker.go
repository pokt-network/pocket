package consensus

import (
	"context"
	"fmt"
	"log"
	"time"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/config"

	"github.com/pokt-network/pocket/shared/modules"
)

// TODO(olshansky): Low priority design: think of a way to make `pacemaker_*` files be a sub-package under consensus.
type Pacemaker interface {
	modules.Module
	PacemakerDebug

	// TODO(olshansky): Rather than exposing the underlying `consensusModule` struct,
	// we could create a `ConsensusModuleDebug` interface that'll expose setters/getters
	// for the height/round/step/etc, and interface with the module that way.
	SetConsensusModule(module *consensusModule)

	ShouldHandleMessage(message *types_consensus.HotstuffMessage) (bool, string)
	RestartTimer()
	NewHeight()
	InterruptRound()
}

var _ modules.Module = &paceMaker{}
var _ PacemakerDebug = &paceMaker{}

type paceMaker struct {
	bus          modules.Bus
	consensusMod *consensusModule

	pacemakerConfigs *config.PacemakerConfig

	stepCancelFunc context.CancelFunc

	// Only used for development and debugging.
	paceMakerDebug
}

func CreatePacemaker(cfg *config.Config) (m *paceMaker, err error) {
	return &paceMaker{
		bus:          nil,
		consensusMod: nil,

		pacemakerConfigs: cfg.Consensus.Pacemaker,

		stepCancelFunc: nil, // Only set on restarts

		paceMakerDebug: paceMakerDebug{
			manualMode:                cfg.Consensus.Pacemaker.Manual,
			debugTimeBetweenStepsMsec: cfg.Consensus.Pacemaker.DebugTimeBetweenStepsMsec,
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

func (p *paceMaker) ShouldHandleMessage(m *types_consensus.HotstuffMessage) (bool, string) {
	// Consensus message is from the past
	if m.Height < p.consensusMod.Height {
		return false, fmt.Sprintf("Hotstuff message is behind the node's height. Current: %d; Message: %d ", p.consensusMod.Height, m.Height)
	}

	// Current node is out of sync
	if m.Height > p.consensusMod.Height {
		// TODO(design): Need to restart state sync
		return false, fmt.Sprintf("Hotstuff message is ahead the node's height. Current: %d; Message: %d ", p.consensusMod.Height, m.Height)
	}

	// Do not handle messages if it is a self proposal
	if p.consensusMod.isLeader() && m.Type == Propose && m.Step != NewRound {
		// TODO(olshansky): This code branch is a function of the optimization in the leader
		// handlers. Since the leader also acts as a replica but doesn't use the replica's
		// handlers given the current implementation, it is safe to drop proposal that the leader made to itself.
		return false, "Hotstuff message is a self proposal"
	}

	// Message is from the past
	if m.Round < p.consensusMod.Round || (m.Round == p.consensusMod.Round && m.Step < p.consensusMod.Step) {
		return false, fmt.Sprintf("Hotstuff message is of the right height but from the past. Current (step, round): (%s, %d); Message (step, round): (%s, %d).", StepToString[p.consensusMod.Step], p.consensusMod.Round, StepToString[m.Step], m.Round)
	}

	// Everything checks out!
	fmt.Println("OLSH", m.Step, p.consensusMod.Step)
	if m.Height == p.consensusMod.Height && m.Step == p.consensusMod.Step && m.Round == p.consensusMod.Round {
		return true, "Hotstuff message received is of the right height, step and round"
	}

	// Pacemaker catch up! Node is synched to the right height, but on a previous step/round so we just jump to the latest state.
	if m.Round > p.consensusMod.Round || (m.Round == p.consensusMod.Round && m.Step > p.consensusMod.Step) {
		reason := fmt.Sprintf("Pacemaker catching up the node's (height, step, round) FROM (%d, %s, %d) TO (%d, %s, %d).", p.consensusMod.Height, StepToString[p.consensusMod.Step], p.consensusMod.Round, m.Height, StepToString[m.Step], m.Round)

		p.consensusMod.Step = m.Step
		p.consensusMod.Round = m.Round

		// TODO(olshansky): TESTS for this. When we catch up to a later step, the leader is still the same.
		// However, when we catch up to a later round, the leader at the same height will be different.
		if p.consensusMod.Round != m.Round || p.consensusMod.LeaderId == nil {
			p.consensusMod.electNextLeader(m)
		}

		return true, reason
	}

	return false, "UNHANDLED PACEMAKER CHECK"
}

func (p *paceMaker) RestartTimer() {
	if p.stepCancelFunc != nil {
		p.stepCancelFunc()
	}

	// TODO(olshansky): This is a hack only used to slow down the progress of the blockchain during development.
	time.Sleep(time.Duration(int64(time.Millisecond) * int64(p.debugTimeBetweenStepsMsec)))

	stepTimeout := p.getStepTimeout(p.consensusMod.Round)

	// NOTE: Not defering a cancel call because this function is asynchronous.
	ctx, cancel := context.WithTimeout(context.TODO(), stepTimeout)
	p.stepCancelFunc = cancel

	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				p.consensusMod.nodeLog(fmt.Sprintf("Timed out at (height, step, round) (%d, %s, %d)!", p.consensusMod.Height, StepToString[p.consensusMod.Step], p.consensusMod.Round))
				p.InterruptRound()
			}
		case <-time.After(stepTimeout + 30*time.Millisecond): // Adding 30ms to the context timeout to avoid race condition.
			return
		}
	}()
}

func (p *paceMaker) InterruptRound() {
	p.consensusMod.nodeLog(fmt.Sprintf("INTERRUPT at (height, step, round): (%d, %s, %d)!", p.consensusMod.Height, StepToString[p.consensusMod.Step], p.consensusMod.Round))

	p.consensusMod.Round++
	p.startNextView(p.consensusMod.HighPrepareQC, false)
}

func (p *paceMaker) NewHeight() {
	p.consensusMod.nodeLog(fmt.Sprintf("Starting first round for new block at height: %d", p.consensusMod.Height+1))

	p.consensusMod.Height++
	p.consensusMod.Round = 0
	p.consensusMod.Block = nil

	p.consensusMod.HighPrepareQC = nil
	p.consensusMod.LockedQC = nil

	p.startNextView(nil, false) // TODO(design): We are omitting CommitQC and TimeoutQC here.
}

func (p *paceMaker) startNextView(qc *types_consensus.QuorumCertificate, forceNextView bool) {
	p.consensusMod.Step = NewRound
	p.consensusMod.clearLeader()
	p.consensusMod.clearMessagesPool()

	// TODO(olshansky): This if structure for debug purposes only; think of a way to externalize it...
	if p.manualMode && !forceNextView {
		p.quorumCertificate = qc
		return
	}

	hotstuffMessage := &types_consensus.HotstuffMessage{
		Type:          Propose,
		Height:        p.consensusMod.Height,
		Step:          NewRound,
		Round:         p.consensusMod.Round,
		Block:         nil,
		Justification: nil, // Set below if qc is not nil
	}

	if qc != nil {
		hotstuffMessage.Justification = &types_consensus.HotstuffMessage_QuorumCertificate{
			QuorumCertificate: qc,
		}
	}

	p.RestartTimer()
	p.consensusMod.broadcastToNodes(hotstuffMessage, HotstuffMessage)
}

func (p *paceMaker) getStepTimeout(round uint64) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(p.pacemakerConfigs.TimeoutMsec))
	// TODO(olshansky): Increase timeout using exponential backoff.
	return baseTimeout
}
