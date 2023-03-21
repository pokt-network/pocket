package p2p

import (
	"io"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Peer interface {
	GetAddress() crypto.Address
	GetPublicKey() crypto.PublicKey
	GetServiceURL() string

	// TECHDEBT(#576): move this to some new `ConnManager` interface.
	GetStream() io.ReadWriteCloser
}

// PeerList is a convenience type for operating on a slice of `Peer`s.
type PeerList []Peer

// Delta returns the difference between two PeerList slices
func (peers PeerList) Delta(comparePeers PeerList) (added, removed PeerList) {
	commonPeers := make(PeerAddrMap)
	for _, p := range peers {
		commonPeers.mustAddPeer(p)
	}

	for _, comparePeer := range comparePeers {
		addedPeer := commonPeers.GetPeer(comparePeer.GetAddress())
		if addedPeer == nil {
			added = append(added, comparePeer)
			continue
		}

		// `comparePeer` exists in both `peers` and `comparePeers`;
		// neither added nor removed.
		commonPeers.mustRemovePeer(comparePeer.GetAddress())
	}

	for _, removedPeer := range commonPeers {
		removed = append(removed, removedPeer)
	}
	return
}
