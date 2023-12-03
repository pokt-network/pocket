package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/ibc/store"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/codec"
	core_types "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	mock_modules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

func TestClientState_Set(t *testing.T) {
	// get provable store prefixed with clients/123
	provableStore := newTestProvableStore(t, "123")

	// create a client state
	clientState := &ClientState{
		Data:         []byte("data"),
		WasmChecksum: make([]byte, 32),
	}
	bz, err := codec.GetCodec().Marshal(clientState)
	require.NoError(t, err)

	// set the client state
	require.NoError(t, setClientState(provableStore, clientState))

	// check cache
	cache := kvstore.NewMemKVStore()

	// flush cache
	require.NoError(t, provableStore.FlushCache(cache))

	// get all from cache
	keys, vals, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Len(t, vals, 1)

	// check key and value set correctly
	require.Equal(t, []byte("clients/123/1/clients/123/clientState"), keys[0])
	require.Equal(t, vals[0], bz)
}

func TestConsensusState_Set(t *testing.T) {
	// get provable store prefixed with clients/123
	provableStore := newTestProvableStore(t, "123")

	// create a consensus state
	consensusState := &ConsensusState{
		Data:      []byte("data"),
		Timestamp: 1,
	}
	height := &Height{
		RevisionNumber: 1,
		RevisionHeight: 1,
	}
	bz, err := codec.GetCodec().Marshal(consensusState)
	require.NoError(t, err)

	// set the client state
	require.NoError(t, setConsensusState(provableStore, consensusState, height))

	// check cache
	cache := kvstore.NewMemKVStore()

	// flush cache
	require.NoError(t, provableStore.FlushCache(cache))

	// get all from cache
	keys, vals, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Len(t, vals, 1)

	// check key and value set correctly
	require.Equal(t, []byte("clients/123/1/clients/123/consensusStates/1-1"), keys[0])
	require.Equal(t, vals[0], bz)
}

func TestClientState_Get(t *testing.T) {
	clientStore := newTestProvableStore(t, "")

	testCases := []struct {
		name        string
		clientId    string
		data        []byte
		checksum    []byte
		expectedErr error
	}{
		{
			name:        "client state not found",
			clientId:    "124",
			data:        nil,
			checksum:    nil,
			expectedErr: core_types.ErrIBCKeyDoesNotExist("clients/124/clientState"),
		},
		{
			name:        "client state found",
			clientId:    "123",
			data:        []byte("data"),
			checksum:    make([]byte, 32),
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientState, err := GetClientState(clientStore, tc.clientId)
			require.ErrorIs(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				require.Equal(t, clientState.GetData(), tc.data)
				require.Equal(t, clientState.GetWasmChecksum(), tc.checksum)
			}
		})
	}
}

func TestConsensusState_Get(t *testing.T) {
	clientStore := newTestProvableStore(t, "123")

	testCases := []struct {
		name        string
		height      *Height
		data        []byte
		timestamp   uint64
		expectedErr error
	}{
		{
			name:        "consensus state not found - wrong height",
			height:      &Height{RevisionNumber: 1, RevisionHeight: 2},
			data:        nil,
			timestamp:   0,
			expectedErr: core_types.ErrIBCKeyDoesNotExist("clients/123/consensusStates/1-2"),
		},
		{
			name:        "consensus state found",
			height:      &Height{RevisionNumber: 1, RevisionHeight: 1},
			data:        []byte("data"),
			timestamp:   1,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			consensusState, err := GetConsensusState(clientStore, tc.height)
			require.ErrorIs(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				require.Equal(t, consensusState.GetData(), tc.data)
				require.Equal(t, consensusState.GetTimestamp(), tc.timestamp)
			}
		})
	}
}

