package modules

import (
	"pocket/shared/events"
	"pocket/shared/messages"
)

type NetworkMessage struct {
	Topic events.PocketEventTopic
	Data  []byte
}

type NetworkModule interface {
	PocketModule

	BroadcastMessage(msg *messages.NetworkMessage) error
	Send(addr string, msg *messages.NetworkMessage) error
}
