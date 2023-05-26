package background

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
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/stretchr/testify/require"
)

const dhtUpdateSleepDuration = time.Millisecond * 500

func TestLibp2pKademliaPeerDiscovery(t *testing.T) {
	ctx := context.Background()

	addr1, host1, _ := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort, nil)

	bootstrapAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", addr1, host1.ID().String()))
	require.NoError(t, err)

	addr2, host2, _ := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort+1, bootstrapAddr)
	addr3, host3, _ := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort+2, bootstrapAddr)

	expectedPeerIDs := []libp2pPeer.ID{host1.ID(), host2.ID(), host3.ID()}

	// TECHDEBT: consider using `host.ConnManager().Notifee()` to avoid sleeping here
	// delay assertions for 500ms
	// NB: wait for peer discovery to complete
	time.Sleep(dhtUpdateSleepDuration)

	// assert that host2 has host3 in its peerstore
	host2DiscoveredAddrs := host2.Peerstore().Addrs(host3.ID())
	require.Greaterf(t, len(host2DiscoveredAddrs), 0, "did not discover host3")
	require.Equalf(t, addr3.String(), host2DiscoveredAddrs[0].String(), "did not discover host3")
	require.ElementsMatchf(t, expectedPeerIDs, host2.Peerstore().Peers(), "host2 peer IDs don't match")

	// assert that host3 has host2 in its peerstore
	host3DiscoveredHost2Addrs := host3.Peerstore().Addrs(host2.ID())
	require.Greaterf(t, len(host3DiscoveredHost2Addrs), 0, "host3 did not discover host2")
	require.Equalf(t, addr2.String(), host3DiscoveredHost2Addrs[0].String(), "host3 did not discover host2")
	require.ElementsMatchf(t, expectedPeerIDs, host3.Peerstore().Peers(), "host3 peer IDs don't match")

	// add another peer to network...
	addr4, host4, _ := setupHostAndDiscovery(t, ctx, defaults.DefaultP2PPort+3, bootstrapAddr)
	expectedPeerIDs = append(expectedPeerIDs, host4.ID())

	// TECHDEBT: consider using `host.ConnManager().Notifee()` to avoid sleeping here
	time.Sleep(dhtUpdateSleepDuration)

	// new host discovers existing hosts...
	host4DiscoveredHost2Addrs := host4.Peerstore().Addrs(host2.ID())
	require.Greaterf(t, len(host4DiscoveredHost2Addrs), 0, "host4 did not discover host2")
	require.Equalf(t, addr2.String(), host4DiscoveredHost2Addrs[0].String(), "host4 did not discover host2")

	host4DiscoveredHost3Addrs := host4.Peerstore().Addrs(host3.ID())
	require.Greaterf(t, len(host4DiscoveredHost3Addrs), 0, "host4 did not discover host3")
	require.Equalf(t, addr3.String(), host4DiscoveredHost3Addrs[0].String(), "host4 did not discover host3")

	// existing hosts discovers host host...
	host2DiscoveredHost4Addrs := host2.Peerstore().Addrs(host4.ID())
	require.Greaterf(t, len(host2DiscoveredHost4Addrs), 0, "host2 did not discover host4")
	require.Equalf(t, addr4.String(), host2DiscoveredHost4Addrs[0].String(), "host2 did not discover host4")

	host3DiscoveredHost4Addrs := host3.Peerstore().Addrs(host4.ID())
	require.Greaterf(t, len(host3DiscoveredHost4Addrs), 0, "host3 did not discover host4")
	require.Equalf(t, addr4.String(), host3DiscoveredHost4Addrs[0].String(), "host3 did not discover host4")

	require.ElementsMatchf(t, expectedPeerIDs, host4.Peerstore().Peers(), "host4 peer IDs don't match")
}

//nolint:unparam // DHT must exist but is otherwise "unused" (i.e. its API)
func setupHostAndDiscovery(t *testing.T,
	ctx context.Context,
	port uint32,
	bootstrapAddr multiaddr.Multiaddr,
) (
	multiaddr.Multiaddr,
	libp2pHost.Host,
	*dht.IpfsDHT,
) {
	t.Helper()

	// CONSIDERATION: perhaps testing with libp2p mocknet would be sufficient
	// listen on loopback interface
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	require.NoError(t, err)

	host, err := setupHost(t, addr)
	require.NoError(t, err)

	kadDHT := setupDHT(t, ctx, host, bootstrapAddr)
	return addr, host, kadDHT
}

func setupHost(t *testing.T, addr multiaddr.Multiaddr) (libp2pHost.Host, error) {
	t.Helper()

	privKey, _, err := libp2pCrypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)

	return libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.Identity(privKey),
	)
}

func setupDHT(t *testing.T, ctx context.Context, host libp2pHost.Host, bootstrapAddr multiaddr.Multiaddr) *dht.IpfsDHT {
	t.Helper()

	kadDHT, err := dht.New(ctx, host, dht.Mode(dht.ModeAutoServer))
	require.NoError(t, err)

	if bootstrapAddr != nil {
		peerInfo, err := libp2pPeer.AddrInfoFromP2pAddr(bootstrapAddr)
		require.NoError(t, err)

		err = host.Connect(ctx, *peerInfo)
		require.NoError(t, err)
	}
	return kadDHT
}
