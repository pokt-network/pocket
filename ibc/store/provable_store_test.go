package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/golang/mock/gomock"
	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/runtime/configs"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	mockModules "github.com/pokt-network/pocket/shared/modules/mocks"
	"github.com/pokt-network/smt"
	"github.com/stretchr/testify/require"
)

func TestProvableStore_Get(t *testing.T) {
	testCase := []struct {
		name          string
		key           []byte
		expectedValue []byte
		expectedError error
	}{
		{
			name:          "key exists",
			key:           []byte("key1"),
			expectedValue: []byte("value1"),
			expectedError: nil,
		},
		{
			name:          "key does not exist",
			key:           []byte("not exists"),
			expectedValue: nil,
			expectedError: coreTypes.ErrIBCKeyDoesNotExist("test/not exists"),
		},
		{
			name:          "value is nil",
			key:           []byte("key2"),
			expectedValue: nil,
			expectedError: coreTypes.ErrIBCKeyDoesNotExist("test/key2"),
		},
	}

	provableStore := newTestProvableStore(t)
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			value, err := provableStore.Get(tc.key)
			if tc.expectedError != nil {
				require.Equal(t, err, tc.expectedError)
				return
			}
			require.Equal(t, value, tc.expectedValue)
			require.NoError(t, err)
		})
	}
}

func TestProvableStore_CommitmentProof(t *testing.T) {
	tc := []struct {
		name          string
		membership    bool
		key           []byte
		value         []byte
		expectedError error
	}{
		{
			name:       "create membership proof",
			membership: true,
			key:        []byte("key1"),
			value:      []byte("value1"),
		},
		{
			name:       "create non membership proof",
			membership: false,
			key:        []byte("not exists"),
			value:      nil,
		},
	}

	provableStore := newTestProvableStore(t)
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			var proof *ics23.CommitmentProof
			var err error
			if tc.membership {
				proof, err = provableStore.CreateMembershipProof(tc.key, tc.value)
			} else {
				proof, err = provableStore.CreateNonMembershipProof(tc.key)
			}
			require.NoError(t, err)
			require.NotNil(t, proof)
		})
	}
}

func TestProvableStore_GetAndProve(t *testing.T) {
	testCase := []struct {
		name          string
		key           []byte
		expectedValue []byte
		expectedError error
	}{
		{
			name:          "key exists",
			key:           []byte("key1"),
			expectedValue: []byte("value1"),
			expectedError: nil,
		},
		{
			name:          "key does not exist",
			key:           []byte("not exists"),
			expectedValue: nil,
			expectedError: coreTypes.ErrIBCKeyDoesNotExist("test/not exists"),
		},
		{
			name:          "value is nil",
			key:           []byte("key2"),
			expectedValue: nil,
			expectedError: coreTypes.ErrIBCKeyDoesNotExist("test/key2"),
		},
	}
	provableStore := newTestProvableStore(t)
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			value, proof, err := provableStore.GetAndProve(tc.key)
			if tc.expectedError != nil {
				require.Equal(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, value, tc.expectedValue)
			require.NotNil(t, proof)
			if tc.expectedValue != nil {
				require.NotNil(t, proof.GetExist())
				require.Equal(t, proof.GetExist().Value, tc.expectedValue)
				require.Equal(t, string(proof.GetExist().Key), "test/"+string(tc.key))
			} else {
				require.NotNil(t, proof.GetExclusion())
				require.Equal(t, string(proof.GetExist().Key), "test/"+string(tc.key))
			}
		})
	}
}

func TestProvableStore_FlushCache(t *testing.T) {
	provableStore := newTestProvableStore(t)
	kvs := []struct {
		key   []byte
		value []byte
	}{
		{
			key:   []byte("testKey1"),
			value: []byte("testValue1"),
		},
		{
			key:   []byte("testKey2"),
			value: []byte("testValue2"),
		},
		{
			key:   []byte("testKey3"),
			value: nil,
		},
	}
	for _, kv := range kvs {
		if bytes.Equal(kv.value, nil) {
			err := provableStore.Delete(kv.key)
			require.NoError(t, err)
		} else {
			err := provableStore.Set(kv.key, kv.value)
			require.NoError(t, err)
		}
	}
	cache := kvstore.NewMemKVStore()
	require.NoError(t, provableStore.FlushCache(cache))
	keys, values, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 3)
	for i, key := range keys {
		require.Equal(t, string(key), prepareTestCacheEntry(1, kvs[i].key))
		if kvs[i].value == nil {
			require.Equal(t, values[i], []byte{})
		} else {
			require.Equal(t, values[i], kvs[i].value)
		}
	}
	require.NoError(t, cache.Stop())
}

func TestProvableStore_PruneCache(t *testing.T) {
	provableStore := newTestProvableStore(t)
	kvs := []struct {
		key   []byte
		value []byte
	}{
		{
			key:   []byte("testKey1"),
			value: []byte("testValue1"),
		},
		{
			key:   []byte("testKey2"),
			value: []byte("testValue2"),
		},
		{
			key:   []byte("testKey3"),
			value: nil,
		},
	}
	for _, kv := range kvs {
		if bytes.Equal(kv.value, nil) {
			err := provableStore.Delete(kv.key)
			require.NoError(t, err)
		} else {
			err := provableStore.Set(kv.key, kv.value)
			require.NoError(t, err)
		}
	}
	cache := kvstore.NewMemKVStore()
	require.NoError(t, provableStore.FlushCache(cache))
	keys, _, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 3) // 3 entries in cache should be flushed to disk
	err = cache.Set([]byte(prepareTestCacheEntry(2, []byte("testKey1"))), []byte("testValue1"))
	require.NoError(t, err)
	require.NoError(t, provableStore.PruneCache(cache, 1))
	keys, values, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 1) // 3 entries from height 1 should be pruned only single height 2 entry remains
	require.Equal(t, string(keys[0]), "test/2/test/testKey1")
	require.Equal(t, values[0], []byte("testValue1"))
	require.NoError(t, cache.Stop())
}

