//go:build test

package p2p_test

import (
	runtime_testutil "github.com/pokt-network/pocket/internal/testutil/runtime"
	telemetry_testutil "github.com/pokt-network/pocket/internal/testutil/telemetry"
	"github.com/pokt-network/pocket/p2p"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/constructors"
	persistence_testutil "github.com/pokt-network/pocket/internal/testutil/persistence"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
)

// TODO(#314): Add the tooling and instructions on how to generate unit tests in this file.

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
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode: {0, 0}, // val_1, the originator, does 0 network reads or writes
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := testutil.NewServiceURL(1)
	// Per the diagram above, in the case of a 2 node network, the originator node (val_1) does a
	// single write to another node (val_2),  also the
	// originator node and never performs any reads or writes during a RainTree broadcast.
	expectedCalls := TestNetworkSimulationConfig{
		// Attempt: I think Validator 1 is sending a message in a 2 (including self) node network originatorNode:                {0, 1}, // val_1 does a single network write (to val_2)
		testutil.NewServiceURL(2): {1, 0}, // val_2 does a single network read (from val_1)
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:            {0, 2}, // val_1 does two network writes (to val_2 and val_3)
		testutil.NewServiceURL(2): {1, 0}, // val_2 does a single network read (from val_1)
		testutil.NewServiceURL(3): {1, 0}, // val_2 does a single network read (from val_3)
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
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:            {0, 3}, // val_1 does 3 network writes (two to val_2 and 1 to val_3)
		testutil.NewServiceURL(2): {2, 1}, // val_2 does 2 network reads (both from val_1) and 1 network write (to val_3)
		testutil.NewServiceURL(3): {2, 1}, // val_2 does 2 network reads (from val_1 and val_2) and 1 network write (to val_4)
		testutil.NewServiceURL(4): {1, 0}, // val_2 does 1 network read (from val_3)
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteNineNodes(t *testing.T) {
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:            {0, 4},
		testutil.NewServiceURL(2): {1, 0},
		testutil.NewServiceURL(3): {1, 0},
		testutil.NewServiceURL(4): {1, 2},
		testutil.NewServiceURL(5): {1, 0},
		testutil.NewServiceURL(6): {1, 0},
		testutil.NewServiceURL(7): {1, 2},
		testutil.NewServiceURL(8): {1, 0},
		testutil.NewServiceURL(9): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

//                                                                                                            val_1
//                                     ┌────────────────────────────────────────────────────────────────────────┴───────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┐
//                                   val_5                                                                                                        val_1                                                                                                     val_9
//            ┌────────────────────────┴────────────┬──────────────────────────────────┐                                     ┌──────────────────────┴────────────┬────────────────────────────────┐                                  ┌────────────────────────┴──────────────┬──────────────────────────────────┐
//          val_7                                 val_5                             val_10                                 val_3                               val_1                            val_6                             val_11                                   val_9                              val_2
//    ┌───────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐            ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐             ┌────────┴────┬──────────┐          ┌───────┴────┬─────────┐
//  val_8        val_7      val_10        val_6        val_5     val_8      val_11         val_10     val_5        val_4        val_3     val_6        val_2        val_1     val_4     val_7        val_6     val_1      val_12         val_11     val_2         val_10        val_9      val_12     val_3        val_2     val_9

func TestRainTreeCompleteTwelveNodes(t *testing.T) {
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:             {1, 6},
		testutil.NewServiceURL(2):  {3, 2},
		testutil.NewServiceURL(3):  {2, 2},
		testutil.NewServiceURL(4):  {2, 0},
		testutil.NewServiceURL(5):  {2, 4},
		testutil.NewServiceURL(6):  {3, 2},
		testutil.NewServiceURL(7):  {2, 2},
		testutil.NewServiceURL(8):  {2, 0},
		testutil.NewServiceURL(9):  {2, 4},
		testutil.NewServiceURL(10): {3, 2},
		testutil.NewServiceURL(11): {2, 2},
		testutil.NewServiceURL(12): {2, 0},
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
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:             {1, 6},
		testutil.NewServiceURL(2):  {1, 0},
		testutil.NewServiceURL(3):  {2, 2},
		testutil.NewServiceURL(4):  {1, 0},
		testutil.NewServiceURL(5):  {2, 2},
		testutil.NewServiceURL(6):  {1, 0},
		testutil.NewServiceURL(7):  {2, 4},
		testutil.NewServiceURL(8):  {1, 0},
		testutil.NewServiceURL(9):  {2, 2},
		testutil.NewServiceURL(10): {1, 0},
		testutil.NewServiceURL(11): {2, 2},
		testutil.NewServiceURL(12): {1, 0},
		testutil.NewServiceURL(13): {2, 4},
		testutil.NewServiceURL(14): {1, 0},
		testutil.NewServiceURL(15): {2, 2},
		testutil.NewServiceURL(16): {1, 0},
		testutil.NewServiceURL(17): {2, 2},
		testutil.NewServiceURL(18): {1, 0},
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
	originatorNode := testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:             {0, 6},
		testutil.NewServiceURL(2):  {1, 0},
		testutil.NewServiceURL(3):  {1, 0},
		testutil.NewServiceURL(4):  {1, 2},
		testutil.NewServiceURL(5):  {1, 0},
		testutil.NewServiceURL(6):  {1, 0},
		testutil.NewServiceURL(7):  {1, 2},
		testutil.NewServiceURL(8):  {1, 0},
		testutil.NewServiceURL(9):  {1, 0},
		testutil.NewServiceURL(10): {1, 4},
		testutil.NewServiceURL(11): {1, 0},
		testutil.NewServiceURL(12): {1, 0},
		testutil.NewServiceURL(13): {1, 2},
		testutil.NewServiceURL(14): {1, 0},
		testutil.NewServiceURL(15): {1, 0},
		testutil.NewServiceURL(16): {1, 2},
		testutil.NewServiceURL(17): {1, 0},
		testutil.NewServiceURL(18): {1, 0},
		testutil.NewServiceURL(19): {1, 4},
		testutil.NewServiceURL(20): {1, 0},
		testutil.NewServiceURL(21): {1, 0},
		testutil.NewServiceURL(22): {1, 2},
		testutil.NewServiceURL(23): {1, 0},
		testutil.NewServiceURL(24): {1, 0},
		testutil.NewServiceURL(25): {1, 2},
		testutil.NewServiceURL(26): {1, 0},
		testutil.NewServiceURL(27): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

// ### RainTree Unit Helpers - To remove redundancy of code in the unit tests ###

// Helper function that can be used for end-to-end P2P module tests that creates a "real" P2P module.
// 1. It creates and configures a "real" P2P module where all the other components of the node are mocked.
// 2. It then triggers a single message and waits for all of the expected messages transmission to complete before announcing failure.
func testRainTreeCalls(t *testing.T, origNode string, networkSimulationConfig TestNetworkSimulationConfig) {
	dnsSrv := testutil.MinimalDNSMock(t)

	// Configure & prepare test module
	numValidators := len(networkSimulationConfig)
	//runtimeConfigs := createMockRuntimeMgrs(t, numValidators)
	//genesisState := runtimeConfigs[0].GetGenesis()

	privKeys := testutil.LoadLocalnetPrivateKeys(t, numValidators)
	pubKeys := make([]cryptoPocket.PublicKey, len(privKeys))
	for i, privKey := range privKeys {
		pubKeys[i] = privKey.PublicKey()
	}

	genesisState := runtime_testutil.GenesisWithSequentialServiceURLs(t, pubKeys)

	var (
		wg         sync.WaitGroup
		busMocks   map[string]*mock_modules.MockBus
		p2pModules map[string]modules.P2PModule
	)
	//busMocks := createMockBuses(t, runtimeConfigs, &wg)
	busEventHandlerFactory := func(t gocuke.TestingT, busMock *mock_modules.MockBus) testutil.BusEventHandler {
		return func(data *messaging.PocketEnvelope) {
			//p2pCfg := busMock.GetRuntimeMgr().GetConfig().P2P
			//
			//// `p2pModule#handleNetworkData()` calls `modules.Bus#PublishEventToBus()`
			//// assumes that P2P module is the only bus event producer running during
			//// the test.
			//t.Logf("[valId: %s:%d] Read", p2pCfg.Hostname, p2pCfg.Port)
			//wg.Done()
		}
	}

	busMocks, _, p2pModules = constructors.NewBusesMocknetAndP2PModules(
		t, numValidators,
		dnsSrv,
		genesisState,
		busEventHandlerFactory,
		nil,
	)

	//for _, busMock := range busMocks {
	//	telemetryMock := busMock.GetTelemetryModule().(*mock_modules.MockTelemetryModule)
	//	telemetryMock.GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent).EXPECT().EmitEvent(
	//		gomock.Any(),
	//		gomock.Any(),
	//		gomock.Any(),
	//		gomock.Any(),
	//	).AnyTimes()
	//}

	//serviceURLs := make([]string, 0, numValidators)
	//for valId := range networkSimulationConfig {
	//	serviceURLs = append(serviceURLs, valId)
	//}
	//
	//// TODO_THIS_COMMIT: need this?
	//// sort `serviceURLs` in ascending order
	//sort.Slice(serviceURLs, func(i, j int) bool {
	//	iId := extractNumericId(serviceURLs[i])
	//	jId := extractNumericId(serviceURLs[j])
	//	return iId < jId
	//})

	// Create connection and bus mocks along with a shared WaitGroup to track the number of expected
	// reads and writes throughout the mocked local network
	for serviceURL, busMock := range busMocks {
		expectedCall := networkSimulationConfig[serviceURL]
		expectedReads := expectedCall.numNetworkReads
		expectedWrites := expectedCall.numNetworkWrites

		log.Printf("[serviceURL: %s] expected reads: %d\n", serviceURL, expectedReads)
		log.Printf("[serviceURL: %s] expected writes: %d\n", serviceURL, expectedWrites)
		wg.Add(expectedReads)
		wg.Add(expectedWrites)

		// TODO_THIS_COMMIT:
		if serviceURL == origNode {
			t.Log("SELF SEND!!!! 1")
			t.Log("SELF SEND!!!! 2")
			t.Log("SELF SEND!!!! 3")
			//...
		}

		// TODO_THIS_COMMIT: MOVE
		// -- option 2a
		//telemetryEventHandler := func(namespace, eventName string, labels ...any) {
		//	t.Log("telemetry event received")
		//	wg.Done()
		//}

		// TODO_THIS_COMMIT: refactor
		// -- option 1
		//eventMetricsAgentMock := telemetry_testutil.PrepareEventMetricsAgentMock(t, serviceURL, &wg, expectedWrites)
		//busMock.GetTelemetryModule().(*mock_modules.MockTelemetryModule).EXPECT().
		//	GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

		// option 1.5
		eventMetricsAgentMock := busMock.
			GetTelemetryModule().
			GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent)

		//telemetry_testutil.WithEventMetricsHandler(
		telemetry_testutil.WhyEventMetricsAgentMock(
			t, eventMetricsAgentMock,
			serviceURL, &wg,
			//telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL,
			//telemetryEventHandler,
			expectedWrites,
		)
		// --

		_ = persistence_testutil.BasePersistenceMock(t, busMock, genesisState)
		//telemetryMock := prepareTelemetryMock(t, busMock, serviceURL, &wg, expectedWrites)

		//telemetryMock := telemetry_testutil.WithTimeSeriesAgent(
		//	t, telemetry_testutil.MinimalTelemetryMock(t, busMock))
		//
		//eventMetricsAgentMock := telemetry_testutil.EventMetricsAgentMockWithHandler(
		//	t, telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL,
		//	telemetryEventHandler,
		//	expectedWrites,
		//)
		//telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()

		// -- option 2b
		//eventMetricsAgentMock := busMock.
		//	GetTelemetryModule().
		//	GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent)
		//
		////telemetry_testutil.WithEventMetricsHandler(
		//telemetry_testutil.WhyEventMetricsAgentMock(
		//	t, eventMetricsAgentMock,
		//	//telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL,
		//	telemetryEventHandler,
		//	expectedWrites,
		//)
		// --

		//telemetryMock := telemetry_testutil.BaseTelemetryMock(t, busMock)
		//telemetryMock.GetEventMetricsAgent().(*mock_modules.MockEventMetricsAgent).EXPECT().EmitEvent(
		//	gomock.Any(),
		//	gomock.Any(),
		//	gomock.Eq(telemetry.P2P_RAINTREE_MESSAGE_EVENT_METRIC_SEND_LABEL),
		//	gomock.Any(),
		//	//).Do(telemetryEventHandler).Times(expectedWrites)
		//).Do(telemetryEventHandler).AnyTimes()
	}

	// Start all p2p modules
	for serviceURL, p2pMod := range p2pModules {
		err := p2pMod.Start()
		require.NoError(t, err)

		// TODO_THIS_COMMIT: consider using BusEventHandler instead...
		sURL := strings.Clone(serviceURL)
		mod := *(p2pMod.(*p2p.P2PModule))
		mod.GetRainTreeRouter().HandlerProxy(t, func(origHandler typesP2P.RouterHandler) typesP2P.RouterHandler {
			return func(data []byte) error {
				log.Printf("[valID: %s] Read\n", sURL)
				wg.Done()
				return nil
			}
		})
	}

	// Wait for completion
	defer waitForNetworkSimulationCompletion(t, &wg)
	t.Cleanup(func() {
		// Stop all p2p modules
		for _, p2pMod := range p2pModules {
			err := p2pMod.Stop()
			require.NoError(t, err)
		}
	})

	// Send the first message (by the originator) to trigger a RainTree broadcast
	p := &anypb.Any{}
	p2pMod := p2pModules[origNode]
	require.NoError(t, p2pMod.Broadcast(p))
}

func extractNumericId(valId string) int64 {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(valId)

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return -1
	}

	return num
}
