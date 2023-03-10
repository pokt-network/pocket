package messaging

import "github.com/pokt-network/pocket/shared/core/types"

const (
	// Node
	NodeStartedEventType            = "pocket.NodeStartedEvent"
	ConsensusNewHeightEventType     = "pocket.ConsensusNewHeightEvent"
	StateMachineTransitionEventType = "pocket.StateMachineTransitionEvent"

	// Utility
	TxGossipMessageContentType = "utility.TxGossipMessage"
)

// Helper logger for state sync tranition events
func EventToMap(stateSyncMsg *StateMachineTransitionEvent) map[string]any {
	fields := map[string]any{
		"previous_state": stateSyncMsg.PreviousState,
		"new_state":      stateSyncMsg.NewState,
	}

	switch types.StateMachineEvent(stateSyncMsg.Event) {
	case types.StateMachineEvent_P2P_IsBootstrapped:
		fields["proto_type"] = "StateMachineEvent_P2P_IsBootstrapped"
	case types.StateMachineEvent_Consensus_IsUnsynched:
		fields["proto_type"] = "StateMachineEvent_Consensus_IsUnsynched"
	case types.StateMachineEvent_Consensus_IsCaughtUp:
		fields["proto_type"] = "StateMachineEvent_Consensus_IsSynched"
	case types.StateMachineEvent_Consensus_IsSyncing:
		fields["proto_type"] = "StateMachineEvent_Consensus_IsSyncing"
	}

	return fields
}
