package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/app/client/cli/cache"
	"github.com/pokt-network/pocket/rpc"
)

const (
	testRelaychainEth = "ETH-Goerli"
	testSessionHeight = 8
	testCurrentHeight = 9
)

func TestGetSessionFromCache(t *testing.T) {
	const app1Addr = "app1Addr"

	testCases := []struct {
		name           string
		cachedSessions []*rpc.Session
		expected       *rpc.Session
		expectedErr    error
	}{
		{
			name:           "cached session is returned",
			cachedSessions: []*rpc.Session{testSession(app1Addr, testSessionHeight)},
			expected:       testSession(app1Addr, testSessionHeight),
		},
		{
			name:        "nil session cache returns an error",
			expectedErr: errNoSessionCache,
		},
		{
			name:           "session not found in cache",
			cachedSessions: []*rpc.Session{testSession("foo", testSessionHeight)},
			expectedErr:    errSessionNotFoundInCache,
		},
		{
			name:           "cached session does not match the provided height",
			cachedSessions: []*rpc.Session{testSession(app1Addr, 9999999)},
			expectedErr:    errNoMatchingSessionInCache,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c cache.SessionCache
			// prepare cache with test session for this unit test
			if len(tc.cachedSessions) > 0 {
				dbPath, err := os.MkdirTemp("", "cliCacheStoragePath")
				require.NoError(t, err)
				defer os.RemoveAll(dbPath)

				c, err = cache.Create(dbPath)
				require.NoError(t, err)

				for _, s := range tc.cachedSessions {
					err := c.Set(s)
					require.NoError(t, err)
				}
			}

			got, err := getSessionFromCache(c, app1Addr, testRelaychainEth, testCurrentHeight)
			require.ErrorIs(t, err, tc.expectedErr)
			require.EqualValues(t, tc.expected, got)
		})
	}
}

func testSession(appAddr string, height int64) *rpc.Session {
	const numSessionBlocks = 4

	return &rpc.Session{
		Application: rpc.ProtocolActor{
			ActorType: rpc.Application,
			Address:   appAddr,
			Chains:    []string{testRelaychainEth},
		},
		Chain:            testRelaychainEth,
		NumSessionBlocks: numSessionBlocks,
		SessionHeight:    height,
		SessionNumber:    (height / numSessionBlocks), // assumes numSessionBlocks never changed
	}
}
