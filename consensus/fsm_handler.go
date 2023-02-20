package consensus

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

// Currently only here for clarity, might not be needed
type FSMMessageHandler interface {
	HandleUnsynched(*consensusModule, *messaging.StateMachineTransitionEvent)
	HandleSyncMode(*consensusModule, *messaging.StateMachineTransitionEvent)
	HandleSynced(*consensusModule, *messaging.StateMachineTransitionEvent)
	HandlePacemaker(*consensusModule, *messaging.StateMachineTransitionEvent)
	HandleServerMode(*consensusModule, *messaging.StateMachineTransitionEvent)
}

// State machine transition event comes to consensus module
// consensus moduel reacts upon the new changed state
// consensus module's reply is a new state machine transition event, which is sent to the state machine module
func (m *consensusModule) handleStateMachineTransitionEvent(msg *messaging.StateMachineTransitionEvent) error {
	fsm_state := msg.NewState
	m.logger.Debug().Fields(map[string]any{
		"event":          msg.Event,
		"previous_state": msg.PreviousState,
		"new_state":      fsm_state,
	}).Msg("Received state machine transition msg")

	switch fsm_state {
	case string(coreTypes.StateMachineState_Consensus_Unsynched):
		return m.HandleUnsynched(msg)

	case string(coreTypes.StateMachineState_Consensus_SyncMode):
		return m.HandleSyncMode(msg)

	case string(coreTypes.StateMachineState_Consensus_Synced):
		return m.HandleSynced(msg)

	case string(coreTypes.StateMachineState_Consensus_Pacemaker):
		return m.HandlePacemaker(msg)

	case string(coreTypes.StateMachineState_Consensus_Server_Enabled), string(coreTypes.StateMachineState_Consensus_Server_Disabled):
		return m.HandleServerMode(msg)
	}

	return nil
}

func (m *consensusModule) HandleUnsynched(msg *messaging.StateMachineTransitionEvent) error {

	return nil
}

// REFACTOR: there are similarities between sync mode and pacemaker modes' transition to unsync mode, consider consolidating
func (m *consensusModule) HandleSyncMode(msg *messaging.StateMachineTransitionEvent) error {

	return nil
}

// Synced mode may change to unsync mode if
func (m *consensusModule) HandleSynced(msg *messaging.StateMachineTransitionEvent) error {
	return nil
}

func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	return nil
}

func (m *consensusModule) HandleServerMode(msg *messaging.StateMachineTransitionEvent) error {
	if msg.Event == string(coreTypes.StateMachineEvent_Consensus_DisableServerMode) {
		return m.DisableServerMode()
	}
	return m.EnableServerMode()

}
