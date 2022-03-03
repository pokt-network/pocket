package consensus

import (
	"context"
	"fmt"
	"log"
	"time"

	// "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/config"
	"github.com/pokt-network/pocket/shared/types"

	"github.com/pokt-network/pocket/shared/modules"
)

// TODO: PaceMaker has some functionsthat are meant only part of the interface
// for development and debugging purposes. Need to think about how to decouple
// it (if needed) to avoid code complexity in the core business logic.

type PaceMaker interface {
	modules.Module

	ShouldHandleMessage(message *HotstuffMessage) bool
	RestartTimer()
	NewHeight()
	InterruptRound()

	SetConsensusMod(module *consensusModule)

	PaceMakerDebug
}

type paceMaker struct {
	modules.Module
	pocketBusMod modules.Bus

	// TODO: Should the PaceMaker have a link back to the consensus module
	// or should they communicate via events or the PocketBusManager?
	consensusMod *consensusModule

	paceMakerParams *types.PaceMakerParams

	stepCancelFunc context.CancelFunc

	paceMakerDebug // Only used for development and debugging.
}

func (m *paceMaker) SetConsensusMod(c *consensusModule) {
	m.consensusMod = c
}

func CreatePaceMaker(_ *config.Config) (m *paceMaker, err error) {
	paceMakerParams := &types.PaceMakerParams{
		TimeoutMSec:               5000,
		RetryTimeoutMSec:          1000,
		MaxTimeoutMSec:            60000,
		MinBlockFreqMSec:          2000,
		DebugTimeBetweenStepsMsec: 500,
	}

	return &paceMaker{
		paceMakerParams: paceMakerParams,
		stepCancelFunc:  nil, // Only set on restarts

		paceMakerDebug: paceMakerDebug{
			manualMode:        true,
			quorumCertificate: nil,
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
	m.pocketBusMod = pocketBus
}

func (m *paceMaker) GetBus() modules.Bus {
	if m.pocketBusMod == nil {
		log.Fatalf("PocketBus is not initialized")
	}
	return m.pocketBusMod
}

func (p *paceMaker) ShouldHandleMessage(m *HotstuffMessage) bool {
	consensusMod := p.consensusMod
	// consensusMod.nodeLog(fmt.Sprintf("[WARN] Pacemaker received unexpected message of Height, Step, Round) from %d:  Expected: (%d, %d, %d); Actual (%d, %d, %d). Discarding...", message.Sender, consensusMod.Height, consensusMod.Step, consensusMod.Round, message.Height, message.Step, message.Round))

	// Chain is out of sync.
	if m.Height != consensusMod.Height {
		// TODO: Need to broadcast message to start state sync.
		consensusMod.nodeLog(fmt.Sprintf("[DEBUG] Discarding hotstuff message: Received message of an unexpected height from %d. Current: %d; Message: %d ", m.Sender, consensusMod.Height, m.Height))
		return false
	}

	// Do not handle messages if it is a self proposal.
	if consensusMod.isLeader() && m.Type == ProposeMessageType && m.Step != NewRound {
		// TODO: Optimization that leads to some code complexity.
		// Since the leader also acts as a replica but doesn't use the replica's handlers given the current
		// implementation, it is safe to drop proposal that the leader made to itself.
		consensusMod.nodeLog("[DEBUG] Discarding hotstuff message: self proposal")
		return false
	}

	if m.Round < consensusMod.Round || (m.Round == consensusMod.Round && m.Step < consensusMod.Step) {
		consensusMod.nodeLog(fmt.Sprintf("[WARN] Discard hotstuff message from the past: Message (step, round) IS (%s, %d) but node is at (%s, %d).", StepToString[m.Step], m.Round, StepToString[consensusMod.Step], consensusMod.Round))
		return false
	}

	// We checked the height above, so if step and round match, everything is fine.
	if m.Step == consensusMod.Step && m.Round == consensusMod.Round {
		consensusMod.nodeLog(fmt.Sprintf("[DEBUG] Received message of the expected (height, step, round): (%d, %s, %d)", m.Height, StepToString[m.Step], m.Round))
		return true
	}

	if m.Round > consensusMod.Round {
		consensusMod.nodeLog("[WARN][TODO] We are catching up the node's round so the leader needs to be updated, but right now we're just setting the leader to the message sender...")
		consensusMod.LeaderId = &m.Sender
	}

	// Advance to the latest step/round.
	if m.Round > consensusMod.Round || (m.Round == consensusMod.Round && m.Step > consensusMod.Step) {
		consensusMod.nodeLog(fmt.Sprintf("[INFO] Catching up the node's (step, round) FROM (%s, %d) TO (%s, %d).", StepToString[consensusMod.Step], consensusMod.Round, StepToString[m.Step], m.Round))
		consensusMod.Step = m.Step
		consensusMod.Round = m.Round
	}

	return true
}

func (p *paceMaker) RestartTimer() {
	if p.stepCancelFunc != nil {
		p.stepCancelFunc()
	}

	// TODO: This is a hack only used to slow down the progress of the blockchain during development.
	time.Sleep(time.Duration(int64(time.Millisecond) * int64(p.paceMakerParams.DebugTimeBetweenStepsMsec)))

	stepTimeout := p.getStepTimeout(p.consensusMod.Round)
	// Not defering a cancel call because this function is asynchronous.
	ctx, cancel := context.WithTimeout(context.TODO(), stepTimeout)
	p.stepCancelFunc = cancel

	go func() {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				p.consensusMod.nodeLog(fmt.Sprintf("[%s][%d] Timed out at step %s!", p.consensusMod.logPrefix, p.consensusMod.NodeId, StepToString[p.consensusMod.Step]))
				p.InterruptRound()
			}
		case <-time.After(stepTimeout + 30*time.Millisecond): // Adding 30ms to the context timeout to avoid race condition.
			return
		}
	}()
}

func (p *paceMaker) InterruptRound() {
	p.consensusMod.nodeLog(fmt.Sprintf("[INTERRUPT] Height: %d; Step: %d; Round: %d!", p.consensusMod.Height, p.consensusMod.Step, p.consensusMod.Round))

	p.consensusMod.Round++
	p.startNextView(p.consensusMod.HighPrepareQC, false)
}

func (p *paceMaker) NewHeight() {
	p.consensusMod.nodeLog(fmt.Sprintf("Starting first round for new block at height: %d", p.consensusMod.Height+1))

	p.consensusMod.Height = p.consensusMod.Height + 1
	p.consensusMod.Round = 0
	p.consensusMod.Block = nil

	p.consensusMod.HighPrepareQC = nil
	p.consensusMod.LockedQC = nil

	p.startNextView(nil, false) // Do we really need to pass commitQC here?
}

func (p *paceMaker) startNextView(qc *QuorumCertificate, forceNextView bool) {
	p.consensusMod.Step = NewRound
	p.consensusMod.clearLeader()
	p.consensusMod.clearMessagesPool()

	if p.manualMode && !forceNextView {
		p.quorumCertificate = qc
		return
	}

	m := &HotstuffMessage{
		Step:      NewRound,
		Height:    p.consensusMod.Height,
		Round:     p.consensusMod.Round,
		JustifyQC: qc,
		Type:      ProposeMessageType,
	}

	p.RestartTimer()
	p.consensusMod.broadcastToNodes(m)
}

func (p *paceMaker) getStepTimeout(round Round) time.Duration {
	baseTimeout := time.Duration(int64(time.Millisecond) * int64(p.paceMakerParams.TimeoutMSec))
	// TODO: Increase timeout using exponential backoff.
	return baseTimeout
}
