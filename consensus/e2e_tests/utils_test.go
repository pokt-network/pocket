package e2e_tests

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"

	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/internal/testutil"
	ibcUtils "github.com/pokt-network/pocket/internal/testutil/ibc"
	persistenceMocks "github.com/pokt-network/pocket/persistence/types/mocks"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/defaults"
	"github.com/pokt-network/pocket/runtime/genesis"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/pokt-network/pocket/state_machine"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// TECHDEBT: Constants in the `e2e_tests` test suite that should be parameterized
const (
	numValidators   = 4    // The number of validators in the testing network created
	dummyStateHash  = "42" // The state hash returned for all committed blocks
	numMockedBlocks = 200  // The number of mocked blocks in in memory for testing purposes
)

var maxTxBytes = defaults.DefaultConsensusMaxMempoolBytes

type idToNodeMapping map[typesCons.NodeId]*shared.Node
type idToPrivKeyMapping map[typesCons.NodeId]cryptoPocket.PrivateKey

/*** Node Generation Helpers ***/

//nolint:unparam // validatorCount will be varied in the future
func generateNodeRuntimeMgrs(t *testing.T, validatorCount int, clockMgr clock.Clock) []*runtime.Manager {
	t.Helper()

	runtimeMgrs := make([]*runtime.Manager, validatorCount)
	genesisState, validatorKeys := test_artifacts.NewGenesisState(validatorCount, 1, 1, 1)
	cfgs := test_artifacts.NewDefaultConfigs(validatorKeys)
	for i, config := range cfgs {
		config.Consensus = &configs.ConsensusConfig{
			PrivateKey:      config.PrivateKey,
			MaxMempoolBytes: maxTxBytes,
			PacemakerConfig: &configs.PacemakerConfig{
				TimeoutMsec:               10000,
				Manual:                    false,
				DebugTimeBetweenStepsMsec: 0,
			},
			ServerModeEnabled: true,
		}
		runtimeMgrs[i] = runtime.NewManager(config, genesisState, runtime.WithClock(clockMgr))
	}
	return runtimeMgrs
}

