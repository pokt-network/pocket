package consensus_tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestStateSyncServer(t *testing.T) {
	clockMock := clock.NewMock()
	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)

	go timeReminder(clockMock, 100*time.Millisecond)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, runtimeMgrs, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	serverNode := pocketNodes[1]
	serverNodeConsensusModImpl := GetConsensusModImpl(serverNode)
	serverNodeConsensusModImpl.MethodByName("EnableServerMode").Call([]reflect.Value{})

	originatorNode := pocketNodes[2]

	stateSyncReq := typesCons.StateSyncMetadataRequest{
		PeerId: originatorNode.GetBus().GetConsensusModule().GetCurrentNodeAddressFromNodeId(),
	}

	stateSyncMessage := &typesCons.StateSyncMessage{
		MsgType: typesCons.StateSyncMessageType_STATE_SYNC_METADATA_REQUEST,
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &stateSyncReq,
		},
	}
	anyProto, err := anypb.New(stateSyncMessage)
	require.NoError(t, err)

	P2PSend(t, serverNode, anyProto)
	advanceTime(clockMock, 10*time.Millisecond)

	_, err = WaitForNetworkStateSyncMessages(t, clockMock, testChannel, serverNode.GetP2PAddress(), typesCons.StateSyncMessageType_STATE_SYNC_METADATA_RESPONSE, numValidators, 1000)
	//_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numValidators, 1000)
	require.NoError(t, err)
	/*
			// NewRound
		newRoundMessages, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numValidators, 1000)
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
		}
		for _, message := range newRoundMessages {
			P2PBroadcast(t, pocketNodes, message)
		}
	*/

	//-- WAIT FOR MESSAGE AFTER SENDING WITH P2PSEND

	//require.Equal(t, false, nodeState.IsLeader)

}
