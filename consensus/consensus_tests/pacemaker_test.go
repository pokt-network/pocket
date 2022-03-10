package consensus_tests

import (
	"log"
	"testing"
	"time"

	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
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
	// numNodes := 4
	// configs := GenerateNodeConfigs(numNodes)

	// // Start test pocket nodes.
	// testPocketBus := make(modules.PocketBus, 100)
	// pocketNodes := CreateTestConsensusPocketNodes(t, configs, testPocketBus)
	// ctx := context.EmptyPocketContext()
	// for _, pocketNode := range pocketNodes {
	// 	go pocketNode.Start(ctx)
	// }
	// time.Sleep(10 * time.Millisecond) // Needed to avoid minor race condition if pocketNode has not finished initialization

	// // Set all nodes to the same step and height but different rounds.
	// testHeight := uint64(3)
	// testStep := uint64(consensus.NewRound)
	// leaderId := s_types.NodeId(1)
	// leader := pocketNodes[leaderId]
	// for _, pocketNode := range pocketNodes {
	// 	consensusModImpl := getConsensusModImplementation(pocketNode)
	// 	consensusModImpl.FieldByName("Height").SetUint(testHeight)
	// 	consensusModImpl.FieldByName("Step").SetUint(testStep)
	// 	// consensusModImpl.FieldByName("LeaderId").Set(reflect.ValueOf(nil)) // Leader is not set because the round update should set it appropriately.
	// }

	// header := &types2.BlockHeaderConsensusTemp{
	// 	Height: int64(testHeight),
	// 	Hash:   "new_hash",

	// 	LastBlockHash:   "prev_hash",
	// 	ProposerAddress: []byte(leader.Address),
	// 	ProposerId:      uint32(leaderId),
	// }
	// leaderBlock := &types2.BlockConsensusTemp{
	// 	BlockHeader:       header,
	// 	Transactions:      make([]*types2.Transaction, 0),
	// 	ConsensusEvidence: make([]*types2.Evidence, 0),
	// }

	// leaderConsensusMod := getConsensusModImplementation(leader)
	// leaderConsensusMod.FieldByName("Block").Set(reflect.ValueOf(leaderBlock))

	// leaderRound := uint64(6)
	// // Set the leader to be in the highest round.
	// getConsensusModImplementation(pocketNodes[leaderId]).FieldByName("Round").SetUint(leaderRound)
	// getConsensusModImplementation(pocketNodes[2]).FieldByName("Round").SetUint(uint64(2))
	// getConsensusModImplementation(pocketNodes[3]).FieldByName("Round").SetUint(uint64(3))
	// getConsensusModImplementation(pocketNodes[4]).FieldByName("Round").SetUint(uint64(4))

	// prepareProposeMessage := &consensus.HotstuffMessage{
	// 	Type:      consensus.ProposeMessageType,
	// 	Height:    consensus.BlockHeight(testHeight),
	// 	Step:      consensus.Prepare,
	// 	Round:     consensus.Round(leaderRound),
	// 	Block:     leaderBlock,
	// 	JustifyQC: nil,
	// 	Sender:    leaderId,
	// }
	// P2PBroadcast(pocketNodes, prepareProposeMessage) // Broadcast the prepare proposal.

	// // numNodes-1 because one of the messages is a self-proposal.
	// WaitForNetworkConsensusMessage(t, testPocketBus, types.P2P_SEND_MESSAGE, consensus.Prepare, numNodes-1, 2000)
	// leaderConsensusMod.FieldByName("LeaderId").Set(reflect.ValueOf(&leaderId))

	// // Check that the leader is in the latest round.
	// for _, pocketNode := range pocketNodes {
	// 	nodeState := GetConsensusNodeState(pocketNode)
	// 	require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
	// 	require.Equal(t, uint64(3), nodeState.Height)
	// 	require.Equal(t, uint8(6), nodeState.Round)
	// 	require.Equal(t, leaderId, nodeState.LeaderId)
	// }
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
