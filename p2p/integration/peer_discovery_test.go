//go:build test

package integration

import (
	"github.com/foxcpp/go-mockdns"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/constructors"
	generics_testutil "github.com/pokt-network/pocket/internal/testutil/generics"
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
	"github.com/pokt-network/pocket/p2p"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	peerDiscoveryFeaturePath = "peer_discovery.feature"
	bootstrapNodeLabelPrefix = "bootstrap"
	otherNodeLabelPrefix     = "other"
)

func TestPeerDiscoveryIntegration(t *testing.T) {
	t.Parallel()

	gocuke.NewRunner(t, new(backgroundPeerDiscoverySuite)).Path(peerDiscoveryFeaturePath).Run()
}

type backgroundPeerDiscoverySuite struct {
	gocuke.TestingT

	dnsSrv            *mockdns.Server
	busMocks          map[string]*mock_modules.MockBus
	libp2pNetworkMock mocknet.Mocknet
	p2pModules        map[string]modules.P2PModule

	// labelServiceURLMap a list of serviceURLs to a set of labels; intended to
	// be access and updated via `#getServiceURLsWithLabel()` and
	// `#addServiceURLWithLabel()`, respectively.
	labelServiceURLMap map[string][]string
}

func (s *backgroundPeerDiscoverySuite) Before(_ gocuke.Scenario) {
	s.labelServiceURLMap = make(map[string][]string)
}

func (s *backgroundPeerDiscoverySuite) ANetworkContainingANode(nodeLabel string) {
	s.dnsSrv = testutil.MinimalDNSMock(s)

	s.busMocks, s.libp2pNetworkMock, s.p2pModules = constructors.NewBusesMocknetAndP2PModules(
		s, 1, s.dnsSrv, nil, nil, nil,
	)

	// i.e. "only" serviceURL as this step definition initializes the network
	firstServiceURL := generics_testutil.GetKeys(s.busMocks)[0]
	s.addServiceURLWithLabel(nodeLabel, firstServiceURL)

	//debugNotifee := testutil.NewDebugNotifee(s)
	bootstrapP2PModule := s.p2pModules[firstServiceURL]
	//bootstrapP2PModule.(*p2p.P2PModule).GetHost().Network().Notify(debugNotifee)

	err := bootstrapP2PModule.Start()
	require.NoError(s, err)

	// TODO_THIS_COMMIT: revisit...
	time.Sleep(time.Second * 1)
}

func (s *backgroundPeerDiscoverySuite) NumberOfNodesJoinTheNetwork(nodeCount int64, nodeLabel string) {
	// TECHDEBT: use an iterator instead..
	// plus 1 to account for the bootstrap node serviceURL which we used earlier
	serviceURLKeyMap := testutil.SequentialServiceURLPrivKeyMap(s, int(nodeCount+1))

	// TODO_THIS_COMMIT: clarify how to use `bootstrapNodeLabelPrefix` / what it is for
	bootstrapNodeServiceURLs := s.getServiceURLsWithLabel(bootstrapNodeLabelPrefix)
	require.Equalf(s, 1, len(bootstrapNodeServiceURLs), "expected exactly one bootstrap node")

	bootstrapNodeServiceURL := bootstrapNodeServiceURLs[0]

	// bootstrap node is the only validator in genesis
	genesisState := runtime_testutil.BaseGenesisStateMockFromServiceURLKeyMap(
		s, map[string]cryptoPocket.PrivateKey{
			bootstrapNodeServiceURL: serviceURLKeyMap[bootstrapNodeServiceURL],
		},
	)

	// remove the bootstrap node from the map
	delete(serviceURLKeyMap, bootstrapNodeServiceURL)

	for serviceURL := range serviceURLKeyMap {
		s.addServiceURLWithLabel(otherNodeLabelPrefix, serviceURL)
	}

	joinersBusMocks, joinersP2PModules := constructors.NewBusesAndP2PModules(
		s, nil,
		s.dnsSrv,
		genesisState,
		s.libp2pNetworkMock,
		serviceURLKeyMap,
		nil,
	)

	err := s.libp2pNetworkMock.LinkAll()
	require.NoError(s, err)

	for _, p2pModule := range joinersP2PModules {
		debugNotifee := testutil.NewDebugNotifee(s)
		p2pModule.(*p2p.P2PModule).GetHost().Network().Notify(debugNotifee)

		err := p2pModule.Start()
		require.NoError(s, err)
	}

	s.addBusMocks(joinersBusMocks)
	s.addP2PModules(joinersP2PModules)

	// TODO_THIS_COMMIT: wait for bootstrapping
	s.Log("STARTING DELAY...")
	time.Sleep(time.Second * 5)
	s.Log("DELAY OVER")
}

