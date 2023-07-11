package ibc

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/persistence/kvstore"
	"github.com/pokt-network/pocket/shared/messaging"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestHandleEvent_FlushCaches(t *testing.T) {
	mgr, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	tmpDir := mgr.GetConfig().IBC.StoresDir + "/caches"
	ibcHost := ibcMod.GetBus().GetIBCHost()
	store, err := ibcHost.GetProvableStore("test")
	require.NoError(t, err)

	// set height
	publishNewHeightEvent(t, ibcMod.GetBus(), 1)

	kvs := []struct {
		key   []byte
		value []byte
	}{
		{
			key:   []byte("key1"),
			value: []byte("value1"),
		},
		{
			key:   []byte("key2"),
			value: []byte("value2"),
		},
		{
			key:   []byte("key3"),
			value: []byte("value3"),
		},
		{
			key:   []byte("key4"),
			value: nil,
		},
	}

	for _, kv := range kvs {
		if kv.value != nil {
			require.NoError(t, store.Set(kv.key, kv.value))
		} else {
			require.NoError(t, store.Delete(kv.key))
		}
	}

	// increment the height
	publishNewHeightEvent(t, ibcMod.GetBus(), 2)

	cache, err := kvstore.NewKVStore(tmpDir)
	require.NoError(t, err)
	keys, values, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, keys, 4)
	require.Len(t, values, 4)
	for i, key := range keys {
		require.Equal(t, string(key), prepareCacheKey(1, kvs[i].key))
		if kvs[i].value == nil {
			require.Equal(t, values[i], []byte{})
			continue
		}
		require.Equal(t, values[i], kvs[i].value)
	}

	err = cache.ClearAll()
	require.NoError(t, err)

	newKeys, newValues, err := cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, newKeys, 0)
	require.Len(t, newValues, 0)

	require.NoError(t, cache.Stop())

	// flush the cache
	err = ibcHost.GetBus().GetBulkStoreCacher().FlushAllEntries()
	require.NoError(t, err)

	cache, err = kvstore.NewKVStore(tmpDir)
	require.NoError(t, err)

	// check in memory cache was cleared (ie nothing flushed)
	newKeys, newValues, err = cache.GetAll([]byte{}, false)
	require.NoError(t, err)
	require.Len(t, newKeys, 0)
	require.Len(t, newValues, 0)

	require.NoError(t, cache.Stop())
}

// using MaxHeightCached = 3
func TestHandleEvent_PruneCaches(t *testing.T) {
	mgr, _, _, _, ibcMod := prepareEnvironment(t, 1, 0, 0, 0)
	tmpDir := mgr.GetConfig().IBC.StoresDir + "/caches"
	ibcHost := ibcMod.GetBus().GetIBCHost()
	store, err := ibcHost.GetProvableStore("test")
	require.NoError(t, err)

	kvs := []struct {
		key   []byte
		value []byte
	}{
		{
			key:   []byte("key1"),
			value: []byte("value1"),
		},
		{
			key:   []byte("key2"),
			value: []byte("value2"),
		},
		{
			key:   []byte("key3"),
			value: []byte("value3"),
		},
		{
			key:   []byte("key4"),
			value: nil,
		},
	}

	testCases := []struct {
		name                  string
		heightsStored         []uint64 // the different heights where entries are written to the cache
		expectedHeightsCached []uint64 // the different heights expected in the cache after pruning
		cacheLength           int      // the length of the cache after pruning
	}{
		{
			name:                  "No pruning after single height increase",
			heightsStored:         []uint64{1, 2},
			expectedHeightsCached: []uint64{1},
			cacheLength:           4,
		},
		{
			name:                  "No pruning after two height increase",
			heightsStored:         []uint64{1, 2, 3},
			expectedHeightsCached: []uint64{1, 2},
			cacheLength:           8,
		},
		{
			name:                  "No pruning at max height stored = 3",
			heightsStored:         []uint64{1, 2, 3, 4},
			expectedHeightsCached: []uint64{1, 2, 3},
			cacheLength:           12,
		},
		{
			name:                  "Pruning after 4 height increase",
			heightsStored:         []uint64{1, 2, 3, 4, 5},
			expectedHeightsCached: []uint64{2, 3, 4},
			cacheLength:           12,
		},
		{
			name:                  "Pruning after 5 height increase",
			heightsStored:         []uint64{1, 2, 3, 4, 5, 6},
			expectedHeightsCached: []uint64{3, 4, 5},
			cacheLength:           12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set height
			publishNewHeightEvent(t, ibcMod.GetBus(), tc.heightsStored[0])

			for _, height := range tc.heightsStored[1:] {
				for _, kv := range kvs {
					if kv.value != nil {
						require.NoError(t, store.Set(kv.key, kv.value))
					} else {
						require.NoError(t, store.Delete(kv.key))
					}
				}

				// increment the height
				publishNewHeightEvent(t, ibcMod.GetBus(), height)
			}

			cache, err := kvstore.NewKVStore(tmpDir)
			//nolint:errcheck // ignore error just make sure closes incase anything else fails
			defer cache.Stop()
			require.NoError(t, err)
			keys, values, err := cache.GetAll([]byte{}, false)
			require.NoError(t, err)
			require.Len(t, keys, tc.cacheLength)
			require.Len(t, values, tc.cacheLength)

			// iterate over all the keys in batches of 4 (for each expected height) confirming the expected height
			// is the same as the height stored in the cache for each batch
			for i, height := range tc.expectedHeightsCached {
				for j, key := range keys[i*4 : (i+1)*4] { // split keys into batches of 4 per height expected
					require.Equal(t, string(key), prepareCacheKey(height, kvs[j%4].key)) // validate the expected height is the actual height
					if kvs[j%4].value == nil {
						require.Equal(t, values[j], []byte{})
						continue
					}
					require.Equal(t, values[j], kvs[j%4].value)
				}
			}

			require.NoError(t, cache.Stop())
		})
	}
}

func prepareCacheKey(height uint64, key []byte) string {
	return fmt.Sprintf("test/%d/test/%s", height, string(key))
}

func publishNewHeightEvent(t *testing.T, bus modules.Bus, height uint64) {
	t.Helper()
	newHeightEvent, err := messaging.PackMessage(&messaging.ConsensusNewHeightEvent{Height: height})
	require.NoError(t, err)
	bus.GetConsensusModule().SetHeight(height)
	require.NoError(t, bus.GetIBCModule().HandleEvent(newHeightEvent.Content))
}
