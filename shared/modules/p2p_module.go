package modules

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module
	Broadcast(msg *anypb.Any, topic debug.PocketTopic) error                       // TODO(TECHDEBT): get rid of topic
	Send(addr cryptoPocket.Address, msg *anypb.Any, topic debug.PocketTopic) error // TODO(TECHDEBT): get rid of topic
	GetAddress() (cryptoPocket.Address, error)
}
