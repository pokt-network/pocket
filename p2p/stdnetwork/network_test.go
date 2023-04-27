package stdnetwork

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/libp2p/go-libp2p"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_types "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
const testIP6ServiceURL = "[2a00:1450:4005:802::2004]:8080"

// TECHDEBT(#609): move & de-dup.
var testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)

func TestLibp2pNetwork_AddPeer(t *testing.T) {
	p2pNet := newTestRouter(t)
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
	p2pNet := newTestRouter(t)
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

func TestNetwork_NetworkBroadcast(t *testing.T) {
	const (
		numPeers            = 4
		testMsg             = "test messsage"
		testTimeoutDuration = time.Second * 5
	)

	var (
		ctx = context.Background()
		// map used as a set to collect IDs of peers which have received a message
		seenMessages = make(map[string]struct{})
		wg           = sync.WaitGroup{}
		done         = make(chan struct{}, 1)
		testTimeout  = time.After(testTimeoutDuration)
		// NB: peerIDs are stringified
		expectedPeerIDs = make([]string, numPeers)
		actualPeerIDs   = make([]string, 0)
		//peerAddrs       = make(map[string][]multiaddr.Multiaddr)
		testHosts = make([]libp2pHost.Host, 0)
	)

	//libp2pMockNet := mocknet.New()

	// setup 4 libp2p hosts to listen for incoming streams from the test router
	for i := 0; i < numPeers; i++ {
		privKey, selfPeer := newTestPeer(t)
		host := newTestHost(t, privKey)
		testHosts = append(testHosts, host)

		expectedPeerIDs[i] = host.ID().String()
		//peerIDStr := testRouterHost.ID().String()
		//expectedPeerIDs[i] = peerIDStr
		//peerAddrs[peerIDStr] = testRouterHost.Addrs()

		t.Log("registering stream handler")
		wg.Add(1)
		rtr := newRouterWithSelfPeerAndHost(t, selfPeer, host)
		go readSubscription(t, ctx, &wg, rtr, seenMessages)
	}

	// TODO_THIS_COMMIT: remove me
	// bootstrap off of some testRouterHost
	var (
		addrInfo          libp2pPeer.AddrInfo
		bootstrapHost     = testHosts[0]
		privKey, selfPeer = newTestPeer(t)
	)

	// set up a test router
	testRouterHost := newTestHost(t, privKey)
	testRouter := newRouterWithSelfPeerAndHost(t, selfPeer, testRouterHost)
	// TODO_THIS_COMMIT: refactor
	testHosts = append(testHosts, testRouterHost)

	// connect each node to one other... (quick & dirty bootstrap)
	// TODO_THIS_COMMIT: refactor
	for _, h := range testHosts {
		if h.ID() == bootstrapHost.ID() {
			addrInfo.ID = testRouterHost.ID()
			addrInfo.Addrs = testRouterHost.Addrs()
		} else {
			p2pAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", bootstrapHost.ID()))
			require.NoError(t, err)

			addrInfo.ID = bootstrapHost.ID()
			addrInfo.Addrs = []multiaddr.Multiaddr{
				bootstrapHost.Addrs()[0].Encapsulate(p2pAddr),
			}
		}
		t.Logf("connecting to %s...", addrInfo.ID.String())
		err := h.Connect(ctx, addrInfo)
		require.NoError(t, err)
		t.Log("connected")
	}
	// end remove me

	go func() {
		time.Sleep(time.Second * 2)
		t.Log("broadcasting...")
		// broadcast message
		err := testRouter.NetworkBroadcast([]byte(testMsg))
		require.NoError(t, err)
	}()

	// wait concurrently
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	// waitgroup done or timeout
	select {
	case <-testTimeout:
		t.Fatalf(
			"timed out waiting for message: got %d; wanted %d",
			len(seenMessages),
			numPeers,
		)
	case <-done:
	}

	actualPeerIDs = getKeys[string](seenMessages)
	require.ElementsMatchf(t, expectedPeerIDs, actualPeerIDs, "peerIDs don't match")
}

func getKeys[K comparable, V any](keyMap map[K]V) (keys []K) {
	for key := range keyMap {
		keys = append(keys, key)
	}
	return keys
}

// TECHDEBT(#609): move & de-duplicate
func newTestRouter(t *testing.T) *router {
	privKey, selfPeer := newTestPeer(t)

	host := newLibp2pMockNetHost(t, privKey, selfPeer)
	t.Cleanup(func() {
		err := host.Close()
		require.NoError(t, err)
	})

	return newRouterWithSelfPeerAndHost(t, selfPeer, host)
}

func newRouterWithSelfPeerAndHost(t *testing.T, selfPeer typesP2P.Peer, host libp2pHost.Host) *router {
	ctrl := gomock.NewController(t)
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	pstore := make(typesP2P.PeerAddrMap)
	pstoreProviderMock := mock_types.NewMockPeerstoreProvider(ctrl)
	pstoreProviderMock.EXPECT().GetStakedPeerstoreAtHeight(gomock.Any()).Return(pstore, nil).AnyTimes()

	err := pstore.AddPeer(selfPeer)
	require.NoError(t, err)

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
	libp2pMockNet := mocknet.New()
	return newMockNetHostFromPeer(t, libp2pMockNet, privKey, peer)
}

// TECHDEBT: move & de-dup
func newTestPeer(t *testing.T) (cryptoPocket.PrivateKey, *typesP2P.NetworkPeer) {
	privKey, err := cryptoPocket.GeneratePrivateKey()
	require.NoError(t, err)

	return privKey, &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}
}

func newMockNetHostFromPeer(
	t *testing.T,
	mockNet mocknet.Mocknet,
	privKey cryptoPocket.PrivateKey,
	peer *typesP2P.NetworkPeer,
) libp2pHost.Host {
	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	host, err := mockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}

func newTestHost(t *testing.T, privKey cryptoPocket.PrivateKey) libp2pHost.Host {
	//host := newMockNetHostFromPeer(t, libp2pMockNet, privKey, selfPeer)

	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	//peerID, err := libp2pPeer.IDFromPrivateKey(libp2pPrivKey)
	//require.NoError(t, err)

	// listen on random port on loopback interface
	//listenAddrStr := fmt.Sprintf("/ip4/127.0.0.1/tcp/0/p2p/%s", peerID.String())
	listenAddrStr := fmt.Sprintf("/ip4/127.0.0.1/tcp/0")
	listenAddr, err := multiaddr.NewMultiaddr(listenAddrStr)

	// construct host
	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(libp2pPrivKey),
		//libp2p.Routing(func(h libp2pHost.Host) (libp2pRouting.PeerRouting, error) {
		//	return dht.New(ctx, h)
		//}),
	)

	require.NoError(t, err)
	return host
}

func TestMultiaddrAssumptions(t *testing.T) {
	ma, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0")
	require.NoError(t, err)

	subMa, err := multiaddr.NewMultiaddr("/tcp/0")
	require.NoError(t, err)

	combinedMa := ma.Encapsulate(subMa)
	require.Equal(t, "/ip4/0.0.0.0/tcp/0", combinedMa.String())
}

func readSubscription(
	t *testing.T,
	ctx context.Context,
	wg *sync.WaitGroup,
	rtr *router,
	seenMsgs map[string]struct{},
) {
	for {
		if err := ctx.Err(); err != nil {
			if err != context.Canceled || err != context.DeadlineExceeded {
				require.NoError(t, err)
			}
			return
		}

		_, err := rtr.subscription.Next(ctx)
		require.NoError(t, err)

		wg.Done()
		seenMsgs[rtr.host.ID().String()] = struct{}{}
	}
}
