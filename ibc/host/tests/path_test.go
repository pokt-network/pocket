package tests

import (
	"testing"

	"github.com/pokt-network/pocket/ibc/host"
	"github.com/stretchr/testify/require"
)

func TestPaths_GenerateValidIdentifiers(t *testing.T) {
	ids := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		switch i % 4 {
		case 0:
			cl := host.GenerateClientIdentifier(int64(i))
			require.NotNil(t, cl)
			_, ok := ids[cl]
			require.False(t, ok)
			ids[cl] = "client"
		case 1:
			co := host.GenerateConnectionIdentifier(int64(i))
			require.NotNil(t, co)
			_, ok := ids[co]
			require.False(t, ok)
			ids[co] = "connection"
		case 2:
			ch := host.GenerateChannelIdentifier(int64(i))
			require.NotNil(t, ch)
			_, ok := ids[ch]
			require.False(t, ok)
			ids[ch] = "channel"
		case 3:
			po := host.GeneratePortIdentifier(int64(i))
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
	prefix := host.CommitmentPrefix([]byte("test"))

	testCases := []struct {
		name     string
		path     string
		prefix   host.CommitmentPrefix
		expected []byte
		result   string
	}{
		{
			name:     "Successfully applies and removes prefix to produce the same path",
			path:     "path",
			prefix:   host.CommitmentPrefix([]byte("test")),
			expected: []byte("test/path"),
			result:   "path",
		},
		{
			name:     "Fails to produce input path when given a different prefix",
			path:     "path",
			prefix:   host.CommitmentPrefix([]byte("test2")),
			expected: []byte("test/path"),
			result:   "ath",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commitment := host.ApplyPrefix(prefix, tc.path)
			require.NotNil(t, commitment)
			require.Equal(t, []byte(commitment), tc.expected)

			path := host.RemovePrefix(tc.prefix, commitment)
			require.NotNil(t, path)
			require.Equal(t, path, tc.result)
		})
	}
}
