package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/p2p_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/debug"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module
	ConfigurableModule

	Broadcast(msg *anypb.Any, topic debug.PocketTopic) error                       // TECHDEBT: get rid of topic
	Send(addr cryptoPocket.Address, msg *anypb.Any, topic debug.PocketTopic) error // TECHDEBT: get rid of topic
	GetAddress() (cryptoPocket.Address, error)
}
