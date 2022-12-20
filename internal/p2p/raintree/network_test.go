package raintree

import (
	"testing"

	"github.com/golang/mock/gomock"
	typesP2P "github.com/pokt-network/pocket/internal/p2p/types"
	"github.com/pokt-network/pocket/internal/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/internal/shared/crypto"
	mockModules "github.com/pokt-network/pocket/internal/shared/modules/mocks"
	"github.com/stretchr/testify/require"
)

func TestRainTreeNetwork_AddPeerToAddrBook(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockP2PCfg := mockModules.NewMockP2PConfig(ctrl)
	mockP2PCfg.EXPECT().GetMaxMempoolCount().Return(defaults.DefaultP2PMaxMempoolCount).AnyTimes()

	// starting with an empty address book and only self
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	addrBook := getAddrBook(nil, 0)
	addrBook = append(addrBook, &typesP2P.NetworkPeer{Address: selfAddr})

	busMock := mockBus(ctrl)
	addrBookProviderMock := mockAddrBookProvider(ctrl, addrBook)

	network := NewRainTreeNetwork(selfAddr, busMock, mockP2PCfg, addrBookProviderMock).(*rainTreeNetwork)

	peerAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)

	peer := &typesP2P.NetworkPeer{Address: peerAddr}

	// adding a peer
	err = network.AddPeerToAddrBook(peer)
	require.NoError(t, err)

	stateView := network.peersManager.getNetworkView()
	require.Equal(t, 2, len(stateView.addrList))
	require.ElementsMatch(t, []string{selfAddr.String(), peerAddr.String()}, stateView.addrList, "addrList does not match")
	require.ElementsMatch(t, []*typesP2P.NetworkPeer{selfPeer, peer}, stateView.addrBook, "addrBook does not match")

	require.Contains(t, stateView.addrBookMap, selfAddr.String(), "addrBookMap does not contain self key")
	require.Equal(t, selfPeer, stateView.addrBookMap[selfAddr.String()], "addrBookMap does not contain self")
	require.Contains(t, stateView.addrBookMap, peerAddr.String(), "addrBookMap does not contain peer key")
	require.Equal(t, peer, stateView.addrBookMap[peerAddr.String()], "addrBookMap does not contain peer")
}

func TestRainTreeNetwork_RemovePeerToAddrBook(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockP2PCfg := mockModules.NewMockP2PConfig(ctrl)
	mockP2PCfg.EXPECT().GetMaxMempoolCount().Return(defaults.DefaultP2PMaxMempoolCount).AnyTimes()

	// starting with an address book having only self and an arbitrary number of peers `numAddressesInAddressBook``
	numAddressesInAddressBook := 3
	addrBook := getAddrBook(nil, numAddressesInAddressBook)
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}
	addrBook = append(addrBook, &typesP2P.NetworkPeer{Address: selfAddr})

	busMock := mockBus(ctrl)
	addrBookProviderMock := mockAddrBookProvider(ctrl, addrBook)

	network := NewRainTreeNetwork(selfAddr, busMock, mockP2PCfg, addrBookProviderMock).(*rainTreeNetwork)
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
