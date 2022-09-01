package p2p

import (
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
)

// IMPROVE(team): Looking into adding more tests and accounting for more edge cases.

// ### RainTree Unit Tests ###

func TestRainTreeOneNodes(t *testing.T) {
	// val_1
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		// TODO (Olshansk) explain how these expected values are derived
		// TODO (Olshansk) explain what is happening here in this first test
		validatorId(t, 1): {0, 0}, // {numReads, numWrites}
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

func TestRainTreeTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		// TODO (Olshansk) explain what is happening here in this second test
		// Attempt: I think Validator 1 is sending a message in a 2 (including self) node network
		validatorId(t, 1): {0, 0}, // Originator
		// TODO (Olshansk) explain why the recipient of a message would also have a network 'write'
		validatorId(t, 2): {1, 1},
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

func TestRainTreeThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
		validatorId(t, 3): {1, 1},
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

func TestRainTreeFourNodes(t *testing.T) {
	// Test configurations (visualization retrieved from simulator)
	// 	                val_1
	// 	  ┌───────────────┴────┬─────────────────┐
	//  val_2                val_1             val_3
	//    └───────┐            └───────┐         └───────┐
	// 		    val_3                val_2             val_4
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {2, 2},
		validatorId(t, 3): {2, 2},
		validatorId(t, 4): {1, 1},
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

func TestRainTreeNineNodes(t *testing.T) {
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		validatorId(t, 1): {0, 0}, // Originator
		validatorId(t, 2): {1, 1},
		validatorId(t, 3): {1, 1},
		validatorId(t, 4): {1, 1},
		validatorId(t, 5): {1, 1},
		validatorId(t, 6): {1, 1},
		validatorId(t, 7): {1, 1},
		validatorId(t, 8): {1, 1},
		validatorId(t, 9): {1, 1},
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

func TestRainTreeEighteenNodes(t *testing.T) {
	// 	                                                                                                              val_1
	// 	                                      ┌──────────────────────────────────────────────────────────────────────────┴─────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                      val_7                                                                                                            val_1                                                                                                     val_13
	//             ┌──────────────────────────┴────────────┬────────────────────────────────────┐                                     ┌────────────────────────┴────────────┬──────────────────────────────────┐                                ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//           val_11                                   val_7                               val_15                                 val_5                                 val_1                              val_9                           val_17                                  val_13                                val_3
	//    ┌────────┴─────┬───────────┐             ┌───────┴────┬──────────┐           ┌────────┴─────┬──────────┐            ┌───────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴────┬─────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_13         val_11      val_16        val_9        val_7      val_12      val_17         val_15     val_8        val_7        val_5      val_10        val_3        val_1     val_6      val_11        val_9     val_2     val_1         val_17     val_4         val_15         val_13      val_18     val_5        val_3      val_14
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		validatorId(t, 1):  {1, 1}, // Originator
		validatorId(t, 2):  {1, 1},
		validatorId(t, 3):  {2, 2},
		validatorId(t, 4):  {1, 1},
		validatorId(t, 5):  {2, 2},
		validatorId(t, 6):  {1, 1},
		validatorId(t, 7):  {2, 2},
		validatorId(t, 8):  {1, 1},
		validatorId(t, 9):  {2, 2},
		validatorId(t, 10): {1, 1},
		validatorId(t, 11): {2, 2},
		validatorId(t, 12): {1, 1},
		validatorId(t, 13): {2, 2},
		validatorId(t, 14): {1, 1},
		validatorId(t, 15): {2, 2},
		validatorId(t, 16): {1, 1},
		validatorId(t, 17): {2, 2},
		validatorId(t, 18): {1, 1},
	}
	// Note that the originator, `val_1` is also messaged by `val_17` outside of continuously
	// demoting itself.
	// TODO (Olshansk) Explain why it's necessary to pass if the originator is pinged or not to the simulation function
	testExpectedRainTree(t, originatorNode, expectedCalls, true)
}

func TestRainTreeTwentySevenNodes(t *testing.T) {
	// 	                                                                                                                    val_1
	// 	                                     ┌────────────────────────────────────────────────────────────────────────────────┴───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────┐
	//                                    val_10                                                                                                                   val_1                                                                                                       val_19
	//            ┌──────────────────────────┴──────────────┬──────────────────────────────────────┐                                         ┌────────────────────────┴────────────┬──────────────────────────────────┐                                  ┌────────────────────────┴──────────────┬────────────────────────────────────┐
	//          val_16                                    val_10                                 val_22                                     val_7                                 val_1                             val_13                             val_25                                  val_19                                val_4
	//   ┌────────┴─────┬───────────┐              ┌────────┴─────┬───────────┐           ┌────────┴─────┬───────────┐              ┌────────┴────┬──────────┐             ┌───────┴────┬─────────┐          ┌────────┴─────┬──────────┐         ┌───────┴─────┬──────────┐             ┌────────┴─────┬───────────┐          ┌───────┴────┬──────────┐
	// val_20         val_16      val_24         val_14         val_10      val_18      val_26         val_22      val_12         val_11        val_7      val_15        val_5        val_1     val_9      val_17         val_13     val_3     val_2         val_25     val_6         val_23         val_19      val_27     val_8        val_4      val_21
	originatorNode := validatorId(t, 1)
	var expectedCalls = TestRainTreeConfig{
		validatorId(t, 1):  {0, 0}, // Originator
		validatorId(t, 2):  {1, 1},
		validatorId(t, 3):  {1, 1},
		validatorId(t, 4):  {1, 1},
		validatorId(t, 5):  {1, 1},
		validatorId(t, 6):  {1, 1},
		validatorId(t, 7):  {1, 1},
		validatorId(t, 8):  {1, 1},
		validatorId(t, 9):  {1, 1},
		validatorId(t, 10): {1, 1},
		validatorId(t, 11): {1, 1},
		validatorId(t, 12): {1, 1},
		validatorId(t, 13): {1, 1},
		validatorId(t, 14): {1, 1},
		validatorId(t, 15): {1, 1},
		validatorId(t, 16): {1, 1},
		validatorId(t, 17): {1, 1},
		validatorId(t, 18): {1, 1},
		validatorId(t, 19): {1, 1},
		validatorId(t, 20): {1, 1},
		validatorId(t, 21): {1, 1},
		validatorId(t, 22): {1, 1},
		validatorId(t, 23): {1, 1},
		validatorId(t, 24): {1, 1},
		validatorId(t, 25): {1, 1},
		validatorId(t, 26): {1, 1},
		validatorId(t, 27): {1, 1},
	}
	testExpectedRainTree(t, originatorNode, expectedCalls, false)
}

// TODO (Olshansk) explain what this function does
// Attempt: This function tests the rain tree implementation against the theoretical truth which is documented [HERE}(?)
func testExpectedRainTree(t *testing.T, origNode string, rainTreeConfig TestRainTreeConfig, isOriginatorPinged bool) {
	// Network configurations
	messageHandledWaitGroup, p2pModules := prepareP2PModulesWithWaitGroup(t, rainTreeConfig, isOriginatorPinged)
	defer cleanupP2PModulesAndWaitGroup(t, p2pModules, messageHandledWaitGroup)
	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[origNode]
	require.NoError(t, p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC))
}
