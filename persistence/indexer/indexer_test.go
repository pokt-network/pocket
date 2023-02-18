package indexer

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/pokt-network/pocket/shared/crypto"
	shared "github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"
)

func FuzzTxIndexer(f *testing.F) {
	operations := []string{
		"GetByHash",
		"GetByHeight",
		"GetBySender",
		"GetByRecipient",
	}
	numOperationTypes := len(operations)
	numOperations := 100
	for i := 0; i < numOperations; i++ {
		f.Add(operations[rand.Intn(numOperationTypes)]) //nolint:gosec // G404 - Weak random source is okay in unit tests
	}
	indexer, err := NewMemTxIndexer()
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *testing.F, indexer TxIndexer) {
		err := indexer.Close()
		require.NoError(f, err)
	}(f, indexer)

	f.Fuzz(func(t *testing.T, op string) {
		// seed random
		rand.Seed(int64(time.Now().Nanosecond())) //nolint:staticcheck // G404 - Weak random source is okay here
		// set height ordering to descending 50% of time
		isDescending := rand.Intn(2) == 0 //nolint:gosec // G404 - Weak random source is okay in unit tests
		// select a height 0 - 9 to index
		height := int64(rand.Intn(10)) //nolint:gosec // G404 - Weak random source is okay in unit tests
		// get index
		heightResult, err := indexer.GetByHeight(height, isDescending)
		require.NoError(t, err)
		// the new txResult is appended to the # of results currently at that height
		// this means the 'index' of the transaction within the block is len(heightResults)
		heightIndex := len(heightResult)
		// create new testing tx
		tx := NewTestingTransactionResult(t, int(height), heightIndex)
		// by sender
		sender := tx.GetSignerAddr()
		senderResult, err := indexer.GetBySender(sender, true)
		require.NoError(t, err)
		senderIndex := len(senderResult)
		// by recipient
		recipient := tx.GetRecipientAddr()
		recipientResult, err := indexer.GetByRecipient(recipient, true)
		require.NoError(t, err)
		recipientIndex := len(recipientResult)
		hash, err := tx.Hash()
		require.NoError(t, err)
		require.NoError(t, indexer.Index(tx))
		switch op {
		case "GetByHash":
			txResult, err := indexer.GetByHash(hash)
			require.NoError(t, err)
			requireTxResultsEqual(t, tx, txResult)
		case "GetByHeight":
			txResult, err := indexer.GetByHeight(height, isDescending)
			require.NoError(t, err)
			if isDescending {
				requireTxResultsEqual(t, tx, txResult[0])
			} else {
				requireTxResultsEqual(t, tx, txResult[heightIndex])
			}
		case "GetBySender":
			txResult, err := indexer.GetBySender(sender, true)
			require.NoError(t, err)
			requireTxResultsEqual(t, tx, txResult[senderIndex])
		case "GetByRecipient":
			txResult, err := indexer.GetByRecipient(recipient, true)
			require.NoError(t, err)
			requireTxResultsEqual(t, tx, txResult[recipientIndex])
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}
	})
}

func TestGetByHash(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup 2 transactions
	txResult := NewTestingTransactionResult(t, 0, 0)
	require.NoError(t, err)
	txResult2 := NewTestingTransactionResult(t, 0, 1)
	require.NoError(t, err)
	// index 2 transactions
	err = txIndexer.Index(txResult)
	require.NoError(t, err)
	err = txIndexer.Index(txResult2)
	require.NoError(t, err)
	// check indexing/get by hash
	hash, err := txResult.Hash()
	require.NoError(t, err)
	txResultFromHash, err := txIndexer.GetByHash(hash)
	require.NoError(t, err)
	requireTxResultsEqual(t, txResult, txResultFromHash)
	// check indexing/get by hash 2
	hash2, err := txResult2.Hash()
	require.NoError(t, err)
	txResultFromHash2, err := txIndexer.GetByHash(hash2)
	require.NoError(t, err)
	requireTxResultsEqual(t, txResult2, txResultFromHash2)
}

