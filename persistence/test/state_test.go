package test

import (
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"strconv"
	"testing"

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

		// txBz := []byte("a tx, i am, which set the app stake amount to " + stakeAmountStr)
		// txResult := indexer.TxRes{
		// 	Tx:            txBz,
		// 	Height:        height,
		// 	Index:         0,
		// 	ResultCode:    0,
		// 	Error:         "TODO",
		// 	SignerAddr:    "TODO",
		// 	RecipientAddr: "TODO",
		// 	MessageType:   "TODO",
		// }

		// err = db.StoreTransaction(modules.TxResult(&txResult))
		// require.NoError(t, err)

		// Update the state hash
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)
		require.Equal(t, expectedAppHash, hex.EncodeToString(appHash))

		// Commit the transactions above
		err = db.Commit([]byte("TODOquorumCert"))
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

type TestReplayableOperation struct {
	methodName string
	args       []reflect.Value
}
type TestReplayableTransaction struct {
	operations []*TestReplayableOperation
	txResult   modules.TxResult
}

type TestReplayableBlock struct {
	height     int64
	txs        []*TestReplayableTransaction
	hash       []byte
	proposer   []byte
	quorumCert []byte
}

func TestStateHash_RandomButDeterministic(t *testing.T) {
	t.Cleanup(clearAllState)
	clearAllState()

	numHeights := 10
	numTxsPerHeight := 2
	numOpsPerTx := 5
	numReplays := 5

	replayableBlocks := make([]*TestReplayableBlock, numHeights)
	for height := int64(0); height < int64(numHeights); height++ {
		db := NewTestPostgresContext(t, height)
		replayableTxs := make([]*TestReplayableTransaction, numTxsPerHeight)
		for txIdx := 0; txIdx < numTxsPerHeight; txIdx++ {
			replayableOps := make([]*TestReplayableOperation, numOpsPerTx)
			for opIdx := 0; opIdx < numOpsPerTx; opIdx++ {
				methodName, args, err := callRandomDatabaseModifierFunc(db, height, true)
				require.NoError(t, err)
				replayableOps[opIdx] = &TestReplayableOperation{
					methodName: methodName,
					args:       args,
				}
			}
			txResult := modules.TxResult(getRandomTxResult(height))
			// err := db.StoreTransaction(txResult)
			// require.NoError(t, err)

			replayableTxs[txIdx] = &TestReplayableTransaction{
				operations: replayableOps,
				txResult:   txResult,
			}
		}
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)

		proposer := getRandomBytes(10)
		quorumCert := getRandomBytes(10)
		err = db.Commit(quorumCert)
		require.NoError(t, err)

		replayableBlocks[height] = &TestReplayableBlock{
			height:     height,
			txs:        replayableTxs,
			hash:       appHash,
			proposer:   proposer,
			quorumCert: quorumCert,
		}
	}

	for i := 0; i < numReplays; i++ {
		t.Run("verify block", func(t *testing.T) {
			verifyReplayableBlocks(t, replayableBlocks)
		})
	}
}

func verifyReplayableBlocks(t *testing.T, replayableBlocks []*TestReplayableBlock) {
	t.Cleanup(clearAllState)
	clearAllState()

	for _, block := range replayableBlocks {
		db := NewTestPostgresContext(t, block.height)
		for _, tx := range block.txs {
			for _, op := range tx.operations {
				require.Nil(t, reflect.ValueOf(db).MethodByName(op.methodName).Call(op.args)[0].Interface())
			}
			// require.NoError(t, db.StoreTransaction(tx.txResult))
		}
		appHash, err := db.UpdateAppHash()
		require.NoError(t, err)
		require.Equal(t, block.hash, appHash)

		err = db.Commit(block.quorumCert)
		require.NoError(t, err)
	}
}

func heightToBytes(height int64) []byte {
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(height))
	return heightBytes
}
