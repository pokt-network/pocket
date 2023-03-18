package types

import (
	"github.com/multiformats/go-multiaddr"

	"github.com/pokt-network/pocket/shared/crypto"
)

var _ Peer = &NetworkPeer{}

type NetworkPeer struct {
	PublicKey crypto.PublicKey
	Address   crypto.Address
	Multiaddr multiaddr.Multiaddr

	// This is only included because it's a more human-friendly differentiator between peers
	ServiceURL string
}

func (peer *NetworkPeer) GetAddress() crypto.Address {
	return peer.Address
}

func (peer *NetworkPeer) GetPublicKey() crypto.PublicKey {
	return peer.PublicKey
}

func (peer *NetworkPeer) GetServiceURL() string {
	return peer.ServiceURL
}
