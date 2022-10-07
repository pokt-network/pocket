package test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/stretchr/testify/require"
)

func TestStateHash_DeterministicStateWhenUpdatingAppStake(t *testing.T) {
	// These hashes were determined manually by running the test, but hardcoded to guarantee
	// that the business logic doesn't change and that they remain deterministic.
	encodedAppHash := []string{
		"a405c3db598c9898c61b76c77f3e1ed94277a2bc683fbc4f9bd502c47633d617",
		"e431c357c0e0d9ef5999b52bc18d36aa0e1bedbd555a82dd5e8a8130b6b8fa6b",
		"a46c8024472f50a4ab887b8b1e06fdc578f0344eada2d68784325c27e74d6529",
	}

	for i := 0; i < 3; i++ {
		// Get the context at the new height and retrieve one of the apps
		height := int64(i + 1)
		heightBz := persistence.HeightToBytes(height)
		expectedAppHash := encodedAppHash[i]

		db := NewTestPostgresContext(t, height)

		apps, err := db.GetAllApps(height)
		require.NoError(t, err)
		app := apps[0]

		addrBz, err := hex.DecodeString(app.GetAddress())
		require.NoError(t, err)

		// Update the app's stake
		newStakeAmount := types.BigIntToString(big.NewInt(height + int64(420000000000)))
		err = db.SetAppStakeAmount(addrBz, newStakeAmount)
		require.NoError(t, err)

		// NOTE: The tx does not currently affect the state hash
		txBz := []byte("a tx, i am, which set the app stake amount to " + newStakeAmount)
		err = db.StoreTransaction(txBz)
		require.NoError(t, err)

		// Update & commit the state hash
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)
		require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

		err = db.Commit([]byte("proposer"), []byte("quorumCert"))
		require.NoError(t, err)

		// Verify the block contents
		blockBz, err := testPersistenceMod.GetBlockStore().Get(heightBz)
		require.NoError(t, err)

		var block types.Block
		err = codec.GetCodec().Unmarshal(blockBz, &block)
		require.NoError(t, err)
		require.Len(t, block.Transactions, 1)
		require.Equal(t, txBz, block.Transactions[0])
		require.Equal(t, expectedAppHash, block.Hash) // block
		if i > 0 {
			require.Equal(t, encodedAppHash[i-1], block.PrevHash) // chain
		}

	}
}
