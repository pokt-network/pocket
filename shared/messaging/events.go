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
