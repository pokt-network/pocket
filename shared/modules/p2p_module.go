package modules

import (
	"pocket/shared/events"
	"pocket/shared/typespb"
)

type NetworkMessage struct {
	Topic events.PocketEventTopic
	Data  []byte
}

type NetworkModule interface {
	PocketModule

	BroadcastMessage(msg *typespb.NetworkMessage) error
	Send(addr string, msg *typespb.NetworkMessage) error
}
