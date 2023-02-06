package raintree

import (
	"testing"

	"github.com/pokt-network/pocket/shared/mempool"
	"github.com/stretchr/testify/require"
)

func TestNonceDeduper(t *testing.T) {
	// pointers in this struct are meant to allow the tester to skip either `initialElements` or `actions` depending on what they are testing for
	type args struct {
		mempoolMaxNonces uint64
		initialElements  *[]uint64
		actions          *[]func(*mempool.GenericFIFOSet[uint64, uint64])
	}
	tests := []struct {
		name      string
		args      args
		wantItems []uint64
	}{
		{
			name: "should correcly implement a FIFO set by removing the oldest element when at capacity",
			args: args{
				mempoolMaxNonces: 3,
				initialElements:  &[]uint64{1, 2, 3},
				actions: &[]func(*mempool.GenericFIFOSet[uint64, uint64]){
					func(nonceDeduper *mempool.GenericFIFOSet[uint64, uint64]) {
						nonceDeduper.Push(4)
					},
				},
			},
			wantItems: []uint64{2, 3, 4},
		},
		{
			name: "an existing element should not be added again",
			args: args{
				mempoolMaxNonces: 3,
				initialElements:  &[]uint64{1, 2, 3},
				actions: &[]func(*mempool.GenericFIFOSet[uint64, uint64]){
					func(nonceDeduper *mempool.GenericFIFOSet[uint64, uint64]) {
						nonceDeduper.Push(1)
					},
				},
			},
			wantItems: []uint64{1, 2, 3},
		},
		{
			name: "removing an element should remove it keeping the order of the rest",
			args: args{
				mempoolMaxNonces: 3,
				initialElements:  &[]uint64{1, 2, 3},
				actions: &[]func(*mempool.GenericFIFOSet[uint64, uint64]){
					func(nonceDeduper *mempool.GenericFIFOSet[uint64, uint64]) {
						nonceDeduper.Remove(2)
					},
				},
			},
			wantItems: []uint64{1, 3},
		},
		{
			name: "clearing the set should remove all elements",
			args: args{
				mempoolMaxNonces: 3,
				initialElements:  &[]uint64{1, 2, 3},
				actions: &[]func(*mempool.GenericFIFOSet[uint64, uint64]){
					func(nonceDeduper *mempool.GenericFIFOSet[uint64, uint64]) {
						nonceDeduper.Clear()
					},
				},
			},
			wantItems: []uint64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonceDeduper := NewNonceDeduper(tt.args.mempoolMaxNonces)

			if tt.args.initialElements != nil {
				for _, item := range *tt.args.initialElements {
					nonceDeduper.Push(item)
				}
			}

			if tt.args.actions != nil {
				for _, action := range *tt.args.actions {
					action(nonceDeduper)
				}
			}

			require.Equal(t, len(tt.wantItems), nonceDeduper.Len(), "mismatching Len (capacity filled with elements)")

			for _, wantItem := range tt.wantItems {
				require.True(t, nonceDeduper.Contains(wantItem), "missing element")
				gotItem, err := nonceDeduper.Pop()
				require.NoError(t, err, "unexpected error while popping element")
				require.Equal(t, wantItem, gotItem, "mismatching element")
			}

			if nonceDeduper.Len() == 0 {
				require.True(t, nonceDeduper.IsEmpty(), "IsEmpty should return true when Len() is 0")
			}
		})
	}
}
