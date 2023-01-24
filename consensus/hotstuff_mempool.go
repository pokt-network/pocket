package consensus

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	typesCons "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/crypto"
	mempool "github.com/pokt-network/pocket/shared/mempool/list"
)

// TODO: index HighQC incrementally
// TODO: expose findHighQC
// TODO: implement iterator
// TODO: tests for list as well

type hotstuffFifoMempool struct {
	g                *mempool.GenericFIFOList[*typesCons.HotstuffMessage]
	m                sync.Mutex
	size             uint32 // current number of transactions in the mempool
	totalMsgBytes    uint64 // current sum of all transactions' sizes (in bytes)
	maxTotalMsgBytes uint64 // maximum total size of all txs allowed in the mempool
}

func NewHotstuffFIFOMempool(maxTransactionBytes uint64) *hotstuffFifoMempool {
	hotstuffFifoMempool := &hotstuffFifoMempool{
		m:                sync.Mutex{},
		size:             0,
		totalMsgBytes:    0,
		maxTotalMsgBytes: maxTransactionBytes,
	}

	hotstuffFifoMempool.g = mempool.NewGenericFIFOList(
		int(1e6), // TODO: make this configurable
		mempool.WithIndexerFn[*typesCons.HotstuffMessage](func(txBz any) string {
			// We are implementing a list and we don't want deduplication (https://discord.com/channels/824324475256438814/997192534168182905/1067115668136284170)
			// otherwise we would hash the message and use that as the key.
			// This is why we are using a nonce as key. In this context. Every message is unique even if it's the same
			return getNonce()
		}),
		mempool.WithCustomIsOverflowingFn(func(g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) bool {
			hotstuffFifoMempool.m.Lock()
			defer hotstuffFifoMempool.m.Unlock()
			// we don't care about the number of messages, only the total size apparently
			return hotstuffFifoMempool.totalMsgBytes >= hotstuffFifoMempool.maxTotalMsgBytes
		}),
		mempool.WithOnCollision(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			// in here we could check if there is double signing...
		}),
		mempool.WithOnAdd(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			hotstuffFifoMempool.m.Lock()
			defer hotstuffFifoMempool.m.Unlock()

			bytes, _ := proto.Marshal(item)

			hotstuffFifoMempool.size++
			hotstuffFifoMempool.totalMsgBytes += uint64(len(bytes))
		}),
		mempool.WithOnRemove(func(item *typesCons.HotstuffMessage, g *mempool.GenericFIFOList[*typesCons.HotstuffMessage]) {
			hotstuffFifoMempool.m.Lock()
			defer hotstuffFifoMempool.m.Unlock()

			bytes, _ := proto.Marshal(item)

			hotstuffFifoMempool.size--
			hotstuffFifoMempool.totalMsgBytes -= uint64(len(bytes))
		}),
	)

	return hotstuffFifoMempool
}

func (mp *hotstuffFifoMempool) Push(msg *typesCons.HotstuffMessage) error {
	return mp.g.Push(msg)
}

func (mp *hotstuffFifoMempool) Clear() {
	mp.g.Clear()
	mp.m.Lock()
	defer mp.m.Unlock()
	mp.size = 0
	mp.totalMsgBytes = 0
}

func (mp *hotstuffFifoMempool) IsEmpty() bool {
	return mp.g.IsEmpty()
}

func (mp *hotstuffFifoMempool) Pop() (*typesCons.HotstuffMessage, error) {
	return mp.g.Pop()
}

func (mp *hotstuffFifoMempool) Remove(tx *typesCons.HotstuffMessage) error {
	mp.g.Remove(tx)
	return nil
}

func (mp *hotstuffFifoMempool) Size() int {
	mp.m.Lock()
	defer mp.m.Unlock()
	return int(mp.size)
}

func (mp *hotstuffFifoMempool) TotalMsgBytes() uint64 {
	mp.m.Lock()
	defer mp.m.Unlock()
	return mp.totalMsgBytes
}

func (mp *hotstuffFifoMempool) GetAll() []*typesCons.HotstuffMessage {
	return mp.g.GetAll()
}

func (mp *hotstuffFifoMempool) Contains(msg *typesCons.HotstuffMessage) bool {
	// since messages are NOT indexed by hash, we need to iterate over all of them
	msgHash := hashMsg(msg)
	for _, m := range mp.GetAll() {
		if hashMsg(m) == msgHash {
			return true
		}
	}
	return false
}

func getNonce() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("%d", rand.Uint64())
}

func hashMsg(msg *typesCons.HotstuffMessage) string {
	msgBytes, err := codec.GetCodec().Marshal(msg)
	if err != nil {
		log.Fatalf("could not marshal message: %v", err)
	}
	return crypto.GetHashStringFromBytes(msgBytes)
}
