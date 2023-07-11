package types

//go:generate mockgen -package=mock_types -destination=./mocks/network_mock.go github.com/pokt-network/pocket/p2p/types Router,RouterConfig

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// TECHDEBT(#880, #811): move these definitions to /shared/modules & /shared/modules/types packages.
// `Peerstore` would have to be moved as well which may create an import cycle.

// NOTE: this is the first case I'm aware of where we need multiple
// instances of a submodule to be dependency-injectable.
//
// CONSIDERATION: this is inconsistent with existing submodule "name" naming
// conventions as the "name" doesn't match the name of the interface. This is
// implied by the fact above, as two distinct "names" are needed to disambiguate
// in the module registry. These names are also distinct from the names of the
// respective `Router` implementations; my thinking is that these names better
// reflect the separation of concerns from the P2P module's perspective.
const (
	StakedActorRouterSubmoduleName   = "staked_actor_router"
	UnstakedActorRouterSubmoduleName = "unstaked_actor_router"
)

type Router interface {
	modules.Submodule

	Broadcast(data []byte) error
	Send(data []byte, address cryptoPocket.Address) error
	Close() error

	// GetPeerstore is used by the P2P module to update the staked actor router's
	// (`RainTreeRouter`) peerstore.
	//
	// TECHDEBT(#859+): remove the need for this group of interface methods.
	// All peer discovery logic should be encapsulated by the router.
	// Adopt `HandleEvent(*anypb.Any) error` here instead and forward events
	// from P2P module to routers.
	// CONSIDERATION: Utility, Conseneus and P2P modules could share an interface
	// containing this method (e.g. `BusEventHandler`).
	GetPeerstore() Peerstore
	// AddPeer is used to add a peer to the routers peerstore. It is intended to
	// support peer discovery.
	AddPeer(peer Peer) error
	// RemovePeer is used to remove a peer to the routers peerstore. It is used
	// during churn to purge offline peers from the routers peerstore.
	RemovePeer(peer Peer) error
}

type MessageHandler func(data []byte) error

// RouterConfig is used to configure `Router` implementations and to test a
// given configuration's validity.
type RouterConfig interface {
	IsValid() error
}
