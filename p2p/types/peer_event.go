package types

type PeerEvent int

const (
	BroadcastDoneEvent PeerEvent = iota
	PeerConnectedEvent
	PeerDisconnectedEvent
	NewMessageReceivedEvent
)
