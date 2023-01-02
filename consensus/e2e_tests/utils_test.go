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
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/codec"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

// TODO(integration): These are temporary variables used in the prototype integration phase that
// will need to be parameterized later once the test framework design matures.
const (
	numValidators = 4
	stateHash     = "42"
	maxTxBytes    = 90000
)

type IdToNodeMapping map[typesCons.NodeId]*shared.Node

/*** Node Generation Helpers ***/

func GenerateNodeRuntimeMgrs(_ *testing.T, validatorCount int, clockMgr clock.Clock) []runtime.Manager {
	runtimeMgrs := make([]runtime.Manager, validatorCount)
	var validatorKeys []string
	genesisState, validatorKeys := test_artifacts.NewGenesisState(validatorCount, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(validatorKeys)
	for i, config := range configs {
		runtime.WithConsensusConfig(&typesCons.ConsensusConfig{
			PrivateKey:      config.GetBaseConfig().GetPrivateKey(),
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &typesCons.PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    false,
				DebugTimeBetweenStepsMsec: 0,
			},
		})(config)
		runtimeMgrs[i] = *runtime.NewManager(config, genesisState, runtime.WithClock(clockMgr))
	}
	return runtimeMgrs
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	runtimeMgrs []runtime.Manager,
	eventsChannel modules.EventsChannel,
) (pocketNodes IdToNodeMapping) {
	pocketNodes = make(IdToNodeMapping, len(runtimeMgrs))
	// TODO(design): The order here is important in order for NodeId to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(runtimeMgrs, func(i, j int) bool {
		pk, err := cryptoPocket.NewPrivateKey(runtimeMgrs[i].GetConfig().GetBaseConfig().GetPrivateKey())
		require.NoError(t, err)
		pk2, err := cryptoPocket.NewPrivateKey(runtimeMgrs[j].GetConfig().GetBaseConfig().GetPrivateKey())
		require.NoError(t, err)
		return pk.Address().String() < pk2.Address().String()
	})
	for i, runtimeMgr := range runtimeMgrs {
		pocketNode := CreateTestConsensusPocketNode(t, &runtimeMgr, eventsChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

func CreateTestConsensusPocketNodesNew(
	t *testing.T,
	runtimeMgrs []runtime.Manager,
	eventsChannel modules.EventsChannel,
) (pocketNodes IdToNodeMapping) {
	pocketNodes = make(IdToNodeMapping, len(runtimeMgrs))
	// TODO(design): The order here is important in order for NodeId to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(runtimeMgrs, func(i, j int) bool {
		pk, err := cryptoPocket.NewPrivateKey(runtimeMgrs[i].GetConfig().GetBaseConfig().GetPrivateKey())
		require.NoError(t, err)
		pk2, err := cryptoPocket.NewPrivateKey(runtimeMgrs[j].GetConfig().GetBaseConfig().GetPrivateKey())
		require.NoError(t, err)
		return pk.Address().String() < pk2.Address().String()
	})
	for i, runtimeMgr := range runtimeMgrs {
		pocketNode := CreateTestConsensusPocketNode(t, &runtimeMgr, eventsChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func CreateTestConsensusPocketNode(
	t *testing.T,
	runtimeMgr *runtime.Manager,
	eventsChannel modules.EventsChannel,
) *shared.Node {
	consensusMod, err := consensus.Create(runtimeMgr)
	require.NoError(t, err)
	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	persistenceMock := basePersistenceMock(t, eventsChannel)
	p2pMock := baseP2PMock(t, eventsChannel)
	utilityMock := baseUtilityMock(t, eventsChannel, runtimeMgr.GetGenesis())
	telemetryMock := baseTelemetryMock(t, eventsChannel)
	loggerMock := baseLoggerMock(t, eventsChannel)
	rpcMock := baseRpcMock(t, eventsChannel)

	bus, err := shared.CreateBus(runtimeMgr, persistenceMock, p2pMock, utilityMock, consensusMod.(modules.ConsensusModule), telemetryMock, loggerMock, rpcMock)

	require.NoError(t, err)

	pk, err := cryptoPocket.NewPrivateKey(runtimeMgr.GetConfig().GetBaseConfig().GetPrivateKey())
	require.NoError(t, err)

	pocketNode := shared.NewNodeWithP2PAddress(pk.Address())

	pocketNode.SetBus(bus)

	return pocketNode
}

// CLEANUP: Reduce package scope visibility in the consensus test module
func StartAllTestPocketNodes(t *testing.T, pocketNodes IdToNodeMapping) {
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
		startEvent := pocketNode.GetBus().GetBusEvent()
		require.Equal(t, startEvent.GetContentType(), messaging.NodeStartedEventType)
	}
}

/*** Node Visibility/Reflection Helpers ***/

// TODO(discuss): Should we use reflections inside the testing module as being done here or explicitly
// define the interfaces used for debug/development. The latter will probably scale more but will
// require more effort and pollute the source code with debugging information.
func GetConsensusNodeState(node *shared.Node) typesCons.ConsensusNodeState {
	return GetConsensusModImpl(node).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(typesCons.ConsensusNodeState)
}

func GetConsensusModElem(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).Elem()
}

func GetConsensusModImpl(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule())
}