func createTestConsensusPocketNodes(
	t *testing.T,
	buses []modules.Bus,
	sharedNetworkChannel modules.EventsChannel,
) (pocketNodes idToNodeMapping) {
	pocketNodes = make(idToNodeMapping, len(buses))
	// TECHDEBT: The order here is important in order for NodeIds to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(buses, func(i, j int) bool {
		pk, err := cryptoPocket.NewPrivateKey(buses[i].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		pk2, err := cryptoPocket.NewPrivateKey(buses[j].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		return pk.Address().String() < pk2.Address().String()
	})

	blocks := &testingBlocks{}

	validatorPrivKeys := make(idToPrivKeyMapping, len(buses))
	for i, bus := range buses {
		nodeId := typesCons.NodeId(i + 1)

		pocketNode := createTestConsensusPocketNode(t, bus, sharedNetworkChannel, blocks)
		pocketNodes[nodeId] = pocketNode

		validatorPrivKey, err := cryptoPocket.NewPrivateKey(pocketNode.GetBus().GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)

		validatorPrivKeys[nodeId] = validatorPrivKey
	}
	blocks.preparePlaceholderBlocks(t, buses[0], validatorPrivKeys, numMockedBlocks)
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func createTestConsensusPocketNode(
	t *testing.T,
	bus modules.Bus,
	sharedNetworkChannel modules.EventsChannel,
	placeholderBlocks *testingBlocks,
) *shared.Node {
	persistenceMock := basePersistenceMock(t, sharedNetworkChannel, bus, placeholderBlocks)
	bus.RegisterModule(persistenceMock)

	consensusMod, err := consensus.Create(bus)
	require.NoError(t, err)
	consensusModule, ok := consensusMod.(modules.ConsensusModule)
	require.True(t, ok)

	_, err = state_machine.Create(bus)
	require.NoError(t, err)

	runtimeMgr := (bus).GetRuntimeMgr()
	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	p2pMock := baseP2PMock(t, sharedNetworkChannel)
	utilityMock := baseUtilityMock(t, sharedNetworkChannel, runtimeMgr.GetGenesis(), consensusModule)
	telemetryMock := baseTelemetryMock(t, sharedNetworkChannel)
	loggerMock := baseLoggerMock(t, sharedNetworkChannel)
	rpcMock := baseRpcMock(t, sharedNetworkChannel)
	ibcMock, hostMock := ibcUtils.IBCMockWithHost(t, bus)
	bus.RegisterModule(hostMock)

	for _, module := range []modules.Module{
		p2pMock,
		utilityMock,
		telemetryMock,
		loggerMock,
		rpcMock,
		ibcMock,
	} {
		bus.RegisterModule(module)
	}

	require.NoError(t, err)

	pk, err := cryptoPocket.NewPrivateKey(runtimeMgr.GetConfig().PrivateKey)
	require.NoError(t, err)

	pocketNode := shared.NewNodeWithP2PAddress(pk.Address())

	pocketNode.SetBus(bus)

	return pocketNode
}

func generateBuses(t *testing.T, runtimeMgrs []*runtime.Manager, opts ...modules.BusOption) (buses []modules.Bus) {
	buses = make([]modules.Bus, len(runtimeMgrs))
	for i := range runtimeMgrs {
		bus, err := runtime.CreateBus(runtimeMgrs[i], opts...)
		require.NoError(t, err)
		buses[i] = bus
	}
	return
}

func startAllTestPocketNodes(t *testing.T, pocketNodes idToNodeMapping) error {
	for _, pocketNode := range pocketNodes {
		go startNode(t, pocketNode)
		stateMachine := pocketNode.GetBus().GetStateMachineModule()
		if err := stateMachine.SendEvent(coreTypes.StateMachineEvent_Start); err != nil {
			return err
		}
		if err := stateMachine.SendEvent(coreTypes.StateMachineEvent_P2P_IsBootstrapped); err != nil {
			return err
		}
	}
	return nil
}

/*** Node Visibility/Reflection Helpers ***/

// HACK: Look for ways to avoid using reflections in the testing package. It was a quick & dirty way to keep going.
func getConsensusNodeState(node *shared.Node) typesCons.ConsensusNodeState {
	return getConsensusModImpl(node).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(typesCons.ConsensusNodeState)
}

func getConsensusModImpl(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule())
}

/*** Debug/Development Message Helpers ***/

func triggerDebugMessage(t *testing.T, node *shared.Node, action messaging.DebugMessageAction) {
	debugMessage := &messaging.DebugMessage{
		Action:  action,
		Message: nil,
	}
	anyProto, err := anypb.New(debugMessage)
	require.NoError(t, err)

	e := &messaging.PocketEnvelope{Content: anyProto}
	node.GetBus().PublishEventToBus(e)
}

/*** P2P Helpers ***/

