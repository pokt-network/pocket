package consensus_tests

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/consensus"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared"
	"github.com/pokt-network/pocket/shared/config"
	pcrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type IdToNodeMapping map[types_consensus.NodeId]*shared.Node

func GenerateNodeConfigs(t *testing.T, n int) (configs []*config.Config) {
	for i := uint32(1); i <= uint32(n); i++ {
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, i)
		pk, err := pcrypto.NewPrivateKeyFromSeed(seed)
		require.NoError(t, err)

		c := config.Config{
			RootDir:    "",                             // left empty intentionally
			PrivateKey: pk.(pcrypto.Ed25519PrivateKey), // deterministic key based on `i`
			Genesis:    genesisJson(t),

			Pre2P:       nil,
			P2P:         nil,
			Consensus:   nil,
			Persistence: nil,
			Utility:     nil,
		}
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
	for i, cfg := range configs {
		pocketNode := CreateTestConsensusPocketNode(t, cfg, testChannel)
		// TODO(olshansky): Figure this part out.
		pocketNodes[types_consensus.NodeId(i)] = pocketNode
	}
	return
}

// Creates a pocket node where all the primary modules, exception for consensus, are mocked.
func CreateTestConsensusPocketNode(
	t *testing.T,
	cfg *config.Config,
	testChannel modules.EventsChannel,
) *shared.Node {
	_ = types.GetTestState(cfg)

	consensusMod, err := consensus.Create(cfg)
	require.NoError(t, err)

	persistenceMock := basePersistenceMock(t, testChannel)
	p2pMock := baseP2PMock(t, testChannel)
	utilityMock := baseUtilityMock(t, testChannel)

	bus, err := shared.CreateBus(nil, persistenceMock, p2pMock, utilityMock, consensusMod)
	require.NoError(t, err)

	pocketNode := &shared.Node{
		Address: cfg.PrivateKey.Address(),
	}
	pocketNode.SetBus(bus)

	// Base persistence mocks
	// persistenceMock.EXPECT().
	// 	Stop(gomock.Any()).
	// 	Do(func(ctx *pcontext.PocketContext) {
	// 		log.Println("[MOCK] Stop persistence mock")
	// 	}).
	// 	AnyTimes()

	// persistenceMock.EXPECT().
	// 	GetLatestBlockHeight().
	// 	Do(func() (uint64, error) {
	// 		log.Println("[MOCK] GetLatestBlockHeight")
	// 		return uint64(0), fmt.Errorf("[MOCK] GetLatestBlockHeight not implemented yet...")
	// 	}).
	// 	AnyTimes()

	// persistenceMock.EXPECT().
	// 	GetBlockHash(gomock.Any()).
	// 	Do(func(height uint64) ([]byte, error) {
	// 		return []byte(strconv.FormatUint(height, 10)), nil
	// 	}).
	// 	AnyTimes()

	// Base network module mocks

	// p2pNetworkMock.EXPECT().
	// 	GetAddrBook().
	// 	DoAndReturn(func() []*p2p_types.NetworkPeer {
	// 		log.Println("[MOCK] Network GetNetwork", addrBook)
	// 		return addrBook
	// 	}).
	// 	AnyTimes()

	// networkMock.EXPECT().
	// 	Stop(gomock.Any()).
	// 	Do(func(ctx *pcontext.PocketContext) {
	// 		log.Println("[MOCK] Stop network mock")
	// 	}).
	// 	AnyTimes()

	// networkMock.EXPECT().
	// 	GetNetwork().
	// 	DoAndReturn(func() p2p_types.Network {
	// 		return p2pNetworkMock
	// 	}).
	// 	AnyTimes()

	// networkMock.EXPECT().
	// 	Send(gomock.Any(), gomock.Any(), gomock.Any()).
	// 	Do(func(ctx *pcontext.PocketContext, message *p2p_types.NetworkMessage, address types2.NodeId) {
	// 		networkMsg, _ := p2p.EncodeNetworkMessage(message)
	// 		e := types.Event{PocketTopic: types.P2P_SEND_MESSAGE, MessageData: networkMsg}
	// 		testPocketBus <- e
	// 	}).
	// 	AnyTimes()

	// networkMock.EXPECT().
	// 	// decoder
	// 	Broadcast(gomock.Any(), gomock.Any()).
	// 	Do(func(ctx *pcontext.PocketContext, message *p2p_types.NetworkMessage) {
	// 		networkMsg, _ := p2p.EncodeNetworkMessage(message)
	// 		e := types.Event{PocketTopic: types.P2P_BROADCAST_MESSAGE, MessageData: networkMsg}
	// 		testPocketBus <- e
	// 	}).
	// 	AnyTimes()

	// Base utility mocks

	// utilityMock.EXPECT().
	// 	Stop(gomock.Any()).
	// 	Do(func(*pcontext.PocketContext) {
	// 		log.Println("[MOCK] Stop utility mock")
	// 	}).
	// 	AnyTimes()

	// utilityMock.EXPECT().
	// 	HandleEvidence(gomock.Any(), gomock.Any()).
	// 	Do(func(*pcontext.PocketContext, *types_consensus.Evidence) {
	// 		log.Println("[MOCK] HandleEvidence utility mock")
	// 	}).
	// 	AnyTimes()

	// utilityMock.EXPECT().
	// 	ReapMempool(gomock.Any()).
	// 	Do(func(*pcontext.PocketContext) {
	// 		log.Println("[MOCK] ReapMempool utility mock")
	// 	}).
	// 	AnyTimes()

	// utilityMock.EXPECT().
	// 	BeginBlock(gomock.Any()).
	// 	Do(func(*pcontext.PocketContext) {
	// 		log.Println("[MOCK] BeginBlock utility mock")
	// 	}).
	// 	AnyTimes()

	// utilityMock.EXPECT().
	// 	DeliverTx(gomock.Any(), gomock.Any()).
	// 	Do(func(*pcontext.PocketContext, *types_consensus.Transaction) {
	// 		log.Println("[MOCK] DeliverTx utility mock")
	// 	}).
	// 	AnyTimes()

	// utilityMock.EXPECT().
	// 	EndBlock(gomock.Any()).
	// 	Do(func(*pcontext.PocketContext) {
	// 		log.Println("[MOCK] Stop EndBlock mock")
	// 	}).
	// 	AnyTimes()

	return pocketNode
}