/*** Debug/Development Message Helpers ***/

func TriggerNextView(t *testing.T, node *shared.Node) {
	triggerDebugMessage(t, node, messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW)
}

func triggerDebugMessage(t *testing.T, node *shared.Node, action messaging.DebugMessageAction) {
	debugMessage := &messaging.DebugMessage{
		Action:  messaging.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
		Message: nil,
	}
	anyProto, err := anypb.New(debugMessage)
	require.NoError(t, err)

	e := &messaging.PocketEnvelope{Content: anyProto}
	node.GetBus().PublishEventToBus(e)
}

/*** P2P Helpers ***/

func P2PBroadcast(_ *testing.T, nodes IdToNodeMapping, any *anypb.Any) {
	e := &messaging.PocketEnvelope{Content: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func P2PSend(_ *testing.T, node *shared.Node, any *anypb.Any) {
	e := &messaging.PocketEnvelope{Content: any}
	node.GetBus().PublishEventToBus(e)
}

// This is a helper for `waitForEventsInternal` that creates the `includeFilter` function based on
// consensus specific parameters.
// failOnExtraMessages:
// 		This flag is useful when running the consensus unit tests. It causes the test to wait up to the
// 		maximum delay specified in the source code and errors if additional unexpected messages are received.
// 		For example, if the test expects to receive 5 messages within 2 seconds:
// 			false: continue if 5 messages are received in 0.5 seconds
// 			true: true: wait for another 1.5 seconds after 5 messages are received in 0.5 seconds, and fail if any additional messages are received.
func WaitForNetworkConsensusEvents(
	t *testing.T,
	clock *clock.Mock,
	eventsChannel modules.EventsChannel,
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
	return waitForEventsInternal(t, clock, eventsChannel, consensus.HotstuffMessageContentType, numExpectedMsgs, millis, includeFilter, errMsg, failOnExtraMessages)
}

// IMPROVE: This function can be extended to testing events outside of just the consensus module.
func waitForEventsInternal(
	t *testing.T,
	clock *clock.Mock,
	eventsChannel modules.EventsChannel,
	eventContentType string,
	numExpectedMsgs int,
	maxWaitTimeMillis time.Duration,
	msgIncludeFilter func(m *anypb.Any) bool,
	errMsg string,
	failOnExtraMessages bool,
) (expectedMsgs []*anypb.Any, err error) {
	expectedMsgs = make([]*anypb.Any, 0)                 // Aggregate and return the messages we're waiting for
	unusedEvents := make([]*messaging.PocketEnvelope, 0) // "Recycle" events back into the events channel if we're not using them

	// Limit the amount of time we're waiting for the messages to be published on the events channel
	ctx, cancel := clock.WithTimeout(context.TODO(), time.Millisecond*maxWaitTimeMillis)
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
				clock.Add(time.Millisecond)
			}
		}
	}()

	numRemainingMsgs := numExpectedMsgs
loop:
	for {
		select {
		case nodeEvent := <-eventsChannel:
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
				return expectedMsgs, fmt.Errorf("Missing '%s' messages; %d expected but %d received. (%s)", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			} else {
				return expectedMsgs, fmt.Errorf("Too many '%s' messages; %d expected but %d received. (%s)", eventContentType, numExpectedMsgs, len(expectedMsgs), errMsg)
			}
		}
	}
	ticker.Stop()
	tickerDone <- true

	for _, u := range unusedEvents {
		eventsChannel <- u
	}
	return
}

/*** Module Mocking Helpers ***/

