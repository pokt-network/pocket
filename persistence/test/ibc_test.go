package test

import (
	"strconv"
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestSetIBCStoreEntry(t *testing.T) {
	db := NewTestPostgresContext(t, 1)

	testCases := []struct {
		name        string
		height      int64
		key         []byte
		value       []byte
		expectedErr string
	}{
		{
			name:        "Successfully set key at height 1",
			height:      1,
			key:         []byte("key"),
			value:       []byte("value"),
			expectedErr: "",
		},
		{
			name:        "Successfully set key at height 2",
			height:      2,
			key:         []byte("key"),
			value:       []byte("value2"),
			expectedErr: "",
		},
		{
			name:        "Successfully set key to nil at height 3",
			height:      3,
			key:         []byte("key"),
			value:       nil,
			expectedErr: "",
		},
		{
			name:        "Fails to set an existing key at height 3",
			height:      3,
			key:         []byte("key"),
			value:       []byte("new value"),
			expectedErr: "ERROR: duplicate key value violates unique constraint \"ibc_entries_pkey\" (SQLSTATE 23505)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db.Height = tc.height
			err := db.SetIBCStoreEntry(tc.key, tc.value)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetIBCStoreEntry(t *testing.T) {
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
		height        int64
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

func TestIBCSetEvent(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	event := new(coreTypes.IBCEvent)
	event.Topic = "test"
	event.Height = 1
	event.Attributes = append(event.Attributes, &coreTypes.Attribute{
		Key:   []byte("testKey"),
		Value: []byte("testValue"),
	})
	require.NoError(t, db.SetIBCEvent(event))

	testCases := []struct {
		name        string
		height      uint64
		topic       string
		attributes  []attribute
		expectedErr string
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
			expectedErr: "",
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
			expectedErr: "",
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
			expectedErr: "",
		},
		{
			name:   "Fails to set a duplicate event at height 1",
			height: 1,
			topic:  "test",
			attributes: []attribute{
				{
					key:   []byte("testKey"),
					value: []byte("testValue"),
				},
			},
			expectedErr: "ERROR: duplicate key value violates unique constraint \"ibc_events_pkey\" (SQLSTATE 23505)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db.Height = int64(tc.height)
			event := new(coreTypes.IBCEvent)
			event.Topic = tc.topic
			event.Height = tc.height
			for _, attr := range tc.attributes {
				event.Attributes = append(event.Attributes, &coreTypes.Attribute{
					Key:   attr.key,
					Value: attr.value,
				})
			}
			err := db.SetIBCEvent(event)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetIBCEvent(t *testing.T) {
	db := NewTestPostgresContext(t, 1)
	events := make([]*coreTypes.IBCEvent, 0, 4)
	for i := 0; i < 4; i++ {
		event := new(coreTypes.IBCEvent)
		event.Topic = "test"
		event.Height = uint64(i + 1)
		if i == 3 {
			event.Height = uint64(i)
		}
		s := strconv.Itoa(i)
		event.Attributes = append(event.Attributes, &coreTypes.Attribute{
			Key:   []byte("testKey" + s),
			Value: []byte("testValue" + s),
		})
		events = append(events, event)
	}
	for _, event := range events {
		db.Height = int64(event.Height)
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
				require.Equal(t, events[index].Height, got[i].Height)
				require.Equal(t, events[index].Topic, got[i].Topic)
				require.Equal(t, events[index].Attributes[0].Key, got[i].Attributes[0].Key)
				require.Equal(t, events[index].Attributes[0].Value, got[i].Attributes[0].Value)
			}
		})
	}
}
