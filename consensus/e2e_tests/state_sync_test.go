package e2e_tests

import (
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

func TestStateSync_ServerGetMetaDataReq_Success(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	testHeight := uint64(4)

	// Choose node 1 as the server node
	// Set server node's height to test height.
	serverNode := pocketNodes[1]
	serverNodePeerId := serverNode.GetBus().GetConsensusModule().GetNodeAddress()
	serverNode.GetBus().GetConsensusModule().SetHeight(testHeight)

	// Choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Test MetaData Req
	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: requesterNodePeerAddress,
			},
		},
	}
	anyProto, err := anypb.New(stateSyncMetaDataReqMessage)
	require.NoError(t, err)

	// Send metadata request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the metadata request on server node,
	errMsg := "StateSync Metadata Request"
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncMetaDataResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	metaDataRes := stateSyncMetaDataResMessage.GetMetadataRes()
	require.NotEmpty(t, metaDataRes)

	require.Equal(t, uint64(4), metaDataRes.MaxHeight)
	require.Equal(t, uint64(1), metaDataRes.MinHeight)
	require.Equal(t, serverNodePeerId, metaDataRes.PeerAddress)
}

func TestStateSync_ServerGetBlock_Success(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	testHeight := uint64(5)
	serverNode := pocketNodes[1]
	serverNode.GetBus().GetConsensusModule().SetHeight(testHeight)

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Passing Test
	// Test GetBlock Req
	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: requesterNodePeerAddress,
				Height:      1,
			},
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// Send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the get block request on server node, expect to return error
	errMsg := "StateSync Get Block Request Message"
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	require.NotEmpty(t, getBlockRes)

	require.Equal(t, uint64(1), getBlockRes.Block.GetBlockHeader().Height)
}

func TestStateSync_ServerGetBlock_FailNonExistingBlock(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	testHeight := uint64(5)

	serverNode := pocketNodes[1]
	serverNode.GetBus().GetConsensusModule().SetHeight(testHeight)

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Failing Test
	// Get Block Req is current block height + 1
	requestHeight := testHeight + 1
	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: requesterNodePeerAddress,
				Height:      requestHeight,
			},
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// Send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the get block request on server node, expect to return error
	errMsg := "StateSync Get Block Request Message"
	_, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false)
	require.Error(t, err)
}

func TestStateSync_UnsyncedPeerSyncs_Success(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)

	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	// Prepare leader info
	testHeight := uint64(3)
	testRound := uint64(0)
	testStep := uint8(consensus.NewRound)

	// Prepare unsynced node info
	unsyncedNode := pocketNodes[2]
	unsyncedNodeId := typesCons.NodeId(2)
	unsyncedNodeHeight := uint64(2)

	// Set the unsynced node to height (2) and rest of the nodes to height (3)
	for id, pocketNode := range pocketNodes {
		if id == unsyncedNodeId {
			pocketNode.GetBus().GetConsensusModule().SetHeight(unsyncedNodeHeight)
		} else {
			pocketNode.GetBus().GetConsensusModule().SetHeight(testHeight)
		}
		pocketNode.GetBus().GetConsensusModule().SetStep(testStep)
		pocketNode.GetBus().GetConsensusModule().SetRound(testRound)

		utilityUnitOfWork, err := pocketNode.GetBus().GetUtilityModule().NewUnitOfWork(int64(testHeight))
		require.NoError(t, err)
		pocketNode.GetBus().GetConsensusModule().SetUtilityUnitOfWork(utilityUnitOfWork)
	}

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}
	currentRound := testRound + 1

	// Get leaderId for the given height, round and step, by using the Consensus Modules' GetLeaderForView() function.
	// Any node in pocketNodes mapping can be used to call GetLeaderForView() function.
	leaderId := typesCons.NodeId(pocketNodes[1].GetBus().GetConsensusModule().GetLeaderForView(testHeight, currentRound, testStep))
	leader := pocketNodes[leaderId]
	leaderPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	block := generatePlaceholderBlock(testHeight, leaderPK.Address())
	leader.GetBus().GetConsensusModule().SetBlock(block)

	// Assert that unsynced node has a different view of the network than the rest of the nodes
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 500, true)
	require.NoError(t, err)

	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == unsyncedNodeId {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: unsyncedNodeHeight,
					Step:   testStep,
					Round:  uint8(currentRound),
				},
				nodeState)
		} else {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: testHeight,
					Step:   testStep,
					Round:  uint8(currentRound),
				},
				nodeState)
		}
		require.Equal(t, false, nodeState.IsLeader)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	metadataReceived := &typesCons.StateSyncMetadataResponse{
		PeerAddress: "unused_peer_addr_in_tests",
		MinHeight:   uint64(1),
		MaxHeight:   testHeight,
	}

	// Simulate state sync metadata response by pushing metadata to the unsynced node's consensus module
	consensusModImpl := GetConsensusModImpl(unsyncedNode)
	consensusModImpl.MethodByName("PushStateSyncMetadataResponse").Call([]reflect.Value{reflect.ValueOf(metadataReceived)})

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 2. Propose
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Propose, numValidators, 500, true)
	require.NoError(t, err)

	// TODO(#352): This function will be updated once state sync implementation is complete
	err = WaitForNodeToSync(t, clockMock, eventsChannel, unsyncedNode, pocketNodes, testHeight)
	require.NoError(t, err)

	// TODO(#352): Add height check once state sync implmentation is complete
}

// TODO(#352): Implement these tests

func TestStateSync_UnsyncedPeerSyncsABlock_Success(t *testing.T) {
	t.Skip()
}

func TestStateSync_UnsyncedPeerSyncsMultipleBlocksInOrder_Success(t *testing.T) {
	t.Skip()
}

func TestStateSync_UnsyncedPeerSyncsMultipleUnorderedBlocks_Success(t *testing.T) {
	t.Skip()
}

func TestStateSync_TemporarilyOfflineValidatorCatchesUp(t *testing.T) {
	t.Skip()
}

func TestStateSync_4of10UnsyncedPeersCatchUp(t *testing.T) {
	t.Skip()
}

func TestStateSync_9of10UnsyncedPeersCatchUp(t *testing.T) {
	t.Skip()
}
