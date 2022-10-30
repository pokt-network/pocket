package test

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
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
		"3078d5c1dc45f3f76f5daef585097c4029e6d5837e2d6bc2bfb8c2c3d3766e4c",
		"021b96cd367323c1d97832580d47ad3e54bfe79141aa507b7d60e3b0ddd107d6",
		"70db812fb2b397252fb49b189d405d6e001bc7e2452914ca5c231af1166f2675",
	}

	stakeAmount := initialStakeAmount
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

		// txResult := mockTxResult(t, height, txBz)
		err = db.StoreTransaction(modules.TxResult(&txResult))
		require.NoError(t, err)

		// Update & commit the state hash
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)
		require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

		err = db.Commit([]byte("TODOproposer"), []byte("TODOquorumCert"))
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

func getTxBytes(seed, size int64) []byte {
	rand.Seed(seed)
	bz := make([]byte, size)
	rand.Read(bz)
	return bz
}

func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
