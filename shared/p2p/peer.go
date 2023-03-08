package p2p

import (
	"io"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Peer interface {
	GetAddress() crypto.Address
	GetPublicKey() crypto.PublicKey
	GetServiceURL() string

	// TECHDEBT: move this to some new `ConnManager` interface.
	GetStream() io.ReadWriteCloser
}

type PeerList []Peer

// getPeerListDelta returns the difference between two PeerList slices
func (peerList PeerList) Delta(compare PeerList) (added, removed PeerList) {
	tempPStore := make(PeerAddrMap)
	for _, np := range peerList {
		tempPStore.mustAddPeer(np)
	}

	for _, comparePeer := range compare {
		if addedPeer := tempPStore.GetPeer(comparePeer.GetAddress()); addedPeer == nil {
			added = append(added, comparePeer)
			continue
		}
		tempPStore.mustRemovePeer(comparePeer.GetAddress())
	}

	for _, removedPeer := range tempPStore {
		removed = append(removed, removedPeer)
	}
	return
}
