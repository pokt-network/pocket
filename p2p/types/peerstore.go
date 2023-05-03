package types

import (
	"fmt"
	"log"

	"github.com/pokt-network/pocket/shared/crypto"
)

type Peerstore interface {
	AddPeer(peer Peer) error
	RemovePeer(addr crypto.Address) error

	GetPeer(addr crypto.Address) Peer
	GetPeerFromString(addrHex string) Peer
	GetPeerList() PeerList

	Size() int
}

// PeerAddrMap implements the `Peerstore` interface.
type PeerAddrMap map[string]Peer

// GetPeer implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetPeer(addr crypto.Address) Peer {
	return paMap[addr.String()]
}

// GetPeerFromString implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetPeerFromString(addrHex string) Peer {
	return paMap.GetPeer(crypto.AddressFromString(addrHex))
}

// GetAllPeers implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) GetPeerList() (peerList PeerList) {
	for _, peer := range paMap {
		peerList = append(peerList, peer)
	}
	return peerList
}

// AddPeer implements the respective `Peerstore` interface member.
func (paMap PeerAddrMap) AddPeer(peer Peer) error {
	addr := peer.GetAddress().String()
	if _, ok := paMap[addr]; ok {
		return fmt.Errorf("peer exists, remove first; addr: %s", addr)
	}

	paMap[addr] = peer
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

// mustAddrPeer calls `#AddPeer` and panics if it returns an error.
func (paMap PeerAddrMap) mustAddPeer(peer Peer) {
	if err := paMap.AddPeer(peer); err != nil {
		log.Fatalf("in PeerAddrMap#mustAddPeer: %s", err)
	}
}

// mustRemovePeer calls `#RemovePeer` and panics if it returns an error.
func (paMap PeerAddrMap) mustRemovePeer(addr crypto.Address) {
	if err := paMap.RemovePeer(addr); err != nil {
		log.Fatalf("in PeerAddrMap#mustRemovePeer: %s", err)
	}
}
