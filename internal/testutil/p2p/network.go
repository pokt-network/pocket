package p2p_testutil

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
)

// TODO: remove if not needed
func NewMocknetWithNPeers(t gocuke.TestingT, peerCount int) (mocknet.Mocknet, []string) {
	t.Helper()

	// load pre-generated validator keypairs
	libp2pNetworkMock := mocknet.New()
	privKeys := testutil.LoadLocalnetPrivateKeys(t, peerCount)
	serviceURLs := SequentialServiceURLs(t, peerCount)
	_ = SetupMockNetPeers(t, libp2pNetworkMock, privKeys, serviceURLs)

	return libp2pNetworkMock, serviceURLs
}

func NewMocknetHost(
	t gocuke.TestingT,
	netMock mocknet.Mocknet,
	privKey cryptoPocket.PrivateKey,
) libp2pHost.Host {
	t.Helper()

	// TODO_THIS_COMMIT: move to const
	addrMock, err := multiaddr.NewMultiaddr("/ip4/10.0.0.1/tcp/0")
	require.NoError(t, err)

	libp2pPrivKey, err := crypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	host, err := netMock.AddPeer(libp2pPrivKey, addrMock)
	require.NoError(t, err)

	return host
}

func SetupMockNetPeers(
	t gocuke.TestingT,
	netMock mocknet.Mocknet,
	privKeys []cryptoPocket.PrivateKey,
	serviceURLs []string,
) (peerIDs []peer.ID) {
	t.Helper()

	// TODO_THIS_COMMIT: return these
	//
	// MUST add mockdns before any libp2p host comes online. Otherwise, it will
	// error while attempting to resolve its own hostname.
	_, dnsSrvDone := testutil.DNSMockFromServiceURLs(t, serviceURLs)
	t.Cleanup(dnsSrvDone)

	// Add a libp2p peers/hosts to the `MockNet` with the keypairs corresponding
	// to the genesis validators' keypairs
	for i, peerInfo := range PeersFromPrivKeysAndServiceURLs(t, privKeys, serviceURLs) {
		libp2pPrivKey, err := crypto.UnmarshalEd25519PrivateKey(privKeys[i].Bytes())
		require.NoError(t, err)

		// TODO_THIS_COMMIT: add mock DNS zone per peer instead of all at once
		_, err = netMock.AddPeer(libp2pPrivKey, peerInfo.Addrs[0])
		require.NoError(t, err)

		peerIDs = append(peerIDs, peerInfo.ID)
	}

	// Link all peers such that any may dial/connect to any other.
	err := netMock.LinkAll()
	require.NoError(t, err)

	return peerIDs
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

func PeersFromPrivKeysAndServiceURLs(
	t gocuke.TestingT,
	privKeys []cryptoPocket.PrivateKey,
	serviceURLs []string,
) (peersInfo []libp2pPeer.AddrInfo) {
	t.Helper()

	serviceURLCount, privKeyCount := len(serviceURLs), len(privKeys)
	maxCount := serviceURLCount
	if privKeyCount < serviceURLCount {
		maxCount = privKeyCount
	}

	for i, privKey := range privKeys[:maxCount] {
		peerInfo := peerFromPrivKeyAndServiceURL(t, privKey, NewServiceURL(i+1))
		peersInfo = append(peersInfo, peerInfo)
	}
	return peersInfo
}

func peerFromPrivKeyAndServiceURL(
	t gocuke.TestingT,
	privKey cryptoPocket.PrivateKey,
	serviceURL string,
) libp2pPeer.AddrInfo {
	t.Helper()

	peerInfo, err := utils.Libp2pAddrInfoFromPeer(&types.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: serviceURL,
	})
	require.NoError(t, err)

	return peerInfo
}

const ServiceURLFormat = "node%d.consensus:42069"

// TECHDEBT: rename `validatorId()` to `serviceURL()`
func NewServiceURL(i int) string {
	return fmt.Sprintf(ServiceURLFormat, i)
}
