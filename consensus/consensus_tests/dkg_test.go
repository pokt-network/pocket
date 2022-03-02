package consensus_tests

import (
	"testing"
	"time"

	"github.com/pokt-network/pocket/consensus/dkg"
	"github.com/pokt-network/pocket/shared/types"

	"github.com/pokt-network/pocket/shared/modules"

	"github.com/stretchr/testify/require"
)

func TestDistributedKeyGenerationCrypto(t *testing.T) {
	// Test configs.
	numNodes := 4
	configs := GenerateNodeConfigs(numNodes)

	// Start test pocket nodes.

	testPocketBus := make(modules.PocketBus, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testPocketBus)
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
	}
	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	for _, pocketNode := range pocketNodes {
		TriggerDKG(t, pocketNode)
	}

	rnd1Bcast := WaitFoNetworkDKGMessages(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, dkg.DKGRound2, numNodes, 1000)
	rnd1P2p := WaitFoNetworkDKGMessages(t, testPocketBus, types.P2P_SEND_MESSAGE, dkg.DKGRound2, numNodes*(numNodes-1), 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint64(0), nodeState.Height)
	}
	for _, message := range rnd1Bcast {
		require.Nil(t, message.Recipient, "DKG broadcast message must not have a recipient.")
		P2PBroadcast(pocketNodes, message)
	}
	for _, message := range rnd1P2p {
		require.NotNil(t, message.Recipient, "DKG send message must have a recipient.")
		P2PSend(pocketNodes[*message.Recipient], message)
	}

	rnd2Bcast := WaitFoNetworkDKGMessages(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, dkg.DKGRound3, numNodes, 500)
	for _, message := range rnd2Bcast {
		require.Nil(t, message.Recipient, "DKG broadcast message must not have a recipient.")
		P2PBroadcast(pocketNodes, message)
	}

	rnd3Bcast := WaitFoNetworkDKGMessages(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, dkg.DKGRound4, numNodes, 500)
	for _, message := range rnd3Bcast {
		require.Nil(t, message.Recipient, "DKG broadcast message must not have a recipient.")
		P2PBroadcast(pocketNodes, message)
	}

	// WaitFoNetworkDKGMessages(t, testPocketBus, events.P2P_BROADCAST_MESSAGE, consensus.DKGRound4, numNodes, 500)
	// TODO: Add signing and verification here too.
}
