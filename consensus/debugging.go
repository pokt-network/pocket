package consensus

import (
	"log"

	"github.com/pokt-network/pocket/shared/debug"

	typesCons "github.com/pokt-network/pocket/consensus/types"
)

func (m *ConsensusModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	m.m.Lock()
	defer m.m.Unlock()
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		m.resetToGenesis(debugMessage)
	case debug.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		m.printNodeState(debugMessage)
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		m.triggerNextView(debugMessage)
	case debug.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		m.togglePacemakerManualMode(debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *ConsensusModule) GetNodeState() typesCons.ConsensusNodeState {
	m.m.RLock()
	defer m.m.RUnlock()
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

func (m *ConsensusModule) resetToGenesis(_ *debug.DebugMessage) {
	m.nodeLog(typesCons.DebugResetToGenesis)

	m.Height = 0
	m.resetForNewHeight()
	m.clearLeader()
	m.clearMessagesPool()
	m.GetBus().GetPersistenceModule().HandleDebugMessage(&debug.DebugMessage{
		Action:  debug.DebugMessageAction_DEBUG_CLEAR_STATE,
		Message: nil,
	})
	m.GetBus().GetPersistenceModule().Start() // reload genesis state
}

func (m *ConsensusModule) printNodeState(_ *debug.DebugMessage) {
	state := m.GetNodeState()
	m.nodeLog(typesCons.DebugNodeState(state))
}

func (m *ConsensusModule) triggerNextView(_ *debug.DebugMessage) {
	m.nodeLog(typesCons.DebugTriggerNextView)

	currentheight := m.Height
	currentStep := m.Step
	if currentheight == 0 || (currentStep == Decide && m.paceMaker.IsManualMode()) {
		m.paceMaker.NewHeight()
	} else {
		m.paceMaker.InterruptRound()
	}

	if m.paceMaker.IsManualMode() {
		m.paceMaker.ForceNextView()
	}
}

func (m *ConsensusModule) togglePacemakerManualMode(_ *debug.DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.nodeLog(typesCons.DebugTogglePacemakerManualMode("MANUAL"))
	} else {
		m.nodeLog(typesCons.DebugTogglePacemakerManualMode("AUTOMATIC"))
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
