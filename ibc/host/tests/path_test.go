package tests

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/ibc/host"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestPaths_GenerateValidIdentifiers(t *testing.T) {
	ids := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		switch i % 4 {
		case 0:
			cl := host.GenerateClientIdentifier()
			require.NotNil(t, cl)
			_, ok := ids[cl]
			require.False(t, ok)
			ids[cl] = "client"
		case 1:
			co := host.GenerateConnectionIdentifier()
			require.NotNil(t, co)
			_, ok := ids[co]
			require.False(t, ok)
			ids[co] = "connection"
		case 2:
			ch := host.GenerateChannelIdentifier()
			require.NotNil(t, ch)
			_, ok := ids[ch]
			require.False(t, ok)
			ids[ch] = "channel"
		case 3:
			po := host.GeneratePortIdentifier()
			require.NotNil(t, po)
			_, ok := ids[po]
			require.False(t, ok)
			ids[po] = "port"
		}
	}
	for k, v := range ids {
		var err error
		switch v {
		case "client":
			err = host.ValidateClientID(k)
		case "connection":
			err = host.ValidateConnectionID(k)
		case "channel":
			err = host.ValidateChannelID(k)
		case "port":
			err = host.ValidatePortID(k)
		}
		require.NoError(t, err)
	}
}

func TestPaths_CommitmentPrefix(t *testing.T) {
	prefix := &coreTypes.CommitmentPrefix{Prefix: []byte("test")}

	testCases := []struct {
		path     string
		prefix   *coreTypes.CommitmentPrefix
		expected []byte
		result   string
	}{
		{ // Successfully applies and removes prefix to produce the same path
			path:     "path",
			prefix:   &coreTypes.CommitmentPrefix{Prefix: []byte("test")},
			expected: []byte("test/path"),
			result:   "path",
		},
		{ // Fails to produce input path when given a different prefix
			path:     "path",
			prefix:   &coreTypes.CommitmentPrefix{Prefix: []byte("test2")},
			expected: []byte("test/path"),
			result:   "ath",
		},
	}

	for _, tc := range testCases {
		commitment := host.ApplyPrefix(prefix, tc.path)
		require.NotNil(t, commitment.GetPath())
		require.True(t, bytes.Equal(commitment.GetPath(), tc.expected))

		path := host.RemovePrefix(tc.prefix, commitment)
		require.NotNil(t, path)
		require.Equal(t, path, tc.result)
	}
}
