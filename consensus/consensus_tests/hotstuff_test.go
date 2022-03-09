package consensus_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/pokt-network/pocket/consensus"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestHotstuff4Nodes1BlockHappyPath(t *testing.T) {
	// Test configs
	numNodes := 4
	configs := GenerateNodeConfigs(t, numNodes)

	// Start test pocket nodes.
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testChannel)
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
	}
	time.Sleep(10 * time.Millisecond) // Avoids minor race conditions if pocket node has not finished starting/initialization

	// Debug message to start consensus by triggering first view change.
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// NewRound
	newRoundMessages := WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, false, nodeState.IsLeader)
	}

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}

	leaderId := types_consensus.NodeId(2)
	leader := pocketNodes[leaderId]

	// Prepare
	prepareProposal := WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Propose, 1, 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint8(consensus.Prepare), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range prepareProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Precommit
	prepareVotes := WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Vote, numNodes, 1000)
	for _, vote := range prepareVotes {
		P2PSend(t, leader, vote)
	}

	preCommitProposal := WaitForNetworkConsensusMessages(t, testChannel, consensus.PreCommit, consensus.Propose, 1, 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range preCommitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Commit
	preCommitVotes := WaitForNetworkConsensusMessages(t, testChannel, consensus.PreCommit, consensus.Vote, numNodes, 1000)
	for _, vote := range preCommitVotes {
		P2PSend(t, leader, vote)
	}

	commitProposal := WaitForNetworkConsensusMessages(t, testChannel, consensus.Commit, consensus.Propose, 1, 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint8(consensus.Commit), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range commitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Decide
	commitVotes := WaitForNetworkConsensusMessages(t, testChannel, consensus.Commit, consensus.Vote, numNodes, 1000)
	for _, vote := range commitVotes {
		P2PSend(t, leader, vote)
	}

	decideProposal := WaitForNetworkConsensusMessages(t, testChannel, consensus.Decide, consensus.Propose, 1, 1000)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		// Leader has already committed the block and hence moved to the next height.
		if pocketId == leaderId {
			require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
			require.Equal(t, uint64(2), nodeState.Height)
			require.Equal(t, uint8(0), nodeState.Round)
			require.Equal(t, nodeState.LeaderId, types_consensus.NodeId(0), "Leader should be empty")
			continue
		}
		require.Equal(t, uint8(consensus.Decide), nodeState.Step)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range decideProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	_ = WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 1000)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint64(2), nodeState.Height)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, nodeState.LeaderId, types_consensus.NodeId(0), "Leader should be empty")
	}
}

// func TestHotstuffSignatures(t *testing.T) {
// 	// Test configs.
// 	numNodes := 4
// 	configs := GenerateNodeConfigs(numNodes)

// 	// Start test pocket nodes.
// 	testPocketBus := make(modules.Bus, 100)
// 	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testPocketBus)
// 	ctx := context.EmptyPocketContext()
// 	for _, pocketNode := range pocketNodes {
// 		go pocketNode.Start(ctx)
// 	}
// 	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

// 	// Debug message to start consensus by triggering next view.
// 	for _, pocketNode := range pocketNodes {
// 		TriggerNextView(t, pocketNode)
// 	}

// 	// NewRound
// 	newRoundMessages := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
// 	for _, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
// 		require.Equal(t, uint64(1), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 	}
// 	for _, message := range newRoundMessages {
// 		P2PBroadcast(t, pocketNodes, message)
// 	}

// 	leaderId := s_types.NodeId(1)
// 	leader := pocketNodes[leaderId]

// 	// Prepare
// 	prepareProposal := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.Prepare, 1, 500)
// 	for _, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		require.Equal(t, uint8(consensus.Prepare), nodeState.Step)
// 		require.Equal(t, uint64(1), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 	}
// 	for _, message := range prepareProposal {
// 		P2PBroadcast(t, pocketNodes, message)
// 	}

// 	// Precommit
// 	prepareVotes := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_SEND_MESSAGE, consensus.Prepare, numNodes, 500)
// 	for _, vote := range prepareVotes {
// 		P2PSend(t, leader, vote)
// 	}

// 	preCommitProposal := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.PreCommit, 1, 500)
// 	for _, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
// 		require.Equal(t, uint64(1), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
// 	}
// 	for _, message := range preCommitProposal {
// 		P2PBroadcast(t, pocketNodes, message)
// 	}

// 	// Commit
// 	preCommitVotes := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_SEND_MESSAGE, consensus.PreCommit, numNodes, 500)
// 	for _, vote := range preCommitVotes {
// 		P2PSend(t, leader, vote)
// 	}

// 	commitProposal := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.Commit, 1, 500)
// 	for _, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		require.Equal(t, uint8(consensus.Commit), nodeState.Step)
// 		require.Equal(t, uint64(1), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
// 	}
// 	for _, message := range commitProposal {
// 		P2PBroadcast(t, pocketNodes, message)
// 	}

// 	// Decide
// 	commitVotes := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_SEND_MESSAGE, consensus.Commit, numNodes, 500)
// 	for _, vote := range commitVotes {
// 		P2PSend(t, leader, vote)
// 	}

// 	decideProposal := WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.Decide, 1, 500)
// 	for pocketId, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		// Leader has already committed the block and hence moved to the next height.
// 		if pocketId == leaderId {
// 			require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
// 			require.Equal(t, uint64(2), nodeState.Height)
// 			require.Equal(t, uint8(0), nodeState.Round)
// 			require.Equal(t, nodeState.LeaderId, s_types.NodeId(0), "Leader should be 0 - no one is the leader.")
// 			continue
// 		}
// 		require.Equal(t, uint8(consensus.Decide), nodeState.Step)
// 		require.Equal(t, uint64(1), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
// 	}
// 	for _, message := range decideProposal {
// 		P2PBroadcast(t, pocketNodes, message)
// 	}

// 	// Block has been committed and new round has begun.
// 	WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_BROADCAST_MESSAGE, consensus.NewRound, numNodes, 500)
// 	for _, pocketNode := range pocketNodes {
// 		nodeState := GetConsensusNodeState(pocketNode)
// 		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
// 		require.Equal(t, uint64(2), nodeState.Height)
// 		require.Equal(t, uint8(0), nodeState.Round)
// 		require.Equal(t, nodeState.LeaderId, s_types.NodeId(0), "Leader should be 0 - no one is the leader.")
// 	}

// }

// func TestHotstuff4Nodes1Byzantine1Block(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4Nodes2Byzantine1Block(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4Nodes1BlockNetworkPartition(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4Nodes1Block4Rounds(t *testing.T) {
// 	// TODO
// }
// func TestHotstuff4Nodes2Blocks(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4Nodes2NewNodes1Block(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4Nodes2DroppedNodes1Block(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4NodesFailOnPrepare(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4NodesFailOnPrecommit(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4NodesFailOnCommit(t *testing.T) {
// 	// TODO
// }

// func TestHotstuff4NodesFailOnDecide(t *testing.T) {
// 	// TODO
// }

// func TestHotstuffAvailableQCFromLockedValidator(t *testing.T) {
// 	// TODO
// }

// func TestHotstuffMissingNewRoundMsgFromLockedValidator(t *testing.T) {
// 	// TODO
// }
