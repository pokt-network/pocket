package host

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestPaths_GenerateValidIdentifiers(t *testing.T) {
	ids := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		switch i % 4 {
		case 0:
			cl := GenerateClientIdentifier()
			require.NotNil(t, cl)
			_, ok := ids[cl]
			require.False(t, ok)
			ids[cl] = "client"
		case 1:
			co := GenerateConnectionIdentifier()
			require.NotNil(t, co)
			_, ok := ids[co]
			require.False(t, ok)
			ids[co] = "connection"
		case 2:
			ch := GenerateChannelIdentifier()
			require.NotNil(t, ch)
			_, ok := ids[ch]
			require.False(t, ok)
			ids[ch] = "channel"
		case 3:
			po := GeneratePortIdentifier()
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
			err = ValidateClientID(k)
		case "connection":
			err = ValidateConnectionID(k)
		case "channel":
			err = ValidateChannelID(k)
		case "port":
			err = ValidatePortID(k)
		}
		require.NoError(t, err)
	}
}

func TestPaths_CommitmentPrefix(t *testing.T) {
	prefix := coreTypes.CommitmentPrefix([]byte("test"))

	testCases := []struct {
		name     string
		path     string
		prefix   coreTypes.CommitmentPrefix
		expected []byte
		result   string
	}{
		{
			name:     "Successfully applies and removes prefix to produce the same path",
			path:     "path",
			prefix:   coreTypes.CommitmentPrefix([]byte("test")),
			expected: []byte("test/path"),
			result:   "path",
		},
		{
			name:     "Fails to produce input path when given a different prefix",
			path:     "path",
			prefix:   coreTypes.CommitmentPrefix([]byte("test2")),
			expected: []byte("test/path"),
			result:   "ath",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			commitment := ApplyPrefix(prefix, tc.path)
			require.NotNil(t, commitment)
			require.Equal(t, []byte(commitment), tc.expected)

			path := RemovePrefix(tc.prefix, commitment)
			require.NotNil(t, path)
			require.Equal(t, path, tc.result)
		})
	}
}