// TODO(discuss): Should we use reflections inside the testing module as being done here or explicitly
// define the interfaces used for debug/development. The latter will probably scale more but will
// require more effort.
func GetConsensusNodeState(node *shared.Node) types_consensus.ConsensusNodeState {
	return reflect.ValueOf(node.GetBus().GetConsensusModule()).MethodByName("GetNodeState").Call([]reflect.Value{})[0].Interface().(types_consensus.ConsensusNodeState)
}

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

func P2PBroadcast(t *testing.T, nodes IdToNodeMapping, message *types_consensus.ConsensusMessage) {
	any, err := anypb.New(message)
	require.NoError(t, err)

	e := &types.PocketEvent{Topic: types.PocketTopic_CONSENSUS_MESSAGE_TOPIC, Data: any}
	for _, node := range nodes {
		node.GetBus().PublishEventToBus(e)
	}
}

func WaitForNetworkConsensusMessages(
	t *testing.T,
	testChannel modules.EventsChannel,
	step types_consensus.HotstuffStep,
	numMessages int,
	millis time.Duration,
) (messages []*types_consensus.ConsensusMessage) {
	decoder := func(any *anypb.Any) *types_consensus.ConsensusMessage {
		var consensusMessage types_consensus.ConsensusMessage
		err := anypb.UnmarshalTo(any, &consensusMessage, proto.UnmarshalOptions{})
		require.NoError(t, err)

		return &consensusMessage
	}

	includeFilter := func(m *types_consensus.ConsensusMessage) bool {
		return m.Type == types_consensus.ConsensusMessageType_CONSENSUS_HOTSTUFF_MESSAGE
	}

	// errorMessage := fmt.Sprintf("HotStuff step: %s", types_consensus.HotstuffStep_name[int32(step)])
	return WaitForNetworkConsensusMessagesInternal(t, testChannel, types.PocketTopic_CONSENSUS_MESSAGE_TOPIC, numMessages, millis, decoder, includeFilter, "error")
}

// TODO(olshansky): Translate this to use generics.
func WaitForNetworkConsensusMessagesInternal(
	t *testing.T,
	testChannel modules.EventsChannel,
	topic types.PocketTopic,
	numMessages int,
	millis time.Duration,
	decoder func(*anypb.Any) *types_consensus.ConsensusMessage,
	includeFilter func(m *types_consensus.ConsensusMessage) bool,
	errorMessage string,
) (messages []*types_consensus.ConsensusMessage) {
	messages = make([]*types_consensus.ConsensusMessage, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*millis)
	unused := make([]*types.PocketEvent, 0) // TODO: Move this into a pool rather than resending back to the eventbus.
loop:
	for {
		select {
		case testEvent := <-testChannel:
			if testEvent.Topic != topic {
				unused = append(unused, testEvent)
				continue
			}

			message := decoder(testEvent.Data)
			if message == nil || !includeFilter(message) {
				unused = append(unused, testEvent)
				continue
			}

			if numMessages <= 0 {
				t.Fatalf("Was only expecting to wait ofr %d messages, but got more...", numMessages)
			}

			messages = append(messages, message)
			numMessages--

			// NOTE: The if structure below "breaks early" when we get enough messages. However, it does not captur
			// the case where we could be receiving more messages than expected. To make sure the latter doesn't
			// happen, the if structure needs to be removed.
			// TODO(design): discuss the comment above with the team.
			if numMessages <= 0 {
				break loop
			}
		case <-ctx.Done():
			if numMessages == 0 {
				break loop
			}
			t.Fatalf("Missing %s messages; missing: %d, received: %d; (%s)", topic, numMessages, len(messages), errorMessage)
		}
	}
	cancel()
	for _, u := range unused {
		testChannel <- u
	}
	return
}

