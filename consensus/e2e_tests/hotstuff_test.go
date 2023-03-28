package e2e_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestHotstuff4Nodes1BlockHappyPath(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// Set starting height, round and step of the test.
	startingHeight := uint64(1)
	startingRound := uint64(0)
	startingStep := uint8(consensus.NewRound)

	// Get leaderId for the given height, round and step, by using the Consensus Modules' GetLeaderForView() function.
	// Any node in pocketNodes mapping can be used to gather leader election result via GetLeaderForView() function.
	leaderId := typesCons.NodeId(pocketNodes[1].GetBus().GetConsensusModule().GetLeaderForView(startingHeight, startingRound, startingStep))
	leader := pocketNodes[leaderId]

	// 1. NewRound
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 500, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: startingHeight,
				Step:   startingStep,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, false, nodeState.IsLeader)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 2. Prepare
	prepareProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Propose, numValidators, 500, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: startingHeight,
				Step:   startingStep + 1,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}

	for _, message := range prepareProposal {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 3. PreCommit
	prepareVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Vote, numValidators, 500, true)
	require.NoError(t, err)

	for _, vote := range prepareVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	preCommitProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Propose, numValidators, 500, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: startingHeight,
				Step:   startingStep + 2,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}

	for _, message := range preCommitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 4. Commit
	preCommitVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Vote, numValidators, 500, true)
	require.NoError(t, err)

	for _, vote := range preCommitVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	commitProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Commit, consensus.Propose, numValidators, 500, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: startingHeight,
				Step:   startingStep + 3,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}

	for _, message := range commitProposal {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 5. Decide
	commitVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Commit, consensus.Vote, numValidators, 500, true)
	require.NoError(t, err)

	for _, vote := range commitVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	decideProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Decide, consensus.Propose, numValidators, 500, true)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		// Leader has already committed the block and hence moved to the next height.
		if pocketId == leaderId {
			assertNodeConsensusView(t, pocketId,
				typesCons.ConsensusNodeState{
					Height: startingHeight + 1,
					Step:   startingStep,
					Round:  uint8(startingRound),
				},
				nodeState)
			require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId, "Leader should be empty")
			continue
		}
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: startingHeight,
				Step:   startingStep + 4,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}

	for _, message := range decideProposal {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 1. NewRound - begin again
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 500, true)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: startingHeight + 1,
				Step:   startingStep,
				Round:  uint8(startingRound),
			},
			nodeState)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId, "Leader should be empty")
	}

	// TODO(#615): Add QC verification here after valid block mocking is implemented with issue #352.

	// Test state synchronisation's get block functionality
	// At this stage, first round is finished, get block request for block height 1 must return non-nill block
	serverNode := pocketNodes[1]

	// We choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	stateSyncGetBlockReq := typesCons.GetBlockRequest{
		PeerAddress: requesterNodePeerAddress,
		Height:      1,
	}

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &stateSyncGetBlockReq,
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// Send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// Server node is waiting for the Get Block Request.
	numExpectedMsgs := 1
	errMsg := "StateSync Get Block Request Message"
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 500, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	require.NotEmpty(t, getBlockRes)

	require.Equal(t, uint64(1), getBlockRes.Block.GetBlockHeader().Height)
}

// TODO: Implement these tests and use them as a starting point for new ones. Consider using ChatGPT to help you out :)

// TODO(#615): Implement this test
func TestQuorumCertificate_Valid(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_InsufficientSignature(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromInvalidValidatorSet(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromJailedValidators(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromUnJailedValidators_Valid(t *testing.T) {
	// Unjailed validators should be able to sign a valid QC.
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_SignatureFromValidAndInvalidValidatorSet(t *testing.T) {
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_QuorumCertificateIsModified(t *testing.T) {
	// Leader modifies the QC after sending the proposal, therefore sent QC is invalid.
	t.Skip()
}

// TODO(#615): Implement this test
func TestQuorumCertificate_InvalidSignaturesFromValidValidatorSet(t *testing.T) {
	t.Skip()
}

// DISCUSS: This test scenario is currently more exploratory, and it may or may not be relevant.
func TestQuorumCertificate_ResistenceToSignatureMalleability(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes1Byzantine1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes2Byzantine1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes1BlockNetworkPartition(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes1Block4Rounds(t *testing.T) {
	t.Skip()
}
func TestHotstuff4Nodes2Blocks(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes2NewNodes1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff4Nodes2DroppedNodes1Block(t *testing.T) {
	t.Skip()
}

func TestHotstuff4NodesFailOnPrepare(t *testing.T) {
	t.Skip()
}

func TestHotstuff4NodesFailOnPrecommit(t *testing.T) {
	t.Skip()
}

func TestHotstuff4NodesFailOnCommit(t *testing.T) {
	t.Skip()
}

func TestHotstuff4NodesFailOnDecide(t *testing.T) {
	t.Skip()
}

func TestHotstuffValidatorWithLockedQC(t *testing.T) {
	t.Skip()
}

func TestHotstuffValidatorWithLockedQCMissingNewRoundMsg(t *testing.T) {
	t.Skip()
}
