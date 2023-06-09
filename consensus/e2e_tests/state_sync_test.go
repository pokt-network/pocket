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

func TestStateSync_MetadataRequestResponse_Success(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Choose node 1 as the server node
	serverNode := pocketNodes[1]
	serverNodePeerId := serverNode.GetBus().GetConsensusModule().GetNodeAddress()
	// Set server node's height to test height.
	serverNode.GetBus().GetConsensusModule().SetHeight(uint64(4))

	// Choose node 2 as the requester node.
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Prepare StateSyncMetadataRequest
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

	// Wait for response from the server node
	receivedMsgs, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "did not receive response to state sync metadata request", 1, 500, false)
	require.NoError(t, err)
	require.Len(t, receivedMsgs, 1)

	// Validate the response
	msg, err := codec.GetCodec().FromAny(receivedMsgs[0])
	require.NoError(t, err)

	stateSyncMetaDataResMsg, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	stateSyncMetaDataRes := stateSyncMetaDataResMsg.GetMetadataRes()
	require.NotEmpty(t, stateSyncMetaDataRes)

	require.Equal(t, uint64(3), stateSyncMetaDataRes.MaxHeight) // 3 because node sends the last persisted height
	require.Equal(t, uint64(1), stateSyncMetaDataRes.MinHeight)
	require.Equal(t, serverNodePeerId, stateSyncMetaDataRes.PeerAddress)
}

func TestStateSync_BlockRequestResponse_Success(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Choose node 1 as the server node
	serverNode := pocketNodes[1]
	// Set server node's height to test height.
	serverNode.GetBus().GetConsensusModule().SetHeight(uint64(5))

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Prepare GetBlockRequest
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
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "error waiting on response to a get block request", 1, 500, false)
	require.NoError(t, err)

	// validate the response
	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	require.NotEmpty(t, getBlockRes)

	require.Equal(t, uint64(1), getBlockRes.Block.GetBlockHeader().Height)
	// IMPROVE: What other data should we validate from the response?
}

func TestStateSync_BlockRequestResponse_FailNonExistingBlock(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Choose node 1 as the server node
	serverNode := pocketNodes[1]
	// Set server node's height to test height.
	serverNode.GetBus().GetConsensusModule().SetHeight(uint64(5))

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerAddress := requesterNode.GetBus().GetConsensusModule().GetNodeAddress()

	// Prepare a get block request for a non existing block (server is only at height 5)
	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: requesterNodePeerAddress,
				Height:      uint64(6),
			},
		},
	}
	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// Send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the get block request on server node, expect to return error
	errMsg := "expecting to time out waiting on a response from a non existent"
	_, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false)
	require.Error(t, err)
}

func TestStateSync_UnsyncedPeerSyncs_Success(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Select node 2 as the unsynched node that will catch up
	unsyncedNodeId := typesCons.NodeId(pocketNodes[2].GetBus().GetConsensusModule().GetNodeId())
	unsyncedNode := pocketNodes[unsyncedNodeId]

	// Set the unsynced node to height (2) and rest of the nodes to height (3)
	for id, pocketNode := range pocketNodes {
		var height uint64
		if id == unsyncedNodeId {
			height = uint64(2)
		} else {
			height = uint64(3)
		}
		pocketNode.GetBus().GetConsensusModule().SetHeight(height)
		pocketNode.GetBus().GetConsensusModule().SetStep(uint8(consensus.NewRound))
		pocketNode.GetBus().GetConsensusModule().SetRound(uint64(0))
	}

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// Assert that unsynced node has a different view of the network than the rest of the nodes
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 500, true)
	require.NoError(t, err)

	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == unsyncedNodeId {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: uint64(2),
					Step:   uint8(consensus.NewRound),
					Round:  uint8(1),
				},
				nodeState)
		} else {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: uint64(3),
					Step:   uint8(consensus.NewRound),
					Round:  uint8(1),
				},
				nodeState)
		}
		require.Equal(t, false, nodeState.IsLeader)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	metadataReceived := &typesCons.StateSyncMetadataResponse{
		PeerAddress: "unused_peer_addr_in_tests",
		MinHeight:   uint64(1),
		MaxHeight:   uint64(2), // 2 because unsynced node last persisted height 2
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

	waitForNodeToSync(t, clockMock, eventsChannel, unsyncedNode, pocketNodes, 3)
	require.NoError(t, err)
}

// TODO: Implement these tests

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

func prepareStateSyncTestEnvironment(t *testing.T) (*clock.Mock, modules.EventsChannel, idToNodeMapping) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := createTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	return clockMock, eventsChannel, pocketNodes
}