// Creates a persistence module mock with mock implementations of some basic functionality
func basePersistenceMock(t *testing.T, _ modules.EventsChannel) *mock_modules.MockPersistenceModule {
	ctrl := gomock.NewController(t)
	persistenceMock := mock_modules.NewMockPersistenceModule(ctrl)

	// Basic NOOP operations
	persistenceMock.EXPECT().Start().Do(func() {}).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()

	return persistenceMock
}

// Creates a p2p module mock with mock implementations of some basic functionality
func baseP2PMock(t *testing.T, testChannel modules.EventsChannel) *mock_modules.MockP2PModule {
	ctrl := gomock.NewController(t)
	p2pMock := mock_modules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Do(func() {}).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any(), gomock.Any()).
		Do(func(msg *anypb.Any, topic types.PocketTopic) {
			e := &types.PocketEvent{Topic: topic, Data: msg}
			testChannel <- e
		}).
		AnyTimes()

	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(addr pcrypto.Address, msg *anypb.Any, topic types.PocketTopic) {
			e := &types.PocketEvent{Topic: topic, Data: msg}
			testChannel <- e
		}).
		AnyTimes()

	return p2pMock
}

// Creates a utility module mock with mock implementations of some basic functionality
func baseUtilityMock(t *testing.T, _ modules.EventsChannel) *mock_modules.MockUtilityModule {
	ctrl := gomock.NewController(t)
	utilityMock := mock_modules.NewMockUtilityModule(ctrl)
	utilityContext := mock_modules.NewMockUtilityContext(ctrl)

	// TODO(integration): This is only valid while we are still integrating and will likely break soon...
	emptyByzValidators := make([][]byte, 0)
	// emptyTxs := make([][]byte, 0)

	appHash, err := hex.DecodeString("31")
	require.NoError(t, err)

	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(gomock.Any()).Do(func(modules.Bus) {}).AnyTimes()
	utilityMock.EXPECT().
		NewContext(int64(1)).
		Return(utilityContext, nil).
		// Times(4)
		AnyTimes()
	utilityContext.EXPECT().
		GetTransactionsForProposal(gomock.Any(), 90000, gomock.AssignableToTypeOf(emptyByzValidators)).
		Return(make([][]byte, 0), nil).
		AnyTimes()
	utilityContext.EXPECT().
		// ApplyBlock(int64(1), gomock.Any(), gomock.AssignableToTypeOf(emptyTxs), gomock.AssignableToTypeOf(emptyByzValidators)).
		ApplyBlock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(appHash, nil).
		AnyTimes()

	return utilityMock
}

// The genesis file is hardcoded for test purposes, but is also
// validated before returning the string in case changes in the
// configurations are made.
func genesisJson(t *testing.T) string {
	genesisJsonStr := `{
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"consensus_params": {
			"max_mempool_bytes": 50000000,
			"max_block_bytes": 4000000,
			"max_transaction_bytes": 100000,
			"vrf_key_refresh_freq_block": 5,
			"vrf_key_validity_block": 5,
			"pace_maker": {
				"timeout_msec": 5000,
				"retry_timeout_msec": 1000,
				"max_timeout_msec": 60000,
				"min_block_freq_msec": 2000,
				"debug_time_between_steps_msec": 3000
			}
		},
		"validators": [
			{
				"address": "fa4d86c3b551aa6cd7c3759d040c037ef2c6379f",
				"public_key": "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
				"private_key": "0100000000000000000000000000000000000000000000000000000000000000cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "1",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "e3c1b362c0df36f6b370b8b1479b67dad96392b2",
				"public_key": "6b79c57e6a095239282c04818e96112f3f03a4001ba97a564c23852a3f1ea5fc",
				"private_key": "02000000000000000000000000000000000000000000000000000000000000006b79c57e6a095239282c04818e96112f3f03a4001ba97a564c23852a3f1ea5fc",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node2",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "db0743e2dcba9ebf2419bde0881beea966689a26",
				"public_key": "dadbd184a2d526f1ebdd5c06fdad9359b228759b4d7f79d66689fa254aad8546",
				"private_key": "0300000000000000000000000000000000000000000000000000000000000000dadbd184a2d526f1ebdd5c06fdad9359b228759b4d7f79d66689fa254aad8546",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node3",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			},
			{
				"address": "6f1e5b61ed9a821457aa6b4d7c2a2b37715ffb16",
				"public_key": "9be3287795907809407e14439ff198d5bfc7dce6f9bc743cb369146f610b4801",
				"private_key": "04000000000000000000000000000000000000000000000000000000000000009be3287795907809407e14439ff198d5bfc7dce6f9bc743cb369146f610b4801",
				"jailed": false,
				"upokt": 5000000000000,
				"host": "node4",
				"port": 8080,
				"debug_port": 9080,
				"chains": ["0001", "0021"]
			}
		]
	}`

	_, err := types.PocketGenesisFromJSON([]byte(genesisJsonStr))
	require.NoError(t, err)

	return genesisJsonStr
}
