package pre2p

import (
	"sync"
	"testing"

	typesPre2P "github.com/pokt-network/pocket/p2p/pre2p/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/types"
	"google.golang.org/protobuf/types/known/anypb"
)

// TODO(team): add more tests with different configurations:
// - varying number of levels (i.e. 1, 2, 3, 4, 5)
// - partial/incomplete network views
// - dead nodes; both ephemeral and constant
// - redundancy related tests

func TestRainTreeSmall(t *testing.T) {
	// Network configurations
	numValidators := 4
	configs := createConfigs(t, numValidators)

	// Test configurations (visualization retrieved from simulator)
	// 	                 val_1
	// 	   ┌───────────────┴────┬─────────────────┐
	//   val_2                val_1             val_3
	//     └───────┐            └───────┐         └───────┐
	// 		     val_3                val_2             val_4
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		// valId: {numReads, numWrites}
		validatorId(1): {0, 0},
		validatorId(2): {2, 2},
		validatorId(3): {2, 2},
		validatorId(4): {1, 1},
	}
	var messageHandeledWaitGroup sync.WaitGroup
	messageHandeledWaitGroup.Add(numValidators - 1) // -1 because the originator node implicitly handles the message

	// Network initialization
	connMocks := make(map[string]typesPre2P.TransportLayerConn)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range expectedCalls {
		connMocks[valId] = prepareConnMock(t, valId, expectedCall.numReads, expectedCall.numWrites)
		busMocks[valId] = prepareBusMock(t, &messageHandeledWaitGroup)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, configs)
	for validatorId, mod := range p2pModules {
		mod.listener = connMocks[validatorId]
		mod.SetBus(busMocks[validatorId])
		for _, peer := range mod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		mod.Start()
		defer mod.Stop()
	}

	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[originatorNode]
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	messageHandeledWaitGroup.Wait()
}

func TestRainTreeLarge(t *testing.T) {
	// Network configurations
	numValidators := 9
	configs := createConfigs(t, numValidators)

	// Test configurations
	// 	                              val_1
	// 	         ┌──────────────────────┴────────────┬────────────────────────────────┐
	//         val_4                               val_1                            val_7
	//   ┌───────┴────┬─────────┐            ┌───────┴────┬─────────┐         ┌───────┴────┬─────────┐
	// val_6        val_4     val_8        val_3        val_1     val_5     val_9        val_7     val_2
	originatorNode := validatorId(1)
	var expectedCalls = TestRainTreeCommConfig{
		// valId: {numReads, numWrites}
		validatorId(1): {0, 0},
		validatorId(2): {1, 1},
		validatorId(3): {1, 1},
		validatorId(4): {1, 1},
		validatorId(5): {1, 1},
		validatorId(6): {1, 1},
		validatorId(7): {1, 1},
		validatorId(8): {1, 1},
		validatorId(9): {1, 1},
	}
	var messageHandeledWaitGroup sync.WaitGroup
	messageHandeledWaitGroup.Add(numValidators - 1) // -1 because the originator node implicitly handles the message

	// Network initialization
	connMocks := make(map[string]typesPre2P.TransportLayerConn)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range expectedCalls {
		connMocks[valId] = prepareConnMock(t, valId, expectedCall.numReads, expectedCall.numWrites)
		busMocks[valId] = prepareBusMock(t, &messageHandeledWaitGroup)
	}

	// Module injection
	p2pModules := prepareP2PModules(t, configs)
	for validatorId, mod := range p2pModules {
		mod.listener = connMocks[validatorId]
		mod.SetBus(busMocks[validatorId])
		for _, peer := range mod.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
		mod.Start()
		defer mod.Stop()
	}

	// Trigger originator message
	p := &anypb.Any{}
	p2pMod := p2pModules[originatorNode]
	p2pMod.Broadcast(p, types.PocketTopic_DEBUG_TOPIC)

	// Wait for completion
	messageHandeledWaitGroup.Wait()
}
