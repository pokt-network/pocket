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
	hostfuffFIFOMempoolCapacity = int(1e5)
)

type hotstuffFIFOMempool struct {
	g                *mempool.GenericFIFOList[*typesCons.HotstuffMessage]
	m                sync.Mutex
	size             uint32           // current number of transactions in the mempool
	totalMsgBytes    uint64           // current sum of all transactions' sizes (in bytes)
	maxTotalMsgBytes uint64           // maximum total size of all txs allowed in the mempool
	hashCounterSet   map[string]uint8 // the set of hashes of messages in the mempool which tracks not only the presence but also the number of occurrences of a message in the mempool. Used to check for duplicates and potentially react to them
}

func NewHotstuffFIFOMempool(maxTotalMsgBytes uint64) *hotstuffFIFOMempool {
	hotstuffFIFOMempool := &hotstuffFIFOMempool{
		m:                sync.Mutex{},
		size:             0,
		totalMsgBytes:    0,
		maxTotalMsgBytes: maxTotalMsgBytes,
		hashCounterSet:   make(map[string]uint8, hostfuffFIFOMempoolCapacity),
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
			hash := hashMsg(item)
			if _, ok := hotstuffFIFOMempool.hashCounterSet[hash]; ok {
				onDuplicateMessageDetected(item)
			}
			hotstuffFIFOMempool.hashCounterSet[hash]++

			incrementCounters(item, hotstuffFIFOMempool)
		}),
		mempool.WithOnRemove(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			hotstuffFIFOMempool.m.Lock()
			defer hotstuffFIFOMempool.m.Unlock()
			hash := hashMsg(item)
			if prevHashCount, ok := hotstuffFIFOMempool.hashCounterSet[hash]; ok {
				hotstuffFIFOMempool.hashCounterSet[hash]--
				if prevHashCount == 1 {
					delete(hotstuffFIFOMempool.hashCounterSet, hash)
				}
			}

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
	resetCounters(mp)
}

// resetCounters resets size and totalMsgBytes to 0
func resetCounters(mp *hotstuffFIFOMempool) {
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
	mp.m.Lock()
	defer mp.m.Unlock()

	_, ok := mp.hashCounterSet[hashMsg(msg)]
	return ok
}

func incrementCounters(item *typesCons.HotstuffMessage, hotstuffFIFOMempool *hotstuffFIFOMempool) {
	bytes, err := codec.GetCodec().Marshal(item)
	if err != nil {
		log.Fatalf("could not marshal message: %v", err)
		return
	}
	hotstuffFIFOMempool.size++
	hotstuffFIFOMempool.totalMsgBytes += uint64(len(bytes))
}

func decrementCounters(item *typesCons.HotstuffMessage, hotstuffFIFOMempool *hotstuffFIFOMempool) {
	bytes, err := codec.GetCodec().Marshal(item)
	if err != nil {
		log.Fatalf("could not marshal message: %v", err)
		return
	}
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

func onDuplicateMessageDetected(item *typesCons.HotstuffMessage) {
	// TODO(#432): Potential place to check for double signing
	log.Printf("duplicate message detected - hash: %s", hashMsg(item))
}
