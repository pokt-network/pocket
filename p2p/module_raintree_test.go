package p2p

import (
	"github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
)

// IMPROVE(team): Looking into adding more tests and accounting for more edge cases.

// ~~~~~~ RainTree Unit Tests ~~~~~~

func TestRainTreeNetworkCompleteOneNodes(t *testing.T) {
	// val_1
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeConfig{
		// TODO (Olshansk) explain how these expected values are derived
		// TODO (Olshansk) explain what is happening here in this first test
		originatorNode: {0, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteTwoNodes(t *testing.T) {
	// val_1
	//   └───────┐
	// 	       val_2
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeConfig{
		// TODO (Olshansk) explain what is happening here in this second test
		// Attempt: I think Validator 1 is sending a message in a 2 (including self) node network
		originatorNode: {0, 1}, // TODO (Olshansk) why did these values change in between commits?
		validatorId(2): {1, 0},
	}
	testRainTreeCalls(t, originatorNode, expectedCalls)
}

func TestRainTreeNetworkCompleteThreeNodes(t *testing.T) {
	// 	          val_1
	// 	   ┌───────┴────┬─────────┐
	//   val_2        val_1     val_3
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeConfig{
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
	var expectedCalls = TestRainTreeConfig{
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
	var expectedCalls = TestRainTreeConfig{
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
	var expectedCalls = TestRainTreeConfig{
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
	var expectedCalls = TestRainTreeConfig{
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

// TODO (Olshansk) explain what this function does
// Attempt: This function tests the rain tree implementation against the theoretical truth which is documented [HERE}(?)
func testRainTreeCalls(t *testing.T, origNode string, rainTreeConfig TestRainTreeConfig) {
	// Network configurations
	messageHandledWaitGroup, p2pModules := prepareP2PModulesWithWaitGroup(t, rainTreeConfig)
	defer cleanupP2PModulesAndWaitGroup(t, p2pModules, messageHandledWaitGroup)
	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[origNode]
	require.NoError(t, p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC))
}
