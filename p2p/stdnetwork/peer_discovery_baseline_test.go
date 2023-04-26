package stdnetwork

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	libp2pDiscovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	libp2pDiscoveryUtil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/runtime/defaults"
)

func TestLibp2pKademliaPeerDiscovery(t *testing.T) {
	ctx := context.Background()

	addr1, host1, discovery1 := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort, nil)

	bootstrapAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr1, host1.ID().String()))
	require.NoError(t, err)

	_, host2, discovery2 := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort+1, bootstrapAddr)
	addr3, host3, discovery3 := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort+2, bootstrapAddr)

	expectedPeerIDs := []libp2pPeer.ID{host1.ID(), host2.ID(), host3.ID()}

	go discoverAndAdvertise(t, ctx, host1, discovery1)
	go discoverAndAdvertise(t, ctx, host2, discovery2)
	go discoverAndAdvertise(t, ctx, host3, discovery3)

	// delay assertions for 1s
	time.Sleep(time.Second * 1)

	// assert that host2 has host3 in its peerstore
	discoveredAddrs := host2.Peerstore().Addrs(host3.ID())
	require.Lenf(t, discoveredAddrs, 1, "did not discover host3")
	require.Equalf(t, addr3.String(), discoveredAddrs[0].String(), "did not discover host3")
	require.ElementsMatchf(t, expectedPeerIDs, host2.Peerstore().Peers(), "peer IDs don't match")
}

func setupHostAndDiscovery(t *testing.T, ctx context.Context, port uint32, bootstrapAddr multiaddr.Multiaddr) (
	multiaddr.Multiaddr,
	libp2pHost.Host,
	*libp2pDiscovery.RoutingDiscovery,
) {
	t.Helper()

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	require.NoError(t, err)

	host, err := setupHost(t, addr)
	require.NoError(t, err)

	discovery, err := setupDiscovery(t, ctx, host, bootstrapAddr)
	require.NoError(t, err)

	return addr, host, discovery
}

func setupHost(t *testing.T, addr multiaddr.Multiaddr) (libp2pHost.Host, error) {
	t.Helper()

	priv, _, err := libp2pCrypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)

	return libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.Identity(priv),
	)
}

func setupDiscovery(t *testing.T, ctx context.Context, host libp2pHost.Host, bootstrapAddr multiaddr.Multiaddr) (*libp2pDiscovery.RoutingDiscovery, error) {
	t.Helper()

	kdht, err := dht.New(ctx, host, dht.Mode(dht.ModeAutoServer))
	require.NoError(t, err)

	//err = kdht.Bootstrap(ctx)
	//require.NoError(t, err)

	if bootstrapAddr != nil {
		peerInfo, err := libp2pPeer.AddrInfoFromP2pAddr(bootstrapAddr)
		require.NoError(t, err)

		go func() {
			err := host.Connect(ctx, *peerInfo)
			require.NoError(t, err)
		}()
	}

	return libp2pDiscovery.NewRoutingDiscovery(kdht), nil
}

func discoverAndAdvertise(t *testing.T, ctx context.Context, host libp2pHost.Host, dht *libp2pDiscovery.RoutingDiscovery) {
	t.Helper()

	discovery := libp2pDiscovery.NewRoutingDiscovery(dht)
	libp2pDiscoveryUtil.Advertise(ctx, discovery, protocol.PeerDiscoveryNamespace)
	discover(t, ctx, host, discovery)
}

func discover(t *testing.T, ctx context.Context, host libp2pHost.Host, discovery *libp2pDiscovery.RoutingDiscovery) {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := discovery.FindPeers(ctx, protocol.PeerDiscoveryNamespace)
			require.NoError(t, err)

			for peer := range peers {
				if peer.ID == host.ID() {
					continue
				}
				if host.Network().Connectedness(peer.ID) != libp2pNetwork.Connected {
					if _, err = host.Network().DialPeer(ctx, peer.ID); err != nil {
						continue
					}
				}
			}
		}
	}
}
