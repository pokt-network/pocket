//go:build integration && test

package integration

import (
	"fmt"
	"github.com/foxcpp/go-mockdns"
	libp2pNetwork "github.com/libp2p/go-libp2p/core/network"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
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

type PeerIDSet map[libp2pPeer.ID]struct{}

const (
	backgroundGossipFeaturePath = "background_gossip.feature"
	//broadcastTimeoutDuration    = time.Millisecond * 250
	broadcastTimeoutDuration = time.Second * 3
	// TODO_THIS_COMMIT: move
	bootstrapTimeoutDuration = time.Second * 3
)

func TestMinimal(t *testing.T) {
	t.Parallel()

	// a new step definition suite is constructed for every scenario
	gocuke.NewRunner(t, new(backgroundGossipSuite)).Path(backgroundGossipFeaturePath).Run()
}

type peerConnectionEvent struct {
	localID  libp2pPeer.ID
	remoteID libp2pPeer.ID
}

type backgroundGossipSuite struct {
	// special arguments like TestingT are injected automatically into exported fields
	gocuke.TestingT
	dnsSrv *mockdns.Server

	timeoutDuration time.Duration
	// TODO_THIS_COMMIT: rename
	mu                   sync.Mutex
	receivedServiceURLCh chan string
	// receivedServiceURLMap is used as a map to track which messages have been
	// received by which nodes.
	receivedServiceURLMap map[string]struct{}
	bootstrapMutex        sync.Mutex

	// bootstrapPeerIDChMap is a mapping between the peerID string of each node to
	// a channel that will be used to signal the peer ID strings of each node it
	// has discovered.
	//bootstrapPeerIDChMap map[libp2pPeer.ID]chan libp2pPeer.ID

	bootstrapPeerIDCh chan peerConnectionEvent

	// bootstrapPeerIDsMap is a mapping between the peerID string of each node to a
	// set of peerID strings that node has discovered. This set is represented as
	// a map with the peerID string as the key and an empty struct as the value.
	bootstrapPeerIDsMap map[libp2pPeer.ID]PeerIDSet

	bootstrapNetworkWaitGroup sync.WaitGroup
	receivedCount             int
	receivedWaitGroup         sync.WaitGroup
	p2pModules                map[string]modules.P2PModule
	busMocks                  map[string]*mock_modules.MockBus
	libp2pNetworkMock         libp2pMocknet.Mocknet
	sender                    *p2p.P2PModule
}

func (s *backgroundGossipSuite) Before(_ gocuke.Scenario) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dnsSrv = testutil.MinimalDNSMock(s)
	//s.bootstrapPeerIDChMap = make(map[libp2pPeer.ID]chan libp2pPeer.ID)
	s.bootstrapPeerIDCh = make(chan peerConnectionEvent)
	s.receivedServiceURLCh = make(chan string)
	s.receivedServiceURLMap = make(map[string]struct{})
}

func (s *backgroundGossipSuite) AFaultyNetworkOfPeers(a int64) {
	panic("PENDING")
}

func (s *backgroundGossipSuite) NumberOfFaultyPeers(a int64) {
	panic("PENDING")
}

func (s *backgroundGossipSuite) NumberOfNodesJoinTheNetwork(a int64) {
	panic("PENDING")
}

func (s *backgroundGossipSuite) NumberOfNodesLeaveTheNetwork(a int64) {
	panic("PENDING")
}

