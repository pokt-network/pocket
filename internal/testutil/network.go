package testutil

import (
	"fmt"

	crypto2 "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/shared/crypto"
)

const ServiceURLFormat = "node%d.consensus:42069"

func NewMocknetHost(
	t gocuke.TestingT,
	libp2pNetworkMock mocknet.Mocknet,
	privKey crypto.PrivateKey,
	notifiee libp2pNetwork.Notifiee,
) host.Host {
	t.Helper()

	// TODO_THIS_COMMIT: move to const
	addrMock, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/10.0.0.1/tcp/%d", defaults.DefaultP2PPort))
	require.NoError(t, err)

	libp2pPrivKey, err := crypto2.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	host, err := libp2pNetworkMock.AddPeer(libp2pPrivKey, addrMock)
	require.NoError(t, err)

	if notifiee != nil {
		host.Network().Notify(notifiee)
	}

	return host
}

func SequentialServiceURLPrivKeyMap(t gocuke.TestingT, count int) map[string]crypto.PrivateKey {
	t.Helper()

	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// ID collisions
	privKeys := LoadLocalnetPrivateKeys(t, count)
	// CONSIDERATION: using an iterator/generator would prevent unintentional
	// serviceURL collisions
	serviceURLs := SequentialServiceURLs(t, count)

	require.GreaterOrEqualf(t, len(privKeys), len(serviceURLs), "not enough private keys for service URLs")

	serviceURLKeysMap := make(map[string]crypto.PrivateKey, len(serviceURLs))

	for i, serviceURL := range serviceURLs {
		serviceURLKeysMap[serviceURL] = privKeys[i]
	}
	return serviceURLKeysMap
}

// CONSIDERATION: serviceURLs are only unique within their respective slice;
// consider building an iterator/generator instead.
func SequentialServiceURLs(t gocuke.TestingT, count int) (serviceURLs []string) {
	t.Helper()

	for i := 0; i < count; i++ {
		serviceURLs = append(serviceURLs, NewServiceURL(i+1))
	}
	return serviceURLs
}

// TECHDEBT: rename `validatorId()` to `serviceURL()`
func NewServiceURL(i int) string {
	return fmt.Sprintf(ServiceURLFormat, i)
}

// TODO_THIS_COMMIT: move
func NewDebugNotifee(t gocuke.TestingT) libp2pNetwork.Notifiee {
	t.Helper()

	return &libp2pNetwork.NotifyBundle{
		ConnectedF: func(_ libp2pNetwork.Network, conn libp2pNetwork.Conn) {
			t.Logf("connected: local: %s; remote: %s",
				conn.LocalPeer().String(),
				conn.RemotePeer().String(),
			)
		},
		DisconnectedF: func(_ libp2pNetwork.Network, conn libp2pNetwork.Conn) {
			t.Logf("disconnected: local: %s; remote: %s",
				conn.LocalPeer().String(),
				conn.RemotePeer().String(),
			)
		},
		ListenF: func(_ libp2pNetwork.Network, addr multiaddr.Multiaddr) {
			t.Logf("listening: %s", addr.String())
		},
		ListenCloseF: func(_ libp2pNetwork.Network, addr multiaddr.Multiaddr) {
			t.Logf("closed: %s", addr.String())
		},
	}
}
