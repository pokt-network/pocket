package consensus_tests

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/pokt-network/pocket/consensus"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestTinyPacemakerTimeouts(t *testing.T) {
	// There can be race conditions related to having a small paceMaker time out, so we skip this test
	// when `failOnExtraMessages` is set to true to simplify things for now. However, we still validate
	// that the rounds are incremented as expected when `failOnExtraMessages` is false.
	if failOnExtraMessages == true {
		log.Println("[DEBUG] Skipping TestPacemakerTimeouts because `failOnExtraMessages` is set to true.")
		t.Skip()
	}

	// Test configs
	numNodes := 4
	paceMakerTimeoutMsec := uint64(50) // Set a very small pacemaker timeout
	paceMakerTimeout := 50 * time.Millisecond
	configs := GenerateNodeConfigs(t, numNodes)
	for _, config := range configs {
		config.Consensus.Pacemaker.TimeoutMsec = paceMakerTimeoutMsec
	}

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testChannel)
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
	}
	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	// Debug message to start consensus by triggering next view.
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// paceMakerTimeout
	_, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(0), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// Check that a new round starts at the same height.
	_, err = WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(1), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// // Check that a new round starts at the same height.
	_, err = WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(2), nodeState.Round)
	}

	// Cause the pacemaker to timeout.
	time.Sleep(paceMakerTimeout)

	// Check that a new round starts at the same height.
	newRoundMessages, err := WaitForNetworkConsensusMessages(t, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.NewRound), nodeState.Step)
		require.Equal(t, uint8(3), nodeState.Round)
	}

	// Continue to the next step at the current round
	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}

	// Confirm we are at the next step
	_, err = WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Propose, 1, 500)
	require.NoError(t, err)
	for _, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		require.Equal(t, uint64(1), nodeState.Height)
		require.Equal(t, uint8(consensus.Prepare), nodeState.Step)
		require.Equal(t, uint8(3), nodeState.Round)
	}

}

func TestPacemakerCatchupSameStepDifferentRounds(t *testing.T) {
	numNodes := 4
	configs := GenerateNodeConfigs(t, numNodes)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, configs, testChannel)
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
	}
	time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	// Starting point
	testHeight := uint64(3)
	testStep := int64(consensus.NewRound)

	// Leader info
	leaderId := types_consensus.NodeId(3) // TODO(olshansky): Same as height % numValidators until we add back leader election
	leader := pocketNodes[leaderId]
	leaderRound := uint64(6)

	// Placeholder block
	blockHeader := &types_consensus.BlockHeaderConsensusTemp{
		Height:            int64(testHeight),
		Hash:              appHash,
		NumTxs:            0,
		LastBlockHash:     "",
		ProposerAddress:   []byte(leader.Address),
		QuorumCertificate: nil,
	}
	block := &types_consensus.BlockConsensusTemp{
		BlockHeader:  blockHeader,
		Transactions: emptyTxs,
	}

	leaderConsensusMod := GetConsensusModImplementation(leader)
	leaderConsensusMod.FieldByName("Block").Set(reflect.ValueOf(block))

	// Set all nodes to the same STEP and HEIGHT BUT different ROUNDS.
	for _, pocketNode := range pocketNodes {
		consensusModImpl := GetConsensusModImplementation(pocketNode)
		consensusModImpl.FieldByName("Height").SetUint(testHeight)
		consensusModImpl.FieldByName("Step").SetInt(testStep)
		consensusModImpl.FieldByName("LeaderId").Set(reflect.Zero(reflect.TypeOf(&leaderId))) // This is re-elected during paceMaker catchup
	}

	// Set the leader to be in the highest round.
	GetConsensusModImplementation(pocketNodes[1]).FieldByName("Round").SetUint(uint64(leaderRound - 2))
	GetConsensusModImplementation(pocketNodes[2]).FieldByName("Round").SetUint(uint64(leaderRound - 3))
	GetConsensusModImplementation(pocketNodes[leaderId]).FieldByName("Round").SetUint(uint64(leaderRound))
	GetConsensusModImplementation(pocketNodes[4]).FieldByName("Round").SetUint(uint64(leaderRound - 4))

	prepareProposal := &types_consensus.HotstuffMessage{
		Type:          consensus.Propose,
		Height:        testHeight,
		Step:          consensus.Prepare, //types_consensus.HotstuffStep(testStep),
		Round:         leaderRound,
		Block:         block,
		Justification: nil,
	}
	anyMsg, err := anypb.New(prepareProposal)
	require.NoError(t, err)
	consensusMessage := &types_consensus.ConsensusMessage{
		Type:    consensus.HotstuffMessage,
		Message: anyMsg,
	}

	P2PBroadcast(t, pocketNodes, consensusMessage)

	// numNodes-1 because one of the messages is a self-proposal that is not passed through the network
	_, err = WaitForNetworkConsensusMessages(t, testChannel, consensus.Prepare, consensus.Vote, numNodes, 2000)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	// Check that the leader is in the latest round.
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		fmt.Println("OLSH", nodeId, nodeState.Step, leaderId)
		if nodeId == leaderId {
			require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
		} else {
			// require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
		}
		require.Equal(t, uint64(3), nodeState.Height)
		require.Equal(t, uint8(6), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId)
	}
}

/*
func TestPacemakerDifferentHeightsCatchup(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerDifferentStepsCatchup(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerDifferentRoudnsCatchup(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerWithLockedQC(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerWithHighPrepareQC(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerNoQuorum(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerNotSafeProposal(t *testing.T) {
	t.Skip() // TODO: Implement
}

func TestPacemakerExponentialTimeouts(t *testing.T) {
	t.Skip() // TODO: Implement
}
*/
