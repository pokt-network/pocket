package e2e_tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
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

	//targetBlockHeight := 3

	// //for i := 0; i < 3; i++ {
	// ConsensusProceedToNextBlock(
	// 	t,
	// 	clockMock,
	// 	eventsChannel,
	// 	pocketNodes,
	// 	targetBlockHeight,
	// )
	//advanceTime(t, clockMock, 10*time.Millisecond)

	// Starting point
	testHeight := uint64(4)
	testStep := uint8(consensus.NewRound)
	testRound := uint64(0)

	leaderId := typesCons.NodeId(2)
	leader := pocketNodes[leaderId]

	consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	// Placeholder block
	blockHeader := &typesCons.BlockHeader{
		Height:            int64(testHeight),
		Hash:              stateHash,
		NumTxs:            0,
		LastBlockHash:     "",
		ProposerAddress:   consensusPK.Address(),
		QuorumCertificate: nil,
	}
	block := &typesCons.Block{
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

	//}

	// _, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, 250, true)
	// require.NoError(t, err)
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
	if err != nil {
		//TODO check if this must be failed, rather than logged.
		t.Logf("Can't get the peerId%s", err)
	}

	stateSyncReq := typesCons.StateSyncMetadataRequest{
		PeerId: requesterNodePeerId,
	}

	stateSyncMessage := &typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST,
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &stateSyncReq,
		},
	}
	anyProto, err := anypb.New(stateSyncMessage)
	require.NoError(t, err)

	// send metadata request to the server node
	P2PSend(t, serverNode, anyProto)

	// start waiting for the state sync message on server node,
	numExpectedMsgs = 1
	_, err = WaitForNetworkStateSyncEvents(t, clockMock, eventsChannel, typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE, serverNode.GetP2PAddress(), numExpectedMsgs, 250, false)
	require.NoError(t, err)
}
