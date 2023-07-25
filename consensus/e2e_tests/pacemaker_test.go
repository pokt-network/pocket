package e2e_tests

import (
	"runtime"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestPacemakerTimeoutIncreasesRound(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// UnitTestNet configs
	// IMPROVE(#295): Remove time specific suffixes as outlined by go-staticcheck (ST1011)
	paceMakerTimeoutMsec := uint64(10000) // Set a small pacemaker timeout
	paceMakerTimeout := time.Duration(paceMakerTimeoutMsec) * time.Millisecond
	consensusMessageTimeout := time.Duration(paceMakerTimeoutMsec / 5) // Must be smaller than pacemaker timeout because we expect a deterministic number of consensus messages.
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	for _, runtimeConfig := range runtimeMgrs {
		consCfg := runtimeConfig.GetConfig().Consensus.PacemakerConfig
		consCfg.TimeoutMsec = paceMakerTimeoutMsec
	}
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	// Debug message to start consensus by triggering next view
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}

	// Advance time by an amount shorter than the pacemaker timeout
	advanceTime(t, clockMock, 10*time.Millisecond)

	_, err = waitForProposalMsgs(t, clockMock, eventsChannel, pocketNodes, 1, uint8(consensus.NewRound), 0, 0, numValidators*numValidators, consensusMessageTimeout, true)
	require.NoError(t, err)

	// Force the pacemaker to time out
	forcePacemakerTimeout(t, clockMock, paceMakerTimeout)
	// Wait for the round=1 to fail
	_, err = waitForProposalMsgs(t, clockMock, eventsChannel, pocketNodes, 1, uint8(consensus.NewRound), 1, 0, numValidators*numValidators, consensusMessageTimeout, true)
	require.NoError(t, err)

	forcePacemakerTimeout(t, clockMock, paceMakerTimeout)
	// Wait for the round=2 to fail
	_, err = waitForProposalMsgs(t, clockMock, eventsChannel, pocketNodes, 1, uint8(consensus.NewRound), 2, 0, numValidators*numValidators, consensusMessageTimeout, true)
	require.NoError(t, err)

	forcePacemakerTimeout(t, clockMock, paceMakerTimeout)
	// Wait for the round=3 to succeed
	newRoundMessages, err := waitForProposalMsgs(t, clockMock, eventsChannel, pocketNodes, 1, uint8(consensus.NewRound), 3, 0, numValidators*numValidators, consensusMessageTimeout, true)
	require.NoError(t, err)
	broadcastMessages(t, newRoundMessages, pocketNodes)
	advanceTime(t, clockMock, 10*time.Millisecond)

	// Get the expected leader id for round=3
	leaderId := typesCons.NodeId(pocketNodes[1].GetBus().GetConsensusModule().GetLeaderForView(1, 3, uint8(consensus.NewRound)))
	// Wait for nodes to proceed to Propose step in round=3
	_, err = waitForProposalMsgs(t, clockMock, eventsChannel, pocketNodes, 1, uint8(consensus.Prepare), 3, leaderId, numValidators, consensusMessageTimeout, true)
	require.NoError(t, err)
}

