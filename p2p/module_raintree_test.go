package p2p

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"testing"

	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
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

//                                                                                                            val_1
//                                     ┌────────────────────────────────────────────────────────────────────────┴───────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┐
//                                   val_5                                                                                                        val_1                                                                                                     val_9
//            ┌────────────────────────┴────────────┬──────────────────────────────────┐                                     ┌──────────────────────┴────────────┬────────────────────────────────┐                                  ┌────────────────────────┴──────────────┬──────────────────────────────────┐
//          val_7                                 val_5                             val_10                                 val_3                               val_1                            val_6                             val_11                                   val_9                              val_2
//    ┌───────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐            ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐             ┌────────┴────┬──────────┐          ┌───────┴────┬─────────┐
//  val_8        val_7      val_10        val_6        val_5     val_8      val_11         val_10     val_5        val_4        val_3     val_6        val_2        val_1     val_4     val_7        val_6     val_1      val_12         val_11     val_2         val_10        val_9      val_12     val_3        val_2     val_9

func TestRainTreeCompleteTwelveNodes(t *testing.T) {
	originatorNode := validatorId(1)
	var expectedCalls = TestNetworkSimulationConfig{
		originatorNode:  {1, 6},
		validatorId(2):  {3, 2},
		validatorId(3):  {2, 2},
		validatorId(4):  {2, 0},
		validatorId(5):  {2, 4},
		validatorId(6):  {3, 2},
		validatorId(7):  {2, 2},
		validatorId(8):  {2, 0},
		validatorId(9):  {2, 4},
		validatorId(10): {3, 2},
		validatorId(11): {2, 2},
		validatorId(12): {2, 0},
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

// Helper function that can be used for end-to-end P2P module tests that creates a "real" P2P module.
// 1. It creates and configures a "real" P2P module where all the other components of the node are mocked.
// 2. It then triggers a single message and waits for all of the expected messages transmission to complete before announcing failure.
func testRainTreeCalls(t *testing.T, origNode string, networkSimulationConfig TestNetworkSimulationConfig) {
	// Configure & prepare test module
	numValidators := len(networkSimulationConfig)
	runtimeConfigs := createMockRuntimeMgrs(t, numValidators)
	busMocks := createMockBuses(t, runtimeConfigs)

	valIds := make([]string, 0, numValidators)
	for key := range networkSimulationConfig {
		valIds = append(valIds, key)
	}

	sort.Slice(valIds, func(i, j int) bool {
		iId := extractNumericId(valIds[i])
		jId := extractNumericId(valIds[j])
		return iId < jId
	})

	// Create connection and bus mocks along with a shared WaitGroup to track the number of expected
	// reads and writes throughout the mocked local network
	var wg sync.WaitGroup
	connMocks := make(map[string]typesP2P.Transport)
	count := 0
	for _, valId := range valIds {
		expectedCall := networkSimulationConfig[valId]
		expectedReads := expectedCall.numNetworkReads + 1
		expectedWrites := expectedCall.numNetworkWrites
		log.Printf("[valId: %s] expected reads: %d\n", valId, expectedReads)
		log.Printf("[valId: %s] expected writes: %d\n", valId, expectedWrites)

		wg.Add(expectedReads)
		connMocks[valId] = prepareConnMock(t, valId, &wg, expectedCall.numNetworkReads)

		wg.Add(expectedWrites)

		persistenceMock := preparePersistenceMock(t, busMocks[count], runtimeConfigs[0].GetGenesis())
		consensusMock := prepareConsensusMock(t, busMocks[count], runtimeConfigs[0].GetGenesis())
		telemetryMock := prepareTelemetryMock(t, busMocks[count], valId, &wg, expectedWrites)

		prepareBusMock(busMocks[count].(*mockModules.MockBus), persistenceMock, consensusMock, telemetryMock)

		count++
	}

	// Inject the connection and bus mocks into the P2P modules
	p2pModules := createP2PModules(t, busMocks)
	for validatorId, p2pMod := range p2pModules {
		p2pMod.listener = connMocks[validatorId]
		p2pMod.Start()
		for _, peer := range p2pMod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		defer p2pMod.Stop()
	}

	// Wait for completion
	defer waitForNetworkSimulationCompletion(t, p2pModules, &wg)

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
