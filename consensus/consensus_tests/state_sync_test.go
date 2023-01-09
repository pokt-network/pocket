package consensus_tests

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

// ! TODO IMPLEMENT UNIT TEST
func TestStateSync(t *testing.T) {
	clockMock := clock.NewMock()
	// Test configs
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)

	go timeReminder(clockMock, 100*time.Millisecond)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, runtimeMgrs, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	advanceTime(clockMock, 10*time.Millisecond)

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
}
