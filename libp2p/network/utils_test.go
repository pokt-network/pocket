package network

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/p2p/providers/addrbook_provider"
	"github.com/pokt-network/pocket/p2p/providers/current_height_provider"
	typesP2P "github.com/pokt-network/pocket/p2p/types"
	mock_typesP2P "github.com/pokt-network/pocket/p2p/types/mocks"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	configTypes "github.com/pokt-network/pocket/runtime/configs/types"
	"github.com/pokt-network/pocket/runtime/genesis"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/modules/mocks"
)

const (
	maxNumKeys             = 42
	genesisConfigSeedStart = 42
)

var (
	keys                 []cryptoPocket.PrivateKey
	serviceUrlFormat     = "node%d.consensus:8080"
	testServiceUrlFormat = "10.0.0.%d:8080"
)

func init() {
	keys = generateKeys(nil, maxNumKeys)
}

func generateKeys(_ *testing.T, numValidators int) []cryptoPocket.PrivateKey {
	keys := make([]cryptoPocket.PrivateKey, numValidators)

	for i := range keys {
		seedInt := genesisConfigSeedStart + i
		keys[i] = cryptoPocket.GetPrivKeySeed(seedInt)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Address().String() < keys[j].Address().String()
	})
	return keys
}

// CLEANUP: This could (should?) be a codebase-wide shared test helper
func validatorId(i int) string {
	return fmt.Sprintf(serviceUrlFormat, i)
}

// createMockRuntimeMgrs creates `numValidators` instances of mocked `RuntimeMgr` that are essentially
// representing the runtime environments of the validators that we will use in our tests
func createMockRuntimeMgrs(t *testing.T, numValidators int) []modules.RuntimeMgr {
	ctrl := gomock.NewController(t)
	mockRuntimeMgrs := make([]modules.RuntimeMgr, numValidators)
	valKeys := make([]cryptoPocket.PrivateKey, numValidators)
	copy(valKeys, keys[:numValidators])
	mockGenesisState := createMockGenesisState(valKeys)
	for i := range mockRuntimeMgrs {
		cfg := &configs.Config{
			RootDirectory: "",
			PrivateKey:    valKeys[i].String(),
			P2P: &configs.P2PConfig{
				PrivateKey:     valKeys[i].String(),
				Port:           8080,
				UseRainTree:    true,
				ConnectionType: configTypes.ConnectionType_EmptyConnection,
			},
		}

		mockRuntimeMgr := mock_modules.NewMockRuntimeMgr(ctrl)
		mockRuntimeMgr.EXPECT().GetConfig().Return(cfg).AnyTimes()
		mockRuntimeMgr.EXPECT().GetGenesis().Return(mockGenesisState).AnyTimes()
		mockRuntimeMgrs[i] = mockRuntimeMgr
	}
	return mockRuntimeMgrs
}

func newTestAddrBookProvider(t *testing.T, ctrl *gomock.Controller, numPeers int) addrbook_provider.AddrBookProvider {
	addrBook := make(typesP2P.AddrBook, numPeers)
	// No expectations, transport is not used in current network test.
	transport := mock_typesP2P.NewMockTransport(ctrl)
	publicKey, err := cryptoPocket.GeneratePublicKey()
	require.NoError(t, err)

	for i := range addrBook {
		addrBook[i] = &typesP2P.NetworkPeer{
			Dialer:     transport,
			PublicKey:  publicKey,
			Address:    publicKey.Address(),
			ServiceUrl: fmt.Sprintf(testServiceUrlFormat, i),
		}
	}

	mockAddrBookProvider := mock_typesP2P.NewMockAddrBookProvider(ctrl)
	mockAddrBookProvider.EXPECT().GetStakedAddrBookAtHeight(gomock.Any()).Return(addrBook, nil)
	return mockAddrBookProvider
}

func createMockBus(t *testing.T, runtimeMgr modules.RuntimeMgr, numPeers int) *mock_modules.MockBus {
	ctrl := gomock.NewController(t)
	mockBus := mock_modules.NewMockBus(ctrl)
	mockBus.EXPECT().GetRuntimeMgr().Return(runtimeMgr).AnyTimes()
	mockBus.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	mockBus.EXPECT().RegisterModule(gomock.Any()).DoAndReturn(func(m modules.Module) {
		m.SetBus(mockBus)
	}).AnyTimes()

	mockAddrBookProvider := newTestAddrBookProvider(t, ctrl, numPeers)

	mockModulesRegistry := mock_modules.NewMockModulesRegistry(ctrl)
	mockModulesRegistry.EXPECT().GetModule(addrbook_provider.ModuleName).Return(mockAddrBookProvider.(modules.Module), nil).AnyTimes()
	mockModulesRegistry.EXPECT().GetModule(current_height_provider.ModuleName).Return(nil, runtime.ErrModuleNotRegistered(current_height_provider.ModuleName)).AnyTimes()
	mockBus.EXPECT().GetModulesRegistry().Return(mockModulesRegistry).AnyTimes()
	mockBus.EXPECT().PublishEventToBus(gomock.Any()).AnyTimes()
	return mockBus
}

// createMockGenesisState configures and returns a mocked GenesisState
func createMockGenesisState(valKeys []cryptoPocket.PrivateKey) *genesis.GenesisState {
	genesisState := new(genesis.GenesisState)
	validators := make([]*coreTypes.Actor, len(valKeys))
	for i, valKey := range valKeys {
		addr := valKey.Address().String()
		mockActor := &coreTypes.Actor{
			ActorType:       coreTypes.ActorType_ACTOR_TYPE_VAL,
			Address:         addr,
			PublicKey:       valKey.PublicKey().String(),
			GenericParam:    validatorId(i + 1),
			StakedAmount:    "1000000000000000",
			PausedHeight:    int64(0),
			UnstakingHeight: int64(0),
			Output:          addr,
		}
		validators[i] = mockActor
	}
	genesisState.Validators = validators

	return genesisState
}

// Bus Mock - needed to return the appropriate modules when accessed
func prepareBusMock(busMock *mock_modules.MockBus,
	consensusMock *mock_modules.MockConsensusModule,
) {
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
}

// Consensus mock - only needed for validatorMap access
func prepareConsensusMock(t *testing.T, busMock *mock_modules.MockBus) *mock_modules.MockConsensusModule {
	ctrl := gomock.NewController(t)
	consensusMock := mock_modules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	consensusMock.EXPECT().GetBus().Return(busMock).AnyTimes()
	consensusMock.EXPECT().SetBus(busMock).AnyTimes()
	consensusMock.EXPECT().GetModuleName().Return(modules.ConsensusModuleName).AnyTimes()
	busMock.RegisterModule(consensusMock)

	return consensusMock
}

func MockBus(ctrl *gomock.Controller) *mock_modules.MockBus {
	consensusMock := mock_modules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().CurrentHeight().Return(uint64(0)).AnyTimes()

	runtimeMgrMock := mock_modules.NewMockRuntimeMgr(ctrl)
	runtimeMgrMock.EXPECT().GetConfig().Return(configs.NewDefaultConfig()).AnyTimes()

	busMock := mock_modules.NewMockBus(ctrl)
	busMock.EXPECT().GetPersistenceModule().Return(nil).AnyTimes()
	busMock.EXPECT().GetConsensusModule().Return(consensusMock).AnyTimes()
	busMock.EXPECT().GetRuntimeMgr().Return(runtimeMgrMock).AnyTimes()

	return busMock
}
