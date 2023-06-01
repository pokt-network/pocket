//go:build integration

package integration

import (
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	p2p_testutil "github.com/pokt-network/pocket/internal/testutil/p2p"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"

	"github.com/regen-network/gocuke"
)

const backgroundGossipFeaturePath = "background_gossip.feature"

func TestMinimal(t *testing.T) {
	// a new step definition suite is constructed for every scenario
	gocuke.NewRunner(t, &suite{}).Path(backgroundGossipFeaturePath).Run()
}

type suite struct {
	// special arguments like TestingT are injected automatically into exported fields
	gocuke.TestingT

	// seenMessages is used as a map to track which messages have been seen
	// by which nodes
	seenMessages map[libp2pPeer.ID]struct{}
	p2pModules   []modules.P2PModule
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

func (s *suite) AFullyConnectedNetworkOfPeers(peerCount int64) {
	// setup mock network
	s.p2pModules = p2p_testutil.NewP2PModules(s, int(peerCount))
}

func (s *suite) ANodeBroadcastsATestMessageViaItsBackgroundRouter() {
	// select arbitrary sender & store in context for reference later
	sender := s.p2pModules[0]

	// broadcast a test message
	msg := &anypb.Any{}
	err := sender.Broadcast(msg)
	require.NoError(s, err)
}

func (s *suite) NumberOfNodesShouldReceiveTheTestMessage(receivedCount int64) {
	done := make(chan struct{}, 1)

	go func() {
		s.wg.Wait()
		require.Len(s.seenMessages, receivedCount)
		done <- struct{}{}
	}()

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
