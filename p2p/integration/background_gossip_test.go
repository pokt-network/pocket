//go:build integration && test

package integration

import (
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	telemetry_testutil "github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/runtime/defaults"
	"sync"
	"testing"
	"time"

	libp2pMocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/constructors"
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
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
		pubKeys   = make([]cryptoPocket.PublicKey, peerCount)
	)
	s.wg.Add(peerCount - 1)
	s.seenServiceURLs = make(map[string]struct{})

	for i, privKey := range testutil.LoadLocalnetPrivateKeys(s, peerCount) {
		pubKeys[i] = privKey.PublicKey()
	}
	genesisState := runtime_testutil.GenesisWithSequentialServiceURLs(s, pubKeys)
	busEventHandlerFactory := func(t gocuke.TestingT, busMock *mock_modules.MockBus) testutil.BusEventHandler {
		// event handler is called when a p2p module receives a network message
		return func(data *messaging.PocketEnvelope) {
			s.mu.Lock()
			defer s.mu.Unlock()

			defer func() {
				if r := recover(); r != nil {
					t.Logf("seenServiceURLs: %v", s.seenServiceURLs)
					//panic(r)
					t.Fatalf("panic: %v", r)
				}
			}()

			p2pCfg := busMock.GetRuntimeMgr().GetConfig().P2P
			serviceURL := fmt.Sprintf("%s:%d", p2pCfg.Hostname, defaults.DefaultP2PPort)
			t.Logf("received message by %s", serviceURL)
			if _, ok := s.seenServiceURLs[serviceURL]; ok {
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
	for _, p2pModule := range s.p2pModules {
		// (NOPE) WIP: pubsub-level intercept...
		//p2pModule.GetBackgroundRouter().Get

		err := p2pModule.(*p2p.P2PModule).Start()
		require.NoError(s, err)
	}

	// (NOPE) WIP: host-level intercept...
	for _, host := range s.libp2pNetworkMock.Hosts() {
		s.Logf("host protocols: %v", host.Mux().Protocols())
		//host.SetStreamHandler(protocol.PoktProtocolID, func(stream libp2pNetwork.Stream) {
		host.SetStreamHandler(pubsub.FloodSubID, func(stream libp2pNetwork.Stream) {
			s.Logf("inbound stream protocol: %s", stream.Protocol())
			//	//s.seenServiceURLs[stream.Conn().RemotePeer()] = struct{}{}
			//	data, err := io.ReadAll(stream)
			//	require.NoError(s, err)
			//
			//	s.Logf("stream data: %s", data)
		})
		host.SetStreamHandler(dht.ProtocolDHT, func(stream libp2pNetwork.Stream) {
			s.Logf("inbound stream protocol: %s", stream.Protocol())
			//s.seenServiceURLs[stream.Conn().RemotePeer()] = struct{}{}
			//data, err := io.ReadAll(stream)
			//require.NoError(s, err)
			//
			//s.Logf("stream data: %s", data)
		})
	}
}

func (s *suite) ANodeBroadcastsATestMessageViaItsBackgroundRouter() {
	s.timeoutDuration = broadcastTimeoutDuration

	// select arbitrary sender & store in context for reference later
	sender := s.p2pModules[testutil.GetKeys(s.p2pModules)[0]].(*p2p.P2PModule)

	// broadcast a test message
	msg := &anypb.Any{}
	err := sender.Broadcast(msg)
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
	}
}
