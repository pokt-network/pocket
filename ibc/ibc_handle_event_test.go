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

// using MaxStoredHeight = 3
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
		name            string
		heights         []uint64
		expectedHeights []uint64
		length          int
	}{
		{
			name:            "No pruning after single height increase",
			heights:         []uint64{1, 2},
			expectedHeights: []uint64{1},
			length:          4,
		},
		{
			name:            "No pruning after two height increase",
			heights:         []uint64{1, 2, 3},
			expectedHeights: []uint64{1, 2},
			length:          8,
		},
		{
			name:            "No pruning at max height stored = 3",
			heights:         []uint64{1, 2, 3, 4},
			expectedHeights: []uint64{1, 2, 3},
			length:          12,
		},
		{
			name:            "Pruning after 4 height increase",
			heights:         []uint64{1, 2, 3, 4, 5},
			expectedHeights: []uint64{2, 3, 4},
			length:          12,
		},
		{
			name:            "Pruning after 5 height increase",
			heights:         []uint64{1, 2, 3, 4, 5, 6},
			expectedHeights: []uint64{3, 4, 5},
			length:          12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set height
			publishNewHeightEvent(t, ibcMod.GetBus(), tc.heights[0])

			for _, height := range tc.heights[1:] {
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
			defer cache.Stop() // incase of errors close store
			require.NoError(t, err)
			keys, values, err := cache.GetAll([]byte{}, false)
			require.NoError(t, err)
			require.Len(t, keys, tc.length)
			require.Len(t, values, tc.length)

			// iterate over the expected heights and check the keys and values are correct
			for i, height := range tc.expectedHeights {
				for j, key := range keys[i*4 : (i+1)*4] {
					require.Equal(t, string(key), prepareCacheKey(height, kvs[j%4].key))
					if kvs[j%4].value == nil {
						require.Equal(t, values[j], []byte{})
						continue
					}
					require.Equal(t, values[j], kvs[j%4].value)
				}
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
