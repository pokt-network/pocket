package raintree

import (
	"testing"

	"github.com/pokt-network/pocket/p2p/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestRainTreeNetwork_AddPeerToAddrBook(t *testing.T) {
	// starting with an empty address book and only self
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	addrBook := getAddrBook(nil, 0)
	addrBook = append(addrBook, &types.NetworkPeer{Address: selfAddr})
	network := NewRainTreeNetwork(selfAddr, addrBook).(*rainTreeNetwork)

	peerAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)

	peer := &typesP2P.NetworkPeer{Address: peerAddr}

	// adding a peer
	err = network.AddPeerToAddrBook(peer)
	require.NoError(t, err)

	stateView := network.peersManager.getNetworkView()
	require.Equal(t, 2, len(stateView.addrList))
	require.ElementsMatch(t, []string{selfAddr.String(), peerAddr.String()}, stateView.addrList, "addrList does not match")
	require.ElementsMatch(t, []*types.NetworkPeer{selfPeer, peer}, stateView.addrBook, "addrBook does not match")

	require.Contains(t, stateView.addrBookMap, selfAddr.String(), "addrBookMap does not contain self key")
	require.Equal(t, selfPeer, stateView.addrBookMap[selfAddr.String()], "addrBookMap does not contain self")
	require.Contains(t, stateView.addrBookMap, peerAddr.String(), "addrBookMap does not contain peer key")
	require.Equal(t, peer, stateView.addrBookMap[peerAddr.String()], "addrBookMap does not contain peer")
}

func TestRainTreeNetwork_RemovePeerToAddrBook(t *testing.T) {
	// starting with an address book having only self and an arbitrary number of peers `numAddressesInAddressBook``
	numAddressesInAddressBook := 3
	addrBook := getAddrBook(nil, numAddressesInAddressBook)
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	addrBook = append(addrBook, &types.NetworkPeer{Address: selfAddr})
	network := NewRainTreeNetwork(selfAddr, addrBook).(*rainTreeNetwork)

	stateView := network.peersManager.getNetworkView()
	require.Equal(t, numAddressesInAddressBook+1, len(stateView.addrList)) // +1 to account for self in the addrBook as well

	// removing a peer
	peer := addrBook[1]
	err = network.RemovePeerToAddrBook(peer)
	require.NoError(t, err)

	stateView = network.peersManager.getNetworkView()
	require.Equal(t, numAddressesInAddressBook+1-1, len(stateView.addrList)) // +1 to account for self and the peer removed

	require.Contains(t, stateView.addrBookMap, selfAddr.String(), "addrBookMap does not contain self key")
	require.Equal(t, selfPeer, stateView.addrBookMap[selfAddr.String()], "addrBookMap contains self")
	require.NotContains(t, stateView.addrBookMap, peer.Address.String(), "addrBookMap contains removed peer key")
}