func broadcast(t *testing.T, nodes idToNodeMapping, any *anypb.Any) {
	t.Helper()

	e := &messaging.PocketEnvelope{Content: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func send(t *testing.T, node *shared.Node, any *anypb.Any) {
	t.Helper()

	e := &messaging.PocketEnvelope{Content: any}
	node.GetBus().PublishEventToBus(e)
}

// This is a helper for `waitForEventsInternal` that creates the `includeFilter` function based on
// consensus specific parameters.
// failOnExtraMessages:
//
//	This flag is useful when running the consensus unit tests. It causes the test to wait up to the
//	maximum delay specified in the source code and errors if additional unexpected messages are received.
//	For example, if the test expects to receive 5 messages within 2 seconds:
//		false: continue if 5 messages are received in 0.5 seconds
//		true: wait for another 1.5 seconds after 5 messages are received in 0.5 seconds, and fail if any additional messages are received.
func waitForNetworkConsensusEvents(
	t *testing.T,
	clck *clock.Mock,
	sharedNetworkChannel modules.EventsChannel,
	step typesCons.HotstuffStep,
	msgType typesCons.HotstuffMessageType,
	numExpectedMsgs int,
	millis time.Duration,
	failOnExtraMessages bool,
) (messages []*anypb.Any, err error) {
	includeFilter := func(anyMsg *anypb.Any) bool {
		msg, err := codec.GetCodec().FromAny(anyMsg)
		require.NoError(t, err)

		hotstuffMessage, ok := msg.(*typesCons.HotstuffMessage)
		require.True(t, ok)

		return hotstuffMessage.Type == msgType && hotstuffMessage.Step == step
	}

	errMsg := fmt.Sprintf("HotStuff step: %s, type: %s", typesCons.HotstuffStep_name[int32(step)], typesCons.HotstuffMessageType_name[int32(msgType)])
	return waitForEventsInternal(clck, sharedNetworkChannel, messaging.HotstuffMessageContentType, numExpectedMsgs, millis, includeFilter, errMsg, failOnExtraMessages)
}

// IMPROVE: Consider unifying this function with WaitForNetworkConsensusEvents
// This is a helper for 'waitForEventsInternal' that creates the `includeFilter` function based on state sync message specific parameters.
//
//nolint:unparam // failOnExtraMessages will be varied in the future
func waitForNetworkStateSyncEvents(
	t *testing.T,
	clck *clock.Mock,
	sharedNetworkChannel modules.EventsChannel,
	errMsg string,
	numExpectedMsgs int,
	maxWaitTime time.Duration,
	failOnExtraMessages bool,
	stateSyncMsgType any,
) (messages []*anypb.Any, err error) {
	includeFilter := func(anyMsg *anypb.Any) bool {
		msg, err := codec.GetCodec().FromAny(anyMsg)
		require.NoError(t, err)

		stateSyncMsg, ok := msg.(*typesCons.StateSyncMessage)
		require.True(t, ok)

		if stateSyncMsgType != nil {
			return reflect.TypeOf(stateSyncMsg.Message) == stateSyncMsgType
		}
		return true
	}

	return waitForEventsInternal(clck, sharedNetworkChannel, messaging.StateSyncMessageContentType, numExpectedMsgs, maxWaitTime, includeFilter, errMsg, failOnExtraMessages)
}

// RESEARCH(#462): Research ways to eliminate time-based non-determinism from the test framework
// IMPROVE: This function can be extended to testing events outside of just the consensus module.
func waitForEventsInternal(
	clck *clock.Mock,
	sharedNetworkChannel modules.EventsChannel,
	eventContentType string,
	numExpectedMsgs int,
	maxWaitTime time.Duration,
	msgIncludeFilter func(m *anypb.Any) bool,
	errMsg string,
	failOnExtraMessages bool,
) (expectedMsgs []*anypb.Any, err error) {
	expectedMsgs = make([]*anypb.Any, 0)                 // Aggregate and return the messages we're waiting for
	unusedEvents := make([]*messaging.PocketEnvelope, 0) // "Recycle" events back into the events channel if we're not using them

	// Limit the amount of time we're waiting for the messages to be published on the events channel
	ctx, cancel := clck.WithTimeout(context.TODO(), time.Millisecond*maxWaitTime)
	defer cancel()

	// Since the tests use a mock clock, we need to manually advance the clock to trigger the timeout
	ticker := time.NewTicker(time.Millisecond)
	tickerDone := make(chan bool)
	go func() {
		for {
			select {
			case <-tickerDone:
				return
			case <-ticker.C:
				clck.Add(time.Millisecond)
			}
		}
	}()
	defer ticker.Stop()
	defer func() {
		tickerDone <- true
	}()

	numRemainingMsgs := numExpectedMsgs
loop:
	for {
		select {
		case nodeEvent := <-sharedNetworkChannel:
			if nodeEvent.GetContentType() != eventContentType {
				unusedEvents = append(unusedEvents, nodeEvent)
				continue
			}

			message := nodeEvent.Content
			if message == nil || !msgIncludeFilter(message) {
				unusedEvents = append(unusedEvents, nodeEvent)
				continue
			}

			expectedMsgs = append(expectedMsgs, message)
			numRemainingMsgs--
			// Break if both of the following are true:
			// 1. We are not expecting any more messages
			// 2. We do not want to fail in the case of extra unexpected messages that pass the filter
			if numRemainingMsgs == 0 && !failOnExtraMessages {
				break loop
			}
		// The reason we return we format and return an error message rather than using t.Fail(...)
		// is to allow the called to run `require.NoError(t, err)` and have the output point to the
		// specific line where the test failed.
		case <-ctx.Done():
			if numRemainingMsgs == 0 {
				break loop
			} else if numRemainingMsgs > 0 {
				return expectedMsgs, fmt.Errorf("Missing '%s' messages; %d expected but %d received. (%s) \n\t !!!IMPORTANT(#462)!!!: Consider increasing `maxWaitTime` as a workaround", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			} else {
				return expectedMsgs, fmt.Errorf("Too many '%s' messages; %d expected but %d received. (%s)", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			}
		}
	}

	for _, u := range unusedEvents {
		sharedNetworkChannel <- u
	}
	return
}

/*** Module Mocking Helpers ***/

// Creates a persistence module mock with mock implementations of some basic functionality
func basePersistenceMock(t *testing.T, _ modules.EventsChannel, bus modules.Bus, testBlocks *testingBlocks) *mockModules.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()
	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()

	blockStoreMock := persistenceMocks.NewMockBlockStore(ctrl)
	blockStoreMock.EXPECT().Get(gomock.Any()).DoAndReturn(func(height []byte) ([]byte, error) {
		heightInt := utils.HeightFromBytes(height)
		if bus.GetConsensusModule().CurrentHeight() < heightInt {
			return nil, fmt.Errorf("requested height is higher than current height of the node's consensus module")
		}
		return codec.GetCodec().Marshal(testBlocks.getBlock(heightInt))
	}).AnyTimes()
	blockStoreMock.
		// NB: The business logic in this mock and below is vital for testing state-sync end-to-end
		EXPECT().
		GetBlock(gomock.Any()).
		DoAndReturn(func(height uint64) (*coreTypes.Block, error) {
			if bus.GetConsensusModule().CurrentHeight() < height {
				return nil, fmt.Errorf("requested height is higher than current height of the node's consensus module")
			}
			return testBlocks.getBlock(height), nil
		}).
		AnyTimes()

	persistenceReadContextMock.EXPECT().GetMaximumBlockHeight().DoAndReturn(func() (uint64, error) {
		// Check that we are retrieving a block at a height that was mocked by our test suite
		if int(bus.GetConsensusModule().CurrentHeight()) <= len(testBlocks.blocks) {
			return bus.GetConsensusModule().CurrentHeight() - 1, nil
		}
		t.Error("Trying to retrieve a block at a height that was not mocked.")
		return 0, nil
	}).AnyTimes()

	persistenceMock.
		EXPECT().
		GetBlockStore().
		Return(blockStoreMock).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		GetMaximumBlockHeight().
		DoAndReturn(func() (uint64, error) {
			height := bus.GetConsensusModule().CurrentHeight()
			return height, nil
		}).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		GetMinimumBlockHeight().
		DoAndReturn(func() (uint64, error) {
			// mock minimum block height in persistence module to 1 if current height is equal or more than 1, else return 0 as the minimum height
			if bus.GetConsensusModule().CurrentHeight() >= 1 {
				return 1, nil
			}
			return 0, nil
		}).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		GetAllValidators(gomock.Any()).
		Return(bus.GetRuntimeMgr().GetGenesis().Validators, nil).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		GetAllStakedActors(gomock.Any()).
		DoAndReturn(func(height int64) ([]*coreTypes.Actor, error) {
			genesisState := bus.GetRuntimeMgr().GetGenesis()
			return testutil.Concatenate[*coreTypes.Actor](
				genesisState.Validators,
				genesisState.Servicers,
				genesisState.Fishermen,
				genesisState.Applications,
			), nil
		}).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		GetBlockHash(gomock.Any()).
		Return("", nil).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		Release().
		AnyTimes()

	persistenceReadContextMock.EXPECT().GetValidatorExists(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, sharedNetworkChannel modules.EventsChannel) *mockModules.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := mockModules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			sharedNetworkChannel <- e
		}).
		AnyTimes()
	// CONSIDERATION: Adding a check to not to send message to itself
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			sharedNetworkChannel <- e
		}).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()
	p2pMock.EXPECT().HandleEvent(gomock.Any()).Return(nil).AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel, genesisState *genesis.GenesisState, consensusMod modules.ConsensusModule) *mockModules.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := mockModules.NewMockUtilityModule(ctrl)
	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	utilityMock.EXPECT().
		NewUnitOfWork(gomock.Any()).
		DoAndReturn(
			// mimicking the behavior of the utility module's NewUnitOfWork method
			func(height int64) (modules.UtilityUnitOfWork, error) {
				if consensusMod.IsLeader() {
					return baseLeaderUtilityUnitOfWorkMock(t, genesisState), nil
				}
				return baseReplicaUtilityUnitOfWorkMock(t, genesisState), nil
			}).
		AnyTimes()

	utilityMock.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()

	return utilityMock
}

