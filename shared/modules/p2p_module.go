package modules

import (
	"pocket/shared/types"
)

type NetworkMessage struct {
	Topic types.PocketEventTopic
	Data  []byte
}

type NetworkModule interface {
	Module
	BroadcastMessage(msg *types.NetworkMessage) error
	Send(addr string, msg *types.NetworkMessage) error
}
