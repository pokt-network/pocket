//go:build integration && test

package integration

import (
	"fmt"
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

	for _, p2pModule := range s.p2pModules {
		err := p2pModule.(*p2p.P2PModule).Start()
		require.NoError(s, err)
	}

	for _, busMock := range s.busMocks {
		//eventMetricsAgentMock := telemetry_testutil.PrepareEventMetricsAgentMock(s, serviceURL, &s.wg, peerCount)
		eventMetricsAgentMock := telemetry_testutil.BaseEventMetricsAgentMock(s)
		busMock.GetTelemetryModule().(*mock_modules.MockTelemetryModule).EXPECT().
			GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	}

	//for _, host := range libp2pNetworkMock.Hosts() {
	//	host.SetStreamHandler(protocol.PoktProtocolID, func(stream libp2pNetwork.Stream) {
	//		s.Logf("inbound stream protocol: %s", stream.Protocol())
	//		s.seenServiceURLs[stream.Conn().RemotePeer()] = struct{}{}
	//	})
	//}
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

		require.Lenf(
			s, s.seenServiceURLs,
			int(receivedCountPlus1-1),
			"expected to see %d peers, got: %v",
			receivedCountPlus1-1,
			len(s.seenServiceURLs),
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

//func TestBackgroundGossipIntegration(t *testing.T) {
//	t.Parallel()
//
//	testutil.RunGherkinFeature(t, backgroundGossipFeaturePath, initBackgroundGossipScenarios(t))
//}
//
//func aNodeBroadcastsATestMessageViaItsBackgroundRouter(ctx context.Context) (context.Context, error) {
//	t, err := testutil.GetTestingTFromContext(ctx)
//	if err != nil {
//		return ctx, err
//	}
//
//	p2pModules := ctx.Value("p2pModules").([]modules.P2PModule)
//
//	// select arbitrary sender & store in context for reference later
//	sender := p2pModules[0]
//	ctx = context.WithValue(ctx, "sender", sender)
//
//	// broadcast a test message
//	msg := &anypb.Any{}
//	err = sender.Broadcast(msg)
//	require.NoError(t, err)
//
//	return ctx, nil
//}
//
//func numberOfFaultyPeers(arg1 int) error {
//	return godog.ErrPending
//}
//
//func numberOfNodesShouldReceiveTheTestMessage(expectedReceivedCount int) error {
//	// wait for all nodes to receive the test message
//	// TODO
//
//	//require.Equal(t)
//
//	return godog.ErrPending
//}
//
//func initBackgroundGossipScenarios(t *testing.T) func(ctx *godog.ScenarioContext) {
//	return func(ctx *godog.ScenarioContext) {
//		ctx.Step(`^a faulty network of (\d+) peers$`, aFaultyNetworkOfPeers)
//		ctx.Step(`^a fully connected network of (\d+) peers$`, aFullyConnectedNetworkOfPeers)
//		ctx.Step(`^a node broadcasts a test message via its background router$`, aNodeBroadcastsATestMessageViaItsBackgroundRouter)
//		ctx.Step(`^(\d+) number of faulty peers$`, numberOfFaultyPeers)
//		ctx.Step(`^(\d+) number of nodes join the network$`, numberOfNodesJoinTheNetwork)
//		ctx.Step(`^(\d+) number of nodes leave the network$`, numberOfNodesLeaveTheNetwork)
//		ctx.Step(`^(\d+) number of nodes should receive the test message$`, numberOfNodesShouldReceiveTheTestMessage)
//	}
//}
