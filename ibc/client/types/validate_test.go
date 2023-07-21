package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientState_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		clientState *ClientState
		expectedErr error
	}{
		{
			name: "valid client state",
			clientState: &ClientState{
				Data:         []byte("data"),
				WasmChecksum: make([]byte, 32),
			},
			expectedErr: nil,
		},
		{
			name: "invalid client state: empty data",
			clientState: &ClientState{
				Data:         nil,
				WasmChecksum: make([]byte, 32),
			},
			expectedErr: errors.New("data cannot be empty"),
		},
		{
			name: "invalid client state: empty wasm checksum",
			clientState: &ClientState{
				Data:         []byte("data"),
				WasmChecksum: nil,
			},
			expectedErr: errors.New("wasm checksum cannot be empty"),
		},
		{
			name: "invalid client state: invalid wasm checksum",
			clientState: &ClientState{
				Data:         []byte("data"),
				WasmChecksum: []byte("invalid"),
			},
			expectedErr: errors.New("expected 32, got 7"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.clientState.Validate()
			if tc.expectedErr != nil {
				require.ErrorAs(t, err, &tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestConsensusState_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name           string
		consensusState *ConsensusState
		expectedErr    error
	}{
		{
			name: "valid consensus state",
			consensusState: &ConsensusState{
				Timestamp: 1,
				Data:      []byte("data"),
			},
			expectedErr: nil,
		},
		{
			name: "invalid consensus state: zero timestamp",
			consensusState: &ConsensusState{
				Timestamp: 0,
				Data:      []byte("data"),
			},
			expectedErr: errors.New("timestamp must be a positive Unix time"),
		},
		{
			name: "invalid consensus state: empty data",
			consensusState: &ConsensusState{
				Timestamp: 1,
				Data:      nil,
			},
			expectedErr: errors.New("data cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.consensusState.ValidateBasic()
			if tc.expectedErr != nil {
				require.ErrorAs(t, err, &tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestHeader_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name        string
		header      *Header
		expectedErr error
	}{
		{
			name: "valid header",
			header: &Header{
				Data: []byte("data"),
			},
			expectedErr: nil,
		},
		{
			name: "invalid header: empty data",
			header: &Header{
				Data: nil,
			},
			expectedErr: errors.New("data cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.header.ValidateBasic()
			if tc.expectedErr != nil {
				require.ErrorAs(t, err, &tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMisbehaviour_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name         string
		misbehaviour *Misbehaviour
		expectedErr  error
	}{
		{
			name: "valid misbehaviour",
			misbehaviour: &Misbehaviour{
				Data: []byte("data"),
			},
			expectedErr: nil,
		},
		{
			name: "invalid misbehaviour: empty data",
			misbehaviour: &Misbehaviour{
				Data: nil,
			},
			expectedErr: errors.New("data cannot be empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.misbehaviour.ValidateBasic()
			if tc.expectedErr != nil {
				require.ErrorAs(t, err, &tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
