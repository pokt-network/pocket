package indexer

import (
	"log"
	"math/rand"
	"testing"
	"time"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
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
		// the new idxTx is appended to the # of results currently at that height
		// this means the 'index' of the transaction within the block is len(heightResults)
		heightIndex := len(heightResult)
		// create new testing tx
		tx := NewTestingIndexedTransaction(t, int(height), heightIndex)
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
		hash := tx.HashFromBytes(tx.GetTx())
		require.NoError(t, indexer.Index(tx))
		switch op {
		case "GetByHash":
			idxTx, err := indexer.GetByHash(hash)
			require.NoError(t, err)
			requireIdxTxsEqual(t, tx, idxTx)
		case "GetByHeight":
			idxTx, err := indexer.GetByHeight(height, isDescending)
			require.NoError(t, err)
			if isDescending {
				requireIdxTxsEqual(t, tx, idxTx[0])
			} else {
				requireIdxTxsEqual(t, tx, idxTx[heightIndex])
			}
		case "GetBySender":
			idxTx, err := indexer.GetBySender(sender, true)
			require.NoError(t, err)
			requireIdxTxsEqual(t, tx, idxTx[senderIndex])
		case "GetByRecipient":
			idxTx, err := indexer.GetByRecipient(recipient, true)
			require.NoError(t, err)
			requireIdxTxsEqual(t, tx, idxTx[recipientIndex])
		default:
			t.Errorf("Unexpected operation fuzzing operation %s", op)
		}
	})
}

func TestGetByHash(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup 2 transactions
	idxTx := NewTestingIndexedTransaction(t, 0, 0)
	require.NoError(t, err)
	idxTx2 := NewTestingIndexedTransaction(t, 0, 1)
	require.NoError(t, err)
	// index 2 transactions
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)
	err = txIndexer.Index(idxTx2)
	require.NoError(t, err)
	// check indexing/get by hash
	hash := idxTx.HashFromBytes(idxTx.GetTx())
	idxTxFromHash, err := txIndexer.GetByHash(hash)
	require.NoError(t, err)
	requireIdxTxsEqual(t, idxTx, idxTxFromHash)
	// check indexing/get by hash 2
	hash2 := idxTx2.HashFromBytes(idxTx2.GetTx())
	idxTxFromHash2, err := txIndexer.GetByHash(hash2)
	require.NoError(t, err)
	requireIdxTxsEqual(t, idxTx2, idxTxFromHash2)
}

func TestGetByHeight(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup 3 transactions
	idxTx := NewTestingIndexedTransaction(t, 0, 0)
	require.NoError(t, err)
	idxTx2 := NewTestingIndexedTransaction(t, 0, 1)
	require.NoError(t, err)
	idxTx3 := NewTestingIndexedTransaction(t, 1, 0)
	require.NoError(t, err)
	// index all 3 transactions
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)
	err = txIndexer.Index(idxTx2)
	require.NoError(t, err)
	err = txIndexer.Index(idxTx3)
	require.NoError(t, err)
	// check indexing/get by height
	idxTxsFromHeight, err := txIndexer.GetByHeight(0, false)
	require.NoError(t, err)
	idxTxsFromHeight1, err := txIndexer.GetByHeight(1, false)
	require.NoError(t, err)
	expectedNumOfTxsAtHeight0 := 2
	expectedNumOfTxsAtHeight1 := 1
	require.Equal(t, expectedNumOfTxsAtHeight0, len(idxTxsFromHeight))
	require.Equal(t, expectedNumOfTxsAtHeight1, len(idxTxsFromHeight1))
}

func TestGetBySender(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup transaction
	idxTx := NewTestingIndexedTransaction(t, 1, 0)
	require.NoError(t, err)
	// index transaction
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)
	// check indexing by sender / recipient
	sender := idxTx.GetSignerAddr()
	require.NoError(t, err)
	idxTxsFromSender, err := txIndexer.GetBySender(sender, false)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 1, len(idxTxsFromSender))
	requireIdxTxsEqual(t, idxTx, idxTxsFromSender[0])
	// ensure it's not indexed elsewhere
	idxTxsFromRecipientBad, err := txIndexer.GetByRecipient(sender, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(idxTxsFromRecipientBad))
}

func TestGetByRecipient(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer closeIndexer(t, txIndexer)
	// setup tx
	idxTx := NewTestingIndexedTransaction(t, 1, 0)
	require.NoError(t, err)
	// index transactions
	err = txIndexer.Index(idxTx)
	require.NoError(t, err)
	recipient := idxTx.GetRecipientAddr()
	require.NoError(t, err)
	idxTxsFromRecipient, err := txIndexer.GetByRecipient(recipient, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(idxTxsFromRecipient))
	requireIdxTxsEqual(t, idxTx, idxTxsFromRecipient[0])
	// ensure it's not indexed elsewhere
	idxTxsFromSenderBad, err := txIndexer.GetBySender(recipient, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(idxTxsFromSenderBad))
}

func requireIdxTxsEqual(t *testing.T, txR1, txR2 *coreTypes.IndexedTransaction) {
	bz, err := txR1.Bytes()
	require.NoError(t, err)
	bz2, err := txR2.Bytes()
	require.NoError(t, err)
	require.Equal(t, bz, bz2)
}

// utility helpers

func NewTestingIndexedTransaction(t *testing.T, height, index int) *coreTypes.IndexedTransaction {
	testingTransaction := randLetterBytes()
	resultCode, err := randomErr()
	return &coreTypes.IndexedTransaction{
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
