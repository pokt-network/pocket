package p2p

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"github.com/golang/mock/gomock"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mocksP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	modulesMock "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/pocket/shared/types"
	"github.com/pokt-network/pocket/shared/types/genesis"
	"github.com/stretchr/testify/require"
	"sort"
	"sync"
	"testing"
	"time"
)

// ### RainTree Unit Utils - Configurations & constants and such ###

const (
	// TODO (olshansk) explain these values
	// Attempt: Since we simulate up to a 27 node network, we will pre-generate a n >= 27 number of keys to avoid generation everytime
	// The genesis config seed start must begin at the max_keys value because...? and 42 is chosen because...?
	genesisConfigSeedStart = 42
	maxNumKeys             = 42 // The number of keys generated for all the unit tests. Optimization to avoid regenerating every time.
	serviceUrlFormat       = "val_%d"
	testChannelSize        = 10000
)

// TODO(olshansky): Add configurations tests for dead and partially visible nodes
type TestRainTreeConfig map[string]struct {
	numNetworkReads  uint16
	numNetworkWrites uint16
}

var keys []cryptoPocket.PrivateKey

func init() {
	keys = generateKeys(nil, maxNumKeys)
}

func prepareP2PModulesWithWaitGroup(t *testing.T, rainTreeConfig TestRainTreeConfig, isOriginatorPinged bool) (*sync.WaitGroup, map[string]*p2pModule) {
	numValidators := len(rainTreeConfig)
	configs, genesisState := createConfigs(t, numValidators)

	// Test configurations
	wg := new(sync.WaitGroup)
	if isOriginatorPinged {
		// TODO (Olshansk) explain better here
		// Attempt: If the originator is 'pinged' during the process - the entire set is covered - else it's implicit and
		// not represented by the wait group... (Why is it designed this way)?
		wg.Add(numValidators)
	} else {
		wg.Add(numValidators - 1) // -1 because the originator node implicitly handles the message
	}
	// Network initialization
	consensusMock := prepareTestConsensusMock(t, genesisState)
	telemetryMock := prepareTestTelemetryMock(t)
	connMocks := make(map[string]typesP2P.Transport)
	busMocks := make(map[string]modules.Bus)
	for valId, expectedCall := range rainTreeConfig {
		connMocks[valId] = prepareTestConnMock(t, expectedCall.numNetworkReads, expectedCall.numNetworkWrites)
		busMocks[valId] = prepareTestBusMock(t, wg, consensusMock, telemetryMock)
	}

	// TODO (Olshansk) explain why we need multiple p2p modules - but only one respective consensus and telemetry module
	// Attempt: Because we are simulating a combination of nodes inside of rain tree, we create a P2P instance for each node
	// consensus, telemetry, and bus for these tests nodes are identical, while the p2p instance requires unique listener configuration
	return wg, prepareTestP2PModules(t, configs, connMocks, busMocks)
}

func cleanupP2PModulesAndWaitGroup(t *testing.T, p2pModules map[string]*p2pModule, messageHandledWaitGroup *sync.WaitGroup) {
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
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message to be handled")
	}
	for _, p2pMod := range p2pModules {
		p2pMod.Stop()
	}
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

