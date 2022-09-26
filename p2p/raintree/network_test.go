package raintree

import (
	"testing"

	"github.com/pokt-network/pocket/p2p/types"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func Test_rainTreeNetwork_AddPeerToAddrBook(t *testing.T) {
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

	stateView := network.peersManager.getStateView()
	require.Equal(t, 2, len(stateView.addrList))
	require.ElementsMatch(t, []string{selfAddr.String(), peerAddr.String()}, stateView.addrList, "addrList")
	require.ElementsMatch(t, []*types.NetworkPeer{selfPeer, peer}, stateView.addrBook, "addrBook")

	require.Contains(t, stateView.addrBookMap, selfAddr.String(), "addrBookMap contains self key")
	require.Equal(t, selfPeer, stateView.addrBookMap[selfAddr.String()], "addrBookMap contains self")
	require.Contains(t, stateView.addrBookMap, peerAddr.String(), "addrBookMap contains peer key")
	require.Equal(t, peer, stateView.addrBookMap[peerAddr.String()], "addrBookMap contains peer")
}

func Test_rainTreeNetwork_RemovePeerToAddrBook(t *testing.T) {
	// starting with an address book having only self and an arbitrary number of peers `numAddressesInAddressBook``
	numAddressesInAddressBook := 3
	addrBook := getAddrBook(nil, numAddressesInAddressBook)
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	addrBook = append(addrBook, &types.NetworkPeer{Address: selfAddr})
	network := NewRainTreeNetwork(selfAddr, addrBook).(*rainTreeNetwork)

	stateView := network.peersManager.getStateView()
	require.Equal(t, numAddressesInAddressBook+1, len(stateView.addrList))

	// removing a peer
	peer := addrBook[1]
	err = network.RemovePeerToAddrBook(peer)
	require.NoError(t, err)

	stateView = network.peersManager.getStateView()
	require.Equal(t, numAddressesInAddressBook+1-1, len(stateView.addrList))

	require.Contains(t, stateView.addrBookMap, selfAddr.String(), "addrBookMap contains self key")
	require.Equal(t, selfPeer, stateView.addrBookMap[selfAddr.String()], "addrBookMap contains self")
	require.NotContains(t, stateView.addrBookMap, peer.Address.String(), "addrBookMap contains removed peer key")
}
