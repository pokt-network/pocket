package consensus

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

// State machine transition event comes to consensus module
// onsensus moduel reacts upon the new changed state
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
		return m.HandleSync(msg)

	case string(coreTypes.StateMachineState_Consensus_Synced):
		return m.HandleSynced(msg)

	case string(coreTypes.StateMachineState_Consensus_Pacemaker):
		return m.HandlePacemaker(msg)

	case string(coreTypes.StateMachineState_Consensus_Server_Enabled), string(coreTypes.StateMachineState_Consensus_Server_Disabled):
		return m.HandleServerMode(msg)
	}

	return nil
}

// Unsynched mode is when the node (validator or non-valdiator) is out of sync with the rest of the network
// This mode is a transition mode from node being up-to-date (i.e. Pacemaker mode, Synched mode) to the latest state to node being out-of-sync
// As soon as node transitions to this mode, it will transition to the sync mode.
func (m *consensusModule) HandleUnsynched(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("In HandleUnsyched, as node is out of sync, sending syncmode event to start syncing")
	if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing); err != nil {
		return err
	}

	return nil
}

// Sync mode is when the node (validator or non-valdiator) is syncing with the rest of the network
func (m *consensusModule) HandleSync(msg *messaging.StateMachineTransitionEvent) error {
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

// Currently we never transition to this state.
// Basically, a non-validator node always stays in syncmode.
// CONSIDER: when a non-validator sync is implemented, maybe there is a case that requires transitioning to this state
func (m *consensusModule) HandleSynced(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("Non-validator node is in Synced mode")
	return nil
}

// Pacemaker mode is when the validator is synced and it is waiting for a new block proposal to come in
func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("Validator is synced and in Pacemaker mode. It will stay in this mode until it receives a new block proposal that has a higher height than the current block height")
	// validator receives a new block proposal, and it understands that it doesn't have block and it transitions to unsycnhed state
	// transitioning out of this state happens when a new block proposal is received by the hotstuff_replica
	return nil
}

// Server mode runs mutually exclusive to the rest of the modes, thus its state changes doesn't affect the other modes
// Server mode changes only happen through the node config and EnableServerMode() and DisableServerMode() functions
func (m *consensusModule) HandleServerMode(msg *messaging.StateMachineTransitionEvent) error {
	if msg.Event == string(coreTypes.StateMachineEvent_Consensus_DisableServerMode) {
		return m.DisableServerMode()
	}
	return m.EnableServerMode()

}
