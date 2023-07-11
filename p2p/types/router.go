package types

//go:generate mockgen -package=mock_types -destination=./mocks/network_mock.go github.com/pokt-network/pocket/p2p/types Router,RouterConfig

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// TECHDEBT(olshansky): When we delete `stdnetwork` and only go with `raintree`, this interface
// can be simplified greatly.
type Router interface {
	modules.IntegratableModule

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
