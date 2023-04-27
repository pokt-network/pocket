package types

//go:generate mockgen -source=$GOFILE -destination=./mocks/network_mock.go -package=mock_types github.com/pokt-network/pocket/p2p/types Network

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
)

// TECHDEBT(olshansky): When we delete `stdnetwork` and only go with `raintree`, this interface
// can be simplified greatly.
type Router interface {
	modules.IntegratableModule

	Broadcast(data []byte) error
	NetworkSend(data []byte, address cryptoPocket.Address) error

	// Address book helpers
	// TECHDEBT: simplify - remove `GetPeerstore`
	GetPeerstore() Peerstore
	AddPeer(peer Peer) error
	RemovePeer(peer Peer) error

	// This function was added to specifically support the RainTree implementation.
	// Handles the raw data received from the network and returns the data to be processed
	// by the application layer.
	HandleNetworkData(data []byte) ([]byte, error)
}
