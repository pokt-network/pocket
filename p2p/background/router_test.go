package background

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/internal/testutil"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_types "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
const testIP6ServiceURL = "[2a00:1450:4005:802::2004]:8080"

// TECHDEBT(#609): move & de-dup.
var testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)

func TestBackgroundRouter_AddPeer(t *testing.T) {
	testRouter := newTestRouter(t, nil)
	libp2pPStore := testRouter.host.Peerstore()

	// NB: assert initial state
	require.Equal(t, 1, testRouter.pstore.Size())

	existingPeer := testRouter.pstore.GetPeerList()[0]
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
	err = testRouter.AddPeer(newPeer)
	require.NoError(t, err)

	require.Len(t, testRouter.pstore, 2)
	require.Equal(t, testRouter.pstore.GetPeer(existingPeer.GetAddress()), existingPeer)
	require.Equal(t, testRouter.pstore.GetPeer(newPeer.Address), newPeer)

	existingPeerstoreAddrs = libp2pPStore.Addrs(existingPeerInfo.ID)
	newPeerstoreAddrs := libp2pPStore.Addrs(newPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)
	require.Len(t, newPeerstoreAddrs, 1)
	require.Equal(t, newPeerstoreAddrs[0].String(), newPeerMultiaddr.String())
}

func TestBackgroundRouter_RemovePeer(t *testing.T) {
	testRouter := newTestRouter(t, nil)
	peerstore := testRouter.host.Peerstore()

	// NB: assert initial state
	require.Len(t, testRouter.pstore, 1)

	existingPeer := testRouter.pstore.GetPeerList()[0]
	require.NotNil(t, existingPeer)

	existingPeerInfo, err := utils.Libp2pAddrInfoFromPeer(existingPeer)
	require.NoError(t, err)

	existingPeerstoreAddrs := peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)

	existingPeerMultiaddr, err := utils.Libp2pMultiaddrFromServiceURL(existingPeer.GetServiceURL())
	require.NoError(t, err)
	require.Equal(t, existingPeerstoreAddrs[0].String(), existingPeerMultiaddr.String())

	err = testRouter.RemovePeer(existingPeer)
	require.NoError(t, err)

	require.Len(t, testRouter.pstore, 0)

	// NB: libp2p peerstore implementations only remove peer keys and metadata
	// but continue to resolve multiaddrs until their respective TTLs expire.
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoremem/peerstore.go#L108)
	// (see: https://github.com/libp2p/go-libp2p/blob/v0.25.1/p2p/host/peerstore/pstoreds/peerstore.go#L187)

	existingPeerstoreAddrs = peerstore.Addrs(existingPeerInfo.ID)
	require.Len(t, existingPeerstoreAddrs, 1)
}

