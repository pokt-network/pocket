package types

import (
	"testing"

	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
)

func TestMempool(t *testing.T) {
	type args struct {
		maxTxBytes uint64
		maxTxs     uint32
		initialTxs *[][]byte
		actions    *[]func(*fIFOMempool)
	}
	tests := []struct {
		name      string
		args      args
		wantItems [][]byte
	}{
		//TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFIFOMempool := NewTxFIFOMempool(tt.args.maxTxBytes, tt.args.maxTxs)

			if tt.args.initialTxs != nil {
				for _, item := range *tt.args.initialTxs {
					txFIFOMempool.AddTx(item)
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(txFIFOMempool)
				}
			}

			require.Equal(t, len(tt.wantItems), txFIFOMempool.TxCount(), "mismatching TxCount (capacity filled with elements)")

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
