package raintree

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

// TECHDEBT(#609): move & de-dup.
var testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)

func TestRainTreeNetwork_AddPeer(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Start with a peerstore containing self.
	selfPeer, host := newTestPeer(t)
	selfAddr := selfPeer.GetAddress()

	expectedPStoreSize := 0
	pstore := getPeerstore(t, expectedPStoreSize)
	peers := pstore.GetPeerList()
	for _, peer := range peers {
		libp2pPeerInfo, err := utils.Libp2pAddrInfoFromPeer(peer)
		require.NoError(t, err)

		host.Peerstore().AddAddrs(libp2pPeerInfo.ID, libp2pPeerInfo.Addrs, time.Hour)

		libp2pPeerPubKey, err := utils.Libp2pPublicKeyFromPeer(peer)
		require.NoError(t, err)

		err = host.Peerstore().AddPubKey(libp2pPeerInfo.ID, libp2pPeerPubKey)
		require.NoError(t, err)
	}

	// Add self to peerstore.
	err := pstore.AddPeer(selfPeer)
	require.NoError(t, err)
	expectedPStoreSize++

	busMock := mockBus(ctrl)
	peerstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)

	netCfg := RainTreeConfig{
		Host:                  host,
		Addr:                  selfAddr,
		PeerstoreProvider:     peerstoreProviderMock,
		CurrentHeightProvider: currentHeightProviderMock,
	}

	network, err := NewRainTreeNetwork(busMock, netCfg)
	require.NoError(t, err)

	rainTreeNet := network.(*rainTreeNetwork)

	privKey, err := cryptoPocket.GeneratePrivateKey()
	require.NoError(t, err)
	peerToAdd := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}

	// Add peerToAdd.
	err = rainTreeNet.AddPeer(peerToAdd)
	require.NoError(t, err)
	expectedPStoreSize++

	peerAddrs, peers := getPeersViewParts(rainTreeNet.peersManager)

	// Ensure size / lengths are consistent.
	require.Equal(t, expectedPStoreSize, network.GetPeerstore().Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	libp2pPStore := host.Peerstore()
	require.Len(t, libp2pPStore.Peers(), expectedPStoreSize)

	require.ElementsMatch(t, []string{selfAddr.String(), peerToAdd.GetAddress().String()}, peerAddrs, "addresses do not match")
	require.ElementsMatch(t, []*typesP2P.NetworkPeer{selfPeer, peerToAdd}, peers, "peers do not match")

	require.Equal(t, selfPeer, network.GetPeerstore().GetPeer(selfAddr), "Peerstore does not contain self")
	require.Equal(t, peerToAdd, network.GetPeerstore().GetPeer(peerToAdd.GetAddress()), "Peerstore does not contain added peer")
}

func TestRainTreeNetwork_RemovePeer(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Start with a peerstore which contains self and some number of peers: the
	// initial value of `expectedPStoreSize`.
	expectedPStoreSize := 3
	pstore := getPeerstore(t, expectedPStoreSize)

	selfPeer, host := newTestPeer(t)
	selfAddr := selfPeer.GetAddress()

	// Add self to peerstore as a control to ensure existing peers persist after
	// removing the target peer.
	err := pstore.AddPeer(selfPeer)
	require.NoError(t, err)
	expectedPStoreSize++

	busMock := mockBus(ctrl)
	peerstoreProviderMock := mockPeerstoreProvider(ctrl, pstore)
	currentHeightProviderMock := mockCurrentHeightProvider(ctrl, 0)
	netCfg := RainTreeConfig{
		Host:                  host,
		Addr:                  selfAddr,
		PeerstoreProvider:     peerstoreProviderMock,
		CurrentHeightProvider: currentHeightProviderMock,
	}

	network, err := NewRainTreeNetwork(busMock, netCfg)
	require.NoError(t, err)
	rainTree := network.(*rainTreeNetwork)

	// Ensure expected starting size / lengths are consistent.
	peerAddrs, peers := getPeersViewParts(rainTree.peersManager)
	require.Equal(t, expectedPStoreSize, pstore.Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	libp2pPStore := host.Peerstore()
	require.Len(t, libp2pPStore.Peers(), expectedPStoreSize)

	var peerToRemove typesP2P.Peer
	// Ensure we don't remove selfPeer. `Peerstore` interface isn't aware
	// of the concept of "self" so we have to find it.
	for _, peer := range pstore.GetPeerList() {
		if peer.GetAddress().Equals(selfAddr) {
			continue
		}
		peerToRemove = peer
		break
	}
	require.NotNil(t, peerToRemove, "did not find selfAddr in peerstore")

	// Remove peerToRemove
	err = rainTree.RemovePeer(peerToRemove)
	require.NoError(t, err)
	expectedPStoreSize--

	peerAddrs, peers = getPeersViewParts(rainTree.peersManager)
	removedAddr := peerToRemove.GetAddress()
	getPeer := func(addr cryptoPocket.Address) typesP2P.Peer {
		return rainTree.GetPeerstore().GetPeer(addr)
	}

	// Ensure updated sizes are consistent.
	require.Equal(t, expectedPStoreSize, pstore.Size())
	require.Equal(t, expectedPStoreSize, len(peerAddrs))
	require.Equal(t, expectedPStoreSize, len(peers))

	require.Equal(t, selfPeer, getPeer(selfAddr), "Peerstore does not contain self")
	require.Nil(t, getPeer(removedAddr), "Peerstore contains removed peer")
}

func getPeersViewParts(pm typesP2P.PeerManager) (
	addrs []string,
	peers typesP2P.PeerList,
) {
	view := pm.GetPeersView()
	addrs = view.GetAddrs()
	peers = view.GetPeers()

	return addrs, peers
}

func newTestPeer(t *testing.T) (*typesP2P.NetworkPeer, libp2pHost.Host) {
	selfPrivKey, err := cryptoPocket.GeneratePrivateKey()
	require.NoError(t, err)

	selfAddr := selfPrivKey.Address()
	selfPeer := &typesP2P.NetworkPeer{
		PublicKey:  selfPrivKey.PublicKey(),
		Address:    selfAddr,
		ServiceURL: testLocalServiceURL,
	}
	return selfPeer, newLibp2pMockNetHost(t, selfPrivKey, selfPeer)
}

// TECHDEBT(#609): move & de-duplicate
func newLibp2pMockNetHost(t *testing.T, privKey cryptoPocket.PrivateKey, peer *typesP2P.NetworkPeer) libp2pHost.Host {
	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	libp2pMockNet := mocknet.New()
	host, err := libp2pMockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}