func TestBackgroundRouter_Broadcast(t *testing.T) {
	const (
		numPeers            = 4
		testMsg             = "test messsage"
		testTimeoutDuration = time.Second * 5
	)

	var (
		ctx = context.Background()
		mu  sync.Mutex
		// map used as a set to collect IDs of peers which have received a message
		seenMessages = make(map[string]struct{})
		wg           = sync.WaitGroup{}
		done         = make(chan struct{}, 1)
		testTimeout  = time.After(testTimeoutDuration)
		// NB: peerIDs are stringified
		actualPeerIDs   []string
		expectedPeerIDs = make([]string, numPeers)
		testHosts       = make([]libp2pHost.Host, 0)
		libp2pMockNet   = mocknet.New()
	)

	// setup 4 libp2p hosts to listen for incoming streams from the test backgroundRouter
	for i := 0; i < numPeers; i++ {
		wg.Add(1)

		privKey, selfPeer := newTestPeer(t)
		host := newTestHost(t, libp2pMockNet, privKey)
		testHosts = append(testHosts, host)
		expectedPeerIDs[i] = host.ID().String()
		rtr := newRouterWithSelfPeerAndHost(t, selfPeer, host)
		go readSubscription(t, ctx, &wg, rtr, &mu, seenMessages)
	}

	// bootstrap off of arbitrary testHost
	privKey, selfPeer := newTestPeer(t)

	// set up a test backgroundRouter
	testRouterHost := newTestHost(t, libp2pMockNet, privKey)
	testRouter := newRouterWithSelfPeerAndHost(t, selfPeer, testRouterHost)
	testHosts = append(testHosts, testRouterHost)

	// simulate network links between each to every other
	// (i.e. fully-connected network)
	err := libp2pMockNet.LinkAll()
	require.NoError(t, err)

	bootstrap(t, ctx, testHosts)

	go func() {
		// wait for hosts to listen and peer discovery
		time.Sleep(time.Second * 2)

		// broadcast message
		t.Log("broadcasting...")
		err := testRouter.Broadcast([]byte(testMsg))
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

	actualPeerIDs = testutil.GetKeys[string](seenMessages)
	require.ElementsMatchf(t, expectedPeerIDs, actualPeerIDs, "peerIDs don't match")
}

// bootstrap connects each host to one other except for the arbitrarily chosen "bootstrap host"
func bootstrap(t *testing.T, ctx context.Context, testHosts []libp2pHost.Host) {
	t.Helper()

	t.Log("bootstrapping...")
	bootstrapHost := testHosts[0]
	for _, h := range testHosts {
		if h.ID() == bootstrapHost.ID() {
			continue
		}

		p2pAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", bootstrapHost.ID()))
		require.NoError(t, err)

		addrInfo := libp2pPeer.AddrInfo{
			ID: bootstrapHost.ID(),
			Addrs: []multiaddr.Multiaddr{
				bootstrapHost.Addrs()[0].Encapsulate(p2pAddr),
			},
		}

		t.Logf("connecting to %s...", addrInfo.ID.String())
		err = h.Connect(ctx, addrInfo)
		require.NoError(t, err)
	}
}

// TECHDEBT(#609): move & de-duplicate
func newTestRouter(t *testing.T, libp2pMockNet mocknet.Mocknet) *backgroundRouter {
	t.Helper()

	privKey, selfPeer := newTestPeer(t)

	if libp2pMockNet == nil {
		libp2pMockNet = mocknet.New()
	}

	host := newMockNetHostFromPeer(t, libp2pMockNet, privKey, selfPeer)
	t.Cleanup(func() {
		err := host.Close()
		require.NoError(t, err)
	})

	return newRouterWithSelfPeerAndHost(t, selfPeer, host)
}

func newRouterWithSelfPeerAndHost(t *testing.T, selfPeer typesP2P.Peer, host libp2pHost.Host) *backgroundRouter {
	t.Helper()

	ctrl := gomock.NewController(t)
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{
		P2P: &configs.P2PConfig{
			IsClientOnly: false,
		},
	})

	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	pstore := make(typesP2P.PeerAddrMap)
	pstoreProviderMock := mock_types.NewMockPeerstoreProvider(ctrl)
	pstoreProviderMock.EXPECT().GetStakedPeerstoreAtHeight(gomock.Any()).Return(pstore, nil).AnyTimes()

	busMock := mockModules.NewMockBus(ctrl)
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()

	err := pstore.AddPeer(selfPeer)
	require.NoError(t, err)

	router, err := NewBackgroundRouter(busMock, &utils.RouterConfig{
		Addr:                  selfPeer.GetAddress(),
		PeerstoreProvider:     pstoreProviderMock,
		CurrentHeightProvider: consensusMock,
		Host:                  host,
	})
	require.NoError(t, err)

	libp2pNet, ok := router.(*backgroundRouter)
	require.Truef(t, ok, "unexpected router type: %T", router)

	return libp2pNet
}

// TECHDEBT: move & de-dup
func newTestPeer(t *testing.T) (cryptoPocket.PrivateKey, *typesP2P.NetworkPeer) {
	t.Helper()

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
	t.Helper()

	libp2pPrivKey, err := libp2pCrypto.UnmarshalEd25519PrivateKey(privKey.Bytes())
	require.NoError(t, err)

	libp2pMultiAddr, err := utils.Libp2pMultiaddrFromServiceURL(peer.ServiceURL)
	require.NoError(t, err)

	host, err := mockNet.AddPeer(libp2pPrivKey, libp2pMultiAddr)
	require.NoError(t, err)

	return host
}

func newTestHost(t *testing.T, mockNet mocknet.Mocknet, privKey cryptoPocket.PrivateKey) libp2pHost.Host {
	t.Helper()

	// listen on random port on loopback interface
	peer := &typesP2P.NetworkPeer{
		PublicKey:  privKey.PublicKey(),
		Address:    privKey.Address(),
		ServiceURL: testLocalServiceURL,
	}

	// construct mock host
	return newMockNetHostFromPeer(t, mockNet, privKey, peer)
}

func readSubscription(
	t *testing.T,
	ctx context.Context,
	wg *sync.WaitGroup,
	rtr *backgroundRouter,
	mu *sync.Mutex,
	seenMsgs map[string]struct{},
) {
	t.Helper()

	for {
		if err := ctx.Err(); err != nil {
			if err != context.Canceled || err != context.DeadlineExceeded {
				require.NoError(t, err)
			}
			return
		}

		_, err := rtr.subscription.Next(ctx)
		require.NoError(t, err)

		mu.Lock()
		wg.Done()
		seenMsgs[rtr.host.ID().String()] = struct{}{}
		mu.Unlock()
	}
}
