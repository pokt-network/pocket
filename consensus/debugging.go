package consensus

import (
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

func (m *consensusModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	m.m.Lock()
	defer m.m.Unlock()

	switch debugMessage.Action {
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
		m.resetToGenesis(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
		m.printNodeState(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
		m.triggerNextView(debugMessage)
	case messaging.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
		m.togglePacemakerManualMode(debugMessage)
	default:
		m.logger.Debug().Msgf("Debug message: %s", debugMessage.Message)
	}
	return nil
}

func (m *consensusModule) GetNodeState() typesCons.ConsensusNodeState {
	leaderId := typesCons.NodeId(0)
	if m.leaderId != nil {
		leaderId = *m.leaderId
	}

	return typesCons.ConsensusNodeState{
		NodeId:   m.nodeId,
		Height:   m.height,
		Round:    uint8(m.round),
		Step:     uint8(m.step),
		IsLeader: m.isLeader(),
		LeaderId: leaderId,
	}
}

func (m *consensusModule) resetToGenesis(_ *messaging.DebugMessage) {
	m.logger.Debug().Msg(typesCons.DebugResetToGenesis)

	m.height = 0
	m.resetForNewHeight()
	m.clearLeader()
	m.clearMessagesPool()
	m.GetBus().GetPersistenceModule().HandleDebugMessage(&messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS,
		Message: nil,
	})
	m.GetBus().GetPersistenceModule().Start() // reload genesis state
}

func (m *consensusModule) printNodeState(_ *messaging.DebugMessage) {
	state := m.GetNodeState()
	m.logger.Debug().
		Fields(map[string]interface{}{
			"step":   state.Step,
			"height": state.Height,
			"round":  state.Round,
		}).Msg("Node state")
}

func (m *consensusModule) triggerNextView(_ *messaging.DebugMessage) {
	m.logger.Debug().Msg(typesCons.DebugTriggerNextView)

	currentHeight := m.height
	currentStep := m.step
	if currentHeight == 0 || (currentStep == Decide && m.paceMaker.IsManualMode()) {
		m.paceMaker.NewHeight()
	} else {
		m.paceMaker.InterruptRound("manual trigger")
	}

	if m.paceMaker.IsManualMode() {
		m.paceMaker.ForceNextView()
	}
}

func (m *consensusModule) togglePacemakerManualMode(_ *messaging.DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.logger.Debug().Str("pacemaker_mode", "MANUAL").Msg("Toggle pacemaker to MANUAL mode")
	} else {
		m.logger.Debug().Str("pacemaker_mode", "AUTOMATIC").Msg("Toggle pacemaker to AUTOMATIC mode")
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

	// IMPROVE: Consider renaming to `previousRoundQC`
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
