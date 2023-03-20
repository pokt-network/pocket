package e2e_tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestStateSync_ServerGetMetaDataReq_Success(t *testing.T) {
	t.Skip()
	/*
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

		// Choose node 1 as the server node
		// Set server node's height to test height.
		serverNode := pocketNodes[1]
		serverNodePeerId := serverNode.GetBus().GetConsensusModule().GetNodeAddress()
		serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
		serverNodeConsensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})

		// We choose node 2 as the requester node.
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
		require.Equal(t, serverNodePeerId, metaDataRes.PeerAddress)

	*/
}

func TestStateSync_ServerGetBlock_Success(t *testing.T) {
	t.Skip()
	/*
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

	*/
}

func TestStateSync_ServerGetBlock_FailNonExistingBlock(t *testing.T) {
	t.Skip()
	/*
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

		numExpectedMsgs := 1
		// Start waiting for the get block request on server node, expect to return error
		errMsg := "StateSync Get Block Request Message"
		_, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
		require.Error(t, err)
	*/
}

func TestStateSync_UnsynchedPeerSynchs_Success(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	numberOfValidators := 4
	numberOfPersistedDummyBlocks := uint64(10)
	// current height of the node is one plus the number of persisted dummy blocks
	testHeight := numberOfPersistedDummyBlocks + 1
	testStep := uint8(consensus.NewRound)
	testRound := uint64(1)

	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numberOfValidators, clockMock)
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)

	GenerateDummyBlocksWithQC(t, testHeight, uint64(numberOfValidators), pocketNodes)

	StartAllTestPocketNodes(t, pocketNodes)

	// Prepare leader info
	leaderId := typesCons.NodeId(3)
	require.Equal(t, uint64(leaderId), testHeight%uint64(numberOfValidators)) // Uses our deterministic round robin leader election
	leader := pocketNodes[leaderId]
	consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	// Prepare unsynched node info
	unsynchedNode := pocketNodes[2]
	unsynchedNodeId := typesCons.NodeId(2)
	unsynchedNodeHeight := testHeight - 1
	unsynchedNodeModImpl := GetConsensusModImpl(unsynchedNode)

	// Placeholder block
	blockHeader := &coreTypes.BlockHeader{
		Height:            testHeight,
		StateHash:         stateHash,
		PrevStateHash:     "",
		ProposerAddress:   consensusPK.Address(),
		QuorumCertificate: nil,
	}
	block := &coreTypes.Block{
		BlockHeader:  blockHeader,
		Transactions: make([][]byte, 0),
	}

	leaderConsensusModImpl := GetConsensusModImpl(leader)
	leaderConsensusModImpl.MethodByName("SetBlock").Call([]reflect.Value{reflect.ValueOf(block)})

	// Set the unsynched node to last persisted height 9
	// Set rest of the nodes to the last persisted height 11
	for id, pocketNode := range pocketNodes {
		consensusModImpl := GetConsensusModImpl(pocketNode)
		if id == unsynchedNodeId {
			consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(unsynchedNodeHeight)})
			utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(unsynchedNodeHeight))
			require.NoError(t, err)
			consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
		} else {
			consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
			utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(testHeight))
			require.NoError(t, err)
			consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
		}
		consensusModImpl.MethodByName("SetStep").Call([]reflect.Value{reflect.ValueOf(testStep)})
		consensusModImpl.MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(testRound)})
	}

	//Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// // Assert that unsynched node has a separate view of the network than the rest of the nodes
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numberOfValidators*numberOfValidators, 250, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == unsynchedNodeId {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: unsynchedNodeHeight,
					Step:   testStep,
					Round:  uint8(testRound + 1),
				},
				nodeState)
		} else {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: testHeight,
					Step:   testStep,
					Round:  uint8(testRound + 1),
				},
				nodeState)
		}
		require.Equal(t, false, nodeState.IsLeader)
		require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId)
	}

	unsynchedNodeModImpl.MethodByName("SetAggregatedStateSyncMetadata").Call([]reflect.Value{reflect.ValueOf(uint64(1)), reflect.ValueOf(numberOfPersistedDummyBlocks), reflect.ValueOf(string(consensusPK.Address()))})

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// Node must request blocks from all other validators.
	// Ensure it sends requests to "numberOfValidators - 1" getBlockReq messages.
	errMsg := "StateSync Get Block Request Messages"
	numExpectedMsgs := numberOfValidators - 1
	//receivedStateSyncMsgs, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
	msgs, err := WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
	require.NoError(t, err)

	// for _, msg := range msgs {
	// 	msg, err := codec.GetCodec().FromAny(msg)
	// 	require.NoError(t, err)

	// 	stateSyncBlockReqMessage, ok := msg.(*typesCons.StateSyncMessage)
	// 	require.True(t, ok)

	// 	blockReq := stateSyncBlockReqMessage.GetGetBlockReq()
	// 	require.NotEmpty(t, blockReq)

	// 	fmt.Println("Block req is: ", blockReq)
	// }

	// send the block request sent by unsynched node to all nodes
	P2PBroadcast(t, pocketNodes, msgs[0])
	advanceTime(t, clockMock, 10*time.Millisecond)

	// We mock that all validators will be replying to its request
	// So, ensure it sends requests to "numberOfValidators - 1" getBlockReq messages.
	errMsg = "StateSync Get Block Response Messages"
	msgs, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, errMsg, numExpectedMsgs, 250, false)
	require.NoError(t, err)

	// get a valid block. One of them should be valid
	// for _, msg := range msgs {

	// 	//fmt.Println("RECEIVED STATE SYNC MSG: ", msg.TypeUrl)
	// 	msg, err := codec.GetCodec().FromAny(msg)
	// 	require.NoError(t, err)

	// 	stateSyncBlockResMessage, ok := msg.(*typesCons.StateSyncMessage)
	// 	require.True(t, ok)

	// 	fmt.Println("Received Get Block Response: ", stateSyncBlockResMessage)

	// 	blockReq := stateSyncBlockResMessage.GetGetBlockRes()
	// 	require.NotEmpty(t, blockReq)

	// }

	// send the first reply to the unsynched node
	P2PSend(t, unsynchedNode, msgs[0])
	advanceTime(t, clockMock, 10*time.Millisecond)

	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		//fmt.Println("Node state is h s r: ", nodeState.Height, nodeState.Step, nodeState.Round)
		assertHeight(t, nodeId, testHeight, nodeState.Height)
	}
}
func TestStateSync_UnsynchedPeerSynchs(t *testing.T) {
	t.Skip()
}

// Test if unsynched peer can catch up with the rest of the network
func TestStateSync_UnsynchedPeerCatchsUpConsensus_ConsensusProceedsSuccessfully(t *testing.T) {
	t.Skip()
}
