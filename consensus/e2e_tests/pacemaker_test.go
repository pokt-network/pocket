package e2e_tests

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

/*
	func TestPacemakerTimeoutIncreasesRound(t *testing.T) {
		// Test preparation
		clockMock := clock.NewMock()
		timeReminder(t, clockMock, time.Second)

		// UnitTestNet configs
		// IMPROVE(#295): Remove time specific suffixes as outlined by go-staticcheck (ST1011)
		paceMakerTimeoutMsec := uint64(10000) // Set a small pacemaker timeout
		paceMakerTimeout := time.Duration(paceMakerTimeoutMsec) * time.Millisecond
		consensusMessageTimeoutMsec := time.Duration(paceMakerTimeoutMsec / 5) // Must be smaller than pacemaker timeout because we expect a deterministic number of consensus messages.
		runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
		for _, runtimeConfig := range runtimeMgrs {
			consCfg := runtimeConfig.GetConfig().Consensus.PacemakerConfig
			consCfg.TimeoutMsec = paceMakerTimeoutMsec
		}
		buses := GenerateBuses(t, runtimeMgrs)

		// Create & start test pocket nodes
		eventsChannel := make(modules.EventsChannel, 100)
		pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
		StartAllTestPocketNodes(t, pocketNodes)

		// Debug message to start consensus by triggering next view
		for _, pocketNode := range pocketNodes {
			TriggerNextView(t, pocketNode)
		}

		// Advance time by an amount shorter than the pacemaker timeout
		advanceTime(t, clockMock, 10*time.Millisecond)

		// Verify consensus started - NewRound messages have an N^2 complexity
		_, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, consensusMessageTimeoutMsec, true)
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

		// Force the pacemaker to time out
		forcePacemakerTimeout(t, clockMock, paceMakerTimeout)

		// Verify that a new round started at the same height
		_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, consensusMessageTimeoutMsec, true)
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

		forcePacemakerTimeout(t, clockMock, paceMakerTimeout)

		// Check that a new round starts at the same height
		_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, consensusMessageTimeoutMsec, true)
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

		forcePacemakerTimeout(t, clockMock, paceMakerTimeout)

		// Check that a new round starts at the same height.
		newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numValidators*numValidators, consensusMessageTimeoutMsec, true)
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
		advanceTime(t, clockMock, 10*time.Millisecond)

		// Confirm we are at the next step (NewRound -> Prepare)
		_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Propose, numValidators, consensusMessageTimeoutMsec, true)
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
		// Test preparation
		clockMock := clock.NewMock()
		timeReminder(t, clockMock, time.Second)

		runtimeConfigs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
		buses := GenerateBuses(t, runtimeConfigs)

		// Create & start test pocket nodes
		eventsChannel := make(modules.EventsChannel, 100)
		pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
		StartAllTestPocketNodes(t, pocketNodes)

		// Starting point
		testHeight := uint64(3)
		testStep := uint8(consensus.NewRound)

		// UnitTestNet configs
		paceMakerTimeoutMsec := uint64(500) // Set a small pacemaker timeout
		runtimeMgrs := GenerateNodeRuntimeMgrs(t, numValidators, clockMock)
		for _, runtimeConfig := range runtimeMgrs {
			runtimeConfig.GetConfig().Consensus.PacemakerConfig.TimeoutMsec = paceMakerTimeoutMsec
		}

		// Prepare leader info
		leaderId := typesCons.NodeId(3)
		require.Equal(t, uint64(leaderId), testHeight%numValidators) // Uses our deterministic round robin leader election
		leaderRound := uint64(6)
		leader := pocketNodes[leaderId]
		consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
		require.NoError(t, err)

		// Placeholder block
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

			// utilityContext is only set on new rounds, which is skipped in this test
			utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(testHeight))
			require.NoError(t, err)
			consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
		}

		// Set the leader to be in the highest round.
		GetConsensusModImpl(pocketNodes[1]).MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(leaderRound - 2)})
		GetConsensusModImpl(pocketNodes[2]).MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(leaderRound - 3)})
		GetConsensusModImpl(pocketNodes[leaderId]).MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(leaderRound)})
		GetConsensusModImpl(pocketNodes[4]).MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(leaderRound - 4)})

		prepareProposal := &typesCons.HotstuffMessage{
			Type:          consensus.Propose,
			Height:        testHeight,
			Step:          consensus.Prepare, // typesCons.HotstuffStep(testStep),
			Round:         leaderRound,
			Block:         block,
			Justification: nil,
		}
		anyMsg, err := anypb.New(prepareProposal)
		require.NoError(t, err)

		P2PBroadcast(t, pocketNodes, anyMsg)

		numExpectedMsgs := numValidators - 1   // -1 because one of the messages is a self proposal (leader to itself as a replica) that is not passed through the network
		msgTimeout := paceMakerTimeoutMsec / 2 // /2 because we do not want the pacemaker to trigger a new timeout
		_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Vote, numExpectedMsgs, time.Duration(msgTimeout), true)
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
			require.Equal(t, uint8(6), nodeState.Round)
			require.Equal(t, leaderId, nodeState.LeaderId)
		}
	}
*/
func forcePacemakerTimeout(t *testing.T, clockMock *clock.Mock, paceMakerTimeout time.Duration) {
	go func() {
		// Cause the pacemaker to timeout
		sleep(t, clockMock, paceMakerTimeout)
	}()
	runtime.Gosched()
	// advance time by an amount longer than the timeout
	advanceTime(t, clockMock, paceMakerTimeout+10*time.Millisecond)
}

