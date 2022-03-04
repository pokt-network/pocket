package consensus

import (
	"fmt"
	"log"

	types_consensus "github.com/pokt-network/pocket/consensus/types"
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
		m.togglePaceMakerManualMode(debugMessage)
	default:
		log.Printf("Debug message: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *consensusModule) GetNodeState() types_consensus.ConsensusNodeState {
	leaderId := types_consensus.NodeId(0)
	if m.LeaderId != nil {
		leaderId = *m.LeaderId
	}
	return types_consensus.ConsensusNodeState{
		NodeId:   m.NodeId,
		Height:   uint64(m.Height),
		Round:    uint8(m.Round),
		Step:     uint8(m.Step),
		IsLeader: m.isLeader(),
		LeaderId: leaderId,
	}
}

func (m *consensusModule) resetToGenesis(message *types.DebugMessage) {
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

func (m *consensusModule) printNodeState(message *types.DebugMessage) {
	state := m.GetNodeState()
	fmt.Printf("\tCONSENSUS STATE: [%s] Node %d is at (Height, Step, Round): (%d, %s, %d)\n", m.logPrefix, state.NodeId, state.Height, StepToString[types_consensus.HotstuffStep(state.Step)], state.Round)
}

// func (m *consensusModule) handleDebugMessage(message *types.DebugMessage) {
// 	switch message.Action {
// 	case types.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS:
// 		m.resetToGenesis(message)
// 	case types.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE:
// 		m.resetToGenesis(message)
// 	case types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW:
// 		m.handleTriggerNextView(message)
// 	case types.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE:
// 		m.handleTriggerNextView(message)
// 	default:
// 		log.Fatalf("Unsupported debug message: %s \n", types.DebugMessageAction_name[int32(message.Action)])
// 	}
// }

func (m *consensusModule) triggerNextView(message *types.DebugMessage) {
	m.nodeLog("[DEBUG] Triggering next view...")

	// Assuming that block was applied if DECIDE step is reached.
	if m.Height == 0 || m.Step == Decide {
		m.paceMaker.NewHeight()
		m.paceMaker.ForceNextView()
	} else {
		m.paceMaker.InterruptRound()
		m.paceMaker.ForceNextView()
	}
}

func (m *consensusModule) togglePaceMakerManualMode(message *types.DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to MANUAL")
	} else {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to AUTOMATIC")
	}
	m.paceMaker.SetManualMode(newMode)
}
