package messaging

const (
	NodeStartedEventType            = "pocket.NodeStartedEvent"
	ConsensusNewHeightEventType     = "pocket.ConsensusNewHeightEvent"
	StateMachineTransitionEventType = "pocket.StateMachineTransitionEvent"

	HotstuffMessageContentType  = "consensus.HotstuffMessage"
	StateSyncMessageContentType = "consensus.StateSyncMessage"
)
