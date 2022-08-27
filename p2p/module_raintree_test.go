package p2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	p2pTelemetry "github.com/pokt-network/pocket/p2p/telemetry"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

// IMPROVE(team): Looking into adding more tests and accounting for more edge cases.

// ~~~~~~ RainTree Unit Tests ~~~~~~

func TestRainTreeNetworkCompleteOneNodes(t *testing.T) {
	// val_1
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode: {0, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode: {0, 1},
		validatorId(2): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode: {0, 2},
		validatorId(2): {1, 0},
		validatorId(3): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteFourNodes(t *testing.T) {
	// Test configurations (visualization retrieved from simulator)
	// 	                val_1
	// 	  ┌───────────────┴────┬─────────────────┐
	//  val_2                val_1             val_3
	//    └───────┐            └───────┐         └───────┐
	// 		    val_3                val_2             val_4
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode: {0, 3},
		validatorId(2): {2, 1},
		validatorId(3): {2, 1},
		validatorId(4): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteNineNodes(t *testing.T) {
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode: {0, 4},
		validatorId(2): {1, 0},
		validatorId(3): {1, 0},
		validatorId(4): {1, 2},
		validatorId(5): {1, 0},
		validatorId(6): {1, 0},
		validatorId(7): {1, 2},
		validatorId(8): {1, 0},
		validatorId(9): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteEighteenNodes(t *testing.T) {
	// 	                                                                                                              val_1
	// 	                                      ┌──────────────────────────────────────────────────────────────────────────┴─────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                      val_7                                                                                                            val_1                                                                                                     val_13
	//             ┌──────────────────────────┴────────────┬────────────────────────────────────┐                                     ┌────────────────────────┴────────────┬──────────────────────────────────┐                                ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//           val_11                                   val_7                               val_15                                 val_5                                 val_1                              val_9                           val_17                                  val_13                                val_3
	//    ┌────────┴─────┬───────────┐             ┌───────┴────┬──────────┐           ┌────────┴─────┬──────────┐            ┌───────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴────┬─────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_13         val_11      val_16        val_9        val_7      val_12      val_17         val_15     val_8        val_7        val_5      val_10        val_3        val_1     val_6      val_11        val_9     val_2     val_1         val_17     val_4         val_15         val_13      val_18     val_5        val_3      val_14
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode:  {1, 6},
		validatorId(2):  {1, 0},
		validatorId(3):  {2, 2},
		validatorId(4):  {1, 0},
		validatorId(5):  {2, 2},
		validatorId(6):  {1, 0},
		validatorId(7):  {2, 4},
		validatorId(8):  {1, 0},
		validatorId(9):  {2, 2},
		validatorId(10): {1, 0},
		validatorId(11): {2, 2},
		validatorId(12): {1, 0},
		validatorId(13): {2, 4},
		validatorId(14): {1, 0},
		validatorId(15): {2, 2},
		validatorId(16): {1, 0},
		validatorId(17): {2, 2},
		validatorId(18): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwentySevenNodes(t *testing.T) {
	// 	                                                                                                                    val_1
	// 	                                     ┌────────────────────────────────────────────────────────────────────────────────┴───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                    val_10                                                                                                                   val_1                                                                                                       val_19
	//            ┌──────────────────────────┴──────────────┬──────────────────────────────────────┐                                         ┌────────────────────────┴────────────┬──────────────────────────────────┐                                  ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//          val_16                                    val_10                                 val_22                                     val_7                                 val_1                             val_13                             val_25                                  val_19                                val_4
	//   ┌────────┴─────┬───────────┐              ┌────────┴─────┬───────────┐           ┌────────┴─────┬───────────┐              ┌────────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_20         val_16      val_24         val_14         val_10      val_18      val_26         val_22      val_12         val_11        val_7      val_15        val_5        val_1     val_9      val_17         val_13     val_3     val_2         val_25     val_6         val_23         val_19      val_27     val_8        val_4      val_21
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		originatorNode:  {0, 6},
		validatorId(2):  {1, 0},
		validatorId(3):  {1, 0},
		validatorId(4):  {1, 2},
		validatorId(5):  {1, 0},
		validatorId(6):  {1, 0},
		validatorId(7):  {1, 2},
		validatorId(8):  {1, 0},
		validatorId(9):  {1, 0},
		validatorId(10): {1, 4},
		validatorId(11): {1, 0},
		validatorId(12): {1, 0},
		validatorId(13): {1, 2},
		validatorId(14): {1, 0},
		validatorId(15): {1, 0},
		validatorId(16): {1, 2},
		validatorId(17): {1, 0},
		validatorId(18): {1, 0},
		validatorId(19): {1, 4},
		validatorId(20): {1, 0},
		validatorId(21): {1, 0},
		validatorId(22): {1, 2},
		validatorId(23): {1, 0},
		validatorId(24): {1, 0},
		validatorId(25): {1, 2},
		validatorId(26): {1, 0},
		validatorId(27): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

// ~~~~~~ RainTree Unit Test Helpers ~~~~~~

// TODO(drewsky): Add configurations tests for dead and partially visible nodes
type TestRainTreeCommConfig map[string]struct {
	// The number of asynchronous reads the node's P2P listener made (i.e. # of messages it received over the network)
	numNetworkReads int
	// The number of asynchronous writes the node's P2P listener made (i.e. # of messages it tried to send over the network)
	numNetworkWrites int
}

func validatorId(i int) string {
	return test_artifacts.GetServiceUrl(i)
}

func testRainTreeCalls(t *testing.T, origNode string, testCommConfig TestRainTreeCommConfig) {
	// Network configurations
	numValidators := len(testCommConfig)
	configs, genesisState := createConfigs(t, numValidators)

	// Network initialization
	consensusMock := prepareConsensusMock(t, genesisState)
	connMocks := make(map[string]typesP2P.Transport)
	busMocks := make(map[string]modules.Bus)
	var messageHandledWaitGroup sync.WaitGroup
	for valId, expectedCall := range testCommConfig {
		messageHandledWaitGroup.Add(expectedCall.numNetworkReads + 1)
		connMocks[valId] = prepareConnMock(t, &messageHandledWaitGroup, expectedCall.numNetworkReads)
		messageHandledWaitGroup.Add(expectedCall.numNetworkWrites)
		telemetryMock := prepareTelemetryMock(t, &messageHandledWaitGroup, expectedCall.numNetworkWrites)
		busMocks[valId] = prepareBusMock(t, configs[valId], consensusMock, telemetryMock)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, configs)
	for validatorId, p2pMod := range p2pModules {
		p2pMod.listener = connMocks[validatorId]
		p2pMod.SetBus(busMocks[validatorId])
		p2pMod.Start()
		for _, peer := range p2pMod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		defer p2pMod.Stop()
	}

	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[origNode]
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	done := make(chan struct{})
	go func() {
		messageHandledWaitGroup.Wait()
		close(done)
	}()

	// Timeout or succeed
	select {
	case <-done:
	// All done!
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message to be handled")
	}
}

// ~~~~~~ RainTree Unit Test Mocks ~~~~~~

// A mock of the application specific to know if a message was sent to be handled by the application
func prepareBusMock(t *testing.T, config *genesis.Config, consensusMock *modulesMock.MockConsensusModule, telemetryMock *modulesMock.MockTelemetryModule) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)

	busMock.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	busMock.EXPECT().GetConfig().Return(config).Times(1)
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return busMock
}

func prepareConsensusMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := modulesMock.NewMockConsensusModule(ctrl)

	validators := genesisState.Utility.Validators
	m := make(modules.ValidatorMap, len(validators))
	for _, v := range validators {
		m[v.Address] = v
	}

	consensusMock.EXPECT().ValidatorMap().Return(m).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	return consensusMock
}

func prepareTelemetryMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkWrites int) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareEventMetricsAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Eq(p2pTelemetry.RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).Do(func(n, e interface{}, l ...interface{}) {
		wg.Done()
	}).Times(expectedNumNetworkWrites)
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Not(p2pTelemetry.RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL), gomock.Any()).AnyTimes()

	return telemetryMock
}

func prepareTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)
	return timeseriesAgentMock
}

func prepareEventMetricsAgentMock(t *testing.T) *modulesMock.MockEventMetricsAgent {
	ctrl := gomock.NewController(t)
	eventMetricsAgentMock := modulesMock.NewMockEventMetricsAgent(ctrl)
	return eventMetricsAgentMock
}

func prepareConnMock(t *testing.T, wg *sync.WaitGroup, expectedNumNetworkReads int) typesP2P.Transport {
	testChannel := make(chan []byte, 1000)
	ctrl := gomock.NewController(t)
	connMock := mocksP2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		wg.Done()
		data := <-testChannel
		return data, nil
	}).Times(expectedNumNetworkReads + 1) // +1 is necessary because there is one extra read of empty data by every channel when it starts

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).Times(expectedNumNetworkReads)

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}

func prepareP2PModules(t *testing.T, configs map[string]*genesis.Config) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(configs))
	for valId, config := range configs {
		p2pMod, err := Create(config, nil)
		require.NoError(t, err)
		p2pModules[valId] = p2pMod.(*p2pModule)
	}
	return
}

// ~~~~~~ RainTree Unit Test Configurations ~~~~~~

// TODO(drewsky): We should leverage the shared `test_artifacts` package and replace the code below.
//                The reason it is necessary right now is because the funcionality of RainTree is dependant
//                on the order of the addresses, which is a function of the public key, so we need to make
//                sure that the validatorId order corresponds to that of the public keys.

