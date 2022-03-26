package consensus

import (
	"log"
	"time"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

func (m *consensusModule) HandleDebugMessage(debugMessage *types.DebugMessage) error {
	switch debugMessage.Action {
	case types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		m.resetToGenesis(debugMessage)
	case types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		m.printNodeState(debugMessage)
	case types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		m.triggerNextView(debugMessage)
	case types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		m.togglePacemakerManualMode(debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *consensusModule) GetNodeState() typesCons.ConsensusNodeState {
	leaderId := typesCons.NodeId(0)
	if m.LeaderId != nil {
		leaderId = *m.LeaderId
	}
	return typesCons.ConsensusNodeState{
		NodeId:   m.NodeId,
		Height:   m.Height,
		Round:    uint8(m.Round),
		Step:     uint8(m.Step),
		IsLeader: m.isLeader(),
		LeaderId: leaderId,
	}
}

func (m *consensusModule) resetToGenesis(_ *types.DebugMessage) {
	m.nodeLog("[DEBUG] Resetting to genesis...")

	m.Height = 0
	m.Round = 0
	m.Step = 0
	m.Block = nil

	m.HighPrepareQC = nil
	m.LockedQC = nil

	m.clearLeader()
	m.clearMessagesPool()
}

func (m *consensusModule) printNodeState(_ *types.DebugMessage) {
	state := m.GetNodeState()
	log.Printf("\t[DEBUG] NODE STATE: [%s] Node %d is at (Height, Step, Round): (%d, %s, %d)\n", m.logPrefix, state.NodeId, state.Height, StepToString[typesCons.HotstuffStep(state.Step)], state.Round)
}

func (m *consensusModule) triggerNextView(_ *types.DebugMessage) {
	m.nodeLog("[DEBUG] Triggering next view...")

	if m.Height == 0 || (m.Step == Decide && m.paceMaker.IsManualMode()) {
		m.paceMaker.NewHeight()
	} else {
		m.paceMaker.InterruptRound()
	}

	if m.paceMaker.IsManualMode() {
		m.paceMaker.ForceNextView()
	}
}

func (m *consensusModule) togglePacemakerManualMode(_ *types.DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to MANUAL")
	} else {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to AUTOMATIC")
	}
	m.paceMaker.SetManualMode(newMode)
}

// This Pacemaker interface is only used for development & debugging purposes.
type PacemakerDebug interface {
	SetManualMode(bool)
	IsManualMode() bool
	ForceNextView()
}

type paceMakerDebug struct {
	manualMode                bool
	debugTimeBetweenStepsMsec uint64

	quorumCertificate *typesCons.QuorumCertificate
}

func (p *paceMaker) IsManualMode() bool {
	return p.manualMode
}

func (p *paceMaker) SetManualMode(manualMode bool) {
	p.manualMode = manualMode
}

func (p *paceMaker) ForceNextView() {
	lastQC := p.quorumCertificate
	p.startNextView(lastQC, true)
}

// This is a hack only used to slow down the progress of the blockchain during development.
func (p *paceMaker) debugSleep() {
	time.Sleep(time.Duration(int64(time.Millisecond) * int64(p.debugTimeBetweenStepsMsec)))
}
