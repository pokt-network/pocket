package consensus_tests

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/consensus"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	typesGenesis "github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// The number at which to start incrementing the seeds
// used for random key generation.
const genesisConfigSeedStart = uint32(42)

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

type IdToNodeMapping map[typesCons.NodeId]*shared.Node

/*** Node Generation Helpers ***/

func GenerateNodeConfigs(t *testing.T, n int) (configs []*config.Config) {
	for i := uint32(1); i <= uint32(n); i++ {
		// Deterministically generate a private key for the node
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, i+genesisConfigSeedStart)
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		require.NoError(t, err)

		// Generate test config
		c := config.Config{
			RootDir:    "",                                  // left empty intentionally
			PrivateKey: pk.(cryptoPocket.Ed25519PrivateKey), // deterministic key based on `i`
			Genesis:    genesisJson(t),
			Pre2P:      nil,
			P2P:        nil,
			Consensus: &config.ConsensusConfig{
				MaxMempoolBytes: 500000000,
				MaxBlockBytes:   4000000,
				Pacemaker: &config.PacemakerConfig{
					TimeoutMsec:               5000,
					Manual:                    false,
					DebugTimeBetweenStepsMsec: 0,
				},
			},
			Persistence: nil,
			Utility:     nil,
		}
		err = c.ValidateAndHydrate()
		require.NoError(t, err)

		// Append the config to the list of configurations used in the test
		configs = append(configs, &c)
	}
	return
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	configs []*config.Config,
	testChannel modules.EventsChannel,
) (pocketNodes IdToNodeMapping) {
	pocketNodes = make(IdToNodeMapping, len(configs))
	// TODO(design): The order here is important in order for NodeId to be set correctly below.
	// This logic will need to change once proper leader election is implemented.
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].PrivateKey.Address().String() < configs[j].PrivateKey.Address().String()
	})
	for i, cfg := range configs {
		pocketNode := CreateTestConsensusPocketNode(t, cfg, testChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[typesCons.NodeId(i+1)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked
func CreateTestConsensusPocketNode(
	t *testing.T,
	cfg *config.Config,
	testChannel modules.EventsChannel,
) *shared.Node {
	_ = typesGenesis.GetNodeState(cfg)

	consensusMod, err := consensus.Create(cfg)
	require.NoError(t, err)

	// TODO(olshansky): At the moment we are using the same base mocks for all the tests,
	// but note that they will need to be customized on a per test basis.
	persistenceMock := basePersistenceMock(t, testChannel)
	p2pMock := baseP2PMock(t, testChannel)
	utilityMock := baseUtilityMock(t, testChannel)

	bus, err := shared.CreateBus(persistenceMock, p2pMock, utilityMock, consensusMod)
	require.NoError(t, err)

	pocketNode := &shared.Node{
		Address: cfg.PrivateKey.Address(),
	}
	pocketNode.SetBus(bus)

	return pocketNode
}

func StartAllTestPocketNodes(t *testing.T, pocketNodes IdToNodeMapping) {
	for _, pocketNode := range pocketNodes {
		go pocketNode.Start()
		startEvent := pocketNode.GetBus().GetBusEvent()
		require.Equal(t, startEvent.Topic, types.PocketTopic_POCKET_NODE_TOPIC)
	}
}

/*** Node Visibility/Reflection Helpers ***/

// TODO(discuss): Should we use reflections inside the testing module as being done here or explicitly
// define the interfaces used for debug/development. The latter will probably scale more but will
// require more effort and pollute the source code with debugging information.
func GetConsensusNodeState(node *shared.Node) typesCons.ConsensusNodeState {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(typesCons.ConsensusNodeState)
}

func GetConsensusModImplementation(node *shared.Node) reflect.Value {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).Elem()
}

/*** Debug/Development Message Helpers ***/

func TriggerNextView(t *testing.T, node *shared.Node) {
	triggerDebugMessage(t, node, types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW)
}

func triggerDebugMessage(t *testing.T, node *shared.Node, action types.DebugMessageAction) {
	debugMessage := &types.DebugMessage{
		Action:  types.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
		Message: nil,
	}
	anyProto, err := anypb.New(debugMessage)
	require.NoError(t, err)

	e := &types.PocketEvent{Topic: types.PocketTopic_DEBUG_TOPIC, Data: anyProto}
	node.GetBus().PublishEventToBus(e)
}

/*** P2P Helpers ***/

func P2PBroadcast(_ *testing.T, nodes IdToNodeMapping, any *anypb.Any) {
	e := &types.PocketEvent{Topic: types.PocketTopic_CONSENSUS_MESSAGE_TOPIC, Data: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func P2PSend(_ *testing.T, node *shared.Node, any *anypb.Any) {
	e := &types.PocketEvent{Topic: types.PocketTopic_CONSENSUS_MESSAGE_TOPIC, Data: any}
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
	return waitForNetworkConsensusMessagesInternal(t, testChannel, types.PocketTopic_CONSENSUS_MESSAGE_TOPIC, numMessages, millis, includeFilter, errorMessage)
}

// IMPROVE(olshansky): Translate this to use generics.
func waitForNetworkConsensusMessagesInternal(
	_ *testing.T,
	testChannel modules.EventsChannel,
	topic types.PocketTopic,
	numMessages int,
	millis time.Duration,
	includeFilter func(m *anypb.Any) bool,
	errorMessage string,
) (messages []*anypb.Any, err error) {
	messages = make([]*anypb.Any, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*millis)
	unused := make([]*types.PocketEvent, 0) // TODO: Move this into a pool rather than resending back to the eventbus.
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

	persistenceMock.EXPECT().Start().Do(func() {}).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()

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
		Do(func(msg *anypb.Any, topic types.PocketTopic) {
			e := &types.PocketEvent{Topic: topic, Data: msg}
			testChannel <- *e
		}).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(addr cryptoPocket.Address, msg *anypb.Any, topic types.PocketTopic) {
			e := &types.PocketEvent{Topic: topic, Data: msg}
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
	persistenceContextMock := modulesMock.NewMockPersistenceContext(ctrl)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	utilityMock.EXPECT().
		// NewContext(int64(1)).
		NewContext(gomock.Any()).
		Return(utilityContextMock, nil).
		MaxTimes(4)

	utilityContextMock.EXPECT().GetPersistenceContext().Return(persistenceContextMock).AnyTimes()
	utilityContextMock.EXPECT().ReleaseContext().Return().AnyTimes()
	utilityContextMock.EXPECT().
		GetTransactionsForProposal(gomock.Any(), maxTxBytes, gomock.AssignableToTypeOf(emptyByzValidators)).
		Return(make([][]byte, 0), nil).
		AnyTimes()
	utilityContextMock.EXPECT().
		// ApplyBlock(int64(1), gomock.Any(), gomock.AssignableToTypeOf(emptyTxs), gomock.AssignableToTypeOf(emptyByzValidators)).
		ApplyBlock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(appHash, nil).
		AnyTimes()

	persistenceContextMock.EXPECT().Commit().Return(nil).AnyTimes()

	return utilityMock
}

/*** Genesis Helpers ***/

// The genesis file is hardcoded for test purposes, but is also
// validated before returning the string in case changes in the
// configurations are made.
func genesisJson(t *testing.T) string {
	genesisJsonStr := fmt.Sprintf(`{
		"genesis_state_configs": {
			"num_validators": 4,
			"num_applications": 0,
			"num_fisherman": 0,
			"num_servicers": 0,
			"keys_seed_start": %d
		},
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash"
	}`, genesisConfigSeedStart)

	_, err := typesGenesis.PocketGenesisFromJSON([]byte(genesisJsonStr))
	require.NoError(t, err)

	return genesisJsonStr
}
