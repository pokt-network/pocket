package e2e_tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestStateSync_ServerGetMetaDataReq_SuccessfulTest(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	//pocketNodes := CreateTestStateSyncPocketNodes(t, buses, eventsChannel)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	testHeight := uint64(4)

	// Choose node 1 as the server node, enable server mode of the node 1
	// Set server node's height to test height.
	serverNode := pocketNodes[1]
	serverNodePeerId, err := serverNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
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
		//MsgType: typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST,
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &stateSyncMetaDataReq,
		},
	}
	anyProto, err := anypb.New(stateSyncMetaDataReqMessage)
	require.NoError(t, err)

	// Send metadata request to the server node
	P2PSend(t, serverNode, anyProto)

	// Start waiting for the metadata request on server node,
	errMsg := "StateSync Metadata Request"
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, 1, 250, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncMetaDataResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	metaDataRes := stateSyncMetaDataResMessage.GetMetadataRes()
	require.NotEmpty(t, metaDataRes)

	require.Equal(t, uint64(4), metaDataRes.MaxHeight)
	require.Equal(t, uint64(1), metaDataRes.MinHeight)
	require.Equal(t, serverNodePeerId, metaDataRes.PeerId)
}

func TestStateSync_ServerGetBlock_SuccessfulTest(t *testing.T) {
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

	testHeight := uint64(5)

	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerId, err := requesterNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)

	// Passing Test
	// Test GetBlock Req
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

	// Start waiting for the get block request on server node, expect to return error
	errMsg := "StateSync Get Block Request Message"
	numExpectedMsgs := 1
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

func TestStateSync_ServerGetBlock_FailingTest(t *testing.T) {
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

	testHeight := uint64(5)

	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	// Choose node 2 as the requester node
	requesterNode := pocketNodes[2]
	requesterNodePeerId, err := requesterNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId()
	require.NoError(t, err)

	// Failing Test
	// Get Block Req is current block height + 1
	stateSyncGetBlockReq := typesCons.GetBlockRequest{
		PeerId: requesterNodePeerId,
		Height: testHeight + 1,
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

	numExpectedMsgs := 1
	// Start waiting for the get block request on server node, expect to return error
	errMsg := "StateSync Get Block Request Message"
	_, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
	require.Error(t, err)
}
