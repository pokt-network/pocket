package types

import (
	"io"

	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/shared/crypto"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
)

var _ sharedP2P.Peer = &NetworkPeer{}

type NetworkPeer struct {
	Dialer    Transport
	PublicKey crypto.PublicKey
	Address   crypto.Address
	Multiaddr multiaddr.Multiaddr

	// This is only included because it's a more human-friendly differentiator between peers
	ServiceURL string
}

func (peer *NetworkPeer) GetAddress() crypto.Address {
	return peer.Address
}

func (peer *NetworkPeer) GetStream() io.ReadWriteCloser {
	return peer.Dialer
}

func (peer *NetworkPeer) GetPublicKey() crypto.PublicKey {
	return peer.PublicKey
}

func (peer *NetworkPeer) GetServiceURL() string {
	return peer.ServiceURL
}
