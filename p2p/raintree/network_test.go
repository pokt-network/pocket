package raintree

import (
	"testing"

	"github.com/golang/mock/gomock"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
	"github.com/stretchr/testify/require"
)

func TestRainTreeNetwork_AddPeerToAddrBook(t *testing.T) {
	ctrl := gomock.NewController(t)

	// starting with an empty address book and only self
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	addrBook := getAddrBook(nil, 0)
	addrBook = append(addrBook, &typesP2P.NetworkPeer{Address: selfAddr})

	busMock := mockBus(ctrl)
	addrBookProviderMock := mockAddrBookProvider(ctrl, addrBook)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

	network := NewRainTreeNetwork(selfAddr, busMock, addrBookProviderMock, currentHeightProviderMock).(*rainTreeNetwork)

	peerAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)

	peer := &typesP2P.NetworkPeer{Address: peerAddr}

	// adding a peer
	err = network.AddPeerToAddrBook(peer)
	require.NoError(t, err)

	stateView := network.peersManager.GetPeersView()
	require.Equal(t, 2, len(stateView.GetAddrs()))
	require.ElementsMatch(t, []string{selfAddr.String(), peerAddr.String()}, stateView.GetAddrs(), "addrList does not match")
	require.ElementsMatch(t, []*typesP2P.NetworkPeer{selfPeer, peer}, stateView.GetPeers(), "addrBook does not match")

	require.Equal(t, selfPeer, stateView.GetPeerstore().GetPeer(selfAddr), "Peerstore does not contain self")
	require.Equal(t, peer, stateView.GetPeerstore().GetPeer(peerAddr), "Peerstore does not contain peer")
}

func TestRainTreeNetwork_RemovePeerFromAddrBook(t *testing.T) {
	ctrl := gomock.NewController(t)

	// starting with an address book having only self and an arbitrary number of peers `numAddressesInAddressBook``
	numAddressesInAddressBook := 3
	addrBook := getAddrBook(nil, numAddressesInAddressBook)
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}
	addrBook = append(addrBook, &typesP2P.NetworkPeer{Address: selfAddr})

	busMock := mockBus(ctrl)
	addrBookProviderMock := mockAddrBookProvider(ctrl, addrBook)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

	network := NewRainTreeNetwork(selfAddr, busMock, addrBookProviderMock, currentHeightProviderMock).(*rainTreeNetwork)
	stateView := network.peersManager.GetPeersView()
	require.Equal(t, numAddressesInAddressBook+1, len(stateView.GetAddrs())) // +1 to account for self in the addrBook as well

	// removing a peer
	peer := addrBook[1]
	err = network.RemovePeerFromAddrBook(peer)
	require.NoError(t, err)

	stateView = network.peersManager.GetPeersView()
	require.Equal(t, numAddressesInAddressBook+1-1, len(stateView.GetAddrs())) // +1 to account for self and the peer removed

	require.Contains(t, stateView.GetPeerstore(), selfAddr.String(), "Peerstore does not contain self key")
	require.Equal(t, selfPeer, stateView.GetPeerstore().GetPeer(selfAddr), "Peerstore contains self")
	require.NotContains(t, stateView.GetPeerstore(), peerToRemove.GetAddress().String(), "Peerstore contains removed peer key")
}
