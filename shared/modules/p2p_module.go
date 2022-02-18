package modules

import (
	"google.golang.org/protobuf/types/known/anypb"
	"pocket/shared/types"
)

type NetworkMessage struct {
	Topic types.EventTopic
	Data  []byte
}

type NetworkModule interface {
	Module
	BroadcastMessage(msg *anypb.Any, topic string) error  // TODO get rid of topic
	Send(addr string, msg *anypb.Any, topic string) error // TODO get rid of topic
}
