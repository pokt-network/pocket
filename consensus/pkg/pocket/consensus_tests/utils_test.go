package consensus_tests

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"pocket/consensus/pkg/config"
	"pocket/consensus/pkg/consensus"
	"pocket/consensus/pkg/consensus/dkg"
	consensus_types "pocket/consensus/pkg/consensus/types"
	"pocket/consensus/pkg/p2p"
	"pocket/consensus/pkg/p2p/p2p_types"
	p2p_types_mocks "pocket/consensus/pkg/p2p/p2p_types/mocks"
	"pocket/consensus/pkg/pocket"
	"pocket/consensus/pkg/types"
	"pocket/shared"
	pcontext "pocket/shared/context"
	"pocket/shared/events"
	"pocket/shared/modules"
	mock_modules "pocket/shared/modules/mocks"
	"pocket/shared/typespb"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func GenerateNodeConfigs(n int) (configs []*config.Config) {
	for i := uint32(1); i <= uint32(n); i++ {
		c := config.Config{
			RootDir:    "",
			PrivateKey: types.GeneratePrivateKey(i),
			Genesis:    genesisJson(),

			P2P: &config.P2PConfig{
				ConsensusPort: 0,
				DebugPort:     0,
			},
			Consensus: &config.ConsensusConfig{
				NodeId: types.NodeId(i),
			},
			persistence: &config.persistenceConfig{},
			Utility:     &config.UtilityConfig{},
		}
		configs = append(configs, &c)
	}
	return
}

func CreateTestConsensusPocketNodes(
	t *testing.T,
	configs []*config.Config,
	testPocketBus modules.PocketBus,
) (pocketNodes map[types.NodeId]*pocket.PocketNode) {
	pocketNodes = make(map[types.NodeId]*pocket.PocketNode, len(configs))
	addrBook := getP2PAddrBook(configs)
	for _, config := range configs {
		pocketNode := CreateTestConsensusPocketNode(t, config, testPocketBus, addrBook)
		pocketNodes[config.Consensus.NodeId] = pocketNode
	}
	return
}

