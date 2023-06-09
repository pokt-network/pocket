package e2e_tests

import (
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
	send(t, serverNode, anyProto)

	// Wait for response from the server node
	receivedMsgs, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "did not receive response to state sync metadata request", 1, 500, false)
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
	stateSyncGetBlockMsg := prepareStateSyncGetBlockMessage(t, requesterNodePeerAddress, 1)

	// Send get block request to the server node
	send(t, serverNode, stateSyncGetBlockMsg)

	// Start waiting for the get block request on server node, expect to return error
	receivedMsg, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "error waiting on response to a get block request", 1, 500, false)
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
	stateSyncGetBlockMsg := prepareStateSyncGetBlockMessage(t, requesterNodePeerAddress, 6)

	// Send get block request to the server node
	send(t, serverNode, stateSyncGetBlockMsg)

	// Start waiting for the get block request on server node, expect to return error
	errMsg := "expecting to time out waiting on a response from a non existent"
	_, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false)
	require.Error(t, err)
}

func TestStateSync_UnsyncedPeerSyncs_Success(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Select node 2 as the unsynched node that will catch up
	unsyncedNodeId := typesCons.NodeId(pocketNodes[2].GetBus().GetConsensusModule().GetNodeId())
	unsyncedNode := pocketNodes[unsyncedNodeId]

	// Set the unsynced node to height (2) and rest of the nodes to height (4)
	for id, pocketNode := range pocketNodes {
		var height uint64
		if id == unsyncedNodeId {
			height = uint64(2)
		} else {
			height = uint64(4)
		}
		pocketNode.GetBus().GetConsensusModule().SetHeight(height)
		pocketNode.GetBus().GetConsensusModule().SetStep(uint8(consensus.NewRound))
		pocketNode.GetBus().GetConsensusModule().SetRound(uint64(0))
	}

	// Trigger all the nodes to the next step
	triggerNextView(t, pocketNodes)
	_, err := waitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 500, true)
	require.NoError(t, err)

	// Verify the unsynched node is still behind after NewRound starts
	for nodeId, pocketNode := range pocketNodes {
		nodeState := getConsensusNodeState(pocketNode)
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
					Height: uint64(4),
					Step:   uint8(consensus.NewRound),
					Round:  uint8(1),
				},
				nodeState)
		}
		require.Equal(t, false, nodeState.IsLeader)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	// Wait for unsyncedNode to go from height 2 to height 4
	assertHeight(t, unsyncedNodeId, uint64(2), getConsensusNodeState(unsyncedNode).Height)
	waitForNodeToSync(t, clockMock, eventsChannel, unsyncedNode, pocketNodes, 3)
	assertHeight(t, unsyncedNodeId, uint64(3), getConsensusNodeState(unsyncedNode).Height)
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
	runtimeMgrs := generateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := generateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := createTestConsensusPocketNodes(t, buses, eventsChannel)
	err := startAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	return clockMock, eventsChannel, pocketNodes
}
