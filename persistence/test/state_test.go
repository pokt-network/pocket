package test

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/pokt-network/pocket/persistence/indexer"
	"github.com/pokt-network/pocket/shared/codec"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/shared/utils"
	"github.com/stretchr/testify/require"
)

const (
	proposerBytesSize   = 10
	quorumCertBytesSize = 10

	// This value is arbitrarily selected, but needs to be a constant to guarantee deterministic tests.
	initialStakeAmount = 42
)

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
	hash       string
	proposer   []byte
	quorumCert []byte
}

func TestStateHash_DeterministicStateWhenUpdatingAppStake(t *testing.T) {
	// These hashes were determined manually by running the test, but hardcoded to guarantee
	// that the business logic doesn't change and that they remain deterministic. Anytime the business
	// logic changes, these hashes will need to be updated based on the test output.
	// TODO: Add an explicit updateSnapshots flag to the test to make this more clear.
	stateHashes := []string{
		"4e781020ad767db9bedc9d64bb30c69a5f1d94a57049c74a6f6e26d8ff3046bc",
		"01786ad7ebd82641b5fa040798edf9eaebc07ac34bdcbd7e089c75ef09f730ee",
		"fa3682d967de200686d0567ff2ab3568372decd91a7a94112d0c1ce493eee2c0",
	}

	stakeAmount := initialStakeAmount
	for i := 0; i < len(stateHashes); i++ {
		// Get the context at the new height and retrieve one of the apps
		height := int64(i + 1)
		heightBz := utils.HeightToBytes(uint64(height))
		expectedStateHash := stateHashes[i]

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

		err = db.IndexTransaction(modules.TxResult(&txResult))
		require.NoError(t, err)

		// Update the state hash
		stateHash, err := db.ComputeStateHash()
		require.NoError(t, err)
		require.Equal(t, expectedStateHash, stateHash)

		// Commit the transactions above
		proposer := []byte("placeholderProposer")
		quorumCert := []byte("placeholderQuorumCert")

		err = db.Commit(proposer, quorumCert)
		require.NoError(t, err)

		// Retrieve the block
		blockBz, err := testPersistenceMod.GetBlockStore().Get(heightBz)
		require.NoError(t, err)

		// Verify the block contents
		var block coreTypes.Block
		err = codec.GetCodec().Unmarshal(blockBz, &block)
		require.NoError(t, err)
		require.Equal(t, expectedStateHash, block.BlockHeader.StateHash) // verify block hash
		if i > 0 {
			require.Equal(t, stateHashes[i-1], block.BlockHeader.PrevStateHash) // verify chain chain
		}
	}
}

// This unit test generates random transactions and creates random state changes, but checks
// that replaying them will result in the same state hash, guaranteeing the integrity of the
// state hash.
func TestStateHash_ReplayingRandomTransactionsIsDeterministic(t *testing.T) {
	testCases := []struct {
		numHeights      int64
		numTxsPerHeight int
		numOpsPerTx     int
		numReplays      int
	}{
		{1, 2, 1, 3},
		{10, 2, 5, 5},
	}

	for _, testCase := range testCases {
		numHeights := testCase.numHeights
		numTxsPerHeight := testCase.numTxsPerHeight
		numOpsPerTx := testCase.numOpsPerTx
		numReplays := testCase.numReplays

		t.Run(fmt.Sprintf("ReplayingRandomTransactionsIsDeterministic(%d;%d,%d,%d", numHeights, numTxsPerHeight, numOpsPerTx, numReplays), func(t *testing.T) {
			t.Cleanup(clearAllState)
			clearAllState()

			replayableBlocks := make([]*TestReplayableBlock, numHeights)

			for height := int64(0); height < int64(numHeights); height++ {
				db := NewTestPostgresContext(t, height)
				replayableTxs := make([]*TestReplayableTransaction, numTxsPerHeight)

				for txIdx := 0; txIdx < numTxsPerHeight; txIdx++ {
					replayableOps := make([]*TestReplayableOperation, numOpsPerTx)

					for opIdx := 0; opIdx < numOpsPerTx; opIdx++ {
						methodName, args, err := callRandomDatabaseModifierFunc(db, true)
						require.NoError(t, err)

						replayableOps[opIdx] = &TestReplayableOperation{
							methodName: methodName,
							args:       args,
						}
					}

					txResult := modules.TxResult(getRandomTxResult(height))
					err := db.IndexTransaction(txResult)
					require.NoError(t, err)

					replayableTxs[txIdx] = &TestReplayableTransaction{
						operations: replayableOps,
						txResult:   txResult,
					}
				}

				stateHash, err := db.ComputeStateHash()
				require.NoError(t, err)

				proposer := getRandomBytes(proposerBytesSize)
				quorumCert := getRandomBytes(quorumCertBytesSize)

				err = db.Commit(proposer, quorumCert)
				require.NoError(t, err)

				replayableBlocks[height] = &TestReplayableBlock{
					height:     height,
					txs:        replayableTxs,
					hash:       stateHash,
					proposer:   proposer,
					quorumCert: quorumCert,
				}
			}

			for i := 0; i < numReplays; i++ {
				t.Run("verify block", func(t *testing.T) {
					verifyReplayableBlocks(t, replayableBlocks)
				})
			}
		})
	}
}

func TestStateHash_TreeUpdatesAreIdempotent(t *testing.T) {
	// ADDTEST(#361): Create an issue dedicated to increasing the test coverage for state hashes
}

func TestStateHash_TreeUpdatesNegativeTestCase(t *testing.T) {
	// ADDTEST(#361): Create an issue dedicated to increasing the test coverage for state hashes
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
			require.NoError(t, db.IndexTransaction(tx.txResult))
		}

		stateHash, err := db.ComputeStateHash()
		require.NoError(t, err)
		require.Equal(t, block.hash, stateHash)

		err = db.Commit(block.proposer, block.quorumCert)
		require.NoError(t, err)
	}
}
