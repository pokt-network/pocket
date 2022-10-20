package consensus_tests

import (
	"encoding/hex"
	"reflect"
	"runtime"
	"testing"
	"time"
	timePkg "time"

	"github.com/benbjohnson/clock"

	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestTinyPacemakerTimeouts(t *testing.T) {
	clockMock := clock.NewMock()
	timeReminder(clockMock, 100*time.Millisecond)

	// Test configs
	numNodes := 4
	paceMakerTimeoutMsec := uint64(50) // Set a very small pacemaker timeout
	paceMakerTimeout := 50 * time.Millisecond
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numNodes, clockMock)
	for _, runtimeConfig := range runtimeMgrs {
		if consCfg, ok := runtimeConfig.GetConfig().GetConsensusConfig().(consensus.HasPacemakerConfig); ok {
			consCfg.GetPacemakerConfig().SetTimeoutMsec(paceMakerTimeoutMsec)
		}
	}

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, runtimeMgrs, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Debug message to start consensus by triggering next view.
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// advance time by an amount shorter than the timeout
	advanceTime(clockMock, 10*time.Millisecond)

	// paceMakerTimeout
	_, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.NewRound),
				Round:  0,
			},
			GetConsensusNodeState(pocketNode))
	}

	forcePacemakerTimeout(clockMock, paceMakerTimeout)

	// Check that a new round starts at the same height.
	_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.NewRound),
				Round:  1,
			},
			GetConsensusNodeState(pocketNode))
	}

	forcePacemakerTimeout(clockMock, paceMakerTimeout)
	// // Check that a new round starts at the same height
	_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.NewRound),
				Round:  2,
			},
			GetConsensusNodeState(pocketNode))
	}

	forcePacemakerTimeout(clockMock, paceMakerTimeout)

	// Check that a new round starts at the same height.
	newRoundMessages, err := WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.NewRound, consensus.Propose, numNodes, 500)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.NewRound),
				Round:  3,
			},
			GetConsensusNodeState(pocketNode))
	}

	// Continue to the next step at the current round
	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}

	// advance time by an amount shorter than the timeout
	advanceTime(clockMock, 10*time.Millisecond)

	// Confirm we are at the next step
	_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Prepare, consensus.Propose, 1, 500)
	require.NoError(t, err)
	for pocketId, pocketNode := range pocketNodes {
		assertNodeConsensusView(t, pocketId,
			typesCons.ConsensusNodeState{
				Height: 1,
				Step:   uint8(consensus.Prepare),
				Round:  3,
			},
			GetConsensusNodeState(pocketNode))
	}
}

func TestPacemakerCatchupSameStepDifferentRounds(t *testing.T) {
	clockMock := clock.NewMock()
	numNodes := 4
	runtimeConfigs := GenerateNodeRuntimeMgrs(t, numNodes, clockMock)

	timeReminder(clockMock, 100*time.Millisecond)

	// Create & start test pocket nodes
	testChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, runtimeConfigs, testChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// Starting point
	testHeight := uint64(3)
	testStep := int64(consensus.NewRound)

	// Leader info
	leaderId := typesCons.NodeId(3) // TODO(olshansky): Same as height % numValidators until we add back leader election
	leader := pocketNodes[leaderId]
	leaderRound := uint64(6)

	consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	// Placeholder block
	blockHeader := &typesCons.BlockHeader{
		Height:            int64(testHeight),
		Hash:              hex.EncodeToString(appHash),
		NumTxs:            0,
		LastBlockHash:     "",
		ProposerAddress:   consensusPK.Address(),
		QuorumCertificate: nil,
	}
	block := &typesCons.Block{
		BlockHeader:  blockHeader,
		Transactions: emptyTxs,
	}

	leaderConsensusMod := GetConsensusModElem(leader)
	leaderConsensusMod.FieldByName("Block").Set(reflect.ValueOf(block))

	// Set all nodes to the same STEP and HEIGHT BUT different ROUNDS
	for _, pocketNode := range pocketNodes {
		// utilityContext is only set on new rounds, which is skipped in this test
		utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(testHeight))
		require.NoError(t, err)

		consensusModElem := GetConsensusModElem(pocketNode)
		consensusModElem.FieldByName("Height").SetUint(testHeight)
		consensusModElem.FieldByName("Step").SetInt(testStep)
		consensusModElem.FieldByName("LeaderId").Set(reflect.Zero(reflect.TypeOf(&leaderId))) // This is re-elected during paceMaker catchup

		consensusModImpl := GetConsensusModImpl(pocketNode)
		consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
	}

	// Set the leader to be in the highest round.
	GetConsensusModElem(pocketNodes[1]).FieldByName("Round").SetUint(uint64(leaderRound - 2))
	GetConsensusModElem(pocketNodes[2]).FieldByName("Round").SetUint(uint64(leaderRound - 3))
	GetConsensusModElem(pocketNodes[leaderId]).FieldByName("Round").SetUint(uint64(leaderRound))
	GetConsensusModElem(pocketNodes[4]).FieldByName("Round").SetUint(uint64(leaderRound - 4))

	prepareProposal := &typesCons.HotstuffMessage{
		Type:          consensus.Propose,
		Height:        testHeight,
		Step:          consensus.Prepare, //typesCons.HotstuffStep(testStep),
		Round:         leaderRound,
		Block:         block,
		Justification: nil,
	}
	anyMsg, err := anypb.New(prepareProposal)
	require.NoError(t, err)

	P2PBroadcast(t, pocketNodes, anyMsg)

	// numNodes-1 because one of the messages is a self-proposal that is not passed through the network
	_, err = WaitForNetworkConsensusMessages(t, clockMock, testChannel, consensus.Prepare, consensus.Vote, numNodes-1, 2000)
	require.NoError(t, err)

	forcePacemakerTimeout(clockMock, 600*time.Millisecond)

	// Check that the leader is in the latest round.
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == leaderId {
			require.Equal(t, uint8(consensus.Prepare), nodeState.Step)
		} else {
			require.Equal(t, uint8(consensus.PreCommit), nodeState.Step)
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

func forcePacemakerTimeout(clockMock *clock.Mock, paceMakerTimeout timePkg.Duration) {
	go func() {
		// Cause the pacemaker to timeout
		sleep(clockMock, paceMakerTimeout)
	}()
	runtime.Gosched()
	// advance time by an amount longer than the timeout
	advanceTime(clockMock, paceMakerTimeout+10*time.Millisecond)
}