const (
	genesisConfigSeedStart = 42
	maxNumKeys             = 42 // The number of keys generated for all the unit tests. Optimization to avoid regenerating every time.
)

var keys []cryptoPocket.PrivateKey

func init() {
	keys = generateKeys(nil, maxNumKeys)
}

func generateKeys(_ *testing.T, numValidators int) []cryptoPocket.PrivateKey {
	keys := make([]cryptoPocket.PrivateKey, numValidators)

	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		seed := make([]byte, ed25519.PrivateKeySize)
		binary.LittleEndian.PutUint32(seed, uint32(seedInt))
		pk, err := cryptoPocket.NewPrivateKeyFromSeed(seed)
		if err != nil {
			panic(err)
		}
		keys[i] = pk
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Address().String() < keys[j].Address().String()
	})
	return keys
}

func createConfigs(t *testing.T, numValidators int) (configs map[string]*genesis.Config, genesisState *genesis.GenesisState) {
	configs = make(map[string]*genesis.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	genesisState = createGenesisState(t, valKeys)

	for i := 0; i < numValidators; i++ {
		configs[validatorId(i+1)] = &genesis.Config{
			Base: &genesis.BaseConfig{
				RootDirectory: "",
				PrivateKey:    valKeys[i].String(),
			},
			Consensus:   &genesis.ConsensusConfig{},
			Utility:     &genesis.UtilityConfig{},
			Persistence: &genesis.PersistenceConfig{},
			P2P: &genesis.P2PConfig{
				ConsensusPort:  8080,
				UseRainTree:    true,
				ConnectionType: genesis.ConnectionType_EmptyConnection,
			},
			Telemetry: nil,
		}
	}
	return
}

func createGenesisState(t *testing.T, valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	validators := make([]*genesis.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		val := &genesis.Actor{
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(i + 1),
			StakedAmount:    "1000000000000000",
			PausedHeight:    0,
			UnstakingHeight: 0,
			Output:          addr,
		}
		validators[i] = val
	}
	return &genesis.GenesisState{
		Utility: &genesis.UtilityGenesisState{
			Validators: validators,
		},
	}
}
