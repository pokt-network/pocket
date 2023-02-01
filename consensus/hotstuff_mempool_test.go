package consensus

import (
	"testing"

	"github.com/golang/protobuf/proto"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/core/types"
	"github.com/stretchr/testify/require"
)

func TestMempool(t *testing.T) {
	type args struct {
		maxTotalMsgBytes uint64
		initialMsgs      *[]*typesCons.HotstuffMessage
		actions          *[]func(*hotstuffFIFOMempool)
	}
	tests := []struct {
		name     string
		args     args
		wantMsgs []*typesCons.HotstuffMessage
	}{
		{
			// maxTotalMsgBytes is 20, we are starting with a message of size _at least_ 10 and adding another one of _at least_ 10
			// we expect the first one to be removed
			name: "should correcly implement a FIFO set by removing the oldest message when at capacity",
			args: args{
				maxTotalMsgBytes: 20,
				initialMsgs:      &[]*typesCons.HotstuffMessage{hotstuffMessageFactory(10)},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						txFIFOMempool.Push(hotstuffMessageFactory(10))
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{hotstuffMessageFactory(10)},
		},
		{
			name: "looping over the mempool and Pop()ing all the messages should return an empty mempool",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						for !txFIFOMempool.IsEmpty() {
							txFIFOMempool.Pop()
						}
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Pop()ing a message should return it correctly",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs:      &[]*typesCons.HotstuffMessage{hotstuffMessageFactory(19)},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						msg, err := txFIFOMempool.Pop()
						require.NoError(t, err, "unexpected error while popping message")
						expectedMsg := hotstuffMessageFactory(19)

						require.True(t, proto.Equal(msg, expectedMsg), "mismatching message")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Pop()ing a message from an empty mempool should return an error",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs:      &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						_, err := txFIFOMempool.Pop()
						require.Error(t, err, "expected error while popping message")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{},
		},
		{
			name: "Push()ing a message to an empty mempool should return the correct size",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs:      &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						txFIFOMempool.Push(hotstuffMessageFactory(19))
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{hotstuffMessageFactory(19)},
		},
		{
			name: "Clear should empty the mempool",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						txFIFOMempool.Clear()
						require.Equal(t, 0, txFIFOMempool.Size(), "mismatching size")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{},
		},
		{
			name: "GetAll should return all the messages in the mempool, in the order they were added in the first place",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						expectedMsgs := []*typesCons.HotstuffMessage{
							hotstuffMessageFactory(3),
							hotstuffMessageFactory(19),
							hotstuffMessageFactory(83),
							hotstuffMessageFactory(28),
						}

						for i, msg := range txFIFOMempool.GetAll() {
							require.True(t, proto.Equal(msg, expectedMsgs[i]), "mismatching message")
						}
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{
				hotstuffMessageFactory(3),
				hotstuffMessageFactory(19),
				hotstuffMessageFactory(83),
				hotstuffMessageFactory(28),
			},
		},
		{
			name: "TotalMsgBytes should return 0 for an empty mempool",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs:      &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						require.Equal(t, int(0), int(txFIFOMempool.TotalMsgBytes()), "mismatching total size")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{},
		},
		{
			name: "TotalMsgBytes should return the correct amount",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs:      &[]*typesCons.HotstuffMessage{},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						msg := hotstuffMessageFactory(3)
						bytes, _ := proto.Marshal(msg)
						txFIFOMempool.Push(msg)
						require.Equal(t, len(bytes), int(txFIFOMempool.TotalMsgBytes()), "mismatching total size")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{hotstuffMessageFactory(3)},
		},
		{
			name: "Remove should remove the correct message",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						txFIFOMempool.Remove(hotstuffMessageFactory(19))
						require.Equal(t, 3, txFIFOMempool.Size(), "mismatching size")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{
				hotstuffMessageFactory(3),
				hotstuffMessageFactory(83),
				hotstuffMessageFactory(28),
			},
		},
		{
			name: "Contains should return true for a message that is in the mempool and false otherwise",
			args: args{
				maxTotalMsgBytes: 1000,
				initialMsgs: &[]*typesCons.HotstuffMessage{
					hotstuffMessageFactory(3),
					hotstuffMessageFactory(19),
					hotstuffMessageFactory(83),
					hotstuffMessageFactory(28),
				},
				actions: &[]func(*hotstuffFIFOMempool){
					func(txFIFOMempool *hotstuffFIFOMempool) {
						require.True(t, txFIFOMempool.Contains(hotstuffMessageFactory(19)), "should contain message")
						require.False(t, txFIFOMempool.Contains(hotstuffMessageFactory(17)), "should not contain message")
					},
				},
			},
			wantMsgs: []*typesCons.HotstuffMessage{
				hotstuffMessageFactory(3),
				hotstuffMessageFactory(19),
				hotstuffMessageFactory(83),
				hotstuffMessageFactory(28),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFIFOMempool := NewHotstuffFIFOMempool(tt.args.maxTotalMsgBytes)

			if tt.args.initialMsgs != nil {
				for _, item := range *tt.args.initialMsgs {
					txFIFOMempool.Push(item)
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(txFIFOMempool)
				}
			}

			require.Equal(t, len(tt.wantMsgs), txFIFOMempool.Size(), "mismatching Size (capacity filled with messages)")

			for _, wantItem := range tt.wantMsgs {
				require.True(t, txFIFOMempool.Contains(wantItem), "missing message")
				gotItem, err := txFIFOMempool.Pop()
				require.NoError(t, err, "unexpected error while popping message")
				require.True(t, proto.Equal(wantItem, gotItem), "mismatching message")
			}

			if txFIFOMempool.Size() == 0 {
				require.True(t, txFIFOMempool.IsEmpty(), "IsEmpty should return true when Len() is 0")
			}
		})
	}
}

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
