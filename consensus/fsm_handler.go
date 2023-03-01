package consensus

import (
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

func (m *consensusModule) handleStateMachineTransitionEvent(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Msgf("Begin handling StateMachineTransitionEvent: %s", msg)

	fsm_state := msg.NewState
	m.logger.Debug().Fields(map[string]any{
		"event":          msg.Event,
		"previous_state": msg.PreviousState,
		"new_state":      fsm_state,
	}).Msg("Received state machine transition msg")

	switch coreTypes.StateMachineState(fsm_state) {
	case coreTypes.StateMachineState_P2P_Bootstrapped:
		return m.HandleBootstrapped(msg)

	case coreTypes.StateMachineState_Consensus_Unsynched:
		return m.HandleUnsynched(msg)

	case coreTypes.StateMachineState_Consensus_SyncMode:
		return m.HandleSync(msg)

	case coreTypes.StateMachineState_Consensus_Synced:
		return m.HandleSynced(msg)

	case coreTypes.StateMachineState_Consensus_Pacemaker:
		return m.HandlePacemaker(msg)
	}

	return nil
}

// Bootrstapped mode is when the node (validator or non-valdiator) is out of sync with the rest of the network
// This mode is a transition mode from node bootstrappin to node being out-of-sync.
func (m *consensusModule) HandleBootstrapped(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in bootstrapped state, so it is out of sync, and transitions to unsynched mode")
	if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynched); err != nil {
		return err
	}

	return nil
}

// Unsynched mode is when the node (validator or non-valdiator) is out of sync with the rest of the network
// This mode is a transition mode from node being up-to-date (i.e. Pacemaker mode, Synched mode) to the latest state to node being out-of-sync
// As soon as node transitions to this mode, it will transition to the sync mode.
func (m *consensusModule) HandleUnsynched(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in Unsyched state, as node is out of sync sending syncmode event to start syncing")
	if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing); err != nil {
		return err
	}

	return nil
}

// Sync mode is when the node (validator or non-valdiator) is syncing with the rest of the network
func (m *consensusModule) HandleSync(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM is in Sync Mode, starting syncing")

	m.stateSync.AggregateMetadataResponses()

	// try sycing until node is synced
	// CONSIDER: consider putting a limit on number of tries, or timeout
	err := m.stateSync.Sync()
	for err != nil {
		err = m.stateSync.Sync()
	}

	if m.IsValidator() {
		m.logger.Debug().Msg("Validator node synced to the latest state with the rest of the peers")
		if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsCaughtUpValidator); err != nil {
			return err
		}
	} else {
		m.logger.Debug().Msg("Non-Validator synced to the latest state with the rest of the peers")
		if err := m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsCaughtUpNonValidator); err != nil {
			return err
		}
	}

	return nil
}

// Currently, FSM never transition to this state and a non-validator node always stays in syncmode.
// CONSIDER: when a non-validator sync is implemented, maybe there is a case that requires transitioning to this state
func (m *consensusModule) HandleSynced(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM of non-validator node is in Synced mode")
	return nil
}

// Pacemaker mode is when the validator is synced and it is waiting for a new block proposal to come in
func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Debug().Msg("FSM of validator node is synced and in Pacemaker mode. It will stay in this mode until it receives a new block proposal that has a higher height than the current block height")
	// validator receives a new block proposal, and it understands that it doesn't have block and it transitions to unsycnhed state
	// transitioning out of this state happens when a new block proposal is received by the hotstuff_replica
	return nil
}
