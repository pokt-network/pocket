package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/p2p_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module

	Broadcast(msg *anypb.Any) error
	Send(addr cryptoPocket.Address, msg *anypb.Any) error
	GetAddress() (cryptoPocket.Address, error)
}
