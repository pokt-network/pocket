package messaging

const (
	// Node
	NodeStartedEventType            = "pocket.NodeStartedEvent"
	ConsensusNewHeightEventType     = "pocket.ConsensusNewHeightEvent"
	StateMachineTransitionEventType = "pocket.StateMachineTransitionEvent"

	// Consensus
	HotstuffMessageContentType  = "consensus.HotstuffMessage"
	StateSyncMessageContentType = "consensus.StateSyncMessage"

	// Utility
	TxGossipMessageContentType = "utility.TxGossipMessage"
)

// Helper logger for state sync tranition events
func TransitionEventToMap(stateSyncMsg *StateMachineTransitionEvent) map[string]any {
	return map[string]any{
		"state_machine_event": stateSyncMsg.Event,
		"previous_state":      stateSyncMsg.PreviousState,
		"new_state":           stateSyncMsg.NewState,
	}
}
