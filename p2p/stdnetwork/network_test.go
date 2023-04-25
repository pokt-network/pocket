package stdnetwork

import (
	"fmt"
	"github.com/pokt-network/pocket/runtime/defaults"
	"testing"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_types "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
const testIP6ServiceURL = "[2a00:1450:4005:802::2004]:8080"

// TECHDEBT(#609): move & de-dup.
var testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)

func TestLibp2pNetwork_AddPeer(t *testing.T) {
	p2pNet := newTestNetwork(t)
	libp2pPStore := p2pNet.host.Peerstore()

	// NB: assert initial state
	require.Equal(t, 1, p2pNet.pstore.Size())

	existingPeer := p2pNet.pstore.GetPeerList()[0]
	require.NotNil(t, existingPeer)

	existingPeerInfo, err := utils.Libp2pAddrInfoFromPeer(existingPeer)
	require.NoError(t, err)

	existingPeerstoreAddrs := libp2pPStore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)

	existingPeerMultiaddr, err := utils.Libp2pMultiaddrFromServiceURL(existingPeer.GetServiceURL())
	require.NoError(t, err)
	require.Equal(t, existingPeerstoreAddrs[0].String(), existingPeerMultiaddr.String())

	newPublicKey, err := cryptoPocket.GeneratePublicKey()
	newPoktAddr := newPublicKey.Address()
	require.NoError(t, err)

	newPeer := &typesP2P.NetworkPeer{
		PublicKey:  newPublicKey,
		Address:    newPoktAddr,
		ServiceURL: testIP6ServiceURL,
	}
	newPeerInfo, err := utils.Libp2pAddrInfoFromPeer(newPeer)
	require.NoError(t, err)
	newPeerMultiaddr := newPeerInfo.Addrs[0]

	// NB: add to address book
	err = p2pNet.AddPeer(newPeer)
	require.NoError(t, err)

	require.Len(t, p2pNet.pstore, 2)
	require.Equal(t, p2pNet.pstore.GetPeer(existingPeer.GetAddress()), existingPeer)
	require.Equal(t, p2pNet.pstore.GetPeer(newPeer.Address), newPeer)

	existingPeerstoreAddrs = libp2pPStore.Addrs(existingPeerInfo.ID)
	newPeerstoreAddrs := libp2pPStore.Addrs(newPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)
	require.Len(t, newPeerstoreAddrs, 1)
	require.Equal(t, newPeerstoreAddrs[0].String(), newPeerMultiaddr.String())
}

func TestLibp2pNetwork_RemovePeer(t *testing.T) {
	p2pNet := newTestNetwork(t)
	peerstore := p2pNet.host.Peerstore()

	// NB: assert initial state
	require.Len(t, p2pNet.pstore, 1)

	existingPeer := p2pNet.pstore.GetPeerList()[0]
	require.NotNil(t, existingPeer)

	existingPeerInfo, err := utils.Libp2pAddrInfoFromPeer(existingPeer)
	require.NoError(t, err)

	existingPeerstoreAddrs := peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)

	existingPeerMultiaddr, err := utils.Libp2pMultiaddrFromServiceURL(existingPeer.GetServiceURL())
	require.NoError(t, err)
	require.Equal(t, existingPeerstoreAddrs[0].String(), existingPeerMultiaddr.String())

	err = p2pNet.RemovePeer(existingPeer)
	require.NoError(t, err)

	require.Len(t, p2pNet.pstore, 0)

	// NB: libp2p peerstore implementations only remove peer keys and metadata
	// but continue to resolve multiaddrs until their respective TTLs expire.
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoremem/peerstore.go#L108)
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoreds/peerstore.go#L187)

	existingPeerstoreAddrs = peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)
}

// TECHDEBT(#609): move & de-duplicate
func newTestNetwork(t *testing.T) *router {
	ctrl := gomock.NewController(t)
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	pstore := make(typesP2P.PeerAddrMap)
	pstoreProviderMock := mock_types.NewMockPeerstoreProvider(ctrl)
	pstoreProviderMock.EXPECT().GetStakedPeerstoreAtHeight(gomock.Any()).Return(pstore, nil).AnyTimes()

	privKey, err := cryptoPocket.GeneratePrivateKey()
	require.NoError(t, err)

	selfPeer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}
	err = pstore.AddPeer(selfPeer)
	require.NoError(t, err)

	host := newLibp2pMockNetHost(t, privKey, selfPeer)
	defer host.Close()

	p2pNetwork, err := NewNetwork(
		host,
		pstoreProviderMock,
		consensusMock,
	)
	require.NoError(t, err)

	libp2pNet, ok := p2pNetwork.(*router)
	require.Truef(t, ok, "unexpected p2pNetwork type: %T", p2pNetwork)

	return libp2pNet
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
