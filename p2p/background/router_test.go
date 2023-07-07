package background

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/p2p/config"
	"github.com/pokt-network/pocket/p2p/protocol"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_types "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/p2p/utils"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
)

// https://www.rfc-editor.org/rfc/rfc3986#section-3.2.2
const (
	testIP6ServiceURL     = "[2a00:1450:4005:802::2004]:8080"
	invalidReceiveTimeout = time.Millisecond * 500
)

// TECHDEBT(#609): move & de-dup.
var (
	testLocalServiceURL = fmt.Sprintf("127.0.0.1:%d", defaults.DefaultP2PPort)
	noopHandler         = func(data []byte) error { return nil }
)

func TestBackgroundRouter_AddPeer(t *testing.T) {
	testRouter := newTestRouter(t, nil, nil)
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
	testRouter := newTestRouter(t, nil, nil)
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

func TestBackgroundRouter_Validation(t *testing.T) {
	invalidProtoMessage := anypb.Any{
		TypeUrl: "/notADefinedProtobufType",
		Value:   []byte("not a serialized protobuf"),
	}

	testCases := []struct {
		name  string
		msgBz []byte
	}{
		{
			name: "invalid BackgroundMessage",
			// NB: `msgBz` would normally be a serialized `BackgroundMessage`.
			msgBz: mustMarshal(t, &invalidProtoMessage),
		},
		{
			name: "empty PocketEnvelope",
			msgBz: mustMarshal(t, &typesP2P.BackgroundMessage{
				// NB: `Data` is normally a serialized `PocketEnvelope`.
				Data: nil,
			}),
		},
		{
			name: "invalid PoketEnvelope",
			msgBz: mustMarshal(t, &typesP2P.BackgroundMessage{
				// NB: `Data` is normally a serialized `PocketEnvelope`.
				Data: mustMarshal(t, &invalidProtoMessage),
			}),
		},
	}

	// Set up test router as the receiver.
	ctx := context.Background()
	libp2pMockNet := mocknet.New()

	receivedChan := make(chan []byte, 1)
	receiverPrivKey, receiverPeer := newTestPeer(t)
	receiverHost := newTestHost(t, libp2pMockNet, receiverPrivKey)
	receiverRouter := newRouterWithSelfPeerAndHost(
		t, receiverPeer,
		receiverHost,
		func(data []byte) error {
			receivedChan <- data
			return nil
		},
	)

	t.Cleanup(func() {
		err := receiverRouter.Close()
		require.NoError(t, err)
	})

	// Wrap `receiverRouter#topicValidator` to make assertions by.
	// Existing topic validator must be unregistered first.
	err := receiverRouter.gossipSub.UnregisterTopicValidator(protocol.BackgroundTopicStr)
	require.NoError(t, err)

	// Register topic validator wrapper.
	err = receiverRouter.gossipSub.RegisterTopicValidator(
		protocol.BackgroundTopicStr,
		func(ctx context.Context, peerID libp2pPeer.ID, msg *pubsub.Message) bool {
			msgIsValid := receiverRouter.topicValidator(ctx, peerID, msg)
			require.Falsef(t, msgIsValid, "expected message to be invalid")

			return msgIsValid
		},
	)
	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			senderPrivKey, _ := newTestPeer(t)
			senderHost := newTestHost(t, libp2pMockNet, senderPrivKey)
			gossipPubsub, err := pubsub.NewGossipSub(ctx, senderHost)
			require.NoError(t, err)

			err = libp2pMockNet.LinkAll()
			require.NoError(t, err)

			receiverAddrInfo, err := utils.Libp2pAddrInfoFromPeer(receiverPeer)
			require.NoError(t, err)

			err = senderHost.Connect(ctx, receiverAddrInfo)
			require.NoError(t, err)

			topic, err := gossipPubsub.Join(protocol.BackgroundTopicStr)
			require.NoError(t, err)

			err = topic.Publish(ctx, testCase.msgBz)
			require.NoError(t, err)

			// Destroy previous topic and sender instances to start with new ones
			// for each test case.
			t.Cleanup(func() {
				_ = topic.Close()
				_ = senderHost.Close()
			})

			// Ensure no messages were handled at the end of each test case for
			// async errors.
			select {
			case <-receivedChan:
				t.Fatal("no messages should have been handled by receiver router")
			case <-time.After(invalidReceiveTimeout):
				// no error, continue
			}
		})
	}
}

