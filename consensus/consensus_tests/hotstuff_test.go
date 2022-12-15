package consensus_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestHotstuff4Nodes1BlockHappyPath(t *testing.T) {
	clockMock := clock.NewMock()
	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)

	go timeReminder(clockMock, 100*time.Millisecond)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, runtimeMgrs, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	// NewRound
	newRoundMessages, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numValidators, 1000)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.NewRound),
				Round:  0,
			},
			nodeState)
		require.Equal(t, false, nodeState.IsLeader)
	}
	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Leader election is deterministic for now, so we know its NodeId
	// TODO(olshansky): Use seeding for deterministic leader election in unit tests.
	leaderId := typesCons.NodeId(2)
	leader := pocketNodes[leaderId]

	advanceTime(clockMock, 10*time.Millisecond)

	// Prepare
	prepareProposal, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Prepare, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.Prepare),
				Round:  0,
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range prepareProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	// Precommit
	prepareVotes, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Prepare, consensus.Vote, numValidators, 1000)
	require.NoError(t, err)
	for _, vote := range prepareVotes {
		P2PSend(t, leader, vote)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	preCommitProposal, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.PreCommit, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.PreCommit),
				Round:  0,
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range preCommitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	// Commit
	preCommitVotes, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.PreCommit, consensus.Vote, numValidators, 1000)
	require.NoError(t, err)
	for _, vote := range preCommitVotes {
		P2PSend(t, leader, vote)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	commitProposal, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Commit, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.Commit),
				Round:  0,
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range commitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	// Decide
	commitVotes, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Commit, consensus.Vote, numValidators, 1000)
	require.NoError(t, err)
	for _, vote := range commitVotes {
		P2PSend(t, leader, vote)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	decideProposal, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Decide, consensus.Propose, 1, 1000)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		// Leader has already committed the block and hence moved to the next height.
		if pocketId == leaderId {
			assertNodeConsensusView(t, pocketId,
				typesCons.ConsensusNodeState{
					Height: 2,
					Step:   uint8(consensus.NewRound),
					Round:  0,
				},
				nodeState)
			require.Equal(t, nodeState.LeaderId, typesCons.NodeId(0), "Leader should be empty")
			continue
		}
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.Decide),
				Round:  0,
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	for _, message := range decideProposal {
		P2PBroadcast(t, pocketNodes, message)
	}

	advanceTime(clockMock, 10*time.Millisecond)

	// Block has been committed and new round has begun
	_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numValidators, 1000)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 2,
				Step:   uint8(consensus.NewRound),
				Round:  0,
			},
			nodeState)
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
