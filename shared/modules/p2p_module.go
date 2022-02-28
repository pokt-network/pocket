package modules

import (
	"github.com/pokt-network/pocket/shared/types"

	"google.golang.org/protobuf/types/known/anypb"
)

type NetworkMessage struct {
	Topic types.EventTopic
	Data  []byte
}

type P2PModule interface {
	Module
	BroadcastMessage(msg *anypb.Any, topic string) error  // TODO(derrandz): get rid of topic
	Send(addr string, msg *anypb.Any, topic string) error // TODO(derrandz): get rid of topic
}
