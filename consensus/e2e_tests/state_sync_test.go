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
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestStateSyncServer(t *testing.T) {
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
	leader := pocketNodes[leaderId]

	consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	blockHeader := &coreTypes.BlockHeader{
		Height:            testHeight,
		StateHash:         stateHash,
		PrevStateHash:     "",
		NumTxs:            0,
		ProposerAddress:   consensusPK.Address(),
		QuorumCertificate: nil,
	}

	block := &coreTypes.Block{
		BlockHeader:  blockHeader,
		Transactions: make([][]byte, 0),
	}

	leaderConsensusModImpl := GetConsensusModImpl(leader)
	leaderConsensusModImpl.MethodByName("SetBlock").Call([]reflect.Value{reflect.ValueOf(block)})

	// Set all nodes to the same STEP and HEIGHT BUT different ROUNDS
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

	prepareProposal := &typesCons.HotstuffMessage{
		Type:          consensus.Propose,
		Height:        4,
		Step:          consensus.Prepare,
		Round:         0,
		Block:         block,
		Justification: nil,
	}
	anyMsg, err := anypb.New(prepareProposal)
	require.NoError(t, err)

	P2PBroadcast(t, pocketNodes, anyMsg)

	numExpectedMsgs := numValidators - 1 // -1 because one of the messages is a self proposal (leader to itself as a replica) that is not passed through the network
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Vote, numExpectedMsgs, 250, true)
	require.NoError(t, err)

	nodeState := GetConsensusNodeState(pocketNodes[1])
	require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))

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
	numExpectedMsgs = 1
	receivedMsg, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE, requesterNode.GetP2PAddress(), numExpectedMsgs, 250, false)
	require.NoError(t, err)

	msg, err := codec.GetCodec().FromAny(receivedMsg[0])
	require.NoError(t, err)

	stateSyncMetaDataResMessage, ok := msg.(*typesCons.StateSyncMessage)
	require.True(t, ok)

	metaDataRes := stateSyncMetaDataResMessage.GetMetadataRes()
	require.NotEmpty(t, metaDataRes)

	require.Equal(t, uint64(4), metaDataRes.MaxHeight)

	// // Test GetBlock Req
	stateSyncGetBlockReq := typesCons.GetBlockRequest{
		PeerId: requesterNodePeerId,
		Height: 0,
	}

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_REQUEST,
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &stateSyncGetBlockReq,
		},
	}

	anyProto, err = anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	// send get block request to the server node
	P2PSend(t, serverNode, anyProto)

	// start waiting for the get block request on server node,
	numExpectedMsgs = 1
	receivedMsg, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, typesCons.StateSyncMessageType_STATE_SYNC_GET_BLOCK_RESPONSE, requesterNode.GetP2PAddress(), numExpectedMsgs, 250, false)
	require.NoError(t, err)

	// msg, err = codec.GetCodec().FromAny(receivedMsg[0])
	// require.NoError(t, err)

	// stateSyncGetBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	// require.True(t, ok)

	// getBlockRes := stateSyncGetBlockResMessage.GetGetBlockRes()
	// require.NotEmpty(t, metaDataRes)

	// fmt.Printf("Get Block Response: %s", getBlockRes)
	// require.Equal(t, getBlockRes.Block.GetBlockHeader().Height, testHeight)

}
