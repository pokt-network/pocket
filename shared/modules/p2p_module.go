package modules

import (
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module
	BroadcastMessage(msg *anypb.Any, topic types.PocketTopic) error           // TODO(derrandz): get rid of topic
	Send(addr pcrypto.Address, msg *anypb.Any, topic types.PocketTopic) error // TODO(derrandz): get rid of topic
}
