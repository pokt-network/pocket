package test

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

const (
	txBytesRandSeed = "42"
	txBytesSize     = 42

	// This value is arbitrarily selected, but needs to be a constant to guarantee deterministic tests.
	initialStakeAmount = 42
)

// Tests/debug to implement:
// - Add a tool to easily see what's in the tree (visualize, size, etc...)
// - Benchmark what happens when we add a shit ton of thins into the trie
// - Fuzz a test that results in the same final state but uses different ways to get there
// - Think about:
//       - Thinking about how it can be synched
//       - Playing back several blocks
// - Add TODOs for:
// 		- Atomicity
//      - Bad tests

func TestStateHash_DeterministicStateWhenUpdatingAppStake(t *testing.T) {
	// These hashes were determined manually by running the test, but hardcoded to guarantee
	// that the business logic doesn't change and that they remain deterministic. Anytime the business
	// logic changes, these hashes will need to be updated based on the test output.
	encodedAppHash := []string{
		"a68dbbcddb69355f893000f9ba07dee1d9615cfd1c5db2a41296bff331b4e99d",
		// "1e736e8c94c899f9ac6544744a0f12d2ed29d4e611e7c088f14fc338499fb166",
		// "ce9bf6328228cd8caf138ddc440a8fd512af6a25542c9863562abeb5c793dd82",
	}

	stakeAmount := initialStakeAmount
	for i := 0; i < len(encodedAppHash); i++ {
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
		stakeAmount += 1 // change the stake amount
		stakeAmountStr := strconv.Itoa(stakeAmount)
		err = db.SetAppStakeAmount(addrBz, stakeAmountStr)
		require.NoError(t, err)

		txBz := []byte("a tx, i am, which set the app stake amount to " + stakeAmountStr)
		txResult := indexer.TxRes{
			Tx:            txBz,
			Height:        height,
			Index:         0,
			ResultCode:    0,
			Error:         "TODO",
			SignerAddr:    "TODO",
			RecipientAddr: "TODO",
			MessageType:   "TODO",
		}

		err = db.StoreTransaction(modules.TxResult(&txResult))
		require.NoError(t, err)

		// Update the state hash
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)
		require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

		// Commit the transactions above
		err = db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
		require.NoError(t, err)

		// Retrieve the block
		blockBz, err := testPersistenceMod.GetBlockStore().Get(heightBz)
		require.NoError(t, err)

		// Verify the block contents
		var block types.Block
		err = codec.GetCodec().Unmarshal(blockBz, &block)
		require.NoError(t, err)
		// require.Len(t, block.Transactions, 1)
		// require.Equal(t, txResult.GetTx(), block.Transactions[0])
		require.Equal(t, expectedAppHash, block.Hash) // verify block hash
		if i > 0 {
			require.Equal(t, encodedAppHash[i-1], block.PrevHash) // verify chain chain
		}
	}
}

func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
