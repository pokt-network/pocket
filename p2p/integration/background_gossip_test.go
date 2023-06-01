//go:build integration && test

package integration

import (
	"fmt"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	"sync"
	"testing"
	"time"

	libp2pMocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/constructors"
	"github.com/pokt-network/pocket/internal/testutil/generics"
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
	telemetry_testutil "github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/runtime/defaults"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
)

const (
	backgroundGossipFeaturePath = "background_gossip.feature"
	broadcastTimeoutDuration    = time.Second * 2
)

func TestMinimal(t *testing.T) {
	t.Parallel()

	// a new step definition suite is constructed for every scenario
	gocuke.NewRunner(t, &suite{}).Path(backgroundGossipFeaturePath).Run()
}

type suite struct {
	// special arguments like TestingT are injected automatically into exported fields
	gocuke.TestingT

	timeoutDuration time.Duration
	// TODO_THIS_COMMIT: rename
	mu sync.Mutex
	// seenServiceURLs is used as a map to track which messages have been seen
	// by which nodes
	seenServiceURLs   map[string]struct{}
	receivedCount     int
	p2pModules        map[string]modules.P2PModule
	busMocks          map[string]*mock_modules.MockBus
	libp2pNetworkMock libp2pMocknet.Mocknet
	sender            *p2p.P2PModule
	// TODO_THIS_COMMIT: reanme
	wg sync.WaitGroup
}

func (s *suite) AFaultyNetworkOfPeers(a int64) {
	panic("PENDING")
}

func (s *suite) NumberOfFaultyPeers(a int64) {
	panic("PENDING")
}

func (s *suite) NumberOfNodesJoinTheNetwork(a int64) {
	panic("PENDING")
}

func (s *suite) NumberOfNodesLeaveTheNetwork(a int64) {
	panic("PENDING")
}

