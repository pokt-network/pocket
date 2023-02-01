package e2e_tests

import (
	"fmt"
	"reflect"
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

	// 1. NewRound
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 250, true)
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
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// IMPROVE: Use seeding for deterministic leader election in unit tests.
	// Leader election is deterministic for now, so we know its NodeId
	leaderId := typesCons.NodeId(2)
	leader := pocketNodes[leaderId]

	// 2. Prepare
	prepareProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Propose, numValidators, 250, true)
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
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 3. PreCommit
	prepareVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Vote, numValidators, 250, true)
	require.NoError(t, err)

	for _, vote := range prepareVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	preCommitProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Propose, numValidators, 250, true)
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
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 4. Commit
	preCommitVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Vote, numValidators, 250, true)
	require.NoError(t, err)

	for _, vote := range preCommitVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	commitProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Commit, consensus.Propose, numValidators, 250, true)
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
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 5. Decide
	commitVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Commit, consensus.Vote, numValidators, 250, true)
	require.NoError(t, err)

	for _, vote := range commitVotes {
		P2PSend(t, leader, vote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	decideProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Decide, consensus.Propose, numValidators, 250, true)
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
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 1. NewRound - begin again
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 250, true)
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

	// Test state synchronisation's get block functionality
	// At this stage, the chain has finished the first round
	// So, get block request for block height 1 must return non-nill block
	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	// We choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerId, err := requesterNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)

	stateSyncGetBlockReq := typesCons.GetBlockRequest{
		PeerId: requesterNodePeerId,
		Height: 1,
	}

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		//MsgType: typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_REQUEST,
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &stateSyncGetBlockReq,
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// Send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the get block request on server node,
	numExpectedMsgs := 1
	errMsg := "StateSync Get Block Request Message"
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
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
