package consensus_tests

import (
	"reflect"
	"testing"
	"time"

	"pocket/consensus/pkg/consensus"
	"pocket/consensus/pkg/pocket"
	"pocket/consensus/pkg/types"
	"pocket/shared"
	"pocket/shared/context"
	"pocket/shared/events"
	"pocket/shared/modules"
	"pocket/shared/typespb"

	"github.com/stretchr/testify/require"
)

func TestPaceMakerWithLockedQC(t *testing.T) {
	// TODO
}

func TestPaceMakerWithHighPrepareQC(t *testing.T) {
	// TODO
}

func TestPaceMakerNoQuorum(t *testing.T) {
	// TODO
}

func TestPaceMakerNotSafeProposal(t *testing.T) {
	// TODO
}

func TestPaceMakerExponentialTimeouts(t *testing.T) {
	// TODO
}

func TestPacemakerTimeouts(t *testing.T) {
	// Test configs.
	numNodes := 4
	paceMakerTimeout := 50 * time.Millisecond // Set a very small pacemaker timeout
	configs := GenerateNodeConfigs(numNodes)

	// Create the Pocket Nodes.
	testPocketBus := make(modules.PocketBus, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testPocketBus)

	// TODO: Should this be part of the configs?
	// Update the State Singleton.
	shared.GetPocketState().ConsensusParams.PaceMaker.TimeoutMSec = 50

	// Start test pocket nodes.
	ctx := context.EmptyPocketContext()
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start(ctx)
	}
	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	// Debug message to start consensus by triggering next view.
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// paceMakerTimeout
	WaitForNetworkConsensusMessage(t, testPocketBus, events.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// Check that a new round starts at the same height.
	WaitForNetworkConsensusMessage(t, testPocketBus, events.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(1), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// // Check that a new round starts at the same height.
	WaitForNetworkConsensusMessage(t, testPocketBus, events.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(2), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// Check that a new round starts at the same height.
	WaitForNetworkConsensusMessage(t, testPocketBus, events.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(3), nodeState.Round)
	}
}

func TestPaceMakerDifferentHeightsCatchup(t *testing.T) {
}

func TestPaceMakerDifferentStepsCatchup(t *testing.T) {
}

func TestPaceMakerCatchupSameStepDifferentRounds(t *testing.T) {
	numNodes := 4
	configs := GenerateNodeConfigs(numNodes)

	// Start test pocket nodes.
	testPocketBus := make(modules.PocketBus, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testPocketBus)
	ctx := context.EmptyPocketContext()
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start(ctx)
	}
	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	// Set all nodes to the same step and height but different rounds.
	testHeight := uint64(3)
	testStep := uint64(consensus.NewRound)
	leaderId := types.NodeId(1)
	leader := pocketNodes[leaderId]
	for _, pocketNode := range pocketNodes {
		consensusModImpl := getConsensusModImplementation(pocketNode)
		consensusModImpl.FieldByName("Height").SetUint(testHeight)
		consensusModImpl.FieldByName("Step").SetUint(testStep)
		// consensusModImpl.FieldByName("LeaderId").Set(reflect.ValueOf(nil)) // Leader is not set because the round update should set it appropriately.
	}

	header := &typespb.BlockHeaderConsTemp{
		Height: int64(testHeight),
		Hash:   "new_hash",

		LastBlockHash:   "prev_hash",
		ProposerAddress: []byte(leader.Address),
		ProposerId:      uint32(leaderId),
	}
	leaderBlock := &typespb.BlockConsTemp{
		BlockHeader:       header,
		Transactions:      make([]*typespb.Transaction, 0),
		ConsensusEvidence: make([]*typespb.Evidence, 0),
	}

	leaderConsensusMod := getConsensusModImplementation(leader)
	leaderConsensusMod.FieldByName("Block").Set(reflect.ValueOf(leaderBlock))

	leaderRound := uint64(6)
	// Set the leader to be in the highest round.
	getConsensusModImplementation(pocketNodes[leaderId]).FieldByName("Round").SetUint(leaderRound)
	getConsensusModImplementation(pocketNodes[2]).FieldByName("Round").SetUint(uint64(2))
	getConsensusModImplementation(pocketNodes[3]).FieldByName("Round").SetUint(uint64(3))
	getConsensusModImplementation(pocketNodes[4]).FieldByName("Round").SetUint(uint64(4))

	prepareProposeMessage := &consensus.HotstuffMessage{
		Type:      consensus.ProposeMessageType,
		Step:      consensus.Prepare,
		Height:    consensus.BlockHeight(testHeight),
		Round:     consensus.Round(leaderRound),
		Block:     leaderBlock,
		JustifyQC: nil,
		Sender:    leaderId,
	}
	P2PBroadcast(pocketNodes, prepareProposeMessage) // Broadcast the prepare proposal.

	// numNodes-1 because one of the messages is a self-proposal.
	WaitForNetworkConsensusMessage(t, testPocketBus, events.P2P_SEND_MESSAGE, consensus.Prepare, numNodes-1, 2000)
	leaderConsensusMod.FieldByName("LeaderId").Set(reflect.ValueOf(&leaderId))

	// Check that the leader is in the latest round.
	for _, pocketNode := range pocketNodes {
		nodeState := pocketNode.ConsensusMod.GetNodeState()
		require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
		require.Equal(t, uint64(3), nodeState.Height)
		require.Equal(t, uint8(6), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId)
	}
}

func getConsensusModImplementation(n *pocket.PocketNode) reflect.Value {
	return reflect.ValueOf(n.ConsensusMod).Elem()
}
