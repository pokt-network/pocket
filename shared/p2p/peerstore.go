package p2p

import "github.com/pokt-network/pocket/shared/crypto"

type Peerstore interface {
	AddPeer(peer Peer) error
	RemovePeer(addr crypto.Address) error

	GetPeer(addr crypto.Address) Peer
	GetPeerFromString(string) Peer
	GetAllPeers() PeerList

	Size() int
}

// PeerAddrMap implements the `Peerstore` interface.
type PeerAddrMap map[string]Peer

// GetPeer implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetPeer(addr crypto.Address) Peer {
	return paMap[addr.String()]
}

// GetPeerFromString implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetPeerFromString(addrStr string) Peer {
	return paMap.GetPeer(crypto.AddressFromString(addrStr))
}

// GetAllPeers implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetAllPeers() (peerList PeerList) {
	for _, peer := range paMap {
		peerList = append(peerList, peer)
	}
	return peerList
}

// AddPeer implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) AddPeer(peer Peer) error {
	paMap[peer.GetAddress().String()] = peer
	return nil
}

// RemovePeer implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) RemovePeer(addr crypto.Address) error {
	delete(paMap, addr.String())
	return nil
}

// Size implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) Size() int {
	return len(paMap)
}