func createConfigs(t *testing.T, numValidators int) (configs []*genesis.Config, genesisState *genesis.GenesisState) {
	configs = make([]*genesis.Config, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys[:], keys[:numValidators])
	genesisState = createGenesisState(t, valKeys)

	for i := range configs {
		configs[i] = &genesis.Config{
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

func validatorId(_ *testing.T, i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
}

func createGenesisState(t *testing.T, valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	validators := make([]*genesis.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		val := &genesis.Actor{
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(t, i+1),
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

// ### Testing Mocks Below ###

// A mock of the application specific to know if a message was sent to be handled by the application
// INVESTIGATE(olshansky): Double check that how the expected calls are counted is accurate per the
//                         expectation with RainTree by comparing with Telemetry after updating specs.
func prepareTestBusMock(t *testing.T, wg *sync.WaitGroup, consensusMock *modulesMock.MockConsensusModule,
	telemetryMock *modulesMock.MockTelemetryModule) *modulesMock.MockBus {
	ctrl := gomock.NewController(t)
	busMock := modulesMock.NewMockBus(ctrl)

	busMock.EXPECT().PublishEventToBus(gomock.Any()).Do(func(e *types.PocketEvent) {
		wg.Done()
		fmt.Println("App specific bus mock publishing event to bus")
	}).MaxTimes(1) // Using `MaxTimes` rather than `Times` because originator node implicitly handles the message

	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetTelemetryModule().Return(telemetryMock).AnyTimes()

	return busMock
}

// TODO (Olshansk) explain what the consensus mock does
// Attempt: the consensus mock returns the genesis validator map anytime the .ValidatorMap() function is called
// the consensus mock also returns '1' when the current height is called
func prepareTestConsensusMock(t *testing.T, genesisState *genesis.GenesisState) *modulesMock.MockConsensusModule {
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

// TODO(team): make the test more rigorous but adding MaxTimes `EmitEvent` expectations. Since we are talking about more than one node
// I have decided to do with `AnyTimes` for the moment.
// TODO (Olshansk) explain why it's necessary to mock telemetry here and document the sub telemetry package
func prepareTestTelemetryMock(t *testing.T) *modulesMock.MockTelemetryModule {
	ctrl := gomock.NewController(t)
	telemetryMock := modulesMock.NewMockTelemetryModule(ctrl)

	timeSeriesAgentMock := prepareTestTimeSeriesAgentMock(t)
	eventMetricsAgentMock := prepareTestEventMetricsAgentMock(t)

	telemetryMock.EXPECT().GetTimeSeriesAgent().Return(timeSeriesAgentMock).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterRegister(gomock.Any(), gomock.Any()).AnyTimes()
	timeSeriesAgentMock.EXPECT().CounterIncrement(gomock.Any()).AnyTimes()

	telemetryMock.EXPECT().GetEventMetricsAgent().Return(eventMetricsAgentMock).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	eventMetricsAgentMock.EXPECT().EmitEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return telemetryMock
}

func prepareTestTimeSeriesAgentMock(t *testing.T) *modulesMock.MockTimeSeriesAgent {
	ctrl := gomock.NewController(t)
	timeseriesAgentMock := modulesMock.NewMockTimeSeriesAgent(ctrl)
	return timeseriesAgentMock
}

func prepareTestEventMetricsAgentMock(t *testing.T) *modulesMock.MockEventMetricsAgent {
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
func prepareTestConnMock(t *testing.T, expectedNumNetworkReads, expectedNumNetworkWrites uint16) typesP2P.Transport {
	testChannel := make(chan []byte, testChannelSize)
	ctrl := gomock.NewController(t)
	connMock := mocksP2P.NewMockTransport(ctrl)

	connMock.EXPECT().Read().DoAndReturn(func() ([]byte, error) {
		data := <-testChannel
		return data, nil
	}).MaxTimes(int(expectedNumNetworkReads + 1)) // INVESTIGATE(olshansky): The +1 is necessary because there is one extra read of empty data by every channel...

	connMock.EXPECT().Write(gomock.Any()).DoAndReturn(func(data []byte) error {
		testChannel <- data
		return nil
	}).MaxTimes(int(expectedNumNetworkWrites))

	connMock.EXPECT().Close().Return(nil).Times(1)

	return connMock
}

func prepareTestP2PModules(t *testing.T, configs []*genesis.Config, connMocks map[string]typesP2P.Transport,
	busMocks map[string]modules.Bus) (p2pModules map[string]*p2pModule) {
	p2pModules = make(map[string]*p2pModule, len(configs))
	for i, config := range configs {
		p2pMod, err := Create(config, nil)
		require.NoError(t, err)
		p2p := p2pMod.(*p2pModule)
		vID := validatorId(t, i+1)
		p2p.listener = connMocks[vID]
		p2p.SetBus(busMocks[vID])
		p2pModules[vID] = p2p
		p2pMod.Start()
		for _, peer := range p2p.network.GetAddrBook() {
			peer.Dialer = connMocks[peer.ServiceUrl]
		}
	}
	return
}
