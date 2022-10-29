package test

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func TestStateHash_DeterministicStateWhenUpdatingAppStake(t *testing.T) {
	// These hashes were determined manually by running the test, but hardcoded to guarantee
	// that the business logic doesn't change and that they remain deterministic.
	encodedAppHash := []string{
		"f13ddc447bdebd38b1db7d534915992fa2b6dd4aabdc81868e3420df37b3647f",
		"40ba19443d9c18d12c17ed25e86ae1aa34ecc1080cc0208854dc72e51f9b8b94",
		"b7578e3d5a675effe31475ce0df034a2aec21c983b4a3c34c23ad9b583cb60eb",
	}

	for i := 0; i < 3; i++ {
		// Get the context at the new height and retrieve one of the apps
		height := int64(i + 1)
		heightBz := heightToBytes(height)
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

		txBz := []byte("a tx, i am, which set the app stake amount to " + newStakeAmount)
		txResult := indexer.TxRes{
			Tx:            txBz,
			Height:        height,
			Index:         0,
			ResultCode:    0,
			Error:         "",
			SignerAddr:    "",
			RecipientAddr: "",
			MessageType:   "",
		}

		// txResult := mockTxResult(t, height, txBz)
		err = db.StoreTransaction(modules.TxResult(&txResult))
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
		// require.Equal(t, txResult.GetTx(), block.Transactions[0])
		require.Equal(t, expectedAppHash, block.Hash) // block
		if i > 0 {
			require.Equal(t, encodedAppHash[i-1], block.PrevHash) // chain
		}

	}
}

// Tests/debug to implement:
// - Visibility into what's in the tree
// - Benchmarking many inserts
// - Release / revert mid block and making sure everything is reverted
// - Thinking about how it can be synched
// - Playing back several blocks

func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
