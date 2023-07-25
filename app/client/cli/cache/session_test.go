package cache

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pokt-network/pocket/rpc"
)

func TestGet(t *testing.T) {
	const (
		app1             = "app1Addr"
		relaychainEth    = "ETH-Goerli"
		numSessionBlocks = 4
		sessionHeight    = 8
		sessionNumber    = 2
	)

	session1 := &rpc.Session{
		Application: rpc.ProtocolActor{
			ActorType: rpc.Application,
			Address:   "app1Addr",
			Chains:    []string{relaychainEth},
		},
		Chain:            relaychainEth,
		NumSessionBlocks: numSessionBlocks,
		SessionHeight:    sessionHeight,
		SessionNumber:    sessionNumber,
	}

	testCases := []struct {
		name          string
		cacheContents []*rpc.Session
		app           string
		chain         string
		expected      *rpc.Session
		expectedErr   error
	}{
		{
			name:          "Return cached session",
			cacheContents: []*rpc.Session{session1},
			app:           app1,
			chain:         relaychainEth,
			expected:      session1,
		},
		{
			name:        "Error returned for session not found in cache",
			app:         "foo",
			chain:       relaychainEth,
			expectedErr: errSessionNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbPath, err := os.MkdirTemp("", "cacheStoragePath")
			require.NoError(t, err)
			defer os.RemoveAll(dbPath)

			cache, err := NewSessionCache(dbPath)
			require.NoError(t, err)

			for _, s := range tc.cacheContents {
				err := cache.Set(s)
				require.NoError(t, err)
			}

			got, err := cache.Get(tc.app, tc.chain)
			require.ErrorIs(t, err, tc.expectedErr)
			require.EqualValues(t, tc.expected, got)
		})
	}
}
