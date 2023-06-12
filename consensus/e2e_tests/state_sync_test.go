package e2e_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
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

	// Send metadata request to the server node
	anyProto := prepareStateSyncGetMetadataMessage(t, requesterNodePeerAddress)
	send(t, serverNode, anyProto)

	// Wait for response from the server node
	receivedMsgs, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "did not receive response to state sync metadata request", 1, 500, false, nil)
	require.NoError(t, err)

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
	receivedMsg, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "error waiting on response to a get block request", 1, 500, false, nil)
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
	_, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 500, false, nil)
	require.Error(t, err)
}

func TestStateSync_UnsyncedPeerSyncs_Success(t *testing.T) {
	clockMock, eventsChannel, pocketNodes := prepareStateSyncTestEnvironment(t)

	// Select node 2 as the unsynched node that will catch up
	unsyncedNodeId := typesCons.NodeId(pocketNodes[2].GetBus().GetConsensusModule().GetNodeId())
	unsyncedNode := pocketNodes[unsyncedNodeId]
	unsyncedNodeHeight := uint64(2)
	targetHeight := uint64(5)

	// Set the unsynced node to height (2) and rest of the nodes to height (4)
	for id, pocketNode := range pocketNodes {
		var height uint64
		if id == unsyncedNodeId {
			height = unsyncedNodeHeight
		} else {
			height = targetHeight
		}
		pocketNode.GetBus().GetConsensusModule().SetHeight(height)
		pocketNode.GetBus().GetConsensusModule().SetStep(uint8(consensus.NewRound))
		pocketNode.GetBus().GetConsensusModule().SetRound(uint64(0))
	}

	// Sanity check unsynched node is at height 2
	assertHeight(t, unsyncedNodeId, uint64(2), getConsensusNodeState(unsyncedNode).Height)

	// Broadcast metadata to all the others nodes so the node that's behind has a view of the network
	anyProto := prepareStateSyncGetMetadataMessage(t, unsyncedNode.GetBus().GetConsensusModule().GetNodeAddress())
	broadcast(t, pocketNodes, anyProto)

	// Make sure the unsynched node has a view of the network
	receivedMsgs, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "did not receive response to state sync metadata request", len(pocketNodes), 500, false, nil)
	require.NoError(t, err)
	for _, msg := range receivedMsgs {
		send(t, unsyncedNode, msg)
	}
	// IMPROVE: Look into ways to assert on unsynched.MinHeightViewOfNetwork and unsynched.MaxHeightViewOfNetwork

	// Trigger the next round of consensus so the unsynched nodes is prompted to start synching
	triggerNextView(t, pocketNodes)
	advanceTime(t, clockMock, 10*time.Millisecond)
	proposalMsgs, err := waitForNetworkConsensusEvents(t, clockMock, eventsChannel, typesCons.HotstuffStep(consensus.NewRound), consensus.Propose, numValidators*numValidators, 500, false)
	require.NoError(t, err)
	broadcastMessages(t, proposalMsgs, pocketNodes)
	advanceTime(t, clockMock, 10*time.Millisecond)

	isGetBlockRequest := func(msg *typesCons.StateSyncMessage) bool {
		return msg.GetGetBlockReq() != nil
	}
	isGetBlockResponse := func(msg *typesCons.StateSyncMessage) bool {
		return msg.GetGetBlockRes() != nil
	}

	for unsyncedNodeHeight < targetHeight {
		// Wait for the unsynched node to request the block at the current height
		blockRequests, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "error waiting on response to a get block request", 1, 5000, false, &isGetBlockRequest)
		require.NoError(t, err)

		// Validate the height being requested is correct
		msg, err := codec.GetCodec().FromAny(blockRequests[0])
		require.NoError(t, err)
		heightRequested := msg.(*typesCons.StateSyncMessage).GetGetBlockReq().Height
		require.Equal(t, unsyncedNodeHeight, heightRequested)

		// Broadcast the block request to all nodes
		broadcast(t, pocketNodes, blockRequests[0])
		advanceTime(t, clockMock, 10*time.Millisecond)

		// Wait for the unsynched node to receive the block responses
		blockResponses, err := waitForNetworkStateSyncEvents(t, clockMock, eventsChannel, "error waiting on response to a get block response", numValidators-1, 5000, false, &isGetBlockResponse)
		require.NoError(t, err)

		// Validate that the block is the same from all the validators who send it
		var blockResponse *typesCons.GetBlockResponse
		for _, msg := range blockResponses {
			msgAny, err := codec.GetCodec().FromAny(msg)
			require.NoError(t, err)

			stateSyncMessage, ok := msgAny.(*typesCons.StateSyncMessage)
			require.True(t, ok)

			if blockResponse == nil {
				blockResponse = stateSyncMessage.GetGetBlockRes()
				continue
			}
			require.Equal(t, blockResponse.Block, stateSyncMessage.GetGetBlockRes().Block)
		}

		// Send one of the responses (since they are equal) to the unsynched node to apply it
		send(t, unsyncedNode, blockResponses[0])
		advanceTime(t, clockMock, 10*time.Millisecond)

		fmt.Println("OLSH events channel", eventsChannel)
		// Wait for the unsynched node to commit the block
		_, err = waitForEventsInternal(clockMock, eventsChannel, messaging.StateSyncBlockCommittedEventType, 1, 5000, nil, "error waiting on response to a get block response", false)
		require.NoError(t, err)

		// ensure unsynced node height increased
		nodeState := getConsensusNodeState(unsyncedNode)
		assertHeight(t, unsyncedNodeId, unsyncedNodeHeight+1, nodeState.Height)

		// Same as `unsyncedNodeHeight+=1`
		unsyncedNodeHeight = unsyncedNode.GetBus().GetConsensusModule().CurrentHeight()
	}

	assertHeight(t, unsyncedNodeId, uint64(4), getConsensusNodeState(unsyncedNode).Height)
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
