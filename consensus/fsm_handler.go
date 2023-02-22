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
	m.m.Lock()
	defer m.m.Unlock()

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
	m.logger.Debug().Msg("In HandleUnsyched, as node is out of sync, sending syncmode event to start syncing")
	if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing); err != nil {
		return err
	}

	return nil
}

// CONSIDER: there are similarities between sync mode and pacemaker modes' transition to unsync mode, consider consolidating
func (m *consensusModule) HandleSyncMode(msg *messaging.StateMachineTransitionEvent) error {
	// wait for syncing to finish
	// if the node is validator move to pacemaker state
	// else move the synced state
	// m.GetBus().GetConsensusModule().StartSnycing()

	m.logger.Debug().Msg("In HandleSyncMode, starting syning")

	m.stateSync.AggregateMetadataResponses()
	err, synced := m.stateSync.Sync()
	if err != nil {
		return err
	}

	if synced {
		if m.IsValidator() {
			m.logger.Debug().Msg("Valdiator node synced to the latest state with the rest of the peers")
			if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsCaughtUpValidator); err != nil {
				return err
			}
		} else {
			m.logger.Debug().Msg("Node synced to the latest state with the rest of the peers")
			if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsCaughtUpNonValidator); err != nil {
				return err
			}
		}
	} else {
		m.logger.Debug().Msg("Syncing is not complete, no state transition")
	}
	return nil
}

// Synced mode may change to unsync mode if node is out of sync
func (m *consensusModule) HandleSynced(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("In HandleSyncedMode")
	// node starts and it is not validator.
	// node syncs,
	// and node is now in Synced state
	// node receives a new block and
	// whenever a new block is received if its height is greater than current block height
	// node transitions to unsycnhed state

	return nil
}

// Pacemaker mode may change to unsync mode if node is out of sync
func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("Validator is synced and in HandlePacemaker mode. It will stay in this mode until it receives a new block proposal that has a higher height than the current block height")
	// node starts and it is a validator.
	// node syncs,
	// and node is now in Synced state
	// node receives a new block proposal, and it understands that it doesn't have block
	// node transitions to unsycnhed state

	//CONSIDER: maybe we can consider transitioning the unsynced node along with the received block height
	return nil
}

func (m *consensusModule) HandleServerMode(msg *messaging.StateMachineTransitionEvent) error {
	if msg.Event == string(coreTypes.StateMachineEvent_Consensus_DisableServerMode) {
		return m.DisableServerMode()
	}
	return m.EnableServerMode()

}
