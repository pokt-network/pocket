package integration

import (
	"github.com/pokt-network/pocket/shared/modules"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/pokt-network/pocket/internal/testutil"
)

const (
	peerDiscoveryFeaturePath = "peer_discovery.feature"
	bootstrapNodePrefix      = "bootstrap"
)

var (
	testNodes map[string]modules.P2PModule
)

func TestPeerDiscoveryIntegration(t *testing.T) {
	t.Parallel()

	testutil.RunGherkinFeature(t, peerDiscoveryFeaturePath, initPeerDiscoveryScenarios)
}

func aNode(nodeName string) error {
	if strings.HasPrefix(bootstrapNodePrefix, nodeName) {
		// --
	}

	return godog.ErrPending
}

func allNodesInPartitionShouldDiscoverAllNodesInPartition(partitionA, partitionB string) error {
	return godog.ErrPending
}

func eachNodeShouldHaveNumberOfPeersInTheirRespectivePeerstores(expectedPeerCount int) error {
	return godog.ErrPending
}

func eachNodeShouldNotHaveAnyLeaversInTheirPeerstores() error {
	return godog.ErrPending
}

func otherNodesShouldHaveNumberOfPeersInTheirPeerstores(arg1 int) error {
	return godog.ErrPending
}

func otherNodesShouldNotBeIncludedInTheirRespectivePeerstores() error {
	return godog.ErrPending
}

func theNetworkShouldContainNumberOfNodes(arg1 int) error {
	return godog.ErrPending
}

func theNodeShouldHaveNumberOfPeersInItsPeerstore(arg1 string, arg2 int) error {
	return godog.ErrPending
}

func theNodeShouldNotBeIncludedInItsOwnPeerstore(arg1 string) error {
	return godog.ErrPending
}

func initPeerDiscoveryScenarios(ctx *godog.ScenarioContext) {
	ctx.Step(`^a "([^"]*)" node$`, aNode)
	ctx.Step(`^a "([^"]*)" node in partition "([^"]*)"$`, aNodeInPartition)
	ctx.Step(`^a "([^"]*)" node joins partitions "([^"]*)" and "([^"]*)"$`, aNodeJoinsPartitionsAnd)
	ctx.Step(`^all nodes in partition "([^"]*)" should discover all nodes in partition "([^"]*)"$`, allNodesInPartitionShouldDiscoverAllNodesInPartition)
	ctx.Step(`^each node should have (\d+) number of peers in their respective peerstores$`, eachNodeShouldHaveNumberOfPeersInTheirRespectivePeerstores)
	ctx.Step(`^each node should not have any leavers in their peerstores$`, eachNodeShouldNotHaveAnyLeaversInTheirPeerstores)
	ctx.Step(`^(\d+) number of nodes bootstrap in partition "([^"]*)"$`, numberOfNodesBootstrapInPartition)
	ctx.Step(`^(\d+) number of nodes join the network$`, numberOfNodesJoinTheNetwork)
	ctx.Step(`^(\d+) number of nodes leave the network$`, numberOfNodesLeaveTheNetwork)
	ctx.Step(`^other nodes should have (\d+) number of peers in their peerstores$`, otherNodesShouldHaveNumberOfPeersInTheirPeerstores)
	ctx.Step(`^other nodes should not be included in their respective peerstores$`, otherNodesShouldNotBeIncludedInTheirRespectivePeerstores)
	ctx.Step(`^the network should contain (\d+) number of nodes$`, theNetworkShouldContainNumberOfNodes)
	ctx.Step(`^the "([^"]*)" node leaves the network$`, theNodeLeavesTheNetwork)
	ctx.Step(`^the "([^"]*)" node should have (\d+) number of peers in its peerstore$`, theNodeShouldHaveNumberOfPeersInItsPeerstore)
	ctx.Step(`^the "([^"]*)" node should not be included in its own peerstore$`, theNodeShouldNotBeIncludedInItsOwnPeerstore)
}
