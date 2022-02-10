package consensus

import (
	"fmt"
	"log"

	"pocket/consensus/pkg/consensus/dkg"
	"pocket/consensus/pkg/shared"
	"pocket/consensus/pkg/shared/context"
)

func (m *consensusModule) handleDebugMessage(message *DebugMessage) {
	switch message.Action {
	case TriggerNextView:
		m.handleTriggerNextView(message)
	case TriggerDKG:
		m.handleTriggerDKG(message)
	case TogglePaceMakerManualMode:
		m.handleTogglePaceMakerManualMode(message)
	case ResetToGenesis:
		m.resetToGenesis(message)
	case PrintNodeState:
		shared.GetPocketState().PrintGlobalState()
		m.printNodeState(message)
	default:
		log.Fatalf("Unsupported debug message: %s \n", StepToString[Step(message.Action)])
	}
}

func (m *consensusModule) handleTriggerNextView(debugMessage *DebugMessage) {
	m.nodeLog("[DEBUG] Triggering next view...")

	// Assuming that block was applied if DECIDE step is reached.
	if m.Height == 0 || m.Step == Decide {
		m.paceMaker.NewHeight()
	} else {
		m.paceMaker.ForceNextView()
	}
}

func (m *consensusModule) handleTriggerDKG(debugMessage *DebugMessage) {
	m.nodeLog("[DEBUG] Triggering DKG...")

	message := &dkg.DKGMessage{
		Round: dkg.DKGRound1,
	}

	m.dkgMod.HandleMessage(context.EmptyPocketContext(), message)
}

func (m *consensusModule) handleTogglePaceMakerManualMode(message *DebugMessage) {
	newMode := !m.paceMaker.IsManualMode()
	if newMode {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to MANUAL")
	} else {
		m.nodeLog("[DEBUG] Toggling Pacemaker mode to AUTOMATIC")
	}
	m.paceMaker.SetManualMode(newMode)
}

func (m *consensusModule) resetToGenesis(message *DebugMessage) {
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

func (m *consensusModule) printNodeState(message *DebugMessage) {
	state := m.GetNodeState()
	fmt.Printf("\tCONSENSUS STATE: [%s] Node %d is at (Height, Step, Round): (%d, %s, %d)\n", m.logPrefix, state.NodeId, state.Height, StepToString[Step(state.Step)], state.Round)
}
