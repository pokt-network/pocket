package consensus_tests

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/test_artifacts"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	m.Run()
}

// If this is set to true, consensus unit tests will fail if additional unexpected messages are received.
// This slows down the tests because we always fail until the timeout specified by the test before continuing
// but guarantees more correctness.
var failOnExtraMessages bool

// TODO(integration): These are temporary variables used in the prototype integration phase that
// will need to be parameterized later once the test framework design matures.
var appHash []byte
var maxTxBytes = 90000
var emptyByzValidators = make([][]byte, 0)
var emptyTxs = make([][]byte, 0)

// Initialize certain unit test configurations on startup.
func init() {
	flag.BoolVar(&failOnExtraMessages, "failOnExtraMessages", false, "Fail if unexpected additional messages are received")

	var err error
	appHash, err = hex.DecodeString("31")
	if err != nil {
		log.Fatalf(err.Error())
	}
}

type IdToNodeMapping map[typesCons.NodeId]modules.NodeModule

/*** Node Generation Helpers ***/

func GenerateNodeRuntimeMgrs(_ *testing.T, validatorCount int) []runtime.Manager {
	runtimeMgrs := make([]runtime.Manager, 0)
	var keys []string
	genesisState, keys := test_artifacts.NewGenesisState(validatorCount, 1, 1, 1)
	configs := test_artifacts.NewDefaultConfigs(keys)
	for _, config := range configs {
		runtime.WithConsensusConfig(&typesCons.ConsensusConfig{
			PrivateKey:      config.GetBaseConfig().GetPrivateKey(),
			MaxMempoolBytes: 500000000,
			PacemakerConfig: &typesCons.PacemakerConfig{
				TimeoutMsec:               5000,
				Manual:                    false,
				DebugTimeBetweenStepsMsec: 0,
			},
		})(config)
		runtimeMgrs = append(runtimeMgrs, *runtime.NewManager(config, genesisState))
	}
	return runtimeMgrs
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	runtimeMgrs []runtime.Manager,
	testChannel modules.EventsChannel,
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
	for i, runtimeConfig := range runtimeMgrs {
		pocketNode := CreateTestConsensusPocketNode(t, &runtimeConfig, testChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

func CreateTestConsensusPocketNodesNew(
	t *testing.T,
	runtimeMgrs []runtime.Manager,
	testChannel modules.EventsChannel,
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
		pocketNode := CreateTestConsensusPocketNode(t, &runtimeMgr, testChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func CreateTestConsensusPocketNode(
	t *testing.T,
	runtimeMgr *runtime.Manager,
	testChannel modules.EventsChannel,
) *shared.Node {
	//createTestingGenesisAndConfigFiles(t, cfg, genesisState)

	consensusMod, err := consensus.Create(runtimeMgr)
	require.NoError(t, err)
	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	persistenceMock := basePersistenceMock(t, testChannel)
	p2pMock := baseP2PMock(t, testChannel)
	utilityMock := baseUtilityMock(t, testChannel)
	telemetryMock := baseTelemetryMock(t, testChannel)

	bus, err := shared.CreateBus(runtimeMgr, persistenceMock, p2pMock, utilityMock, consensusMod.(modules.ConsensusModule), telemetryMock)
	require.NoError(t, err)

	pk, err := cryptoPocket.NewPrivateKey(runtimeMgr.GetConfig().GetBaseConfig().GetPrivateKey())
	require.NoError(t, err)

	pocketNode := shared.NewNodeWithAddress(pk.Address())

	pocketNode.SetBus(bus)

	return pocketNode
}

func StartAllTestPocketNodes(t *testing.T, pocketNodes IdToNodeMapping) {
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
		startEvent := pocketNode.GetBus().GetBusEvent()
		require.Equal(t, startEvent.Topic, debug.PocketTopic_POCKET_NODE_TOPIC)
	}
}

/*** Node Visibility/Reflection Helpers ***/

// TODO(discuss): Should we use reflections inside the testing module as being done here or explicitly
// define the interfaces used for debug/development. The latter will probably scale more but will
// require more effort and pollute the source code with debugging information.
func GetConsensusNodeState(node modules.NodeModule) typesCons.ConsensusNodeState {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(typesCons.ConsensusNodeState)
}

func GetConsensusModImplementation(node modules.NodeModule) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).Elem()
}

/*** Debug/Development Message Helpers ***/

func TriggerNextView(t *testing.T, node modules.NodeModule) {
	triggerDebugMessage(t, node, debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW)
}

func triggerDebugMessage(t *testing.T, node modules.NodeModule, action debug.DebugMessageAction) {
	debugMessage := &debug.DebugMessage{
		Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
		Message: nil,
	}
	anyProto, err := anypb.New(debugMessage)
	require.NoError(t, err)

	e := &debug.PocketEvent{Topic: debug.PocketTopic_DEBUG_TOPIC, Data: anyProto}
	node.GetBus().PublishEventToBus(e)
}

/*** P2P Helpers ***/

func P2PBroadcast(_ *testing.T, nodes IdToNodeMapping, any *anypb.Any) {
	e := &debug.PocketEvent{Topic: debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC, Data: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func P2PSend(_ *testing.T, node modules.NodeModule, any *anypb.Any) {
	e := &debug.PocketEvent{Topic: debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC, Data: any}
	node.GetBus().PublishEventToBus(e)
}

func WaitForNetworkConsensusMessages(
	t *testing.T,
	testChannel modules.EventsChannel,
	step typesCons.HotstuffStep,
	hotstuffMsgType typesCons.HotstuffMessageType,
	numMessages int,
	millis time.Duration,
) (messages []*anypb.Any, err error) {

	includeFilter := func(m *anypb.Any) bool {
		var hotstuffMessage typesCons.HotstuffMessage
		err := anypb.UnmarshalTo(m, &hotstuffMessage, proto.UnmarshalOptions{})
		require.NoError(t, err)

		return hotstuffMessage.Type == hotstuffMsgType && hotstuffMessage.Step == step
	}

	errorMessage := fmt.Sprintf("HotStuff step: %s, type: %s", typesCons.HotstuffStep_name[int32(step)], typesCons.HotstuffMessageType_name[int32(hotstuffMsgType)])
	return waitForNetworkConsensusMessagesInternal(t, testChannel, debug.PocketTopic_CONSENSUS_MESSAGE_TOPIC, numMessages, millis, includeFilter, errorMessage)
}

// IMPROVE(olshansky): Translate this to use generics.
func waitForNetworkConsensusMessagesInternal(
	_ *testing.T,
	testChannel modules.EventsChannel,
	topic debug.PocketTopic,
	numMessages int,
	millis time.Duration,
	includeFilter func(m *anypb.Any) bool,
	errorMessage string,
) (messages []*anypb.Any, err error) {
	messages = make([]*anypb.Any, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*millis)
	unused := make([]*debug.PocketEvent, 0) // TODO: Move this into a pool rather than resending back to the eventbus.
loop:
	for {
		select {
		case testEvent := <-testChannel:
			if testEvent.Topic != topic {
				unused = append(unused, &testEvent)
				continue
			}

			message := testEvent.Data
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
				return nil, fmt.Errorf("Missing %s messages; missing: %d, received: %d; (%s)", topic, numMessages, len(messages), errorMessage)
			} else {
				cancel()
				return nil, fmt.Errorf("Too many %s messages received; expected: %d, received: %d; (%s)", topic, numMessages+len(messages), len(messages), errorMessage)
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
func basePersistenceMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := modulesMock.NewMockPersistenceModule(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().Start().Do(func() {}).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	persistenceMock.EXPECT().NewReadContext(int64(-1)).Return(persistenceContextMock, nil).AnyTimes()

	// The persistence context should usually be accessed via the utility module within the context
	// of the consensus module. This one is only used when loading the initial consensus module
	// state; hence the `-1` expectation in the call above.
	persistenceContextMock.EXPECT().Close().Return(nil).AnyTimes()
	persistenceContextMock.EXPECT().GetLatestBlockHeight().Return(uint64(0), nil).AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, testChannel modules.EventsChannel) *modulesMock.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := modulesMock.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Do(func() {}).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any(), gomock.Any()).
		Do(func(msg *anypb.Any, topic debug.PocketTopic) {
			e := &debug.PocketEvent{Topic: topic, Data: msg}
			testChannel <- *e
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any, topic debug.PocketTopic) {
			e := &debug.PocketEvent{Topic: topic, Data: msg}
			testChannel <- *e
		}).
		AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := modulesMock.NewMockUtilityModule(ctrl)
	utilityContextMock := modulesMock.NewMockUtilityContext(ctrl)
	persistenceContextMock := modulesMock.NewMockPersistenceRWContext(ctrl)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	utilityMock.EXPECT().
		NewContext(gomock.Any()).
		Return(utilityContextMock, nil).
		MaxTimes(4)

	utilityContextMock.EXPECT().GetPersistenceContext().Return(persistenceContextMock).AnyTimes()
	utilityContextMock.EXPECT().CommitPersistenceContext().Return(nil).AnyTimes()
	utilityContextMock.EXPECT().ReleaseContext().Return().AnyTimes()
	utilityContextMock.EXPECT().
		GetProposalTransactions(gomock.Any(), maxTxBytes, gomock.AssignableToTypeOf(emptyByzValidators)).
		Return(make([][]byte, 0), nil).
		AnyTimes()
	utilityContextMock.EXPECT().
		ApplyBlock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(appHash, nil).
		AnyTimes()
	utilityContextMock.EXPECT().StoreBlock(gomock.Any()).AnyTimes().Return(nil)

	persistenceContextMock.EXPECT().Commit().Return(nil).AnyTimes()

	return utilityMock
}

func baseTelemetryMock(t *testing.T, _ modules.EventsChannel) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)
	timeSeriesAgentMock := baseTelemetryTimeSeriesAgentMock(t)
	eventMetricsAgentMock := baseTelemetryEventMetricsAgentMock(t)

	telemetryMock.EXPECT().Start().Do(func() {}).AnyTimes()
	telemetryMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).MaxTimes(1)
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return telemetryMock
}

func baseTelemetryTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)
	return timeseriesAgentMock
}

func baseTelemetryEventMetricsAgentMock(t *testing.T) *modulesMock.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := modulesMock.NewMockEventMetricsAgent(ctrl)
	return eventMetricsAgentMock
}
