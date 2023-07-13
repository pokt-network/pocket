package path

import (
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func FuzzIdentifiers_GenerateValidIdentifiers(f *testing.F) {
	for i := 0; i < 100; i++ {
		switch i % 4 {
		case 0:
			f.Add("client")
		case 1:
			f.Add("connection")
		case 2:
			f.Add("channel")
		case 3:
			f.Add("port")
		}
	}
	f.Fuzz(func(t *testing.T, idType string) {
		switch idType {
		case "client":
			id := GenerateClientIdentifier()
			require.NotEmpty(t, id)
			require.GreaterOrEqual(t, len(id), 9)
			require.LessOrEqual(t, len(id), 64)
			err := ValidateClientID(id)
			require.NoError(t, err)
		case "connection":
			id := GenerateConnectionIdentifier()
			require.NotEmpty(t, id)
			require.GreaterOrEqual(t, len(id), 10)
			require.LessOrEqual(t, len(id), 64)
			err := ValidateConnectionID(id)
			require.NoError(t, err)
		case "channel":
			id := GenerateChannelIdentifier()
			require.NotEmpty(t, id)
			require.GreaterOrEqual(t, len(id), 8)
			require.LessOrEqual(t, len(id), 64)
			err := ValidateChannelID(id)
			require.NoError(t, err)
		case "port":
			id := GeneratePortIdentifier()
			require.NotEmpty(t, id)
			require.GreaterOrEqual(t, len(id), 2)
			require.LessOrEqual(t, len(id), 128)
			err := ValidatePortID(id)
			require.NoError(t, err)
		}
	})
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
