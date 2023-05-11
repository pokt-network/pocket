//go:build test

package p2p_test

import (
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/internal/testutil/constructors"
	"github.com/pokt-network/pocket/internal/testutil/p2p"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/regen-network/gocuke"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	consensus_testutil "github.com/pokt-network/pocket/internal/testutil/consensus"
	persistence_testutil "github.com/pokt-network/pocket/internal/testutil/persistence"
	"github.com/pokt-network/pocket/p2p/protocol"
	"github.com/pokt-network/pocket/p2p/raintree"
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
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode: {0, 0}, // val_1, the originator, does 0 network reads or writes
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := p2p_testutil.NewServiceURL(1)
	// Per the diagram above, in the case of a 2 node network, the originator node (val_1) does a
	// single write to another node (val_2),  also the
	// originator node and never performs any reads or writes during a RainTree broadcast.
	expectedCalls := TestNetworkSimulationConfig{
		// Attempt: I think Validator 1 is sending a message in a 2 (including self) node network originatorNode:                {0, 1}, // val_1 does a single network write (to val_2)
		p2p_testutil.NewServiceURL(2): {1, 0}, // val_2 does a single network read (from val_1)
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                {0, 2}, // val_1 does two network writes (to val_2 and val_3)
		p2p_testutil.NewServiceURL(2): {1, 0}, // val_2 does a single network read (from val_1)
		p2p_testutil.NewServiceURL(3): {1, 0}, // val_2 does a single network read (from val_3)
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
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                {0, 3}, // val_1 does 3 network writes (two to val_2 and 1 to val_3)
		p2p_testutil.NewServiceURL(2): {2, 1}, // val_2 does 2 network reads (both from val_1) and 1 network write (to val_3)
		p2p_testutil.NewServiceURL(3): {2, 1}, // val_2 does 2 network reads (from val_1 and val_2) and 1 network write (to val_4)
		p2p_testutil.NewServiceURL(4): {1, 0}, // val_2 does 1 network read (from val_3)
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteNineNodes(t *testing.T) {
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                {0, 4},
		p2p_testutil.NewServiceURL(2): {1, 0},
		p2p_testutil.NewServiceURL(3): {1, 0},
		p2p_testutil.NewServiceURL(4): {1, 2},
		p2p_testutil.NewServiceURL(5): {1, 0},
		p2p_testutil.NewServiceURL(6): {1, 0},
		p2p_testutil.NewServiceURL(7): {1, 2},
		p2p_testutil.NewServiceURL(8): {1, 0},
		p2p_testutil.NewServiceURL(9): {1, 0},
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
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                 {1, 6},
		p2p_testutil.NewServiceURL(2):  {3, 2},
		p2p_testutil.NewServiceURL(3):  {2, 2},
		p2p_testutil.NewServiceURL(4):  {2, 0},
		p2p_testutil.NewServiceURL(5):  {2, 4},
		p2p_testutil.NewServiceURL(6):  {3, 2},
		p2p_testutil.NewServiceURL(7):  {2, 2},
		p2p_testutil.NewServiceURL(8):  {2, 0},
		p2p_testutil.NewServiceURL(9):  {2, 4},
		p2p_testutil.NewServiceURL(10): {3, 2},
		p2p_testutil.NewServiceURL(11): {2, 2},
		p2p_testutil.NewServiceURL(12): {2, 0},
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
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                 {1, 6},
		p2p_testutil.NewServiceURL(2):  {1, 0},
		p2p_testutil.NewServiceURL(3):  {2, 2},
		p2p_testutil.NewServiceURL(4):  {1, 0},
		p2p_testutil.NewServiceURL(5):  {2, 2},
		p2p_testutil.NewServiceURL(6):  {1, 0},
		p2p_testutil.NewServiceURL(7):  {2, 4},
		p2p_testutil.NewServiceURL(8):  {1, 0},
		p2p_testutil.NewServiceURL(9):  {2, 2},
		p2p_testutil.NewServiceURL(10): {1, 0},
		p2p_testutil.NewServiceURL(11): {2, 2},
		p2p_testutil.NewServiceURL(12): {1, 0},
		p2p_testutil.NewServiceURL(13): {2, 4},
		p2p_testutil.NewServiceURL(14): {1, 0},
		p2p_testutil.NewServiceURL(15): {2, 2},
		p2p_testutil.NewServiceURL(16): {1, 0},
		p2p_testutil.NewServiceURL(17): {2, 2},
		p2p_testutil.NewServiceURL(18): {1, 0},
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
	originatorNode := p2p_testutil.NewServiceURL(1)
	expectedCalls := TestNetworkSimulationConfig{
		originatorNode:                 {0, 6},
		p2p_testutil.NewServiceURL(2):  {1, 0},
		p2p_testutil.NewServiceURL(3):  {1, 0},
		p2p_testutil.NewServiceURL(4):  {1, 2},
		p2p_testutil.NewServiceURL(5):  {1, 0},
		p2p_testutil.NewServiceURL(6):  {1, 0},
		p2p_testutil.NewServiceURL(7):  {1, 2},
		p2p_testutil.NewServiceURL(8):  {1, 0},
		p2p_testutil.NewServiceURL(9):  {1, 0},
		p2p_testutil.NewServiceURL(10): {1, 4},
		p2p_testutil.NewServiceURL(11): {1, 0},
		p2p_testutil.NewServiceURL(12): {1, 0},
		p2p_testutil.NewServiceURL(13): {1, 2},
		p2p_testutil.NewServiceURL(14): {1, 0},
		p2p_testutil.NewServiceURL(15): {1, 0},
		p2p_testutil.NewServiceURL(16): {1, 2},
		p2p_testutil.NewServiceURL(17): {1, 0},
		p2p_testutil.NewServiceURL(18): {1, 0},
		p2p_testutil.NewServiceURL(19): {1, 4},
		p2p_testutil.NewServiceURL(20): {1, 0},
		p2p_testutil.NewServiceURL(21): {1, 0},
		p2p_testutil.NewServiceURL(22): {1, 2},
		p2p_testutil.NewServiceURL(23): {1, 0},
		p2p_testutil.NewServiceURL(24): {1, 0},
		p2p_testutil.NewServiceURL(25): {1, 2},
		p2p_testutil.NewServiceURL(26): {1, 0},
		p2p_testutil.NewServiceURL(27): {1, 0},
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
	genesisMock := runtimeConfigs[0].GetGenesis()

	var wg sync.WaitGroup
	//busMocks := createMockBuses(t, runtimeConfigs, &wg)
	busEventHandlerFactory := func(t gocuke.TestingT, bus modules.Bus) testutil.BusEventHandler {
		return func(data *messaging.PocketEnvelope) {
			// `p2pModule#handleNetworkData()` calls `modules.Bus#PublishEventToBus()`
			// assumes that P2P module is the only event producer running during the test
			wg.Done()
		}
	}

	busMocks, _, p2pModules := constructors.NewBusesMocknetAndP2PModules(
		t, numValidators,
		genesisMock,
		busEventHandlerFactory,
	)

	valIds := make([]string, 0, numValidators)
	for valId := range networkSimulationConfig {
		valIds = append(valIds, valId)
	}

	// TODO_THIS_COMMIT: need this?
	// sort `valIds` in ascending order
	sort.Slice(valIds, func(i, j int) bool {
		iId := extractNumericId(valIds[i])
		jId := extractNumericId(valIds[j])
		return iId < jId
	})

	// Create connection and bus mocks along with a shared WaitGroup to track the number of expected
	// reads and writes throughout the mocked local network
	for _, valId := range valIds {
		expectedCall := networkSimulationConfig[valId]
		expectedReads := expectedCall.numNetworkReads
		expectedWrites := expectedCall.numNetworkWrites

		log.Printf("[valId: %s] expected reads: %d\n", valId, expectedReads)
		log.Printf("[valId: %s] expected writes: %d\n", valId, expectedWrites)
		wg.Add(expectedReads)
		wg.Add(expectedWrites)

		persistenceMock := persistence_testutil.BasePersistenceMock(t, busMocks[valId], genesisMock)
		consensusMock := consensus_testutil.PrepareConsensusMock(t, busMocks[valId])
		telemetryMock := prepareTelemetryMock(t, busMocks[valId], valId, &wg, expectedWrites)

		busMocks[valId].EXPECT().GetPersistenceModule().Return(persistenceMock).AnyTimes()
		busMocks[valId].EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
		busMocks[valId].EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()
	}

	// Inject the connection and bus mocks into the P2P modules
	//p2pModules := createP2PModules(t, busMocks, libp2pNetworkMock, valIds)

	for _, p2pMod := range p2pModules {
		err := p2pMod.Start()
		require.NoError(t, err)

		sURL := strings.Clone(serviceURL)
		mod := *p2pMod
		p2pMod.host.SetStreamHandler(protocol.PoktProtocolID, func(stream libp2pNetwork.Stream) {
			log.Printf("[valID: %s] Read\n", sURL)
			(&mod).router.(*raintree.RainTreeRouter).HandleStream(stream)
			wg.Done()
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
