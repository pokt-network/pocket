package consensus_tests

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestHotstuff4Nodes1BlockHappyPath(t *testing.T) {
	// Test configs
	numNodes := 4
	configs, genesisStates := GenerateNodeConfigs(t, numNodes)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, genesisStates, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// NewRound
	newRoundMessages, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 1000)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, false, nodeState.IsLeader)
	}
	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Leader election is deterministic for now, so we know its NodeId
	// TODO(olshansky): Use seeding for deterministic leader election in unit tests.
	leaderId := typesCons.NodeId(2)
	leader := pocketNodes[leaderId]

	// Prepare
	prepareProposal, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.Prepare), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range prepareProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Precommit
	prepareVotes, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Vote, numNodes, 1000)
	require.NoError(t, err)
	for _, vote := range prepareVotes {
		P2PSend(t, leader, vote)
	}

	preCommitProposal, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.PreCommit, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range preCommitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Commit
	preCommitVotes, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.PreCommit, consensus.Vote, numNodes, 1000)
	require.NoError(t, err)
	for _, vote := range preCommitVotes {
		P2PSend(t, leader, vote)
	}

	commitProposal, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.Commit, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.Commit), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range commitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Decide
	commitVotes, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.Commit, consensus.Vote, numNodes, 1000)
	require.NoError(t, err)
	for _, vote := range commitVotes {
		P2PSend(t, leader, vote)
	}

	decideProposal, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.Decide, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		// Leader has already committed the block and hence moved to the next height.
		if pocketId == leaderId {
			require.Equal(t, uint64(2), nodeState.Height)
			require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
			require.Equal(t, uint8(0), nodeState.Round)
			require.Equal(t, nodeState.LeaderId, typesCons.NodeId(0), "Leader should be empty")
			continue
		}
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.Decide), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range decideProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Block has been committed and new round has begun
	_, err = WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 1000)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(2), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
		require.Equal(t, nodeState.LeaderId, typesCons.NodeId(0), "Leader should be empty")
	}
}

/*
func TestHotstuff4Nodes1Byzantine1Block(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4Nodes2Byzantine1Block(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4Nodes1BlockNetworkPartition(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4Nodes1Block4Rounds(t *testing.T) {
	t.Skip() // TODO: Implement
}
func TestHotstuff4Nodes2Blocks(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4Nodes2NewNodes1Block(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4Nodes2DroppedNodes1Block(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4NodesFailOnPrepare(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4NodesFailOnPrecommit(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4NodesFailOnCommit(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuff4NodesFailOnDecide(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuffValidatorWithLockedQC(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestHotstuffValidatorWithLockedQCMissingNewRoundMsg(t *testing.T) {
	t.Skip() // TODO: Implement
}
*/