// TODO_THIS_COMMIT: move below exported methods
func (s *backgroundPeerDiscoverySuite) addServiceURLWithLabel(nodeLabel, serviceURL string) {
	serviceURLs := s.labelServiceURLMap[nodeLabel]
	if serviceURLs == nil {
		serviceURLs = make([]string, 0)
	}
	s.labelServiceURLMap[nodeLabel] = append(serviceURLs, serviceURL)
}
func (s *backgroundPeerDiscoverySuite) getServiceURLsWithLabel(nodeLabel string) []string {
	serviceURLs, ok := s.labelServiceURLMap[nodeLabel]
	if !ok {
		return nil
	}

	if len(serviceURLs) < 1 {
		return nil
	}
	return serviceURLs
}

func (s *backgroundPeerDiscoverySuite) TheNodeShouldHavePlusOneNumberOfPeersInItsPeerstore(nodeLabel string, peerCountMinus1 int64) {
	labelServiceURLs := s.getServiceURLsWithLabel(nodeLabel)
	require.NotEmptyf(s, labelServiceURLs, "node label %q not found", nodeLabel)
	require.Equalf(s, 1, len(labelServiceURLs), "node label %q has more than one service url", nodeLabel)

	serviceURL := labelServiceURLs[0]
	p2pModule, ok := s.p2pModules[serviceURL]
	require.Truef(s, ok, "p2p module for service url %q not found", serviceURL)

	host := p2pModule.(*p2p.P2PModule).GetHost()
	peers := host.Peerstore().Peers()

	require.Equalf(s, int(peerCountMinus1+1), len(peers), "unexpected number of peers in peerstore: %s", peers)
}

func (s *backgroundPeerDiscoverySuite) OtherNodesShouldHavePlusOneNumberOfPeersInTheirPeerstores(peerCountMinus1 int64) {
	// TODO_THIS_COMMIT: clarify how to use `bootstrapNodeLabelPrefix` / what it is for
	otherNodeServiceURLs := s.getServiceURLsWithoutLabel(bootstrapNodeLabelPrefix)

	require.NotEmpty(s, otherNodeServiceURLs, "other nodes not found")
	require.Equal(s, int(peerCountMinus1), len(otherNodeServiceURLs))

	for _, serviceURL := range otherNodeServiceURLs {
		p2pModule, ok := s.p2pModules[serviceURL]
		require.Truef(s, ok, "p2p module for service url %q not found", serviceURL)

		host := p2pModule.(*p2p.P2PModule).GetHost()
		peers := host.Peerstore().Peers()

		require.Equalf(s, int(peerCountMinus1+1), len(peers), "unexpected number of peers in peerstore: %s", peers)
	}
}

//func (s *backgroundPeerDiscoverySuite) EachNodeShouldNotHaveAnyLeaversInTheirPeerstores() {
//	panic("PENDING")
//}
//
//func (s *backgroundPeerDiscoverySuite) LeaverNumberOfNodesLeaveTheNetwork() {
//	panic("PENDING")
//}

func (s *backgroundPeerDiscoverySuite) EachNodeShouldHaveNumberOfPeersInTheirRespectivePeerstores(peerCount int64) {
	panic("PENDING")
}

//func (s *backgroundPeerDiscoverySuite) NumberOfNodesBootstrapInPartition(a int64, b string) {
//	panic("PENDING")
//}

func (s *backgroundPeerDiscoverySuite) TheNetworkShouldContainNumberOfNodes(nodeCount int64) {
	panic("PENDING")
}

//func (s *backgroundPeerDiscoverySuite) ANodeInPartition(a string, b string) {
//	panic("PENDING")
//}
//
//func (s *backgroundPeerDiscoverySuite) ANodeJoinsPartitionsAnd(a string, b string, c string) {
//	panic("PENDING")
//}
//
//func (s *backgroundPeerDiscoverySuite) AllNodesInPartitionShouldDiscoverAllNodesInPartition(a string, b string) {
//	panic("PENDING")
//}

//func (s *backgroundPeerDiscoverySuite) TheNodeLeavesTheNetwork(a string) {
//	panic("PENDING")
//}

func (s *backgroundPeerDiscoverySuite) addBusMocks(busMocks map[string]*mock_modules.MockBus) {
	s.Helper()

	for serviceURL, busMock := range busMocks {
		require.Nilf(s, s.busMocks[serviceURL], "busMock for serviceURL %s already exists", serviceURL)
		s.busMocks[serviceURL] = busMock
	}
}

func (s *backgroundPeerDiscoverySuite) addP2PModules(p2pModules map[string]modules.P2PModule) {
	s.Helper()

	for serviceURL, p2pModule := range p2pModules {
		require.Nilf(s, s.p2pModules[serviceURL], "p2pModule for serviceURL %s already exists", serviceURL)
		s.p2pModules[serviceURL] = p2pModule
	}
}

func (s *backgroundPeerDiscoverySuite) getServiceURLsWithoutLabel(nodeLabel string) (serviceURLs []string) {
	for label, urls := range s.labelServiceURLMap {
		if label == nodeLabel {
			continue
		}
		serviceURLs = append(serviceURLs, urls...)
	}
	return serviceURLs
}