func (s *backgroundGossipSuite) AFullyConnectedNetworkOfPeers(count int64) {
	peerCount := int(count)

	// TODO_THIS_COMMOT: comment, explain
	//s.bootstrapNetworkWaitGroup.Add(peerCount)
	//s.bootstrapNetworkWaitGroup.Add(peerCount * (peerCount - 1) / 2)
	s.bootstrapNetworkWaitGroup.Add(peerCount * (peerCount - 1))
	s.receivedWaitGroup.Add(peerCount - 1)
	s.mu.Lock()
	defer s.mu.Unlock()

	serviceURLKeyMap := testutil.SequentialServiceURLPrivKeyMap(s, peerCount)
	genesisState := runtime_testutil.BaseGenesisStateMockFromServiceURLKeyMap(s, serviceURLKeyMap)

	// TODO_THIS_COMMIT: refactor
	go func() {
		for serviceURL := range s.receivedServiceURLCh {
			// use a channel instead of a map; send a struct like w/ trackBootstrapProgress()
			// THIS IS PER BUS (i.e. run in each node)!!
			_, ok := s.receivedServiceURLMap[serviceURL]
			require.Falsef(s, ok, "received message from duplicate serviceURL: %s", serviceURL)

			s.receivedCount++
			s.receivedServiceURLMap[serviceURL] = struct{}{}
			s.receivedWaitGroup.Done()
		}
	}()

	s.Cleanup(func() {
		close(s.receivedServiceURLCh)
	})

	busEventHandlerFactory := func(t gocuke.TestingT, busMock *mock_modules.MockBus) testutil.BusEventHandler {
		// event handler is called when a p2p module receives a network message
		return func(data *messaging.PocketEnvelope) {
			s.mu.Lock()
			defer s.mu.Unlock()

			//defer func() {
			//	if r := recover(); r != nil {
			//		t.Logf("seenCount: %d; receivedServiceURLMap: %v", len(s.receivedServiceURLMap), s.receivedServiceURLMap)
			//		//panic(r)
			//		t.Fatalf("panic: %v", r)
			//	}
			//}()

			p2pCfg := busMock.GetRuntimeMgr().GetConfig().P2P
			serviceURL := fmt.Sprintf("%s:%d", p2pCfg.Hostname, defaults.DefaultP2PPort)

			peerPrivKey, err := cryptoPocket.NewPrivateKey(p2pCfg.PrivateKey)
			require.NoError(t, err)

			senderAddr, err := s.sender.GetAddress()
			require.NoError(t, err)

			if senderAddr.Equals(peerPrivKey.Address()) {
				t.Logf("ignoring message from self: %s", serviceURL)
				return
			}
			t.Logf("received message from %s", serviceURL)

			// TODO: RESUME HERE!!!
			// TODO: RESUME HERE!!!
			// TODO: RESUME HERE!!!

			s.receivedServiceURLCh <- serviceURL
		}
	}
	// --

	// TODO_THIS_COMMIT: refactor
	debugNotifiee := testutil.NewDebugNotifee(s)
	notifiee := &libp2pNetwork.NotifyBundle{
		//DisconnectedF: func(network libp2pNetwork.Network, conn libp2pNetwork.Conn) {
		//	s.Logf("disconnected: %s", conn.RemotePeer())
		//},
		DisconnectedF: debugNotifiee.Disconnected,
		ConnectedF: func(net libp2pNetwork.Network, conn libp2pNetwork.Conn) {
			//s.Logf("connected: %s", conn.RemotePeer())
			//s.Logf("pstore size: %d", len(p2pModule.GetHost().Peerstore().Peers()))

			s.bootstrapPeerIDCh <- peerConnectionEvent{
				localID:  conn.LocalPeer(),
				remoteID: conn.RemotePeer(),
			}
			//s.bootstrapPeerIDChMap[p2pModule.GetHost().ID()] <- conn.RemotePeer()
			//s.Logf("bootstrap peer ID sent on channel")
			//if len(p2pModule.GetHost().Peerstore().Peers()) == peerCount {
			//	countz++
			//	s.Logf("count: %d", countz)
			//	//s.bootstrapNetworkWaitGroup.Done()
			//	//s.bootstrapNetworkWaitGroup.Done()
			//	//s.bootstrapNetworkWaitGroup.Done()
			//}
			debugNotifiee.Connected(net, conn)
		},
		ListenF:      debugNotifiee.Listen,
		ListenCloseF: debugNotifiee.ListenClose,
	}
	// --

	// setup mock network
	s.busMocks, s.libp2pNetworkMock, s.p2pModules = constructors.NewBusesMocknetAndP2PModules(
		s, peerCount,
		s.dnsSrv,
		genesisState,
		busEventHandlerFactory,
		notifiee,
	)

	// add expectations for P2P events to telemetry module's event metrics agent
	for _, busMock := range s.busMocks {
		// TODO_THIS_COMMIT: ??
		eventMetricsAgentMock := busMock.
			GetTelemetryModule().
			GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent)

		// TODO_THIS_COMMIT: ??
		telemetry_testutil.WithP2PIntegrationEvents(
			s, eventMetricsAgentMock,
		)
	}

	// TODO_THIS_COMMIT: bus event handler based receivedWaitGroup.Done()!

	// concurrently update `s.bootstrapPeerIDsMap` by receiving from the
	// corresponding channel from `s.bootstrapPeerIDChMap` that  `notifee`
	// is sending on.
	go s.trackBootstrapProgress(peerCount - 1)

	// start P2P modules of all peers
	//handleCount := 0
	for _, module := range s.p2pModules {
		//countz := 0
		p2pModule := module.(*p2p.P2PModule)

		// TODO_THIS_COMMIT: refactor
		//debugNotifiee := testutil.NewDebugNotifee(s)
		//notifee := &libp2pNetwork.NotifyBundle{
		//	//DisconnectedF: func(network libp2pNetwork.Network, conn libp2pNetwork.Conn) {
		//	//	s.Logf("disconnected: %s", conn.RemotePeer())
		//	//},
		//	DisconnectedF: debugNotifiee.Disconnected,
		//	ConnectedF: func(net libp2pNetwork.Network, conn libp2pNetwork.Conn) {
		//		//s.Logf("connected: %s", conn.RemotePeer())
		//		//s.Logf("pstore size: %d", len(p2pModule.GetHost().Peerstore().Peers()))
		//		s.bootstrapMutex.Lock()
		//		defer s.bootstrapMutex.Unlock()
		//
		//		s.bootstrapPeerIDCh <- peerConnectionEvent{
		//			localID:  conn.LocalPeer(),
		//			remoteID: conn.RemotePeer(),
		//		}
		//		//s.bootstrapPeerIDChMap[p2pModule.GetHost().ID()] <- conn.RemotePeer()
		//		//s.Logf("bootstrap peer ID sent on channel")
		//		//if len(p2pModule.GetHost().Peerstore().Peers()) == peerCount {
		//		//	countz++
		//		//	s.Logf("count: %d", countz)
		//		//	//s.bootstrapNetworkWaitGroup.Done()
		//		//	//s.bootstrapNetworkWaitGroup.Done()
		//		//	//s.bootstrapNetworkWaitGroup.Done()
		//		//}
		//		debugNotifiee.Connected(net, conn)
		//	},
		//	ListenF:      debugNotifiee.Listen,
		//	ListenCloseF: debugNotifiee.ListenClose,
		//}
		////p2pModule.GetHost().Network().Notify(debugNotifiee)
		//p2pModule.GetHost().Network().Notify(notifee)
		// --

		err := p2pModule.Start()
		require.NoError(s, err)

		// TODO_THIS_COMMIT: fix
		//s.Cleanup(func() {
		//	err := p2pModule.Stop()
		//	require.NoError(s, err)
		//})

		handlerProxyFactory := func(
			origHandler typesP2P.RouterHandler,
		) (proxyHandler typesP2P.RouterHandler) {
			return func(data []byte) error {
				//s.receivedWaitGroup.Done()
				return origHandler(data)
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

		p2pModule.GetRainTreeRouter().HandlerProxy(
			s, noopHandlerProxyFactory,
		)

		p2pModule.GetBackgroundRouter().HandlerProxy(
			s, handlerProxyFactory,
		)

	}

	// wait for bootstrapping to complete
	bootstrapDone := make(chan struct{}, 0)
	go func() {
		s.bootstrapNetworkWaitGroup.Wait()
		close(bootstrapDone)
	}()

	select {
	case <-time.After(bootstrapTimeoutDuration):
		s.Fatal("timed out waiting for bootstrapping")
	case <-bootstrapDone:
	}
}

func (s *backgroundGossipSuite) ANodeBroadcastsATestMessageViaItsBackgroundRouter() {
	// TODO_THIS_COMMIT: refactor
	s.timeoutDuration = broadcastTimeoutDuration

	// select arbitrary sender & store in context for reference later
	s.sender = s.p2pModules[generics_testutil.GetKeys(s.p2pModules)[0]].(*p2p.P2PModule)

	// broadcast a test message
	msg := &anypb.Any{}

	// TODO_THIS_COMMIT: try to remove...
	// wait for DHT bootstrapping
	time.Sleep(time.Millisecond * 250)

	err := s.sender.Broadcast(msg)
	require.NoError(s, err)
}

func (s *backgroundGossipSuite) MinusOneNumberOfNodesShouldReceiveTheTestMessage(receivedCountPlus1 int64) {
	done := make(chan struct{}, 1)

	go func() {
		s.receivedWaitGroup.Wait()
		s.mu.Lock()
		defer s.mu.Unlock()

		receivedCount := int(receivedCountPlus1 - 1)
		require.Lenf(
			s, s.receivedServiceURLMap, receivedCount,
			"expected to see %d peers, got: %v",
			receivedCount, len(s.receivedServiceURLMap),
		)
		done <- struct{}{}
	}()

	select {
	case <-time.After(s.timeoutDuration):
		s.mu.Lock()
		defer s.mu.Unlock()

		s.Fatalf("timed out waiting for messages to be received; received: %d; receivedServiceURLMap: %v", s.receivedCount, s.receivedServiceURLMap)
	case <-done:
		s.Logf("seenCount: %d; receivedServiceURLMap: %v", len(s.receivedServiceURLMap), s.receivedServiceURLMap)
	}
}

//func (s *backgroundGossipSuite) initBootstrapPeerIDChMap(p2pModule *p2p.P2PModule) {
//	s.bootstrapMutex.Lock()
//	defer s.bootstrapMutex.Unlock()
//
//	selfID := p2pModule.GetHost().ID()
//	// initialize `s.bootstrapPeerIDChMap` for each p2pModule
//	if _, ok := s.bootstrapPeerIDChMap[selfID]; !ok {
//		s.bootstrapPeerIDChMap[selfID] = make(chan libp2pPeer.ID)
//	}
//}

func (s *backgroundGossipSuite) initBootstrapPeerIDsMap(selfID libp2pPeer.ID) {
	// TODO_THIS_COMMIT: need this?
	s.bootstrapMutex.Lock()
	defer s.bootstrapMutex.Unlock()

	// initialize `s.bootstrapPeerIDsMap`
	if s.bootstrapPeerIDsMap == nil {
		s.bootstrapPeerIDsMap = make(map[libp2pPeer.ID]PeerIDSet)
	}

	// initialize `s.bootstrapPeerIDsMap` for each p2pModule
	if _, ok := s.bootstrapPeerIDsMap[selfID]; !ok {
		s.bootstrapPeerIDsMap[selfID] = make(PeerIDSet)
	}
}

func (s *backgroundGossipSuite) trackBootstrapProgress(peerCount int) {
	//selfID := p2pModule.GetHost().ID()

	// add unique bootstrap peer IDs to `bootstrapPeerIDsMap` for this
	// p2pModule (`selfID`) as they connect
	for newPeerConnectionEvent := range s.bootstrapPeerIDCh {
		//newBootstrapPeerID := <-s.bootstrapPeerIDChMap[selfID]
		localID, remoteID := newPeerConnectionEvent.localID, newPeerConnectionEvent.remoteID

		s.initBootstrapPeerIDsMap(localID)

		if localID == remoteID {
			// don't count self as a bootstrap peer
			s.Logf("self bootstrap peer ID: %s")
			continue
		}

		if _, ok := s.bootstrapPeerIDsMap[localID][remoteID]; ok {
			// already connected to this peer during bootstrapping
			s.Logf("duplicate bootstrap peer ID: %s", remoteID)
			continue
		}
		s.bootstrapPeerIDsMap[localID][remoteID] = struct{}{}
		s.bootstrapNetworkWaitGroup.Done()

		//p2pModule := s.p2pModules[serviceURL].(*p2p.P2PModule)
		//if len(p2pModule.GetHost().Network().Conns()) == peerCount {
		//	s.Logf("bootstrap peer connections len: %d", len(p2pModule.GetHost().Network().Conns()))
		//	connections := p2pModule.GetHost().Network().Conns()
		//	remoteConnPeers := make([]libp2pPeer.ID, len(connections))
		//	for i, conn := range connections {
		//		remoteConnPeers[i] = conn.RemotePeer()
		//	}
		//	s.Logf("p2pModule.GetHost().Network().Conns(): %v", remoteConnPeers)
		//	s.Logf("s.bootstrapPeerIDsMap[selfID]: %v", s.bootstrapPeerIDsMap[localID])
		//	s.bootstrapNetworkWaitGroup.Done()
		//}
	}
}