// Test node catching up to the leader's height (through state sync) and leader's round (through pacemaker)
func TestPacemakerDifferentHeightsCatchup(t *testing.T) {
	// Test preparation
	clockMock := clock.NewMock()
	timeReminder(t, clockMock, time.Second)

	numberOfValidators := 6
	testHeight := uint64(3)
	testStep := uint8(consensus.NewRound)

	runtimeConfigs := GenerateNodeRuntimeMgrs(t, numberOfValidators, clockMock)

	buses := GenerateBuses(t, runtimeConfigs)

	// Create & start test pocket nodes
	eventsChannel := make(modules.EventsChannel, 100)
	pocketNodes := CreateTestConsensusPocketNodes(t, buses, eventsChannel)
	StartAllTestPocketNodes(t, pocketNodes)

	// UnitTestNet configs
	paceMakerTimeoutMsec := uint64(500) // Set a small pacemaker timeout
	runtimeMgrs := GenerateNodeRuntimeMgrs(t, numberOfValidators, clockMock)
	for _, runtimeConfig := range runtimeMgrs {
		runtimeConfig.GetConfig().Consensus.PacemakerConfig.TimeoutMsec = paceMakerTimeoutMsec
	}

	// Prepare leader info
	leaderId := typesCons.NodeId(3)
	require.Equal(t, uint64(leaderId), testHeight%uint64(numberOfValidators)) // Uses our deterministic round robin leader election
	testRound := uint64(6)
	leader := pocketNodes[leaderId]
	consensusPK, err := leader.GetBus().GetConsensusModule().GetPrivateKey()
	require.NoError(t, err)

	// Prepare unsynched node info
	unsynchedNodeId := typesCons.NodeId(2)
	unSynchedNodeRound := uint64(1)
	unsynchedNodeHeight := uint64(2)

	// Placeholder block
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

	// Set the unsynched node to height (1) and round (1)
	// Set rest of the nodes to the same height (3) and round (6)
	for id, pocketNode := range pocketNodes {
		consensusModImpl := GetConsensusModImpl(pocketNode)
		if id == unsynchedNodeId {
			consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(unsynchedNodeHeight)})
			consensusModImpl.MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(unSynchedNodeRound)})
		} else {
			consensusModImpl.MethodByName("SetHeight").Call([]reflect.Value{reflect.ValueOf(testHeight)})
			consensusModImpl.MethodByName("SetRound").Call([]reflect.Value{reflect.ValueOf(testRound)})
		}
		consensusModImpl.MethodByName("SetStep").Call([]reflect.Value{reflect.ValueOf(testStep)})
		utilityContext, err := pocketNode.GetBus().GetUtilityModule().NewContext(int64(testHeight))
		require.NoError(t, err)
		consensusModImpl.MethodByName("SetUtilityContext").Call([]reflect.Value{reflect.ValueOf(utilityContext)})
	}

	// Debug message to start consensus by triggering first view change
	for _, pocketNode := range pocketNodes {
		TriggerNextView(t, pocketNode)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// Assert that unsynched node has a seperate view of the network than the rest of the nodes
	newRoundMessages, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.NewRound, consensus.Propose, numberOfValidators*numberOfValidators, 250, true)
	require.NoError(t, err)
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == unsynchedNodeId {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: unsynchedNodeHeight,
					Step:   testStep,
					Round:  uint8(unSynchedNodeRound + 1),
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

	// Mock the unsynched node's periodic metadata sync, which is ()
	MockPeriodicMetaDataSynch(testHeight, 1)

	for _, message := range newRoundMessages {
		P2PBroadcast(t, pocketNodes, message)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	// 2. Propose
	numExpectedMsgs := numberOfValidators
	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Propose, numExpectedMsgs, 250, true)
	require.NoError(t, err)

	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)

		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: testHeight,
				Step:   uint8(consensus.Prepare),
				Round:  uint8(testRound + 1),
			},
			nodeState)
	}
}

