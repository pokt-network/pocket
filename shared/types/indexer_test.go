package types

import (
	"bytes"
	"math/rand"
	"testing"
	"time"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestTransactionIndexerIndexAndGetters(t *testing.T) {
	txIndexer, err := NewMemTxIndexer()
	defer txIndexer.Close()
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
	// check indexing/get by hash
	hash, err := txResult.Hash()
	require.NoError(t, err)
	txResultFromHash, err := txIndexer.GetByHash(hash)
	require.NoError(t, err)
	require.True(t, txResultsEqual(t, txResult, txResultFromHash))
	// check indexing/get by height
	txResultsFromHeight, err := txIndexer.GetByHeight(0)
	require.NoError(t, err)
	txResultsFromHeight1, err := txIndexer.GetByHeight(1)
	require.NoError(t, err)
	require.Equal(t, 2, len(txResultsFromHeight))
	require.Equal(t, 1, len(txResultsFromHeight1))
	require.True(t, txResultsEqual(t, txResult, txResultsFromHeight[0]))
	require.True(t, txResultsEqual(t, txResult2, txResultsFromHeight[1]))
	require.True(t, txResultsEqual(t, txResult3, txResultsFromHeight1[0]))
	// check indexing by sender / recipient
	sender := txResult3.GetSigner()
	recipient := txResult3.GetRecipient()
	require.NoError(t, err)
	txResultsFromSender, err := txIndexer.GetBySender(sender, false)
	require.NoError(t, err)
	txResultsFromRecipient, err := txIndexer.GetByRecipient(recipient, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(txResultsFromSender))
	require.Equal(t, 1, len(txResultsFromRecipient))
	require.True(t, txResultsEqual(t, txResult3, txResultsFromSender[0]))
	require.True(t, txResultsEqual(t, txResult3, txResultsFromRecipient[0]))
	// ensure it's not indexed elsewhere
	txResultsFromSenderBad, err := txIndexer.GetBySender(recipient, false)
	require.NoError(t, err)
	txResultsFromRecipientBad, err := txIndexer.GetByRecipient(sender, false)
	require.NoError(t, err)
	require.Equal(t, 0, len(txResultsFromSenderBad))
	require.Equal(t, 0, len(txResultsFromRecipientBad))
}

func txResultsEqual(t *testing.T, txR1, txR2 TxResult) bool {
	bz, err := txR1.Bytes()
	require.NoError(t, err)
	bz2, err := txR2.Bytes()
	require.NoError(t, err)
	return bytes.Equal(bz, bz2)
}

// utility helpers

func NewTestingTransactionResult(t *testing.T, height, index int) TxResult {
	testingTransaction := randLetterBytes()
	resultCode, err := randomErr()
	return &DefaultTxResult{
		Tx:          testingTransaction,
		Height:      int64(height),
		Index:       int32(index),
		ResultCode:  resultCode,
		Error:       err,
		Signer:      randomAddress(t),
		Recipient:   randomAddress(t),
		MessageType: randomMessageType(),
	}
}

func randomMessageType() string {
	// TODO(andrew): Add an enum for the different message types
	msgTypes := []string{"send", "stake", "unstake", "editStake", "unjail"}
	return msgTypes[rand.Intn(len(msgTypes))]
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
	if rand.Intn(4) == 1 {
		code = int32(rand.Intn(len(errors)))
		err = errors[code]
	}
	return
}

// Generates a random alphanumeric sequence of exactly 50 characters
func randLetterBytes() []byte {
	rand.Seed(int64(time.Now().Nanosecond()))
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, rand.Intn(50))
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}