func newTestProvableStore(t *testing.T, clientId string) modules.ProvableStore {
	t.Helper()

	tree, nodeStore, dbMap := setupDB(t)

	runtimeCfg := newTestRuntimeConfig(t)
	bus, err := runtime.CreateBus(runtimeCfg)
	require.NoError(t, err)

	persistenceMock := newPersistenceMock(t, bus, dbMap)
	bus.RegisterModule(persistenceMock)
	consensusMock := newConsensusMock(t, bus)
	bus.RegisterModule(consensusMock)
	treeStoreMock := newTreeStoreMock(t, bus, tree, nodeStore)
	bus.RegisterModule(treeStoreMock)
	p2pMock := newTestP2PModule(t, bus)
	bus.RegisterModule(p2pMock)
	utilityMock := newUtilityMock(t, bus)
	bus.RegisterModule(utilityMock)

	privKey := runtimeCfg.GetConfig().IBC.Host.PrivateKey

	t.Cleanup(func() {
		err := persistenceMock.Stop()
		require.NoError(t, err)
		err = consensusMock.Stop()
		require.NoError(t, err)
		err = p2pMock.Stop()
		require.NoError(t, err)
	})

	if clientId != "" {
		clientId = "/" + clientId
	}

	return store.NewProvableStore(bus, []byte("clients"+clientId), privKey)
}

func setupDB(t *testing.T) (*smt.SMT, kvstore.KVStore, map[string]string) {
	dbMap := make(map[string]string, 0)
	nodeStore := kvstore.NewMemKVStore()
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New())

	clientState := &ClientState{
		Data:         []byte("data"),
		WasmChecksum: make([]byte, 32),
	}
	cliBz, err := codec.GetCodec().Marshal(clientState)
	require.NoError(t, err)
	consensusState := &ConsensusState{
		Data:      []byte("data"),
		Timestamp: 1,
	}
	conBz, err := codec.GetCodec().Marshal(consensusState)
	require.NoError(t, err)

	keys := [][]byte{
		[]byte("clients/123/consensusStates/1-1"),
		[]byte("clients/123/clientState"),
	}
	values := [][]byte{
		conBz,
		cliBz,
	}

	for i, key := range keys {
		dbMap[hex.EncodeToString(key)] = hex.EncodeToString(values[i])
		err := tree.Update(key, values[i])
		require.NoError(t, err)
	}

	require.NoError(t, tree.Commit())

	t.Cleanup(func() {
		err := nodeStore.Stop()
		require.NoError(t, err)
	})

	return tree, nodeStore, dbMap
}

func newConsensusMock(t *testing.T, bus modules.Bus) *mock_modules.MockConsensusModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	consensusMock := mock_modules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().GetModuleName().Return(modules.ConsensusModuleName).AnyTimes()
	consensusMock.EXPECT().Start().Return(nil).AnyTimes()
	consensusMock.EXPECT().Stop().Return(nil).AnyTimes()
	consensusMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	consensusMock.EXPECT().GetBus().Return(bus).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	return consensusMock
}

func newUtilityMock(t *testing.T, bus modules.Bus) *mock_modules.MockUtilityModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	utilityMock := mock_modules.NewMockUtilityModule(ctrl)
	utilityMock.EXPECT().GetModuleName().Return(modules.UtilityModuleName).AnyTimes()
	utilityMock.EXPECT().Start().Return(nil).AnyTimes()
	utilityMock.EXPECT().Stop().Return(nil).AnyTimes()
	utilityMock.EXPECT().SetBus(bus).Return().AnyTimes()
	utilityMock.EXPECT().GetBus().Return(bus).AnyTimes()
	utilityMock.EXPECT().HandleTransaction(gomock.Any()).Return(nil).AnyTimes()

	return utilityMock
}

