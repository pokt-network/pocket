package test

import (
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
