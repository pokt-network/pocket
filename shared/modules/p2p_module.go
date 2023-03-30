package modules

//go:generate mockgen -destination=./mocks/p2p_module_mock.go github.com/pokt-network/pocket/shared/modules P2PModule

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"google.golang.org/protobuf/types/known/anypb"
)

const P2PModuleName = "p2p"

type P2PModule interface {
	Module

	// Returns the public P2P address of this node
	GetAddress() (cryptoPocket.Address, error)

	// A network broadcast to all staked actors on the network using RainTree
	Broadcast(msg *anypb.Any) error

	// A direct asynchronous
	Send(addr cryptoPocket.Address, msg *anypb.Any) error

	// HandleEvent is used to react to events that occur inside the application
	HandleEvent(*anypb.Any) error

	// CONSIDERATION: The P2P module currently does implement a synchronous "request-response" pattern
	//                for core business logic between nodes. Rather, all communication is done
	//                asynchronously via a "fire-and-forget" pattern using `Send` and `Broadcast`.
	//                There are pros and cons to both, and future readers/maintainers/developers may
	//                consider the addition of a function similar to the one below.
	// Request(addr cryptoPocket.Address, msg *anypb.Any) *anypb.Any
}
