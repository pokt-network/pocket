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

	// Address book helpers
	// TECHDEBT: simplify - remove `GetPeerstore`
	GetPeerstore() Peerstore
	AddPeer(peer Peer) error
	RemovePeer(peer Peer) error
}

type MessageHandler func(data []byte) error

// RouterConfig is used to configure `Router` implementations and to test a
// given configuration's validity.
type RouterConfig interface {
	IsValid() error
}