func (s *suite) AFullyConnectedNetworkOfPeers(count int64) {
	var (
		peerCount = int(count)
		//pubKeys   = make([]cryptoPocket.PublicKey, peerCount)
	)
	s.Logf("ADDING peerCount - 1: %d", peerCount-1)
	s.wg.Add(peerCount - 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seenServiceURLs = make(map[string]struct{})

	//for i, privKey := range testutil.LoadLocalnetPrivateKeys(s, peerCount) {
	//	pubKeys[i] = privKey.PublicKey()
	//}
	//genesisState := runtime_testutil.GenesisWithSequentialServiceURLs(s, pubKeys)
	// TODO_THIS_COMMIT: explain
	genesisState := runtime_testutil.GenesisWithSequentialServiceURLs(s, nil)

	busEventHandlerFactory := func(t gocuke.TestingT, busMock *mock_modules.MockBus) testutil.BusEventHandler {
		// event handler is called when a p2p module receives a network message
		return func(data *messaging.PocketEnvelope) {
			s.mu.Lock()
			defer s.mu.Unlock()

			defer func() {
				if r := recover(); r != nil {
					t.Logf("seenCount: %d; seenServiceURLs: %v", len(s.seenServiceURLs), s.seenServiceURLs)
					//panic(r)
					t.Fatalf("panic: %v", r)
				}
			}()

			p2pCfg := busMock.GetRuntimeMgr().GetConfig().P2P
			serviceURL := fmt.Sprintf("%s:%d", p2pCfg.Hostname, defaults.DefaultP2PPort)
			t.Logf("received message by %s", serviceURL)

			peerPrivKey, err := cryptoPocket.NewPrivateKey(p2pCfg.PrivateKey)
			require.NoError(t, err)

			senderAddr, err := s.sender.GetAddress()
			require.NoError(t, err)

			if senderAddr.Equals(peerPrivKey.Address()) {
				t.Logf("SELF: %s", serviceURL)
				return
			}

			if _, ok := s.seenServiceURLs[serviceURL]; ok {
				t.Logf("DUPLICATE SERVICE URL: %s", serviceURL)
				return
			}

			s.receivedCount++
			s.seenServiceURLs[serviceURL] = struct{}{}
			s.wg.Done()
		}
	}

	// setup mock network
	s.busMocks, s.libp2pNetworkMock, s.p2pModules = constructors.NewBusesMocknetAndP2PModules(
		s, peerCount,
		genesisState,
		busEventHandlerFactory,
	)

	// add expectations for P2P events to telemetry module's event metrics agent
	for _, busMock := range s.busMocks {
		eventMetricsAgentMock := busMock.
			GetTelemetryModule().
			GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent)

		telemetry_testutil.WithP2PIntegrationEvents(
			s, eventMetricsAgentMock,
		)
	}

	// TODO_THIS_COMMIT: bus event handler based wg.Done()!

	// start P2P modules of all peers
	handleCount := 0
	for _, p2pModule := range s.p2pModules {
		err := p2pModule.(*p2p.P2PModule).Start()
		require.NoError(s, err)

		handlerProxyFactory := func(
			origHandler typesP2P.RouterHandler,
		) (proxyHandler typesP2P.RouterHandler) {
			return func(data []byte) error {
				s.mu.Lock()
				handleCount++
				s.mu.Unlock()

				s.Logf("handleCount: %d", handleCount)
				//s.wg.Done()
				return origHandler(data)

				//return nil
			}
		}

		// TODO_THIS_COMMIT: look into go-libp2p-pubsub tracing
		// (see: https://github.com/libp2p/go-libp2p-pubsub#tracing)
		noopHandlerProxyFactory := func(_ typesP2P.RouterHandler) typesP2P.RouterHandler {
			return func(_ []byte) error {
				// noop
				return nil
			}
		}

		p2pModule.(*p2p.P2PModule).GetRainTreeRouter().HandlerProxy(
			s, noopHandlerProxyFactory,
		)
		p2pModule.(*p2p.P2PModule).GetBackgroundRouter().HandlerProxy(
			s, handlerProxyFactory,
		)
	}

	//time.Sleep(time.Millisecond * 500)

	// (NOPE) WIP: host-level intercept...
	//for _, host := range s.libp2pNetworkMock.Hosts() {
	//	//s.Logf("host protocols: %v", host.Mux().Protocols())
	//	//host.SetStreamHandler(protocol.PoktProtocolID, func(stream libp2pNetwork.Stream) {
	//	host.SetStreamHandler(pubsub.FloodSubID, func(stream libp2pNetwork.Stream) {
	//		//s.Logf("inbound stream protocol: %s", stream.Protocol())
	//		//	//s.seenServiceURLs[stream.Conn().RemotePeer()] = struct{}{}
	//		//	data, err := io.ReadAll(stream)
	//		//	require.NoError(s, err)
	//		//
	//		//	s.Logf("stream data: %s", data)
	//	})
	//	host.SetStreamHandler(dht.ProtocolDHT, func(stream libp2pNetwork.Stream) {
	//		//s.Logf("inbound stream protocol: %s", stream.Protocol())
	//		//s.seenServiceURLs[stream.Conn().RemotePeer()] = struct{}{}
	//		//data, err := io.ReadAll(stream)
	//		//require.NoError(s, err)
	//		//
	//		//s.Logf("stream data: %s", data)
	//	})
	//}
}

func (s *suite) ANodeBroadcastsATestMessageViaItsBackgroundRouter() {
	s.timeoutDuration = broadcastTimeoutDuration

	// select arbitrary sender & store in context for reference later
	s.sender = s.p2pModules[generics_testutil.GetKeys(s.p2pModules)[0]].(*p2p.P2PModule)

	// broadcast a test message
	msg := &anypb.Any{}

	// TODO:
	// - disable raintree router OR broadcast w/ bg router only

	err := s.sender.Broadcast(msg)
	require.NoError(s, err)
}

func (s *suite) MinusOneNumberOfNodesShouldReceiveTheTestMessage(receivedCountPlus1 int64) {
	done := make(chan struct{}, 1)

	go func() {
		s.wg.Wait()
		s.mu.Lock()
		defer s.mu.Unlock()

		receivedCount := int(receivedCountPlus1 - 1)
		require.Lenf(
			s, s.seenServiceURLs, receivedCount,
			"expected to see %d peers, got: %v",
			receivedCount, len(s.seenServiceURLs),
		)
		done <- struct{}{}
	}()

	select {
	case <-time.After(s.timeoutDuration):
		s.mu.Lock()
		defer s.mu.Unlock()

		s.Fatalf("timed out waiting for messages to be received; received: %d; seenServiceURLs: %v", s.receivedCount, s.seenServiceURLs)
	case <-done:
		s.Logf("seenCount: %d; seenServiceURLs: %v", len(s.seenServiceURLs), s.seenServiceURLs)
	}
}