func TestGetByHeight(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup 3 transactions
	txResult := NewTestingTransactionResult(t, 0, 0)
	require.NoError(t, err)
	txResult2 := NewTestingTransactionResult(t, 0, 1)
	require.NoError(t, err)
	txResult3 := NewTestingTransactionResult(t, 1, 0)
	require.NoError(t, err)
	// index all 3 transactions
	err = txIndexer.Index(txResult)
	require.NoError(t, err)
	err = txIndexer.Index(txResult2)
	require.NoError(t, err)
	err = txIndexer.Index(txResult3)
	require.NoError(t, err)
	// check indexing/get by height
	txResultsFromHeight, err := txIndexer.GetByHeight(0, false)
	require.NoError(t, err)
	txResultsFromHeight1, err := txIndexer.GetByHeight(1, false)
	require.NoError(t, err)
	expectedNumOfTxsAtHeight0 := 2
	expectedNumOfTxsAtHeight1 := 1
	require.Equal(t, expectedNumOfTxsAtHeight0, len(txResultsFromHeight))
	require.Equal(t, expectedNumOfTxsAtHeight1, len(txResultsFromHeight1))
}

func TestGetBySender(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup transaction
	txResult := NewTestingTransactionResult(t, 1, 0)
	require.NoError(t, err)
	// index transaction
	err = txIndexer.Index(txResult)
	require.NoError(t, err)
	// check indexing by sender / recipient
	sender := txResult.GetSignerAddr()
	require.NoError(t, err)
	txResultsFromSender, err := txIndexer.GetBySender(sender, false)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 1, len(txResultsFromSender))
	requireTxResultsEqual(t, txResult, txResultsFromSender[0])
	// ensure it's not indexed elsewhere
	txResultsFromRecipientBad, err := txIndexer.GetByRecipient(sender, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(txResultsFromRecipientBad))
}

func TestGetByRecipient(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup tx
	txResult := NewTestingTransactionResult(t, 1, 0)
	require.NoError(t, err)
	// index transactions
	err = txIndexer.Index(txResult)
	require.NoError(t, err)
	recipient := txResult.GetRecipientAddr()
	require.NoError(t, err)
	txResultsFromRecipient, err := txIndexer.GetByRecipient(recipient, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(txResultsFromRecipient))
	requireTxResultsEqual(t, txResult, txResultsFromRecipient[0])
	// ensure it's not indexed elsewhere
	txResultsFromSenderBad, err := txIndexer.GetBySender(recipient, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(txResultsFromSenderBad))
}

func requireTxResultsEqual(t *testing.T, txR1, txR2 shared.TxResult) {
	bz, err := txR1.Bytes()
	require.NoError(t, err)
	bz2, err := txR2.Bytes()
	require.NoError(t, err)
	require.Equal(t, bz, bz2)
}

// utility helpers

func NewTestingTransactionResult(t *testing.T, height, index int) shared.TxResult {
	testingTransaction := randLetterBytes()
	resultCode, err := randomErr()
	return &TxRes{
		Tx:            testingTransaction,
		Height:        int64(height),
		Index:         int32(index),
		ResultCode:    resultCode,
		Error:         err,
		SignerAddr:    randomAddress(t),
		RecipientAddr: randomAddress(t),
		MessageType:   randomMessageType(),
	}
}

type MessageType int

const (
	SendMessage MessageType = iota + 1
	StakeMessage
	UnstakeMessage
	EditStakeMessage
	UnjailMessage
)

var msgTypes = []MessageType{SendMessage, StakeMessage, UnstakeMessage, EditStakeMessage, UnjailMessage}

func (mt MessageType) String() string {
	switch mt {
	case SendMessage:
		return "send"
	case StakeMessage:
		return "stake"
	case UnstakeMessage:
		return "unstake"
	case EditStakeMessage:
		return "editStake"
	case UnjailMessage:
		return "unjail"
	}
	return "unrecognized message type"
}

func randomMessageType() string {
	return msgTypes[rand.Intn(len(msgTypes))].String() //nolint:gosec // G404 - Weak random source is okay in unit tests
}

func randomAddress(t *testing.T) string {
	add, err := crypto.GenerateAddress()
	require.NoError(t, err)
	return add.String()
}

// Returns an error 25% of the time
func randomErr() (code int32, err string) {
	errors := []string{"insufficient funds", "address not valid", "invalid signature"}
	code = int32(0)
	err = ""
	//nolint:gosec // G404 - Weak random source is okay in unit tests
	if rand.Intn(4) == 1 {
		code = int32(rand.Intn(len(errors)))
		err = errors[code]
	}
	return
}

// Generates a random alphanumeric sequence of exactly 50 characters
//
//nolint:gosec // G404 - Weak random source is okay in unit tests
func randLetterBytes() []byte {
	randBytes := make([]byte, 50)
	rand.Read(randBytes) //nolint:staticcheck // G404 - Weak random source is okay here
	return randBytes
}

func closeIndexer(t *testing.T, indexer TxIndexer) {
	err := indexer.Close()
	require.NoError(t, err)
}
