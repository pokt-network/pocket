package modules

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module
	Broadcast(msg *anypb.Any, topic types.PocketTopic) error                       // TODO(derrandz): get rid of topic
	Send(addr cryptoPocket.Address, msg *anypb.Any, topic types.PocketTopic) error // TODO(derrandz): get rid of topic
}
