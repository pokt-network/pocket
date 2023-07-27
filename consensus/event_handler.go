package consensus

import (
	"fmt"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
	"google.golang.org/protobuf/types/known/anypb"
)

// RESEARCH(#816): Research whether the E2E state sync business logic can be simplified by not using the FSM module at all.
// We were originally intending to make heavier use of the FSM module to handle state sync, but we ended up not using it much as
// seen below.

const (
	consensusFSMHandlerSource = "ConsensusFSMHandler"
)

// Implements the `HandleEvent` function in the `ConsensusModule` interface
func (m *consensusModule) HandleEvent(event *anypb.Any) error {
	m.m.Lock()
	defer m.m.Unlock()

	msg, err := codec.GetCodec().FromAny(event)
	if err != nil {
		return err
	}

	switch event.MessageName() {

	case messaging.StateMachineTransitionEventType:
		stateTransitionMessage, ok := msg.(*messaging.StateMachineTransitionEvent)
		if !ok {
			return fmt.Errorf("failed to cast message to StateSyncMessage")
		}
		return m.handleStateTransitionEvent(stateTransitionMessage)

	case messaging.ConsensusNewHeightEventType:
		blockCommittedEvent, ok := msg.(*messaging.ConsensusNewHeightEvent)
		if !ok {
			return fmt.Errorf("failed to cast event to ConsensusNewHeightEvent")
		}
		m.stateSync.HandleBlockCommittedEvent(blockCommittedEvent)
		return nil
	default:
		return typesCons.ErrUnknownStateSyncMessageType(event.MessageName())
	}
}

// handleStateTransitionEvent handles the state transition event from the state machine module
func (m *consensusModule) handleStateTransitionEvent(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Fields(messaging.TransitionEventToMap(msg)).Msg("Received state machine transition msg")

	switch coreTypes.StateMachineState(msg.NewState) {
	case coreTypes.StateMachineState_P2P_Bootstrapped:
		return m.HandleBootstrapped(msg)

	case coreTypes.StateMachineState_Consensus_Unsynced:
		return m.HandleUnsynced(msg)

	case coreTypes.StateMachineState_Consensus_SyncMode:
		return m.HandleSyncMode(msg)

	case coreTypes.StateMachineState_Consensus_Synced:
		return m.HandleSynced(msg)

	case coreTypes.StateMachineState_Consensus_Pacemaker:
		return m.HandlePacemaker(msg)

	default:
		m.logger.Warn().Msgf("Consensus module not handling this event: %s", msg.Event)

	}

	return nil
}

// HandleBootstrapped handles the FSM event P2P_IsBootstrapped, and when P2P_Bootstrapped is the destination state.
// Bootstrapped mode is when the node (validator or non) is first coming online.
// This is a transition mode from node bootstrapping to a node being out-of-sync.
func (m *consensusModule) HandleBootstrapped(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Str("source", consensusFSMHandlerSource).Msg("Node is in the bootstrapped state. Transitioning to IsUnsynched mode...")
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsUnsynced)
}

// HandleUnsynced handles the FSM event Consensus_IsUnsynced, and when Unsynced is the destination state.
// In Unsynced mode, the node (validator or not) is out of sync with the rest of the network.
// This mode is a transition mode from the node being up-to-date (i.e. Pacemaker mode, Synced mode) with the latest network height to being out-of-sync.
// As soon as a node transitions to this mode, it will transition to the synching mode.
func (m *consensusModule) HandleUnsynced(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Str("source", consensusFSMHandlerSource).Msg("Node is in an Unsynced state. Transitioning to IsSyncing mode...")
	return m.GetBus().GetStateMachineModule().SendEvent(coreTypes.StateMachineEvent_Consensus_IsSyncing)
}

// HandleSyncMode handles the FSM event Consensus_IsSyncing, and when SyncMode is the destination state.
// In Sync mode, the node (validator or not starts syncing with the rest of the network.
func (m *consensusModule) HandleSyncMode(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Str("source", consensusFSMHandlerSource).Msg("Node is in Sync Mode. About to start synchronous sync loop...")
	go m.stateSync.StartSynchronousStateSync()
	return nil
}

// HandleSynced handles the FSM event IsSyncedNonValidator for Non-Validators, and Synced is the destination state.
// Currently, FSM never transition to this state and a non-validator node always stays in SyncMode.
// CONSIDER: when a non-validator sync is implemented, maybe there is a case that requires transitioning to this state.
func (m *consensusModule) HandleSynced(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Str("source", consensusFSMHandlerSource).Msg("Node (non-validator) is Synced. NOOP")
	return nil
}

// HandlePacemaker handles the FSM event IsSyncedValidator, and Pacemaker is the destination state.
// Execution of this state means the validator node is synced and it will stay in this mode until
// it receives a new block proposal that has a higher height than the current consensus height.
func (m *consensusModule) HandlePacemaker(msg *messaging.StateMachineTransitionEvent) error {
	m.logger.Info().Str("source", consensusFSMHandlerSource).Msg("Node (validator) node is Synced and entering Pacemaker mode. About to starting participating in consensus...")

	// if a validator is just bootstrapped and finished state sync, it will not have a nodeId yet, which is 0. Set correct nodeId here.
	if m.nodeId == 0 {
		return m.updateNodeId()
	}

	return nil
}
