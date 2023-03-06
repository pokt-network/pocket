package raintree

import (
	"testing"

	"github.com/golang/mock/gomock"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	sharedP2P "github.com/pokt-network/pocket/shared/p2p"
	"github.com/stretchr/testify/require"
)

func TestRainTreeNetwork_AddPeer(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Start with a peerstore containing self.
	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	expectedPStoreSize := 0
	pstore := getPeerstore(nil, expectedPStoreSize)

	// Add self to peerstore.
	err = pstore.AddPeer(&typesP2P.NetworkPeer{Address: selfAddr})
	require.NoError(t, err)
	expectedPStoreSize++

	busMock := mockBus(ctrl)
	peerstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

	network := NewRainTreeNetwork(selfAddr, busMock, peerstoreProviderMock, currentHeightProviderMock).(*rainTreeNetwork)

	peerAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	peerToAdd := &typesP2P.NetworkPeer{Address: peerAddr}

	// Add peerToAdd.
	err = network.AddPeer(peerToAdd)
	require.NoError(t, err)
	expectedPStoreSize++

	peersView, peerAddrs, peers := getPeersViewParts(network.peersManager)

	// Ensure size / lengths are consistent.
	require.Equal(t, expectedPStoreSize, peersView.GetPeerstore().Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	require.ElementsMatch(t, []string{selfAddr.String(), peerAddr.String()}, peerAddrs, "addresses do not match")
	require.ElementsMatch(t, []*typesP2P.NetworkPeer{selfPeer, peerToAdd}, peers, "peers do not match")

	require.Equal(t, selfPeer, peersView.GetPeerstore().GetPeer(selfAddr), "Peerstore does not contain self")
	require.Equal(t, peerToAdd, peersView.GetPeerstore().GetPeer(peerAddr), "Peerstore does not contain added peer")
}

func TestRainTreeNetwork_RemovePeer(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Start with a peerstore which contains self and some number of peers: the
	// initial value of `expectedPStoreSize`.
	expectedPStoreSize := 3
	pstore := getPeerstore(nil, expectedPStoreSize)

	selfAddr, err := cryptoPocket.GenerateAddress()
	require.NoError(t, err)
	selfPeer := &typesP2P.NetworkPeer{Address: selfAddr}

	// TODO _THIS_COMMIT: finish comment
	// Add self to peerstore because ... it's expected by something
	err = pstore.AddPeer(&typesP2P.NetworkPeer{Address: selfAddr})
	require.NoError(t, err)
	expectedPStoreSize++

	busMock := mockBus(ctrl)
	peerstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

	network := NewRainTreeNetwork(selfAddr, busMock, peerstoreProviderMock, currentHeightProviderMock).(*rainTreeNetwork)

	// Ensure expected starting size / lengths are consistent.
	peersView, peerAddrs, peers := getPeersViewParts(network.peersManager)
	require.Equal(t, expectedPStoreSize, pstore.Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	var peerToRemove sharedP2P.Peer
	// Ensure we don't remove selfPeer. `Peerstore` interface isn't aware
	// of the concept of "self" so we have to find it.
	for _, peer := range pstore.GetAllPeers() {
		if peer.GetAddress().Equals(selfAddr) {
			continue
		}
		peerToRemove = peer
		break
	}
	require.NotNil(t, peerToRemove, "did not find selfAddr in peerstore")

	// Remove peerToRemove
	err = network.RemovePeer(peerToRemove)
	require.NoError(t, err)
	expectedPStoreSize--

	peersView, peerAddrs, peers = getPeersViewParts(network.peersManager)
	removedAddr := peerToRemove.GetAddress()
	getPeer := func(addr cryptoPocket.Address) sharedP2P.Peer {
		return peersView.GetPeerstore().GetPeer(addr)
	}

	// Ensure updated size / lengths are consistent.
	require.Equal(t, expectedPStoreSize, pstore.Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	require.Equal(t, selfPeer, getPeer(selfAddr), "Peerstore does not contain self")
	require.Nil(t, getPeer(removedAddr), "Peerstore contains removed peer")
}

func getPeersViewParts(pm sharedP2P.PeerManager) (
	view sharedP2P.PeersView,
	addrs []string,
	peers sharedP2P.PeerList) {

	view = pm.GetPeersView()
	addrs = view.GetAddrs()
	peers = view.GetPeers()

	return view, addrs, peers
}
