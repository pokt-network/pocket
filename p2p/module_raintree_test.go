package p2p

import (
	"log"
	"os"
	"path/filepath"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/stretchr/testify/require"
	"github.com/golang/mock/gomock"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	files, err := filepath.Glob("*.json")
	if err != nil {
		log.Fatalf("Error finding json file: %v", err)
	}
	for _, f := range files {
		os.Remove(f)
	}
	os.Exit(exitCode)
}

// ### RainTree Unit Tests ###
func TestRainTreeNetworkCompleteOneNodes(t *testing.T) {
	// val_1
	originatorNode := validatorId(1)
	var expectedCalls = TestNetworkSimulationConfig{
		originatorNode: {0, 0}, // val_1, the originator, does 0 network reads or writes
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := validatorId(1)
	// Per the diagram above, in the case of a 2 node network, the originator node (val_1) does a
	// single write to another node (val_2),  also the
	// originator node and never performs any reads or writes during a RainTree broadcast.
	var expectedCalls = TestNetworkSimulationConfig{
		// Attempt: I think Validator 1 is sending a message in a 2 (including self) node network
		originatorNode: {0, 1}, // val_1 does a single network write (to val_2)
		validatorId(2): {1, 0}, // val_2 does a single network read (from val_1)
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := validatorId(1)
	var expectedCalls = TestNetworkSimulationConfig{
		originatorNode: {0, 2}, // val_1 does two network writes (to val_2 and val_3)
		validatorId(2): {1, 0}, // val_2 does a single network read (from val_1)
		validatorId(3): {1, 0}, // val_2 does a single network read (from val_3)
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
	var expectedCalls = TestNetworkSimulationConfig{
		originatorNode: {0, 3}, // val_1 does 3 network writes (two to val_2 and 1 to val_3)
		validatorId(2): {2, 1}, // val_2 does 2 network reads (both from val_1) and 1 network write (to val_3)
		validatorId(3): {2, 1}, // val_2 does 2 network reads (from val_1 and val_2) and 1 network write (to val_4)
		validatorId(4): {1, 0}, // val_2 does 1 network read (from val_3)
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
	var expectedCalls = TestNetworkSimulationConfig{
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
	var expectedCalls = TestNetworkSimulationConfig{
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
	var expectedCalls = TestNetworkSimulationConfig{
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

// ### RainTree Unit Helpers - To remove redundancy of code in the unit tests ###

func testRainTreeCalls(t *testing.T, origNode string, testCommConfig TestRainTreeCommConfig, isOriginatorPinged bool) {
	// Network configurations
	numValidators := len(testCommConfig)
	runtimeConfigs := createMockRuntimeMgrs(t, numValidators)

	// Test configurations
	var messageHandeledWaitGroup sync.WaitGroup
	if isOriginatorPinged {
		messageHandeledWaitGroup.Add(numValidators)
	} else {
		messageHandeledWaitGroup.Add(numValidators - 1) // -1 because the originator node implicitly handles the message
	}

	// Network initialization
	consensusMock := prepareConsensusMock(t, runtimeConfigs[0].GetGenesis())
	telemetryMock := prepareTelemetryMock(t)
	connMocks := make(map[string]typesP2P.Transport)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range testCommConfig {
		connMocks[valId] = prepareConnMock(t, expectedCall.numNetworkReads, expectedCall.numNetworkWrites)
		busMocks[valId] = prepareBusMock(t, &messageHandeledWaitGroup, consensusMock, telemetryMock)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, runtimeConfigs)
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
	p2pMod.Broadcast(p, debug.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	done := make(chan struct{})
	go func() {
		messagesHandledWg.Wait()
		close(done)
	}()

	// Timeout or succeed
	select {
	case <-done:
	// All done!
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message to be handled")
	}
}

// ### RainTree Unit Utils - Configurations & constants and such ###

const (
	genesisConfigSeedStart = 42
	maxNumKeys             = 42 // The number of keys generated for all the unit tests. Optimization to avoid regenerating every time.
	serviceUrlFormat       = "val_%d"
	testChannelSize        = 10000
)

// TODO(olshansky): Add configurations tests for dead and partially visible nodes
type TestRainTreeCommConfig map[string]struct {
	numNetworkReads  uint16
	numNetworkWrites uint16
}

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

// A mock of the application specific to know if a message was sent to be handled by the application
// INVESTIGATE(olshansky): Double check that how the expected calls are counted is accurate per the
//                         expectation with RainTree by comparing with Telemetry after updating specs.
func prepareBusMock(t *testing.T, wg *sync.WaitGroup, consensusMock *modulesMock.MockConsensusModule, telemetryMock *modulesMock.MockTelemetryModule) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)

	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(func(e *debug.PocketEvent) {
		wg.Done()
		log.Println("App specific bus mock publishing event to bus")
	}).MaxTimes(1) // Using `MaxTimes` rather than `Times` because originator node implicitly handles the message

	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return busMock
}

func prepareConsensusMock(t *testing.T, genesisState modules.GenesisState) *modulesMock.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := modulesMock.NewMockConsensusModule(ctrl)

	validators := genesisState.GetPersistenceGenesisState().GetVals()
	m := make(modules.ValidatorMap, len(validators))
	for _, v := range validators {
		m[v.GetAddress()] = v
	}

	consensusMock.EXPECT().ValidatorMap().Return(m).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()
	return consensusMock
}

func prepareTelemetryMock(t *testing.T) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareEventMetricsAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

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

// The reason with use `MaxTimes` instead of `Times` here is because we could have gotten full coverage
// while a message was still being sent that would have later been dropped due to de-duplication. There
// is a race condition here, but it is okay because our goal is to achieve max coverage with an upper limit
// on the number of expected messages propagated.
// INVESTIGATE(olshansky): Double check that how the expected calls are counted is accurate per the
//                         expectation with RainTree by comparing with Telemetry after updating specs.
func prepareConnMock(t *testing.T, expectedNumNetworkReads, expectedNumNetworkWrites uint16) typesP2P.Transport {
	testChannel := make(chan []byte, testChannelSize)
	ctrl := gomock.NewController(t)
	connMock := mocksP2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		data := <-testChannel
		return data, nil
	}).MaxTimes(int(expectedNumNetworkReads + 1))

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).MaxTimes(int(expectedNumNetworkWrites))

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}

// prepareP2PModules returns a map of configured p2pModules keyed by an incremental naming convention (eg: `val_1`, `val_2`, etc.)
func prepareP2PModules(t *testing.T, runtimeConfigs []modules.RuntimeMgr) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(runtimeConfigs))
	for i, runtimeConfig := range runtimeConfigs {
		p2pMod, err := Create(runtimeConfig)
		require.NoError(t, err)
		p2pModules[validatorId(t, i+1)] = p2pMod.(*p2pModule)
	}
	return
}

// createMockRuntimeMgrs creates `numValidators` instances of mocked `RuntimeMgr` that are essentially
// representing the runtime environments of the validators that we will use in our tests
func createMockRuntimeMgrs(t *testing.T, numValidators int) []modules.RuntimeMgr {
	ctrl := gomock.NewController(t)
	mockRuntimeMgrs := make([]modules.RuntimeMgr, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	mockGenesisState := createMockGenesisState(t, valKeys)
	for i := range mockRuntimeMgrs {
		mockConfig := modulesMock.NewMockConfig(ctrl)
		mockConfig.EXPECT().GetBaseConfig().Return(&runtime.BaseConfig{
			RootDirectory: "",
			PrivateKey:    valKeys[i].String(),
		}).AnyTimes()
		mockConfig.EXPECT().GetP2PConfig().Return(&typesP2P.P2PConfig{
			PrivateKey:            valKeys[i].String(),
			ConsensusPort:         8080,
			UseRainTree:           true,
			IsEmptyConnectionType: true,
		}).AnyTimes()

		mockRuntimeMgr := modulesMock.NewMockRuntimeMgr(ctrl)
		mockRuntimeMgr.EXPECT().GetConfig().Return(mockConfig).AnyTimes()
		mockRuntimeMgr.EXPECT().GetGenesis().Return(mockGenesisState).AnyTimes()
		mockRuntimeMgrs[i] = mockRuntimeMgr
	}
	return mockRuntimeMgrs
}

func validatorId(_ *testing.T, i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
}

// createMockGenesisState configures and returns a mocked GenesisState
func createMockGenesisState(t *testing.T, valKeys []cryptoPocket.PrivateKey) modules.GenesisState {
	ctrl := gomock.NewController(t)

	validators := make([]modules.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := modulesMock.NewMockActor(ctrl)
		mockActor.EXPECT().GetAddress().Return(addr).AnyTimes()
		mockActor.EXPECT().GetPublicKey().Return(valKey.PublicKey().String()).AnyTimes()
		mockActor.EXPECT().GetGenericParam().Return(validatorId(t, i+1)).AnyTimes()
		mockActor.EXPECT().GetStakedAmount().Return("1000000000000000").AnyTimes()
		mockActor.EXPECT().GetPausedHeight().Return(int64(0)).AnyTimes()
		mockActor.EXPECT().GetUnstakingHeight().Return(int64(0)).AnyTimes()
		mockActor.EXPECT().GetOutput().Return(addr).AnyTimes()
		validators[i] = mockActor
	}

	mockPersistenceGenesisState := modulesMock.NewMockPersistenceGenesisState(ctrl)
	mockPersistenceGenesisState.EXPECT().
		GetVals().
		Return(validators).AnyTimes()

	mockGenesisState := modulesMock.NewMockGenesisState(ctrl)
	mockGenesisState.EXPECT().
		GetPersistenceGenesisState().
		Return(mockPersistenceGenesisState).AnyTimes()
	return mockGenesisState

}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	files, err := filepath.Glob("*.json")
	if err != nil {
		log.Fatalf("Error finding json file: %v", err)
	}
	for _, f := range files {
		os.Remove(f)
	}
	os.Exit(exitCode)
}
>>>>>>> main