// TODO: Implement these tests and use them as a starting point for new ones. Consider using ChatGPT to help you out :)

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

/*
	prepareProposal := &typesCons.HotstuffMessage{
		Type:          consensus.Propose,
		Height:        testHeight,
		Step:          consensus.Prepare, // typesCons.HotstuffStep(testStep),
		Round:         leaderRound,
		Block:         block,
		Justification: nil,
	}
	anyMsg, err := anypb.New(prepareProposal)
	require.NoError(t, err)


	P2PBroadcast(t, pocketNodes, anyMsg)
	advanceTime(t, clockMock, 10*time.Millisecond)

	// -2 because one of the messages is a self proposal, and also unsynched node will not accept the proposal
	msgTimeout := paceMakerTimeoutMsec / 2 // /2 because we do not want the pacemaker to trigger a new timeout
	prepareVotes, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.Prepare, consensus.Vote, numExpectedMsgs, time.Duration(msgTimeout), true)
	require.NoError(t, err)

	for _, prepareVote := range prepareVotes {
		P2PSend(t, leader, prepareVote)
		//fmt.Print("PREPARE VOTE: ", prepareVote)
	}
	advanceTime(t, clockMock, 10*time.Millisecond)

	_, err = WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Propose, numExpectedMsgs, 250, true)
	//preCommitProposal, err := WaitForNetworkConsensusEvents(t, clockMock, eventsChannel, consensus.PreCommit, consensus.Propose, numExpectedMsgs, 250, true)
	require.NoError(t, err)

	//t.Log("NUMBER OF RECEIVED MESSAGES: ", len(prepareVotes))

	// Check that all the nodes caught up to the leader's (i.e. the latest) round
	for nodeId, pocketNode := range pocketNodes {
		nodeState := GetConsensusNodeState(pocketNode)
		if nodeId == unsynchedNodeId {
			require.Equal(t, uint64(1), nodeState.Height)
			require.Equal(t, uint8(1), nodeState.Round)
		} else {
			// if nodeId == leaderId {
			// 	require.Equal(t, consensus.Prepare.String(), typesCons.HotstuffStep(nodeState.Step).String())
			// } else {
			require.Equal(t, consensus.PreCommit.String(), typesCons.HotstuffStep(nodeState.Step).String())
			//	}
			require.Equal(t, uint64(3), nodeState.Height)
			require.Equal(t, uint8(6), nodeState.Round)
			require.Equal(t, leaderId, nodeState.LeaderId)
		}
	}

*/