func TestBackgroundRouter_Broadcast(t *testing.T) {
	const (
		numPeers            = 4
		testMsg             = "test messsage"
		testTimeoutDuration = time.Second * 5
	)

	var (
		ctx = context.Background()
		// mutex preventing concurrent writes to `seenMessages`
		seenMessagesMutext sync.Mutex
		// map used as a set to collect IDs of peers which have received a message
		seenMessages       = make(map[string]struct{})
		bootstrapWaitgroup = sync.WaitGroup{}
		broadcastWaitgroup = sync.WaitGroup{}
		broadcastDone      = make(chan struct{}, 1)
		testTimeout        = time.After(testTimeoutDuration)
		// NB: peerIDs are stringified
		actualPeerIDs   []string
		expectedPeerIDs = make([]string, numPeers)
		testHosts       = make([]libp2pHost.Host, 0)
		libp2pMockNet   = mocknet.New()
	)

	testPocketEnvelope, err := messaging.PackMessage(&anypb.Any{
		TypeUrl: "/test",
		Value:   []byte(testMsg),
	})
	require.NoError(t, err)

	testPocketEnvelopeBz, err := proto.Marshal(testPocketEnvelope)
	require.NoError(t, err)

	// setup 4 receiver routers to listen for incoming messages from the sender router
	for i := 0; i < numPeers; i++ {
		broadcastWaitgroup.Add(1)
		bootstrapWaitgroup.Add(1)

		privKey, peer := newTestPeer(t)
		host := newTestHost(t, libp2pMockNet, privKey)
		testHosts = append(testHosts, host)
		expectedPeerIDs[i] = host.ID().String()
		newRouterWithSelfPeerAndHost(t, peer, host, func(data []byte) error {
			seenMessagesMutext.Lock()
			defer seenMessagesMutext.Unlock()
			seenMessages[host.ID().String()] = struct{}{}
			broadcastWaitgroup.Done()
			return nil
		})
	}

	// bootstrap off of arbitrary testHost
	privKey, selfPeer := newTestPeer(t)

	// set up a test backgroundRouter
	testRouterHost := newTestHost(t, libp2pMockNet, privKey)
	testRouter := newRouterWithSelfPeerAndHost(t, selfPeer, testRouterHost, nil)
	testHosts = append(testHosts, testRouterHost)

	// simulate network links between each to every other
	// (i.e. fully-connected network)
	err = libp2pMockNet.LinkAll()
	require.NoError(t, err)

	// setup notifee/notify BEFORE bootstrapping
	notifee := &libp2pNetwork.NotifyBundle{
		ConnectedF: func(_ libp2pNetwork.Network, _ libp2pNetwork.Conn) {
			t.Logf("connected!")
			bootstrapWaitgroup.Done()
		},
	}
	testRouter.host.Network().Notify(notifee)

	bootstrap(t, ctx, testHosts)

	// broadcasting in a go routine so that we can wait for bootstrapping to
	// complete before broadcasting.
	go func() {
		// wait for hosts to listen and peer discovery
		bootstrapWaitgroup.Wait()
		// INVESTIGATE: look for a more idiomatic way to wait for DHT peer discovery to complete
		//
		// `bootstrapWaitgroup` isn't quite sufficient; I suspect the DHT
		// needs more time but am unaware of a notify/notifee interface (or
		// something similar) at that level.
		time.Sleep(time.Millisecond * 250)

		// broadcast message
		t.Log("broadcasting...")
		err := testRouter.Broadcast(testPocketEnvelopeBz)
		require.NoError(t, err)

		// wait for broadcast to be received by all peers
		broadcastWaitgroup.Wait()
		broadcastDone <- struct{}{}
	}()

	// waitgroup broadcastDone or timeout
	select {
	case <-testTimeout:
		t.Fatalf(
			"timed out waiting for all expected messages: got %d; wanted %d",
			len(seenMessages),
			numPeers,
		)
	case <-broadcastDone:
	}

	actualPeerIDs = testutil.GetKeys[string](seenMessages)
	require.ElementsMatchf(t, expectedPeerIDs, actualPeerIDs, "peerIDs don't match")
}

// bootstrap connects each host to one other except for the arbitrarily chosen "bootstrap host"
func bootstrap(t *testing.T, ctx context.Context, testHosts []libp2pHost.Host) {
	t.Helper()

	t.Log("bootstrapping...")
	bootstrapHost := testHosts[0]
	bootstrapAddr := bootstrapHost.Addrs()[0]
	for _, h := range testHosts {
		if h.ID() == bootstrapHost.ID() {
			continue
		}

		p2pAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", bootstrapHost.ID()))
		require.NoError(t, err)

		addrInfo := libp2pPeer.AddrInfo{
			ID: bootstrapHost.ID(),
			Addrs: []multiaddr.Multiaddr{
				bootstrapAddr.Encapsulate(p2pAddr),
			},
		}

		t.Logf("connecting to %s...", addrInfo.ID.String())
		err = h.Connect(ctx, addrInfo)
		require.NoError(t, err)
	}
}

// TECHDEBT(#609): move & de-duplicate
func newTestRouter(
	t *testing.T,
	libp2pMockNet mocknet.Mocknet,
	handler typesP2P.MessageHandler,
) *backgroundRouter {
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

	return newRouterWithSelfPeerAndHost(t, selfPeer, host, handler)
}

func newRouterWithSelfPeerAndHost(
	t *testing.T,
	selfPeer typesP2P.Peer,
	host libp2pHost.Host,
	handler typesP2P.MessageHandler,
) *backgroundRouter {
	t.Helper()

	ctrl := gomock.NewController(t)
	runtimeMgrMock := mockModules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(&configs.Config{
		P2P: &configs.P2PConfig{
			IsClientOnly: false,
		},
	}).AnyTimes()

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

	if handler == nil {
		handler = noopHandler
	}

	router, err := Create(busMock, &config.BackgroundConfig{
		Addr:                  selfPeer.GetAddress(),
		PeerstoreProvider:     pstoreProviderMock,
		CurrentHeightProvider: consensusMock,
		Host:                  host,
		Handler:               handler,
	})
	require.NoError(t, err)

	libp2pNet, ok := router.(*backgroundRouter)
	require.Truef(t, ok, "unexpected router type: %T", router)

	return libp2pNet
}

// TECHDEBT(#609): move & de-duplicate
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

func newTestHost(
	t *testing.T,
	mockNet mocknet.Mocknet,
	privKey cryptoPocket.PrivateKey,
) libp2pHost.Host {
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

func mustMarshal(t *testing.T, msg proto.Message) []byte {
	t.Helper()

	msgBz, err := proto.Marshal(msg)
	require.NoError(t, err)

	return msgBz
}
