package consensus

import (
	"testing"

	"github.com/golang/protobuf/proto"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

// hotstuffMessageFactory returns a hotstuff message with a block with a single transaction of the given size
//
// it is used in tests to make sure that the mempool is correctly limiting the size of the transactions it stores
//
// IMPORTANT: the size of the transaction is not the size of the whole message on the wire, there's obviously some overhead in the protobuf
func hotstuffMessageFactory(size int) *typesCons.HotstuffMessage {
	tx := make([]byte, size)
	for i := range tx {
		tx[i] = 1
	}
	return &typesCons.HotstuffMessage{
		Block: &types.Block{
			Transactions: [][]byte{tx},
		},
	}
}

func TestMempool(t *testing.T) {
	type args struct {
		maxTransactionBytes uint64
		initialElements     *[]*typesCons.HotstuffMessage
		actions             *[]func(*hotstuffFIFOMempool)
	}
	tests := []struct {
		name      string
		args      args
		wantItems []*typesCons.HotstuffMessage
	}{
		{
			// maxTransactionBytes is 20, we are starting with a message of size _at least_ 10 and adding another one of _at least_ 10
			// we expect the first one to be removed
			name: "should correcly implement a FIFO set by removing the oldest element when at capacity",
			args: args{
				maxTransactionBytes: 20,
				initialElements:     &[]*typesCons.HotstuffMessage{hotstuffMessageFactory(10)},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						txFifoMempool.Push(hotstuffMessageFactory(10))
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{hotstuffMessageFactory(10)},
		},
		{
			name: "looping over the mempool and Pop()ing all the elements should return an empty mempool",
			args: args{
				maxTransactionBytes: 1000,
				initialElements: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						for !txFifoMempool.IsEmpty() {
							txFifoMempool.Pop()
						}
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Pop()ing an element should return it correctly",
			args: args{
				maxTransactionBytes: 1000,
				initialElements:     &[]*typesCons.HotstuffMessage{hotstuffMessageFactory(19)},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						msg, err := txFifoMempool.Pop()
						require.NoError(t, err, "unexpected error while popping element")
						expectedMsg := hotstuffMessageFactory(19)

						require.True(t, proto.Equal(msg, expectedMsg), "mismatching element")
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Pop()ing an element from an empty mempool should return an error",
			args: args{
				maxTransactionBytes: 1000,
				initialElements:     &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						_, err := txFifoMempool.Pop()
						require.Error(t, err, "expected error while popping element")
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Push()ing an element to an empty mempool should return the correct size",
			args: args{
				maxTransactionBytes: 1000,
				initialElements:     &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						txFifoMempool.Push(hotstuffMessageFactory(19))
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{hotstuffMessageFactory(19)},
		},
		{
			name: "Clear should empty the mempool",
			args: args{
				maxTransactionBytes: 1000,
				initialElements: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						txFifoMempool.Clear()
						require.Equal(t, 0, txFifoMempool.Size(), "mismatching size")
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{},
		},
		{
			name: "GetAll should return all the elements in the mempool, in the order they were added in the first place",
			args: args{
				maxTransactionBytes: 1000,
				initialElements: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						expectedMsgs := []*typesCons.HotstuffMessage{
							hotstuffMessageFactory(3),
							hotstuffMessageFactory(19),
							hotstuffMessageFactory(83),
							hotstuffMessageFactory(28),
						}

						for i, msg := range txFifoMempool.GetAll() {
							require.True(t, proto.Equal(msg, expectedMsgs[i]), "mismatching element")
						}
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{
				hotstuffMessageFactory(3),
				hotstuffMessageFactory(19),
				hotstuffMessageFactory(83),
				hotstuffMessageFactory(28),
			},
		},
		{
			name: "TotalMsgBytes should return 0 for an empty mempool",
			args: args{
				maxTransactionBytes: 1000,
				initialElements:     &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						require.Equal(t, int(0), int(txFifoMempool.TotalMsgBytes()), "mismatching total size")
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{},
		},
		{
			name: "TotalMsgBytes should return the correct amount",
			args: args{
				maxTransactionBytes: 1000,
				initialElements:     &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFifoMempool *hotstuffFIFOMempool) {
						msg := hotstuffMessageFactory(3)
						bytes, _ := proto.Marshal(msg)
						txFifoMempool.Push(msg)
						require.Equal(t, len(bytes), int(txFifoMempool.TotalMsgBytes()), "mismatching total size")
					},
				},
			},
			wantItems: []*typesCons.HotstuffMessage{hotstuffMessageFactory(3)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFifoMempool := NewHotstuffFIFOMempool(tt.args.maxTransactionBytes)

			if tt.args.initialElements != nil {
				for _, item := range *tt.args.initialElements {
					txFifoMempool.Push(item)
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(txFifoMempool)
				}
			}

			require.Equal(t, len(tt.wantItems), txFifoMempool.Size(), "mismatching Size (capacity filled with elements)")

			for _, wantItem := range tt.wantItems {
				require.True(t, txFifoMempool.Contains(wantItem), "missing element")
				gotItem, err := txFifoMempool.Pop()
				require.NoError(t, err, "unexpected error while popping element")
				require.True(t, proto.Equal(wantItem, gotItem), "mismatching element")
			}

			if txFifoMempool.Size() == 0 {
				require.True(t, txFifoMempool.IsEmpty(), "IsEmpty should return true when Len() is 0")
			}
		})
	}
}
