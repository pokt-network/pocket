package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/p2p_module_mock.go -aux_files=github.com/pokt-network/pocket/internal/shared/modules=module.go

import (
	cryptoPocket "github.com/pokt-network/pocket/internal/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

type P2PModule interface {
	Module
	ConfigurableModule

	Broadcast(msg *anypb.Any) error
	Send(addr cryptoPocket.Address, msg *anypb.Any) error
	GetAddress() (cryptoPocket.Address, error)
}