func CreateTestConsensusPocketNode(
	t *testing.T,
	config *config.Config,
	testPocketBus modules.PocketBus,
	addrBook []*p2p_types.NetworkPeer,
) *pocket.PocketNode {
	ctrl := gomock.NewController(nil)

	state := shared.GetPocketState()
	state.LoadStateFromConfig(config)

	ctx := pcontext.EmptyPocketContext()

	persistenceMock := mock_modules.NewMockpersistenceModule(ctrl)
	p2pNetworkMock := p2p_types_mocks.NewMockNetwork(ctrl)
	networkMock := mock_modules.NewMockNetworkModule(ctrl)
	utilityMock := mock_modules.NewMockUtilityModule(ctrl)

	baseMod, err := modules.NewBaseModule(ctx, config)
	require.NoError(t, err)

	consensusMod, err := consensus.Create(ctx, baseMod)
	require.NoError(t, err)

	pocketBusMod, err := shared.CreatePocketBus(nil, persistenceMock, networkMock, utilityMock, consensusMod)
	require.NoError(t, err)

	baseMod.SetPocketBusMod(pocketBusMod)

	pocketNode := &pocket.PocketNode{
		BasePocketModule: baseMod,
		persistenceMod:   persistenceMock,
		NetworkMod:       networkMock,
		UtilityMod:       utilityMock,
		ConsensusMod:     consensusMod,
	}

	// Base persistence mocks

	persistenceMock.EXPECT().
		Start(gomock.Any()).
		Do(func(ctx *pcontext.PocketContext) {
			log.Println("[MOCK] Start persistence mock")
		}).
		AnyTimes()

	persistenceMock.EXPECT().
		Stop(gomock.Any()).
		Do(func(ctx *pcontext.PocketContext) {
			log.Println("[MOCK] Stop persistence mock")
		}).
		AnyTimes()

	persistenceMock.EXPECT().
		GetLatestBlockHeight().
		Do(func() (uint64, error) {
			log.Println("[MOCK] GetLatestBlockHeight")
			return uint64(0), fmt.Errorf("[MOCK] GetLatestBlockHeight not implemented yet...")
		}).
		AnyTimes()

	persistenceMock.EXPECT().
		GetBlockHash(gomock.Any()).
		Do(func(height uint64) ([]byte, error) {
			return []byte(strconv.FormatUint(height, 10)), nil
		}).
		AnyTimes()

	// Base network module mocks

	p2pNetworkMock.EXPECT().
		GetAddrBook().
		DoAndReturn(func() []*p2p_types.NetworkPeer {
			log.Println("[MOCK] Network GetNetwork", addrBook)
			return addrBook
		}).
		AnyTimes()

	networkMock.EXPECT().
		Start(gomock.Any()).
		Do(func(ctx *pcontext.PocketContext) {
			log.Println("[MOCK] Start network mock")
		}).
		AnyTimes()

	networkMock.EXPECT().
		Stop(gomock.Any()).
		Do(func(ctx *pcontext.PocketContext) {
			log.Println("[MOCK] Stop network mock")
		}).
		AnyTimes()

	networkMock.EXPECT().
		GetNetwork().
		DoAndReturn(func() p2p_types.Network {
			return p2pNetworkMock
		}).
		AnyTimes()

	networkMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Any()).
		Do(func(ctx *pcontext.PocketContext, message *p2p_types.NetworkMessage, address types.NodeId) {
			networkMsg, _ := p2p.EncodeNetworkMessage(message)
			e := events.PocketEvent{PocketTopic: events.P2P_SEND_MESSAGE, MessageData: networkMsg}
			testPocketBus <- e
		}).
		AnyTimes()

	networkMock.EXPECT().
		// decoder
		Broadcast(gomock.Any(), gomock.Any()).
		Do(func(ctx *pcontext.PocketContext, message *p2p_types.NetworkMessage) {
			networkMsg, _ := p2p.EncodeNetworkMessage(message)
			e := events.PocketEvent{PocketTopic: events.P2P_BROADCAST_MESSAGE, MessageData: networkMsg}
			testPocketBus <- e
		}).
		AnyTimes()

	// Base utility mocks

	utilityMock.EXPECT().
		Start(gomock.Any()).
		Do(func(*pcontext.PocketContext) {
			log.Println("[MOCK] Start utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		Stop(gomock.Any()).
		Do(func(*pcontext.PocketContext) {
			log.Println("[MOCK] Stop utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		HandleTransaction(gomock.Any(), gomock.Any()).
		Do(func(*pcontext.PocketContext, *typespb.Transaction) {
			log.Println("[MOCK] HandleTransaction utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		HandleEvidence(gomock.Any(), gomock.Any()).
		Do(func(*pcontext.PocketContext, *typespb.Evidence) {
			log.Println("[MOCK] HandleEvidence utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		ReapMempool(gomock.Any()).
		Do(func(*pcontext.PocketContext) {
			log.Println("[MOCK] ReapMempool utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		BeginBlock(gomock.Any()).
		Do(func(*pcontext.PocketContext) {
			log.Println("[MOCK] BeginBlock utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		DeliverTx(gomock.Any(), gomock.Any()).
		Do(func(*pcontext.PocketContext, *typespb.Transaction) {
			log.Println("[MOCK] DeliverTx utility mock")
		}).
		AnyTimes()

	utilityMock.EXPECT().
		EndBlock(gomock.Any()).
		Do(func(*pcontext.PocketContext) {
			log.Println("[MOCK] Stop EndBlock mock")
		}).
		AnyTimes()

	return pocketNode
}

func WaitForNetworkConsensusMessage(
	t *testing.T,
	pocketBus modules.PocketBus,
	pocketEvent events.PocketEventTopic,
	step consensus.Step,
	numMessages int,
	millis time.Duration,
) (messages []consensus_types.GenericConsensusMessage) {
	printStatement := fmt.Sprintf("consensus step %s", consensus.StepToString[step])

	includeFilter := func(m consensus_types.GenericConsensusMessage) bool {
		hotstuffMsg, ok := m.(*consensus.HotstuffMessage)
		return ok && hotstuffMsg.Step == step
	}

	decoder := func(data []byte) consensus_types.GenericConsensusMessage {
		networkMsg, err := p2p.DecodeNetworkMessage(data)
		require.NoError(t, err)

		consensusMsg, err := consensus_types.DecodeConsensusMessage(networkMsg.Data)
		require.NoError(t, err)

		hotstuffMessage, ok := consensusMsg.Message.(*consensus.HotstuffMessage)
		if !ok {
			return nil
		}

		return hotstuffMessage
	}

	return WaitForNetworkConsensusMessageInternal(t, pocketBus, pocketEvent, numMessages, millis, decoder, includeFilter, printStatement)
}

func WaitFoNetworkDKGMessages(
	t *testing.T,
	pocketBus modules.PocketBus,
	pocketEvent events.PocketEventTopic,
	round dkg.DKGRound,
	numMessages int,
	millis time.Duration,
) (messages []*dkg.DKGMessage) {
	printStatement := fmt.Sprintf("DKG round %d", round)

	includeFilter := func(m consensus_types.GenericConsensusMessage) bool {
		dkgMsg, ok := m.(*dkg.DKGMessage)
		return ok && dkgMsg.Round == round
	}

	decoder := func(data []byte) consensus_types.GenericConsensusMessage {
		networkMsg, err := p2p.DecodeNetworkMessage(data)
		require.NoError(t, err)

		consensusMsg, err := consensus_types.DecodeConsensusMessage(networkMsg.Data)
		require.NoError(t, err)

		dkgMsg, ok := consensusMsg.Message.(*dkg.DKGMessage)
		if !ok {
			return nil
		}

		return dkgMsg
	}

	genericMessages := WaitForNetworkConsensusMessageInternal(t, pocketBus, pocketEvent, numMessages, millis, decoder, includeFilter, printStatement)
	for _, genericMsg := range genericMessages {
		messages = append(messages, genericMsg.(*dkg.DKGMessage))
	}

	return
}

func TriggerNextView(t *testing.T, node *pocket.PocketNode) {
	triggerDebugMessage(t, node, consensus.TriggerNextView)
}

func TriggerDKG(t *testing.T, node *pocket.PocketNode) {
	triggerDebugMessage(t, node, consensus.TriggerDKG)
}

func triggerDebugMessage(t *testing.T, node *pocket.PocketNode, action consensus.DebugMessageAction) {
	debugMessage := &consensus.DebugMessage{
		Action: action,
	}
	consensusMessage := &consensus_types.ConsensusMessage{
		Message: debugMessage,
		Sender:  0,
	}
	data, err := consensus_types.EncodeConsensusMessage(consensusMessage)
	require.NoError(t, err)

	event := events.PocketEvent{
		SourceModule: events.TEST,
		PocketTopic:  events.CONSENSUS,
		MessageData:  data,
	}
	node.GetPocketBusMod().PublishEventToBus(&event)
}

func P2PBroadcast(nodes map[types.NodeId]*pocket.PocketNode, message consensus_types.GenericConsensusMessage) {
	m := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  0,
	}

	event := prepareEvent(m)
	for _, node := range nodes {
		node.GetPocketBusMod().PublishEventToBus(&event)
	}
}

func P2PSend(node *pocket.PocketNode, message consensus_types.GenericConsensusMessage) {
	m := &consensus_types.ConsensusMessage{
		Message: message,
		Sender:  0,
	}

	event := prepareEvent(m)
	node.GetPocketBusMod().PublishEventToBus(&event)
}

func prepareEvent(message *consensus_types.ConsensusMessage) events.PocketEvent {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(message); err != nil {
		panic("Failed to encode message")
	}

	return events.PocketEvent{
		SourceModule: events.TEST,
		PocketTopic:  events.CONSENSUS,
		MessageData:  buff.Bytes(),
	}
}

// TODO: This copy-pasted code is just a quick workaround which
// can be very easily generalized using generics in Go 1.18. Leaving
// that for the migration to the main repo.
func WaitForNetworkConsensusMessageInternal(
	t *testing.T,
	testPocketBus modules.PocketBus,
	pocketEvent events.PocketEventTopic,
	numMessages int,
	millis time.Duration,
	decoder func([]byte) consensus_types.GenericConsensusMessage,
	includeFilter func(m consensus_types.GenericConsensusMessage) bool,
	errorMessage string,
) (messages []consensus_types.GenericConsensusMessage) {
	messages = make([]consensus_types.GenericConsensusMessage, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*millis)
	unused := make([]events.PocketEvent, 0) // TODO: Move this into a pool rather than resending back to the eventbus.
loop:
	for {
		select {
		case testEvent := <-testPocketBus:
			if testEvent.PocketTopic != pocketEvent {
				unused = append(unused, testEvent)
				continue
			}

			message := decoder(testEvent.MessageData)
			if message == nil || !includeFilter(message) {
				unused = append(unused, testEvent)
				continue
			}

			messages = append(messages, message)
			numMessages--
			if numMessages <= 0 {
				break loop
			}
		case <-ctx.Done():
			t.Fatalf("Missing %s messages; missing: %d, received: %d; (%s)", pocketEvent, numMessages, len(messages), errorMessage)
		}
	}
	cancel()
	for _, u := range unused {
		testPocketBus <- u
	}
	return
}

func getP2PAddrBook(configs []*config.Config) []*p2p_types.NetworkPeer {
	addrBook := make([]*p2p_types.NetworkPeer, len(configs))
	for idx, config := range configs {
		addrBook[idx] = &p2p_types.NetworkPeer{
			NodeId:    config.Consensus.NodeId,
			PublicKey: config.PrivateKey.Public(),
		}
	}
	return addrBook
}

func genesisJson() string {
	return `
	{
		"genesis_time": "2022-01-19T00:00:00.000000Z",
		"app_hash": "genesis_block_or_state_hash",
		"accounts": [
		  {
			"address": "04c56dfc51c3ec68d90a08a2efaa4b9d3db32b3b",
			"public_key": "03e6b38162ccdd0cd8ed657be73885e0b7b99ca09969729e3390c218cfcff07d",
			"upokt": 69
		  }
		],
		"consensus_params": {
			"max_mempool_bytes": 50000000,

			"max_block_bytes": 4000000,
			"max_transaction_bytes": 100000,

			"vrf_key_refresh_freq_block": 5,
			"vrf_key_validity_block": 5,

			"pace_maker": {
				"timeout_msec": 100000,
				"retry_timeout_msec": 1000,
				"max_timeout_msec": 60000,
				"min_block_freq_msec": 2000
			}
		},
		"validators": [
		  {
			"node_id": 1,
			"address": "71f8be163036c0da94f188bb817d77691869ccff5932059f3c398f2fb92fa08b",
			"public_key": "b1f804dabc68274c1233995c5a9119b56935bcdd83b7de07ec726dcedc4e9ce7",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node1.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  },
		  {
			"node_id": 2,
			"address": "0273a7327f5cd145ae29a12a76ffbfd4d89c0b78ca247450c05f556c24bc264f",
			"public_key": "6a0f6a283a8e4e86d2a3d60ef9e37ec33f2ab6071a30e0a477735128e7571eb0",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node2.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  },
		  {
			"node_id": 3,
			"address": "2a4156d371f8a49a88a6285e9f2ffd77947eac6801c0cfeccdb79ab4b8705f16",
			"public_key": "ab5696551fe1711c3c31669ff20e1e0bc12cb99917c3ab2412e7c13013dee7e7",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node3.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  },
		  {
			"node_id": 4,
			"address": "ffeb214baf0cc1b8019e91a5e5ba0aa71d58de2cc140dd6885147b5b26299fb8",
			"public_key": "d1f87d985adee0c3466ac0458745998fc0f39a9884897ce4c7548d1db8e10642",
			"jailed": false,
			"upokt": 5000000000000,
			"host": "node4.consensus",
			"port": 8080,
			"debug_port": 9080,
			"chains": ["0001", "0021"]
		  }
		]
	  }`
}