func TestPacemakerCatchupSameStepDifferentRounds(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	runtimeConfigs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	buses := GenerateBuses(t, runtimeConfigs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	// Starting point
	testHeight := uint64(3)
	testStep := uint8(consensus.NewRound)

	// UnitTestNet configs
	paceMakerTimeoutMsec := uint64(500) // Set a small pacemaker timeout
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	for _, runtimeConfig := range runtimeMgrs {
		runtimeConfig.GetConfig().Consensus.PacemakerConfig.TimeoutMsec = paceMakerTimeoutMsec
	}

	// Set all nodes to the same STEP and HEIGHT BUT different ROUNDS
	for _, pocketNode := range pocketNodes {
		// Update height, step, leaderId, and utility via setters exposed with the debug interface
		pocketNode.GetBus().GetConsensusModule().SetHeight(testHeight)
		pocketNode.GetBus().GetConsensusModule().SetStep(testStep)

		// utilityUnitOfWork is only set on new rounds, which is skipped in this test
		utilityUnitOfWork, err := pocketNode.GetBus().GetUtilityModule().NewUnitOfWork(int64(testHeight))
		require.NoError(t, err)
		pocketNode.GetBus().GetConsensusModule().SetUtilityUnitOfWork(utilityUnitOfWork)
	}

	// Prepare leader info
	leaderRound := uint64(6)

	// Get leaderId for the given height, round and step, by using the Consensus Modules' GetLeaderForView() function.
	// Any node in pocketNodes mapping can be used to call GetLeaderForView() function.
	leaderId := typesCons.NodeId(pocketNodes[1].GetBus().GetConsensusModule().GetLeaderForView(testHeight, leaderRound, uint8(consensus.Prepare)))
	leader := pocketNodes[leaderId]
	leaderPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	block := generatePlaceholderBlock(testHeight, leaderPK.Address())
	leader.GetBus().GetConsensusModule().SetBlock(block)

	// Set the leader to be in the highest round.
	pocketNodes[1].GetBus().GetConsensusModule().SetRound(leaderRound - 2)
	pocketNodes[2].GetBus().GetConsensusModule().SetRound(leaderRound - 3)
	pocketNodes[leaderId].GetBus().GetConsensusModule().SetRound(leaderRound)
	pocketNodes[4].GetBus().GetConsensusModule().SetRound(leaderRound - 4)

	prepareProposal := &typesCons.HotstuffMessage{
		Type:          consensus.Propose,
		Height:        testHeight,
		Step:          consensus.Prepare,
		Round:         leaderRound,
		Block:         block,
		Justification: nil,
	}
	anyMsg, err := anypb.New(prepareProposal)
	require.NoError(t, err)

	numExpectedMsgs := numValidators - 1   // -1 because one of the messages is a self proposal (leader to itself as a replica) that is not passed through the network
	msgTimeout := paceMakerTimeoutMsec / 2 // /2 because we do not want the pacemaker to trigger a new timeout

	broadcastMessages(t, []*anypb.Any{anyMsg}, pocketNodes)
	advanceTime(t, clockMock, 10*time.Millisecond)
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, 2, consensus.Vote, numExpectedMsgs, time.Duration(msgTimeout), true)

	require.NoError(t, err)

	// Check that all the nodes caught up to the leader's (i.e. the latest) round
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == leaderId {
			require.Equal(t, consensus.Prepare.String(), typesCons.HotstuffStep(nodeState.Step).String())
		} else {
			require.Equal(t, consensus.PreCommit.String(), typesCons.HotstuffStep(nodeState.Step).String())
		}
		require.Equal(t, uint64(3), nodeState.Height)
		require.Equal(t, uint8(leaderRound), nodeState.Round)
		require.Equal(t, leaderId, nodeState.LeaderId)
	}
}

func forcePacemakerTimeout(t *testing.T, clockMock *clock.Mock, paceMakerTimeout time.Duration) {
	go func() {
		// Cause the pacemaker to timeout
		sleep(t, clockMock, paceMakerTimeout)
	}()
	runtime.Gosched()
	// advance time by an amount longer than the timeout
	advanceTime(t, clockMock, paceMakerTimeout+10*time.Millisecond)
}

