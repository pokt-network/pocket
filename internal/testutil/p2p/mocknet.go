package p2p_testutil

import (
	"testing"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/shared/crypto"
)

func NewTestPeer(t *testing.T) (*types.NetworkPeer, libp2pHost.Host) {
	t.Helper()

	selfPrivKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	selfAddr := selfPrivKey.Address()
	selfPeer := &types.NetworkPeer{
		PublicKey:  selfPrivKey.PublicKey(),
		Address:    selfAddr,
		ServiceURL: IP4ServiceURL,
	}
	return selfPeer, NewLibp2pMockNetHost(t, selfPrivKey, selfPeer)
}

func NewLibp2pMockNetHost(t *testing.T, privKey crypto.PrivateKey, peer *types.NetworkPeer) libp2pHost.Host {
	t.Helper()

	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	libp2pMockNet := mocknet.New()
	host, err := libp2pMockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}
