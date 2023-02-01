package consensus

import (
	"log"
	"sync"

	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	mempool "github.com/pokt-network/pocket/shared/mempool/list"
)

const (
	hostfuffFIFOMempoolCapacity = int(1e6)
)

type hotstuffFIFOMempool struct {
	g                *mempool.GenericFIFOList[*typesCons.HotstuffMessage]
	m                sync.Mutex
	size             uint32 // current number of transactions in the mempool
	totalMsgBytes    uint64 // current sum of all transactions' sizes (in bytes)
	maxTotalMsgBytes uint64 // maximum total size of all txs allowed in the mempool
}

func NewHotstuffFIFOMempool(maxTransactionBytes uint64) *hotstuffFIFOMempool {
	hotstuffFIFOMempool := &hotstuffFIFOMempool{
		m:                sync.Mutex{},
		size:             0,
		totalMsgBytes:    0,
		maxTotalMsgBytes: maxTransactionBytes,
	}

	hotstuffFIFOMempool.g = mempool.NewGenericFIFOList(
		hostfuffFIFOMempoolCapacity,
		mempool.WithIsEqual(func(a *typesCons.HotstuffMessage, b *typesCons.HotstuffMessage) bool {
			return hashMsg(a) == hashMsg(b)
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) bool {
			hotstuffFIFOMempool.m.Lock()
			defer hotstuffFIFOMempool.m.Unlock()
			// we don't care about the number of messages, only the total size
			return hotstuffFIFOMempool.totalMsgBytes >= hotstuffFIFOMempool.maxTotalMsgBytes
		}),
		mempool.WithOnAdd(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			hotstuffFIFOMempool.m.Lock()
			defer hotstuffFIFOMempool.m.Unlock()

			incrementCounters(item, hotstuffFIFOMempool)
		}),
		mempool.WithOnRemove(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			hotstuffFIFOMempool.m.Lock()
			defer hotstuffFIFOMempool.m.Unlock()

			decrementCounters(item, hotstuffFIFOMempool)
		}),
	)

	return hotstuffFIFOMempool
}

func (mp *hotstuffFIFOMempool) Push(msg *typesCons.HotstuffMessage) error {
	return mp.g.Push(msg)
}

func (mp *hotstuffFIFOMempool) Clear() {
	mp.g.Clear()
	mp.m.Lock()
	defer mp.m.Unlock()
	mp.size = 0
	mp.totalMsgBytes = 0
}

func (mp *hotstuffFIFOMempool) IsEmpty() bool {
	return mp.g.IsEmpty()
}

func (mp *hotstuffFIFOMempool) Pop() (*typesCons.HotstuffMessage, error) {
	return mp.g.Pop()
}

func (mp *hotstuffFIFOMempool) Remove(tx *typesCons.HotstuffMessage) error {
	mp.g.Remove(tx)
	return nil
}

func (mp *hotstuffFIFOMempool) Size() int {
	mp.m.Lock()
	defer mp.m.Unlock()
	return int(mp.size)
}

func (mp *hotstuffFIFOMempool) TotalMsgBytes() uint64 {
	mp.m.Lock()
	defer mp.m.Unlock()
	return mp.totalMsgBytes
}

func (mp *hotstuffFIFOMempool) GetAll() []*typesCons.HotstuffMessage {
	return mp.g.GetAll()
}

func (mp *hotstuffFIFOMempool) Contains(msg *typesCons.HotstuffMessage) bool {
	// since messages are NOT indexed by hash, we need to iterate over all of them
	msgHash := hashMsg(msg)
	for _, m := range mp.GetAll() {
		if hashMsg(m) == msgHash {
			return true
		}
	}
	return false
}

func incrementCounters(item *typesCons.HotstuffMessage, hotstuffFIFOMempool *hotstuffFIFOMempool) {
	bytes, _ := codec.GetCodec().Marshal(item)
	hotstuffFIFOMempool.size++
	hotstuffFIFOMempool.totalMsgBytes += uint64(len(bytes))
}

func decrementCounters(item *typesCons.HotstuffMessage, hotstuffFIFOMempool *hotstuffFIFOMempool) {
	bytes, _ := codec.GetCodec().Marshal(item)
	hotstuffFIFOMempool.size--
	hotstuffFIFOMempool.totalMsgBytes -= uint64(len(bytes))
}

func hashMsg(msg *typesCons.HotstuffMessage) string {
	msgBytes, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("could not marshal message: %v", err)
	}
	return crypto.GetHashStringFromBytes(msgBytes)
}
