package types

import (
	"bytes"
	"testing"

	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestTxFIFOMempool(t *testing.T) {
	// pointers in this struct are meant to allow the tester to skip either `initialTxs` or `actions` depending on what they are testing for
	type args struct {
		maxTxBytes uint64
		maxTxs     uint32
		initialTxs *[][]byte
		actions    *[]func(*txFIFOMempool)
	}
	tests := []struct {
		name      string
		args      args
		wantItems [][]byte
	}{
		{
			// maxTxBytes is 1, we are starting with a message of size _at least_ 10 and adding another one of _at least_ 10
			// we expect the first one to be removed
			name: "should not allow adding a transaction because it would result in an overflow (in bytes)",
			args: args{
				maxTxBytes: 1,
				maxTxs:     10,
				initialTxs: &[][]byte{messageSendFactory(10)},
			},
			wantItems: [][]byte{},
		},
		{
			// maxTxBytes is 20, we are starting with a message of size _at least_ 10 and adding another one of _at least_ 10
			// we expect the first one to be removed
			name: "should correcly implement a FIFO set by removing the oldest message when at capacity (in bytes)",
			args: args{
				maxTxBytes: 20,
				maxTxs:     10,
				initialTxs: &[][]byte{messageSendFactory(10)},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						err := txFIFOMempool.AddTx(messageSendFactory(10))
						require.Contains(t, err.Error(), "already found in")
					},
				},
			},
			wantItems: [][]byte{messageSendFactory(10)},
		},
		{
			name: "should correctly implement a FIFO set by removing the oldest message when at capacity (in number of txs)",
			args: args{
				maxTxBytes: 100,
				maxTxs:     2,
				initialTxs: &[][]byte{messageSendFactory(10)},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						err := txFIFOMempool.AddTx(messageSendFactory(9))
						require.NoError(t, err)
						err = txFIFOMempool.AddTx(messageSendFactory(8))
						require.NoError(t, err)
					},
				},
			},
			wantItems: [][]byte{messageSendFactory(9), messageSendFactory(8)},
		},
		{
			name: "looping over the mempool and Pop()ing all the transactions should return an empty mempool",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
					messageSendFactory(8),
					messageSendFactory(8),
					messageSendFactory(3),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						for !txFIFOMempool.IsEmpty() {
							_, err := txFIFOMempool.PopTx()
							require.NoError(t, err)
						}
					},
				},
			},
			wantItems: [][]byte{},
		},
		{
			name: "Pop()ing a message should return it correctly",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						tx, err := txFIFOMempool.PopTx()
						require.NoError(t, err, "unexpected error while popping transaction")
						expectedTx := messageSendFactory(9)

						require.True(t, bytes.Equal(tx, expectedTx), "mismatching transaction")

					},
				},
			},
			wantItems: [][]byte{messageSendFactory(10)},
		},
		{
			name: "Pop()ing a message should return an error if the mempool is empty",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						_, err := txFIFOMempool.PopTx()
						require.Error(t, err, "expected error while popping transaction")
					},
				},
			},
			wantItems: [][]byte{},
		},
		{
			name: "Clear should empty the mempool",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						txFIFOMempool.Clear()
					},
				},
			},
			wantItems: [][]byte{},
		},
		{
			name: "TxCount should return 0 for an empty mempool",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						require.Equal(t, 0, int(txFIFOMempool.TxCount()), "mismatching TxCount")
					},
				},
			},
			wantItems: [][]byte{},
		},
		{
			name: "TxCount should return the correct amount of transactions",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						require.Equal(t, 2, int(txFIFOMempool.TxCount()), "mismatching TxCount")
					},
				},
			},
			wantItems: [][]byte{messageSendFactory(9), messageSendFactory(10)},
		},
		{
			name: "TxsBytesTotal should return 0 for an empty mempool",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						require.Equal(t, 0, int(txFIFOMempool.TxsBytesTotal()), "mismatching TxsBytesTotal")
					},
				},
			},
			wantItems: [][]byte{},
		},
		{
			name: "TxsBytesTotal should return the correct amount of bytes",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						require.Equal(t, len(messageSendFactory(9))+len(messageSendFactory(10)), int(txFIFOMempool.TxsBytesTotal()), "mismatching TxsBytesTotal")
					},
				},
			},
			wantItems: [][]byte{messageSendFactory(9), messageSendFactory(10)},
		},
		{
			name: "RemoveTx should remove a transaction from the mempool",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						err := txFIFOMempool.RemoveTx(messageSendFactory(9))
						require.NoError(t, err)
					},
				},
			},
			wantItems: [][]byte{
				messageSendFactory(10),
			},
		},
		{
			name: "Contains should return true for a transaction that is in the mempool and false viceversa",
			args: args{
				maxTxBytes: 1000,
				maxTxs:     100,
				initialTxs: &[][]byte{
					messageSendFactory(9),
					messageSendFactory(10),
				},
				actions: &[]func(*txFIFOMempool){
					func(txFIFOMempool *txFIFOMempool) {
						txHashOK := crypto.GetHashStringFromBytes(messageSendFactory(9))
						require.True(t, txFIFOMempool.Contains(txHashOK), "mismatching Contains")
						txHashKO := crypto.GetHashStringFromBytes(messageSendFactory(19))
						require.False(t, txFIFOMempool.Contains(txHashKO), "mismatching Contains")
					},
				},
			},
			wantItems: [][]byte{
				messageSendFactory(9),
				messageSendFactory(10),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFIFOMempool := NewTxFIFOMempool(tt.args.maxTxBytes, tt.args.maxTxs)

			if tt.args.initialTxs != nil {
				for _, item := range *tt.args.initialTxs {
					txFIFOMempool.AddTx(item) //nolint:errcheck // Do not error check
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(txFIFOMempool)
				}
			}

			require.Equal(t, len(tt.wantItems), int(txFIFOMempool.TxCount()), "mismatching TxCount (capacity filled with elements)")

			for _, wantItem := range tt.wantItems {
				wantHash := crypto.GetHashStringFromBytes(wantItem)
				require.True(t, txFIFOMempool.Contains(wantHash), "missing element")
				gotItem, err := txFIFOMempool.PopTx()
				require.NoError(t, err, "unexpected error while popping element")
				require.Equal(t, wantItem, gotItem, "mismatching element")
			}

			if txFIFOMempool.TxCount() == 0 {
				require.True(t, txFIFOMempool.IsEmpty(), "IsEmpty should return true when TxCount() is 0")
			}
		})
	}
}

// messageSendFactory returns a MessageSend message marshalled in []byte format
//
// it is used in tests to make sure that the mempool is correctly limiting the size of the transactions it stores
//
// IMPORTANT: the size of the transaction is not the size of the whole message on the wire, there's obviously some overhead in the protobuf
func messageSendFactory(amountByteSize int) []byte {
	amountBz := make([]byte, amountByteSize)
	for i := range amountBz {
		amountBz[i] = 1
	}

	message := &MessageSend{
		FromAddress: []byte{},
		ToAddress:   []byte{},
		Amount:      string(amountBz),
	}
	messageBz, _ := codec.GetCodec().Marshal(message)
	return messageBz
}