// Creates a persistence module mock with mock implementations of some basic functionality
func basePersistenceMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := modulesMock.NewMockPersistenceModule(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)
	persistenceReadContextMock := modulesMock.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()
	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()

	// The persistence context should usually be accessed via the utility module within the context
	// of the consensus module. This one is only used when loading the initial consensus module
	// state; hence the `-1` expectation in the call above.
	persistenceContextMock.EXPECT().Close().Return(nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetLatestBlockHeight().Return(uint64(0), nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(makeMockActors(numValidators), nil).AnyTimes()
	persistenceReadContextMock.EXPECT().Close().Return(nil).AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, eventsChannel modules.EventsChannel) *modulesMock.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := modulesMock.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			eventsChannel <- e
		}).
		AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel, genesisState modules.GenesisState) *modulesMock.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := modulesMock.NewMockUtilityModule(ctrl)
	utilityContextMock := baseUtilityContextMock(t, genesisState)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	utilityMock.EXPECT().
		NewContext(gomock.Any()).
		Return(utilityContextMock, nil).
		MaxTimes(4)

	return utilityMock
}

func baseUtilityContextMock(t *testing.T, genesisState modules.GenesisState) *modulesMock.MockUtilityContext {
	ctrl := gomock.NewController(t)
	utilityContextMock := modulesMock.NewMockUtilityContext(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)
	persistenceContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetPersistenceGenesisState().GetVals(), nil).AnyTimes()
	persistenceContextMock.EXPECT().GetBlockHash(gomock.Any()).Return("", nil).AnyTimes()

	utilityContextMock.EXPECT().
		CreateAndApplyProposalBlock(gomock.Any(), maxTxBytes).
		Return(stateHash, make([][]byte, 0), nil).
		AnyTimes()
	utilityContextMock.EXPECT().
		ApplyBlock().
		Return(stateHash, nil).
		AnyTimes()
	utilityContextMock.EXPECT().SetProposalBlock(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	utilityContextMock.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
	utilityContextMock.EXPECT().Release().Return(nil).AnyTimes()
	utilityContextMock.EXPECT().GetPersistenceContext().Return(persistenceContextMock).AnyTimes()

	persistenceContextMock.EXPECT().Release().Return(nil).AnyTimes()

	return utilityContextMock
}

func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)
	timeSeriesAgentMock := baseTelemetryTimeSeriesAgentMock(t)
	eventMetricsAgentMock := baseTelemetryEventMetricsAgentMock(t)

	telemetryMock.EXPECT().Start().Return(nil).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

	return telemetryMock
}

func baseRpcMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockRPCModule {
	ctrl := gomock.NewController(t)
	rpcMock := modulesMock.NewMockRPCModule(ctrl)
	rpcMock.EXPECT().Start().Return(nil).AnyTimes()
	rpcMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()

	return rpcMock
}

func baseTelemetryTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeSeriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).MaxTimes(1)
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()
	return timeSeriesAgentMock
}

func baseTelemetryEventMetricsAgentMock(t *testing.T) *modulesMock.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := modulesMock.NewMockEventMetricsAgent(ctrl)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

func baseLoggerMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockLoggerModule {
	ctrl := gomock.NewController(t)
	loggerMock := modulesMock.NewMockLoggerModule(ctrl)

	loggerMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()

	return loggerMock
}

func logTime(t *testing.T, clock *clock.Mock) {
	t.Logf("[⌚ CLOCK ⌚] the time is: %v ms from UNIX Epoch [%v]", clock.Now().UTC().UnixMilli(), clock.Now().UTC())
}

// advanceTime moves the time forward on the mock clock and logs what just happened.
func advanceTime(t *testing.T, clock *clock.Mock, duration time.Duration) {
	clock.Add(duration)
	t.Logf("[⌚ CLOCK ⏩] advanced by %v", duration)
	logTime(t, clock)
}

// sleep pauses the goroutine for the given duration on the mock clock and logs what just happened.
//
// Note: time has to be moved forward in a separate goroutine, see `advanceTime`.
func sleep(t *testing.T, clock *clock.Mock, duration time.Duration) {
	t.Logf("[⌚ CLOCK 💤] sleeping for %v", duration)
	clock.Sleep(duration)
}

// timeReminder simply prints, at a given interval and in a separate goroutine, the current mocked time to help with events.
func timeReminder(t *testing.T, clock *clock.Mock, frequency time.Duration) {
	go func() {
		tick := time.NewTicker(frequency)
		for {
			<-tick.C
			logTime(t, clock)
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

// makeMockActors creates a slice of modules.Actor with n &modulesMock.MockActor{} in it.
func makeMockActors(n int) []modules.Actor {
	actors := make([]modules.Actor, n)
	for i := 0; i < n; i++ {
		actors[i] = &modulesMock.MockActor{}
	}
	return actors
}
