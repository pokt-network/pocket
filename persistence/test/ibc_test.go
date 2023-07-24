package test

import (
	"fmt"
	"strconv"
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestIBC_SetIBCStoreEntry(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	testCases := []struct {
		name           string
		height         int64
		key            []byte
		value          []byte
		expectedErrStr *string
	}{
		{
			name:           "Successfully set key at height 1",
			height:         1,
			key:            []byte("key"),
			value:          []byte("value"),
			expectedErrStr: nil,
		},
		{
			name:           "Successfully set key at height 2",
			height:         2,
			key:            []byte("key"),
			value:          []byte("value2"),
			expectedErrStr: nil,
		},
		{
			name:           "Successfully set key to nil at height 3",
			height:         3,
			key:            []byte("key"),
			value:          nil,
			expectedErrStr: nil,
		},
		{
			name:           "Fails to set an existing key at height 3",
			height:         3,
			key:            []byte("key"),
			value:          []byte("new value"),
			expectedErrStr: duplicateError("ibc_entries"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db.Height = tc.height
			err := db.SetIBCStoreEntry(tc.key, tc.value)
			if tc.expectedErrStr != nil {
				require.EqualError(t, err, *tc.expectedErrStr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIBC_GetIBCStoreEntry(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	err := db.SetIBCStoreEntry([]byte("key"), []byte("value"))
	require.NoError(t, err)
	db.Height = 2
	err = db.SetIBCStoreEntry([]byte("key"), []byte("value2"))
	require.NoError(t, err)
	db.Height = 3
	err = db.SetIBCStoreEntry([]byte("key"), nil)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		height        uint64
		key           []byte
		expectedValue []byte
		expectedErr   error
	}{
		{
			name:          "Successfully get key at height 1",
			height:        1,
			key:           []byte("key"),
			expectedValue: []byte("value"),
			expectedErr:   nil,
		},
		{
			name:          "Successfully get key updated at height 2",
			height:        2,
			key:           []byte("key"),
			expectedValue: []byte("value2"),
			expectedErr:   nil,
		},
		{
			name:          "Fails to get key nil at height 3",
			height:        3,
			key:           []byte("key"),
			expectedValue: nil,
			expectedErr:   coreTypes.ErrIBCKeyDoesNotExist("key"),
		},
		{
			name:          "Fails to get unset key",
			height:        3,
			key:           []byte("key2"),
			expectedValue: nil,
			expectedErr:   coreTypes.ErrIBCKeyDoesNotExist("key2"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.GetIBCStoreEntry(tc.key, tc.height)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, got, tc.expectedValue)
		})
	}
}

type attribute struct {
	key   []byte
	value []byte
}

var (
	baseAttributeKey   = []byte("testKey")
	baseAttributeValue = []byte("testValue")
)

func TestIBC_SetIBCEvent(t *testing.T) {
	// Setup database
	db := NewTestPostgresContext(t, 1)
	// Add a single event at height 1
	event := new(coreTypes.IBCEvent)
	event.Topic = "test"
	event.Attributes = append(event.Attributes, &coreTypes.Attribute{
		Key:   baseAttributeKey,
		Value: baseAttributeValue,
	})
	require.NoError(t, db.SetIBCEvent(event))

	testCases := []struct {
		name           string
		height         uint64
		topic          string
		attributes     []attribute
		expectedErrStr *string
	}{
		{
			name:   "Successfully set new event at height 1",
			height: 1,
			topic:  "test",
			attributes: []attribute{
				{
					key:   []byte("key"),
					value: []byte("value"),
				},
				{
					key:   []byte("key2"),
					value: []byte("value2"),
				},
			},
			expectedErrStr: nil,
		},
		{
			name:   "Successfully set new event at height 2",
			height: 2,
			topic:  "test",
			attributes: []attribute{
				{
					key:   []byte("key"),
					value: []byte("value"),
				},
				{
					key:   []byte("key2"),
					value: []byte("value2"),
				},
			},
			expectedErrStr: nil,
		},
		{
			name:   "Successfully set a duplicate event new height",
			height: 2,
			topic:  "test",
			attributes: []attribute{
				{
					key:   []byte("testKey"),
					value: []byte("testValue"),
				},
			},
			expectedErrStr: nil,
		},
		{
			name:   "Fails to set a duplicate event at height 1",
			height: 1,
			topic:  "test",
			attributes: []attribute{
				{
					key:   baseAttributeKey,
					value: baseAttributeValue,
				},
			},
			expectedErrStr: duplicateError("ibc_events"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db.Height = int64(tc.height)
			event := new(coreTypes.IBCEvent)
			event.Topic = tc.topic
			for _, attr := range tc.attributes {
				event.Attributes = append(event.Attributes, &coreTypes.Attribute{
					Key:   attr.key,
					Value: attr.value,
				})
			}
			err := db.SetIBCEvent(event)
			if tc.expectedErrStr != nil {
				require.EqualError(t, err, *tc.expectedErrStr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIBC_GetIBCEvent(t *testing.T) {
	// Setup database
	db := NewTestPostgresContext(t, 1)
	// Add events "testKey0", "testKey1", "testKey2", "testKey3"
	// at heights 1, 2, 3, 3 respectively
	events := make([]*coreTypes.IBCEvent, 0, 4)
	for i := 0; i < 4; i++ {
		event := new(coreTypes.IBCEvent)
		event.Topic = "test"
		s := strconv.Itoa(i)
		event.Attributes = append(event.Attributes, &coreTypes.Attribute{
			Key:   []byte("testKey" + s),
			Value: []byte("testValue" + s),
		})
		events = append(events, event)
	}
	for i, event := range events {
		db.Height = int64(i + 1)
		if i == 3 {
			db.Height = int64(i)
		}
		require.NoError(t, db.SetIBCEvent(event))
	}

	testCases := []struct {
		name           string
		height         uint64
		topic          string
		eventsIndexes  []int
		expectedLength int
	}{
		{
			name:           "Successfully get events at height 1",
			height:         1,
			topic:          "test",
			eventsIndexes:  []int{0},
			expectedLength: 1,
		},
		{
			name:           "Successfully get events at height 2",
			height:         2,
			topic:          "test",
			eventsIndexes:  []int{1},
			expectedLength: 1,
		},
		{
			name:           "Successfully get events at height 3",
			height:         3,
			topic:          "test",
			eventsIndexes:  []int{2, 3},
			expectedLength: 2,
		},
		{
			name:           "Successfully returns empty array when no events found",
			height:         3,
			topic:          "test2",
			eventsIndexes:  []int{},
			expectedLength: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.GetIBCEvents(tc.height, tc.topic)
			require.NoError(t, err)
			require.Len(t, got, tc.expectedLength)
			for i, index := range tc.eventsIndexes {
				require.Equal(t, events[index].Topic, got[i].Topic)
				require.Equal(t, events[index].Attributes[0].Key, got[i].Attributes[0].Key)
				require.Equal(t, events[index].Attributes[0].Value, got[i].Attributes[0].Value)
			}
		})
	}
}

func duplicateError(tableName string) *string {
	str := fmt.Sprintf("ERROR: duplicate key value violates unique constraint \"%s_pkey\" (SQLSTATE 23505)", tableName)
	return &str
}
