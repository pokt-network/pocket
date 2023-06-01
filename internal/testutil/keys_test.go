package testutil_test

import (
	"testing"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestLoadLocalnetPrivateKeys(t *testing.T) {
	keyCount := 1000
	privKeys := testutil.LoadLocalnetPrivateKeys(t, keyCount)

	require.Lenf(t, privKeys, keyCount, "expected %d private keys; got %d", keyCount, len(privKeys))

	// ensure each key is unique
	seen := make(map[string]struct{})
	for _, privKey := range privKeys {
		seen[privKey.String()] = struct{}{}
	}

	require.Lenf(t, seen, keyCount, "expected %d unique private keys; got %d", keyCount, len(seen))
}