func baseLeaderUtilityUnitOfWorkMock(t *testing.T, genesisState *genesis.GenesisState) *mockModules.MockLeaderUtilityUnitOfWork {
	ctrl := gomock.NewController(t)
	utilityLeaderUnitOfWorkMock := mockModules.NewMockLeaderUtilityUnitOfWork(ctrl)

	rwContextMock := mockModules.NewMockPersistenceRWContext(ctrl)
	rwContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	rwContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	rwContextMock.EXPECT().Release().AnyTimes()

	utilityLeaderUnitOfWorkMock.EXPECT().
		CreateProposalBlock(gomock.Any(), maxTxBytes).
		Return(dummyStateHash, make([][]byte, 0), nil).
		AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().
		ApplyBlock().
		Return(nil).
		AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().
		GetStateHash().
		Return(dummyStateHash).
		AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().SetProposalBlock(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	utilityLeaderUnitOfWorkMock.EXPECT().Release().Return(nil).AnyTimes()

	return utilityLeaderUnitOfWorkMock
}

func baseReplicaUtilityUnitOfWorkMock(t *testing.T, genesisState *genesis.GenesisState) *mockModules.MockReplicaUtilityUnitOfWork {
	ctrl := gomock.NewController(t)
	utilityReplicaUnitOfWorkMock := mockModules.NewMockReplicaUtilityUnitOfWork(ctrl)

	rwContextMock := mockModules.NewMockPersistenceRWContext(ctrl)
	rwContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
	rwContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()
	rwContextMock.EXPECT().Release().AnyTimes()

	utilityReplicaUnitOfWorkMock.EXPECT().
		ApplyBlock().
		Return(nil).
		AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().
		GetStateHash().
		Return(dummyStateHash).
		AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().SetProposalBlock(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	utilityReplicaUnitOfWorkMock.EXPECT().Release().Return(nil).AnyTimes()

	return utilityReplicaUnitOfWorkMock
}

func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := mockModules.NewMockTelemetryModule(ctrl)
	timeSeriesAgentMock := baseTelemetryTimeSeriesAgentMock(t)
	eventMetricsAgentMock := baseTelemetryEventMetricsAgentMock(t)

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	return telemetryMock
}

func baseRpcMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockRPCModule {
	ctrl := gomock.NewController(t)
	rpcMock := mockModules.NewMockRPCModule(ctrl)
	rpcMock.EXPECT().Start().Return(nil).AnyTimes()
	rpcMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	rpcMock.EXPECT().GetModuleName().Return(modules.RPCModuleName).AnyTimes()

	return rpcMock
}

func WaitForNextBlock(
	t *testing.T,
	clck *clock.Mock,
	sharedNetworkChannel modules.EventsChannel,
	pocketNodes idToNodeMapping,
	height uint64,
	round uint8,
	maxWaitTime time.Duration,
	failOnExtraMessages bool,
) *coreTypes.Block {
	leaderId := typesCons.NodeId(pocketNodes[1].GetBus().GetConsensusModule().GetLeaderForView(height, uint64(round), uint8(consensus.NewRound)))

	// Debug message to start consensus by triggering first view change
	triggerNextView(t, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// 1. NewRound
	newRoundMessages, err := waitForProposalMsgs(t, clck, sharedNetworkChannel, pocketNodes, height, uint8(consensus.NewRound), round, 0, numValidators*numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, newRoundMessages, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// 2. Prepare
	prepareProposals, err := waitForProposalMsgs(t, clck, sharedNetworkChannel, pocketNodes, height, uint8(consensus.Prepare), round, leaderId, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, prepareProposals, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// wait for prepare votes
	prepareVotes, err := waitForNetworkConsensusEvents(t, clck, sharedNetworkChannel, 2, consensus.Vote, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, prepareVotes, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// 3. PreCommit
	preCommitProposals, err := waitForProposalMsgs(t, clck, sharedNetworkChannel, pocketNodes, height, uint8(consensus.PreCommit), round, leaderId, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, preCommitProposals, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// wait for preCommit votes
	preCommitVotes, err := waitForNetworkConsensusEvents(t, clck, sharedNetworkChannel, 3, consensus.Vote, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, preCommitVotes, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// 4. Commit
	commitProposals, err := waitForProposalMsgs(t, clck, sharedNetworkChannel, pocketNodes, height, uint8(consensus.Commit), round, leaderId, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, commitProposals, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// wait for commit votes
	commitVotes, err := waitForNetworkConsensusEvents(t, clck, sharedNetworkChannel, 4, consensus.Vote, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, commitVotes, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	// 5. Decide
	decideProposals, err := waitForProposalMsgs(t, clck, sharedNetworkChannel, pocketNodes, height, uint8(consensus.Decide), round, leaderId, numValidators, maxWaitTime, failOnExtraMessages)
	require.NoError(t, err)
	broadcastMessages(t, decideProposals, pocketNodes)
	advanceTime(t, clck, 10*time.Millisecond)

	blockStore := pocketNodes[1].GetBus().GetPersistenceModule().GetBlockStore()
	block, err := blockStore.GetBlock(height)
	require.NoError(t, err)

	return block
}

func waitForProposalMsgs(
	t *testing.T,
	clck *clock.Mock,
	sharedNetworkChannel modules.EventsChannel,
	pocketNodes idToNodeMapping,
	height uint64,
	step uint8,
	round uint8,
	leaderId typesCons.NodeId,
	numExpectedMsgs int,
	maxWaitTime time.Duration,
	failOnExtraMessages bool,
) ([]*anypb.Any, error) {
	proposalMsgs, err := waitForNetworkConsensusEvents(t, clck, sharedNetworkChannel, typesCons.HotstuffStep(step), consensus.Propose, numExpectedMsgs, maxWaitTime, failOnExtraMessages)
	if err != nil {
		return nil, err
	}

	for nodeId, pocketNode := range pocketNodes {
		nodeState := getConsensusNodeState(pocketNode)
		if (typesCons.HotstuffStep(step) == consensus.Decide) && (nodeId == leaderId) {
			assertNodeConsensusView(t, nodeId,
				typesCons.ConsensusNodeState{
					Height: height + 1,
					Step:   1,
					Round:  round,
				},
				nodeState)
			require.Equal(t, typesCons.NodeId(0), nodeState.LeaderId, "Leader should be empty")
			continue
		}
		assertNodeConsensusView(t, nodeId,
			typesCons.ConsensusNodeState{
				Height: height,
				Step:   step,
				Round:  round,
			},
			nodeState)
		require.Equal(t, leaderId, nodeState.LeaderId, fmt.Sprintf("%d should be the current leader", leaderId))
	}
	return proposalMsgs, nil
}

func broadcastMessages(t *testing.T, msgs []*anypb.Any, pocketNodes idToNodeMapping) {
	for _, message := range msgs {
		broadcast(t, pocketNodes, message)
	}
}

func triggerNextView(t *testing.T, pocketNodes idToNodeMapping) {
	for _, node := range pocketNodes {
		triggerDebugMessage(t, node, messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW)
	}
}

func generatePlaceholderBlock(height uint64, leaderAddrr cryptoPocket.Address) *coreTypes.Block {
	blockHeader := &coreTypes.BlockHeader{
		Height:            height,
		StateHash:         dummyStateHash,
		PrevStateHash:     "",
		ProposerAddress:   leaderAddrr,
		QuorumCertificate: nil,
	}
	return &coreTypes.Block{
		BlockHeader:  blockHeader,
		Transactions: make([][]byte, 0),
	}
}

func baseTelemetryTimeSeriesAgentMock(t *testing.T) *mockModules.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := mockModules.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).MaxTimes(1)
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

func baseTelemetryEventMetricsAgentMock(t *testing.T) *mockModules.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := mockModules.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

func baseLoggerMock(t *testing.T, _ modules.EventsChannel) *mockModules.MockLoggerModule {
	ctrl := gomock.NewController(t)
	loggerMock := mockModules.NewMockLoggerModule(ctrl)

	loggerMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	loggerMock.EXPECT().GetModuleName().Return(modules.LoggerModuleName).AnyTimes()

	return loggerMock
}

/*** Placeholder Block Generation Helpers ***/

type testingBlocks struct {
	blocks []*coreTypes.Block
}

func (p *testingBlocks) getBlock(index uint64) *coreTypes.Block {
	// returning block at index-1, because block 1 is stored at index 0 of the blocks array
	return p.blocks[index-1]
}

func (p *testingBlocks) preparePlaceholderBlocks(t *testing.T, bus modules.Bus, validatorPrivKeys idToPrivKeyMapping, numMockedBlocks uint64) {
	t.Helper()
	for i := uint64(1); i <= numMockedBlocks; i++ {
		leaderId := bus.GetConsensusModule().GetLeaderForView(i, uint64(0), uint8(consensus.NewRound))
		leaderPivKey := validatorPrivKeys[typesCons.NodeId(leaderId)]

		// Construct the block
		blockHeader := &coreTypes.BlockHeader{
			Height:            i,
			StateHash:         dummyStateHash,
			PrevStateHash:     dummyStateHash,
			ProposerAddress:   leaderPivKey.Address(),
			QuorumCertificate: nil, // inserted below
		}
		block := &coreTypes.Block{
			BlockHeader:  blockHeader,
			Transactions: make([][]byte, 0), // we don't care about the transactions in this context
		}

		qc := generateQuorumCertificate(t, validatorPrivKeys, block)
		qcBytes, err := codec.GetCodec().Marshal(qc)
		require.NoError(t, err)

		block.BlockHeader.QuorumCertificate = qcBytes

		p.blocks = append(p.blocks, block)
	}
}

/*** Quorum certificate Generation Helpers ***/

func generateQuorumCertificate(t *testing.T, validatorPrivKeys idToPrivKeyMapping, block *coreTypes.Block) *typesCons.QuorumCertificate {
	// Aggregate partial signatures
	var pss []*typesCons.PartialSignature
	for _, validatorPrivKey := range validatorPrivKeys {
		pss = append(pss, generatePartialSignature(t, block, validatorPrivKey))
	}

	// Generate threshold signature
	thresholdSig := new(typesCons.ThresholdSignature)
	thresholdSig.Signatures = make([]*typesCons.PartialSignature, len(pss))
	copy(thresholdSig.Signatures, pss)

	return &typesCons.QuorumCertificate{
		Height:             block.BlockHeader.Height,
		Round:              1, // assume everything succeeds on the first round for now
		Step:               consensus.NewRound,
		Block:              block,
		ThresholdSignature: thresholdSig,
	}
}

// generate partial signature for the validator
func generatePartialSignature(t *testing.T, block *coreTypes.Block, validatorPrivKey cryptoPocket.PrivateKey) *typesCons.PartialSignature {
	return &typesCons.PartialSignature{
		Signature: getMessageSignature(t, block, validatorPrivKey),
		Address:   validatorPrivKey.PublicKey().Address().String(),
	}
}

// Generates partial signature with given private key
func getMessageSignature(t *testing.T, block *coreTypes.Block, privKey cryptoPocket.PrivateKey) []byte {
	// Signature only over subset of fields in HotstuffMessage
	// For reference, see section 4.3 of the the hotstuff whitepaper, partial signatures are
	// computed over `tsignr(hm.type, m.viewNumber , m.nodei)`. https://arxiv.org/pdf/1803.05069.pdf
	msgToSign := &typesCons.HotstuffMessage{
		Height: block.BlockHeader.Height,
		Step:   1,
		Round:  1,
		Block:  block,
	}

	bytesToSign, err := codec.GetCodec().Marshal(msgToSign)
	require.NoError(t, err)

	signature, err := privKey.Sign(bytesToSign)
	require.NoError(t, err)

	return signature
}

func logTime(t *testing.T, clck *clock.Mock) {
	t.Helper()
	defer func() {
		// this is to recover from a panic that could happen if the goroutine tries to log after the test has finished
		// cause of the panic: https://github.com/golang/go/blob/135c470b2277e1c9514ba8a5478408fea0dee8a2/src/testing/testing.go#L1003
		//
		// spotted for the first time in our CI: https://github.com/pokt-network/pocket/actions/runs/4198025819/jobs/7281103860#step:8:1118
		//nolint:errcheck // ignoring completely
		recover()
	}()
	t.Logf("[⌚ CLOCK ⌚] the time is: %v ms from UNIX Epoch [%v]", clck.Now().UTC().UnixMilli(), clck.Now().UTC())
}

// advanceTime moves the time forward on the mock clock and logs what just happened.
func advanceTime(t *testing.T, clck *clock.Mock, duration time.Duration) {
	t.Helper()
	clck.Add(duration)
	t.Logf("[⌚ CLOCK ⏩] advanced by %v", duration)
	logTime(t, clck)
}

// sleep pauses the goroutine for the given duration on the mock clock and logs what just happened.
//
// Note: time has to be moved forward in a separate goroutine, see `advanceTime`.
func sleep(t *testing.T, clck *clock.Mock, duration time.Duration) {
	t.Helper()
	t.Logf("[⌚ CLOCK 💤] sleeping for %v", duration)
	clck.Sleep(duration)
}

// timeReminder simply prints, at a given interval and in a separate goroutine, the current mocked time to help with events.
// nolint:unparam // we want to keep the frequency as a parameter for clarity
func timeReminder(t *testing.T, clck *clock.Mock, frequency time.Duration) {
	go func() {
		tick := time.NewTicker(frequency)
		for {
			<-tick.C
			logTime(t, clck)
		}
	}()
}

func assertNodeConsensusView(t *testing.T, nodeId typesCons.NodeId, expected, actual typesCons.ConsensusNodeState) {
	assertHeight(t, nodeId, expected.Height, actual.Height)
	assertStep(t, nodeId, typesCons.HotstuffStep(expected.Step), typesCons.HotstuffStep(actual.Step))
	assertRound(t, nodeId, expected.Round, actual.Round)
}

func assertHeight(t *testing.T, nodeId typesCons.NodeId, expected, actual uint64) {
	require.Equal(t, expected, actual, "[NODE][%v] failed assertHeight", nodeId)
}

func assertStep(t *testing.T, nodeId typesCons.NodeId, expected, actual typesCons.HotstuffStep) {
	require.Equal(t, expected.String(), actual.String(), "[NODE][%v] failed assertStep", nodeId)
}

func assertRound(t *testing.T, nodeId typesCons.NodeId, expected, actual uint8) {
	require.Equal(t, expected, actual, "[NODE][%v] failed assertRound", nodeId)
}

func startNode(t *testing.T, pocketNode *shared.Node) {
	err := pocketNode.Start()
	require.NoError(t, err)
}

func prepareStateSyncGetBlockMessage(t *testing.T, peerAddress string, height uint64) *anypb.Any {
	t.Helper()

	stateSyncGetBlockMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_GetBlockReq{
			GetBlockReq: &typesCons.GetBlockRequest{
				PeerAddress: peerAddress,
				Height:      height,
			},
		},
	}

	anyProto, err := anypb.New(stateSyncGetBlockMessage)
	require.NoError(t, err)

	return anyProto
}

func prepareStateSyncGetMetadataMessage(t *testing.T, selfAddress string) *anypb.Any {
	t.Helper()

	stateSyncMetaDataReqMessage := &typesCons.StateSyncMessage{
		Message: &typesCons.StateSyncMessage_MetadataReq{
			MetadataReq: &typesCons.StateSyncMetadataRequest{
				PeerAddress: selfAddress,
			},
		},
	}
	anyProto, err := anypb.New(stateSyncMetaDataReqMessage)
	require.NoError(t, err)

	return anyProto
}
