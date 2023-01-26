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

func TestStateSyncServerGetMetaDataReq(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	testHeight := uint64(4)
	testStep := uint8(consensus.NewRound)
	testRound := uint64(0)

	leaderId := typesCons.NodeId(2)
	for _, pocketNode := range pocketNodes {
		// Update height, step, leaderId, and utility context via setters exposed with the debug interface
		consensusModImpl := GetConsensusModImpl(pocketNode)
		consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
		consensusModImpl.MethodByName("SetStep").Call([]reflect.Value{reflect.ValueOf(testStep)})
		consensusModImpl.MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(testRound)})
		consensusModImpl.MethodByName("SetLeaderId").Call([]reflect.Value{reflect.Zero(reflect.TypeOf(&leaderId))})

		// utilityContext is only set on new rounds, which is skipped in this test
		utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(testHeight))
		require.NoError(t, err)
		consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
	}

	for pocketId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 4,
				Step:   uint8(consensus.NewRound),
				Round:  0,
			},
			nodeState)
		require.Equal(t, nodeState.LeaderId, typesCons.NodeId(0), "Leader should be empty")
	}

	// Choose node 1 as the server node.
	// We first enable server mode of the node 1.
	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	// We choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerId, err := requesterNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)

	// Test MetaData Req
	stateSyncMetaDataReq := typesCons.StateSyncMetadataRequest{
		PeerId: requesterNodePeerId,
	}

	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST,
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &stateSyncMetaDataReq,
		},
	}
	anyProto, err := anypb.New(stateSyncMetaDataReqMessage)
	require.NoError(t, err)

	// send metadata request to the server node
	P2PSend(t, serverNode, anyProto)

	// start waiting for the metadata request on server node,
	numExpectedMsgs := 1
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE, requesterNode.GetP2PAddress(), numExpectedMsgs, 250, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncMetaDataResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	metaDataRes := stateSyncMetaDataResMessage.GetMetadataRes()
	require.NotEmpty(t, metaDataRes)

	require.Equal(t, uint64(4), metaDataRes.MaxHeight)
}

// TODO!
// Remove unneccsaary block generation
// Write test considering heights: 1) with valid block height (response is non-empty) , 2) with invalid block height (response is empty)
// Add get height check to the pacemaker_test as well.
func TestStateSyncServerGetBlock(t *testing.T) {
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

	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	// We choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerId, err := requesterNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)

	// Test GetBlock Req
	stateSyncGetBlockReq := typesCons.GetBlockRequest{
		PeerId: requesterNodePeerId,
		Height: 1,
	}

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_REQUEST,
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &stateSyncGetBlockReq,
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// start waiting for the get block request on server node,
	numExpectedMsgs := 1
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_RESPONSE, requesterNode.GetP2PAddress(), numExpectedMsgs, 250, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	require.NotEmpty(t, getBlockRes)

	fmt.Printf("Get Block Response: %s", getBlockRes)
	//require.Equal(t, uint64(1), getBlockRes.Block.GetBlockHeader().Height)
}
