package types

import (
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestMempool(t *testing.T) {
	type args struct {
		maxTransactionBytes uint64
		maxTransactions     uint32
		initialElements     *[][]byte
		actions             *[]func(*fIFOMempool)
	}
	tests := []struct {
		name      string
		args      args
		wantItems [][]byte
	}{
		//
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFifoMempool := NewTxFIFOMempool(tt.args.maxTransactionBytes, tt.args.maxTransactions)

			if tt.args.initialElements != nil {
				for _, item := range *tt.args.initialElements {
					txFifoMempool.AddTransaction(item)
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(txFifoMempool)
				}
			}

			require.Equal(t, len(tt.wantItems), txFifoMempool.Size(), "mismatching Size (capacity filled with elements)")

			for _, wantItem := range tt.wantItems {
				wantHash := crypto.GetHashStringFromBytes(wantItem)
				require.True(t, txFifoMempool.Contains(wantHash), "missing element")
				gotItem, err := txFifoMempool.PopTransaction()
				require.NoError(t, err, "unexpected error while popping element")
				require.Equal(t, wantItem, gotItem, "mismatching element")
			}

			if txFifoMempool.Size() == 0 {
				require.True(t, txFifoMempool.IsEmpty(), "IsEmpty should return true when Len() is 0")
			}
		})
	}
}