func TestProvableStore_RestoreCache(t *testing.T) {
	provableStore := newTestProvableStore(t)
	kvs := []struct {
		key   []byte
		value []byte
	}{
		{
			key:   []byte("testKey1"),
			value: []byte("testValue1"),
		},
		{
			key:   []byte("testKey2"),
			value: []byte("testValue2"),
		},
		{
			key:   []byte("testKey3"),
			value: nil,
		},
	}
	for _, kv := range kvs {
		if bytes.Equal(kv.value, nil) {
			require.NoError(t, provableStore.Delete(kv.key))
		} else {
			require.NoError(t, provableStore.Set(kv.key, kv.value))
		}
	}

	cache := kvstore.NewMemKVStore()
	require.NoError(t, provableStore.FlushCache(cache))
	keys, values, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 3)
	require.NoError(t, cache.ClearAll())
	require.NoError(t, provableStore.FlushCache(cache))
	newKeys, _, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, newKeys, 0)
	for i, key := range keys {
		require.NoError(t, cache.Set(key, values[i]))
	}
	keys, values, err = cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 3)

	require.NoError(t, provableStore.RestoreCache(cache, 1))
	newKeys, _, err = cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, newKeys, 0)
	require.NoError(t, provableStore.FlushCache(cache))
	newKeys, newValues, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, newKeys, 3)
	for i, key := range newKeys {
		require.Equal(t, key, keys[i])
		require.Equal(t, newValues[i], values[i])
	}
	require.NoError(t, cache.Stop())
}

func TestProvableStore_Root(t *testing.T) {
	provableStore := newTestProvableStore(t)
	root := provableStore.Root()
	require.Equal(t, "91bdac23aa7a63a812a32f73b3ceb61d6128642dd7675bf34543fc3e771d0030", hex.EncodeToString(root))
}

func TestProvableStore_GetCommitmentPrefix(t *testing.T) {
	provableStore := newTestProvableStore(t)
	prefix := provableStore.GetCommitmentPrefix()
	require.True(t, bytes.Equal([]byte("test"), prefix))
}

// Ref: provable_store.go:
// func (c *cachedEntry) prepare() (key, value []byte)
func prepareTestCacheEntry(height uint64, key []byte) string {
	return fmt.Sprintf("test/%d/test/%s", height, string(key))
}

func newTestProvableStore(t *testing.T) modules.ProvableStore {
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

	return newProvableStore(bus, []byte("test"), privKey)
}

func setupDB(t *testing.T) (*smt.SMT, kvstore.KVStore, map[string]string) {
	dbMap := make(map[string]string, 0)
	nodeStore := kvstore.NewMemKVStore()
	tree := smt.NewSparseMerkleTree(nodeStore, sha256.New())

	keys := [][]byte{
		[]byte("test/key1"),
		[]byte("test/key2"),
		[]byte("test/key3"),
	}
	values := [][]byte{
		[]byte("value1"),
		nil,
		[]byte("value3"),
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

func newConsensusMock(t *testing.T, bus modules.Bus) *mockModules.MockConsensusModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	consensusMock := mockModules.NewMockConsensusModule(ctrl)
	consensusMock.EXPECT().GetModuleName().Return(modules.ConsensusModuleName).AnyTimes()
	consensusMock.EXPECT().Start().Return(nil).AnyTimes()
	consensusMock.EXPECT().Stop().Return(nil).AnyTimes()
	consensusMock.EXPECT().SetBus(gomock.Any()).Return().AnyTimes()
	consensusMock.EXPECT().GetBus().Return(bus).AnyTimes()
	consensusMock.EXPECT().CurrentHeight().Return(uint64(1)).AnyTimes()

	return consensusMock
}

func newUtilityMock(t *testing.T, bus modules.Bus) *mockModules.MockUtilityModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	utilityMock := mockModules.NewMockUtilityModule(ctrl)
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
) *mockModules.MockPersistenceModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	persistenceMock := mockModules.NewMockPersistenceModule(ctrl)
	persistenceReadContextMock := mockModules.NewMockPersistenceReadContext(ctrl)

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
			func(key []byte, _ int64) ([]byte, error) {
				value, ok := dbMap[hex.EncodeToString(key)]
				if !ok {
					return nil, coreTypes.ErrIBCKeyDoesNotExist(string(key))
				}
				bz, err := hex.DecodeString(value)
				if err != nil {
					return nil, err
				}
				if bytes.Equal(bz, nil) {
					return nil, coreTypes.ErrIBCKeyDoesNotExist(string(key))
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
) *mockModules.MockTreeStoreModule {
	t.Helper()

	ctrl := gomock.NewController(t)
	treeStoreMock := mockModules.NewMockTreeStoreModule(ctrl)
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
	p2pMock := mockModules.NewMockP2PModule(ctrl)

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
