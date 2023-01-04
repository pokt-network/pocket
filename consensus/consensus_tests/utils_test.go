package consensus_tests

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/genesis"
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

// If this is set to true, consensus unit tests will fail if additional unexpected messages are received.
// This slows down the tests because we always fail until the timeout specified by the test before continuing
// but guarantees more correctness.
var failOnExtraMessages bool

// TODO(integration): These are temporary variables used in the prototype integration phase that
// will need to be parameterized later once the test framework design matures.
var (
	stateHash          = "42"
	maxTxBytes         = 90000
	emptyByzValidators = make([][]byte, 0)
	emptyTxs           = make([][]byte, 0)
)

const numValidators = 4

// Initialize certain unit test configurations on startup.
func init() {
	flag.BoolVar(&failOnExtraMessages, "failOnExtraMessages", false, "Fail if unexpected additional messages are received")

	var err error
	if err != nil {
		log.Fatalf(err.Error())
	}
}

type IdToNodeMapping map[typesCons.NodeId]*shared.Node

/*** Node Generation Helpers ***/

func GenerateNodeRuntimeMgrs(_ *testing.T, validatorCount int, clockMgr clock.Clock) []*runtime.Manager {
	runtimeMgrs := make([]*runtime.Manager, validatorCount)
	var validatorKeys []string
	genesisState, validatorKeys := test_artifacts.NewGenesisState(validatorCount, 1, 1, 1)
	cfgs := test_artifacts.NewDefaultConfigs(validatorKeys)
	for i, config := range cfgs {
		config.Consensus = &configs.ConsensusConfig{
			PrivateKey:      config.PrivateKey,
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &configs.PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    false,
				DebugTimeBetweenStepsMsec: 0,
			},
		}
		runtimeMgrs[i] = runtime.NewManager(config, genesisState, runtime.WithClock(clockMgr))
	}
	return runtimeMgrs
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	buses []modules.Bus,
	testChannel modules.EventsChannel,
) (pocketNodes IdToNodeMapping) {
	pocketNodes = make(IdToNodeMapping, len(buses))
	// TODO(design): The order here is important in order for NodeId to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(buses, func(i, j int) bool {
		pk, err := cryptoPocket.NewPrivateKey(buses[i].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		pk2, err := cryptoPocket.NewPrivateKey(buses[j].GetRuntimeMgr().GetConfig().PrivateKey)
		require.NoError(t, err)
		return pk.Address().String() < pk2.Address().String()
	})

	for i := range buses {
		pocketNode := CreateTestConsensusPocketNode(t, &buses[i], testChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func CreateTestConsensusPocketNode(
	t *testing.T,
	bus *modules.Bus,
	testChannel modules.EventsChannel,
) *shared.Node {
	// persistence is a dependency of consensus, so we need to create it first
	persistenceMock := basePersistenceMock(t, testChannel, *bus)
	(*bus).RegisterModule(persistenceMock)

	_, err := consensus.Create(*bus)
	require.NoError(t, err)

	runtimeMgr := (*bus).GetRuntimeMgr()
	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	p2pMock := baseP2PMock(t, testChannel)
	utilityMock := baseUtilityMock(t, testChannel, runtimeMgr.GetGenesis())
	telemetryMock := baseTelemetryMock(t, testChannel)
	loggerMock := baseLoggerMock(t, testChannel)
	rpcMock := baseRpcMock(t, testChannel)

	for _, module := range []modules.Module{
		p2pMock,
		utilityMock,
		telemetryMock,
		loggerMock,
		rpcMock,
	} {
		(*bus).RegisterModule(module)
	}

	require.NoError(t, err)

	pk, err := cryptoPocket.NewPrivateKey(runtimeMgr.GetConfig().PrivateKey)
	require.NoError(t, err)

	pocketNode := shared.NewNodeWithP2PAddress(pk.Address())

	pocketNode.SetBus(*bus)

	return pocketNode
}

func GenerateBuses(t *testing.T, runtimeMgrs []*runtime.Manager) (buses []modules.Bus) {
	buses = make([]modules.Bus, len(runtimeMgrs))
	for i := range runtimeMgrs {
		bus, err := runtime.CreateBus(runtimeMgrs[i])
		require.NoError(t, err)
		buses[i] = bus
	}
	return
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

func WaitForNetworkConsensusMessages(
	t *testing.T,
	clock clock.Clock,
	testChannel modules.EventsChannel,
	step typesCons.HotstuffStep,
	hotstuffMsgType typesCons.HotstuffMessageType,
	numMessages int,
	millis time.Duration,
) (messages []*anypb.Any, err error) {

	includeFilter := func(m *anypb.Any) bool {
		msg, err := codec.GetCodec().FromAny(m)
		require.NoError(t, err)

		hotstuffMessage, ok := msg.(*typesCons.HotstuffMessage)
		require.True(t, ok)

		return hotstuffMessage.Type == hotstuffMsgType && hotstuffMessage.Step == step
	}

	errorMessage := fmt.Sprintf("HotStuff step: %s, type: %s", typesCons.HotstuffStep_name[int32(step)], typesCons.HotstuffMessageType_name[int32(hotstuffMsgType)])
	return waitForNetworkConsensusMessagesInternal(t, clock, testChannel, consensus.HotstuffMessageContentType, numMessages, millis, includeFilter, errorMessage)
}

// IMPROVE(olshansky): Translate this to use generics.
func waitForNetworkConsensusMessagesInternal(
	_ *testing.T,
	clock clock.Clock,
	testChannel modules.EventsChannel,
	messageContentType string,
	numMessages int,
	millis time.Duration,
	includeFilter func(m *anypb.Any) bool,
	errorMessage string,
) (messages []*anypb.Any, err error) {
	messages = make([]*anypb.Any, 0)
	ctx, cancel := clock.WithTimeout(context.Background(), time.Millisecond*millis)
	unused := make([]*messaging.PocketEnvelope, 0) // TODO: Move this into a pool rather than resending back to the eventbus.

loop:
	for {
		select {
		case testEvent := <-testChannel:
			if testEvent.GetContentType() != messageContentType {
				unused = append(unused, &testEvent)
				continue
			}

			message := testEvent.Content
			if message == nil || !includeFilter(message) {
				unused = append(unused, &testEvent)
				continue
			}

			messages = append(messages, message)
			numMessages--

			// The if structure below "breaks early" when we get enough messages. However, it does not capture
			// the case where we could be receiving more messages than expected. To make sure the latter doesn't
			// happen, the `failOnExtraMessages` flag must be set to true.
			if !failOnExtraMessages && numMessages == 0 {
				break loop
			}
		case <-ctx.Done():
			if numMessages == 0 {
				break loop
			} else if numMessages > 0 {
				cancel()
				return nil, fmt.Errorf("Missing %s messages; missing: %d, received: %d; (%s)", messageContentType, numMessages, len(messages), errorMessage)
			} else {
				cancel()
				return nil, fmt.Errorf("Too many %s messages received; expected: %d, received: %d; (%s)", messageContentType, numMessages+len(messages), len(messages), errorMessage)
			}
		}
	}
	cancel()
	for _, u := range unused {
		testChannel <- *u
	}
	return
}

/*** Module Mocking Helpers ***/

// Creates a persistence module mock with mock implementations of some basic functionality
func basePersistenceMock(t *testing.T, _ modules.EventsChannel, bus modules.Bus) *modulesMock.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := modulesMock.NewMockPersistenceModule(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)
	persistenceReadContextMock := modulesMock.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()
	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()

	// The persistence context should usually be accessed via the utility module within the context
	// of the consensus module. This one is only used when loading the initial consensus module
	// state; hence the `-1` expectation in the call above.
	persistenceContextMock.EXPECT().Close().Return(nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetLatestBlockHeight().Return(uint64(0), nil).AnyTimes()
	persistenceReadContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(bus.GetRuntimeMgr().GetGenesis().Validators, nil).AnyTimes()

	persistenceReadContextMock.EXPECT().Close().Return(nil).AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, testChannel modules.EventsChannel) *modulesMock.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := modulesMock.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Do(func(msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			testChannel <- *e
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any) {
			e := &messaging.PocketEnvelope{Content: msg}
			testChannel <- *e
		}).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel, genesisState *genesis.GenesisState) *modulesMock.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := modulesMock.NewMockUtilityModule(ctrl)
	utilityContextMock := baseUtilityContextMock(t, genesisState)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	utilityMock.EXPECT().
		NewContext(gomock.Any()).
		Return(utilityContextMock, nil).
		MaxTimes(4)
	utilityMock.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()

	return utilityMock
}

func baseUtilityContextMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockUtilityContext {
	ctrl := gomock.NewController(t)
	utilityContextMock := modulesMock.NewMockUtilityContext(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)
	persistenceContextMock.EXPECT().GetAllValidators(gomock.Any()).Return(genesisState.GetValidators(), nil).AnyTimes()
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
	telemetryMock.EXPECT().GetModuleName().Return(modules.TelemetryModuleName).AnyTimes()

	return telemetryMock
}

func baseRpcMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockRPCModule {
	ctrl := gomock.NewController(t)
	rpcMock := modulesMock.NewMockRPCModule(ctrl)
	rpcMock.EXPECT().Start().Return(nil).AnyTimes()
	rpcMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	rpcMock.EXPECT().GetModuleName().Return(modules.RPCModuleName).AnyTimes()

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
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return eventMetricsAgentMock
}

func baseLoggerMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockLoggerModule {
	ctrl := gomock.NewController(t)
	loggerMock := modulesMock.NewMockLoggerModule(ctrl)

	loggerMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	loggerMock.EXPECT().GetModuleName().Return(modules.LoggerModuleName).AnyTimes()

	return loggerMock
}

func logTime(clock *clock.Mock) {
	log.Printf("[⌚ CLOCK ⌚] the time is: %v ms from UNIX Epoch [%v]", clock.Now().UTC().UnixMilli(), clock.Now().UTC())
}

// advanceTime moves the time forward on the mock clock and logs what just happened.
func advanceTime(clock *clock.Mock, duration time.Duration) {
	clock.Add(duration)
	log.Printf("[⌚ CLOCK ⏩] advanced by %v", duration)
	logTime(clock)
}

// sleep pauses the goroutine for the given duration on the mock clock and logs what just happened.
//
// Note: time has to be moved forward in a separate goroutine, see `advanceTime`.
func sleep(clock *clock.Mock, duration time.Duration) {
	log.Printf("[⌚ CLOCK 💤] sleeping for %v", duration)
	clock.Sleep(duration)
}

// timeReminder simply prints, at a given interval and in a separate goroutine, the current mocked time to help with events.
func timeReminder(clock *clock.Mock, frequency time.Duration) {
	go func() {
		tick := time.NewTicker(frequency)
		for {
			<-tick.C
			logTime(clock)
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