func newPersistenceMock(t *testing.T,
	bus modules.Bus,
	dbMap map[string]string,
) *mock_modules.MockPersistenceModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	persistenceMock := mock_modules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mock_modules.NewMockPersistenceReadContext(ctrl)

	persistenceMock.EXPECT().GetModuleName().Return(modules.PersistenceModuleName).AnyTimes()
	persistenceMock.EXPECT().Start().Return(nil).AnyTimes()
	persistenceMock.EXPECT().Stop().Return(nil).AnyTimes()
	persistenceMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	persistenceMock.EXPECT().GetBus().Return(bus).AnyTimes()
	persistenceMock.EXPECT().NewReadContext(gomock.Any()).Return(persistenceReadContextMock, nil).AnyTimes()

	persistenceMock.EXPECT().ReleaseWriteContext().Return(nil).AnyTimes()
	persistenceReadContextMock.
		EXPECT().
		GetIBCStoreEntry(gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(key []byte, _ uint64) ([]byte, error) {
				value, ok := dbMap[hex.EncodeToString(key)]
				if !ok {
					return nil, core_types.ErrIBCKeyDoesNotExist(string(key))
				}
				bz, err := hex.DecodeString(value)
				if err != nil {
					return nil, err
				}
				if bytes.Equal(bz, nil) {
					return nil, core_types.ErrIBCKeyDoesNotExist(string(key))
				}
				return bz, nil
			}).
		AnyTimes()

	persistenceReadContextMock.
		EXPECT().
		Release().
		AnyTimes()

	return persistenceMock
}

func newTreeStoreMock(t *testing.T,
	bus modules.Bus,
	tree *smt.SMT,
	nodeStore kvstore.KVStore,
) *mock_modules.MockTreeStoreModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	treeStoreMock := mock_modules.NewMockTreeStoreModule(ctrl)
	treeStoreMock.EXPECT().GetModuleName().Return(modules.TreeStoreSubmoduleName).AnyTimes()
	treeStoreMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	treeStoreMock.EXPECT().GetBus().Return(bus).AnyTimes()

	treeStoreMock.
		EXPECT().
		GetTree(gomock.Any()).
		DoAndReturn(
			func(_ string) ([]byte, kvstore.KVStore) {
				return tree.Root(), nodeStore
			}).
		AnyTimes()

	return treeStoreMock
}

func newTestP2PModule(t *testing.T, bus modules.Bus) modules.P2PModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	p2pMock := mock_modules.NewMockP2PModule(ctrl)

	p2pMock.EXPECT().Start().Return(nil).AnyTimes()
	p2pMock.EXPECT().Stop().Return(nil).AnyTimes()
	p2pMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	p2pMock.EXPECT().GetBus().Return(bus).AnyTimes()
	p2pMock.EXPECT().
		Broadcast(gomock.Any()).
		Return(nil).
		AnyTimes()
	p2pMock.EXPECT().
		Send(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	p2pMock.EXPECT().GetModuleName().Return(modules.P2PModuleName).AnyTimes()
	p2pMock.EXPECT().HandleEvent(gomock.Any()).Return(nil).AnyTimes()

	return p2pMock
}

// TECHDEBT: centralise these helper functions in internal/testutils
func newTestRuntimeConfig(t *testing.T) *runtime.Manager {
	t.Helper()
	cfg, err := configs.CreateTempConfig(&configs.Config{
		Consensus: &configs.ConsensusConfig{
			PrivateKey: "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
		},
		Utility: &configs.UtilityConfig{
			MaxMempoolTransactionBytes: 1000000,
			MaxMempoolTransactions:     1000,
		},
		Persistence: &configs.PersistenceConfig{
			PostgresUrl:       "",
			NodeSchema:        "test_schema",
			BlockStorePath:    ":memory:",
			TxIndexerPath:     ":memory:",
			TreesStoreDir:     ":memory:",
			MaxConnsCount:     50,
			MinConnsCount:     1,
			MaxConnLifetime:   "5m",
			MaxConnIdleTime:   "1m",
			HealthCheckPeriod: "30s",
		},
		Validator: &configs.ValidatorConfig{Enabled: true},
		IBC: &configs.IBCConfig{
			Enabled:   true,
			StoresDir: ":memory:",
			Host: &configs.IBCHostConfig{
				PrivateKey: "0ca1a40ddecdab4f5b04fa0bfed1d235beaa2b8082e7554425607516f0862075dfe357de55649e6d2ce889acf15eb77e94ab3c5756fe46d3c7538d37f27f115e",
			},
		},
	})
	if err != nil {
		t.Fatalf("Error creating config: %s", err)
	}
	genesisState, _ := test_artifacts.NewGenesisState(0, 0, 0, 0)
	runtimeCfg := runtime.NewManager(cfg, genesisState)
	return runtimeCfg
}