// TODO: Add more tests for minBlockTime behavior:
// 1. Block preparation triggers ASAP if conditions are met AFTER minBlockTime has triggered.
// 2. Block preparation is always discarded if a new one with better QC is received within minBlockTime.
// 3. Mempool reaped is the one present at minBlockTime or later.
// 4. Successive blocks timings are at least minBlockTime apart.
func TestPacemakerMinBlockTime(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	// UnitTestNet configs
	paceMakerTimeoutMsec := uint64(300000)
	consensusMessageTimeout := time.Duration(paceMakerTimeoutMsec / 5) // Must be smaller than pacemaker timeout because we expect a deterministic number of consensus messages.
	paceMakerMinBlockTimeMsec := uint64(5000)                          // Make sure it is larger than the consensus message timeout
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
	for _, runtimeConfig := range runtimeMgrs {
		consCfg := runtimeConfig.GetConfig().Consensus.PacemakerConfig
		consCfg.TimeoutMsec = paceMakerTimeoutMsec
		consCfg.MinBlockTimeMsec = paceMakerMinBlockTimeMsec
	}
	buses := GenerateBuses(t, runtimeMgrs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	err := StartAllTestPocketNodes(t, pocketNodes)
	require.NoError(t, err)

	replicas := IdToNodeMapping{}
	// First round ever has leaderId=2 ((height+round+step-1)%numValidators)+1
	// See: consensus/leader_election/module.go#electNextLeaderDeterministicRoundRobin
	leaderId := typesCons.NodeId(2)
	leader := IdToNodeMapping{}
	numReplicas := len(pocketNodes) - 1

	// Debug message to start consensus by triggering next view
	// Get leader out of replica set
	for id, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
		if id == leaderId {
			leader[id] = pocketNode
		} else {
			replicas[id] = pocketNode
		}

		// Right after triggering the next view
		// Consensus started and all nodes are at NewRound step
		step := typesCons.HotstuffStep(pocketNode.GetBus().GetConsensusModule().CurrentStep())
		require.Equal(t, consensus.NewRound, step)
	}

	newRoundMessages, err := WaitForNetworkConsensusEvents(
		t, clockMock, eventsChannel, typesCons.HotstuffStep(consensus.NewRound), typesCons.HotstuffMessageType(consensus.NewRound),
		numReplicas, // We want new round messages from replicas only
		consensusMessageTimeout, false,
	)
	require.NoError(t, err)

	// Broadcast new round messages to leader to build a block
	broadcastMessages(t, newRoundMessages, leader)

	var step typesCons.HotstuffStep
	var pivotTime = 1 * time.Millisecond // Min time it takes to switch from NewRound to Prepare step

	// Give go routines time to trigger
	advanceTime(t, clockMock, 0)

	// We get consensus module from leader to get its POV
	leaderConsensusModule := leader[leaderId].GetBus().GetConsensusModule()

	// Make sure all nodes are aligned to the same leader
	for _, pocketNode := range pocketNodes {
		nodeLeader := pocketNode.GetBus().GetConsensusModule().GetLeaderForView(1, 0, uint8(consensus.NewRound))
		require.Equal(t, typesCons.NodeId(nodeLeader), leaderId)
	}

	// Timer is blocking the proposal step
	step = typesCons.HotstuffStep(leaderConsensusModule.CurrentStep())
	require.Equal(t, consensus.NewRound, step)

	// Advance time right before minBlockTime triggers
	advanceTime(t, clockMock, time.Duration(paceMakerMinBlockTimeMsec*uint64(time.Millisecond))-pivotTime)

	// Should still be blocking proposal step
	step = typesCons.HotstuffStep(leaderConsensusModule.CurrentStep())
	require.Equal(t, consensus.NewRound, step)

	// Advance time just enough to trigger minBlockTime
	advanceTime(t, clockMock, pivotTime)
	step = typesCons.HotstuffStep(leaderConsensusModule.CurrentStep())

	// Time advanced by minBlockTime
	require.Equal(t, uint64(clockMock.Now().UnixMilli()), paceMakerMinBlockTimeMsec)
	// Leader is at proposal step
	require.Equal(t, consensus.Prepare, step)
}

// TODO: Implement these tests and use them as a starting point for new ones. Consider using ChatGPT to help you out :)

func TestPacemakerDifferentHeightsCatchup(t *testing.T) {
	t.Skip()
}

func TestPacemakerDifferentStepsCatchup(t *testing.T) {
	t.Skip()
}

func TestPacemakerDifferentRoundsCatchup(t *testing.T) {
	t.Skip()
}

func TestPacemakerWithLockedQC(t *testing.T) {
	t.Skip()
}

func TestPacemakerWithHighPrepareQC(t *testing.T) {
	t.Skip()
}

func TestPacemakerNoQuorum(t *testing.T) {
	t.Skip()
}

func TestPacemakerNotSafeProposal(t *testing.T) {
	t.Skip()
}

func TestPacemakerExponentialTimeouts(t *testing.T) {
	t.Skip()
}
