package mempool

import (
	"container/list"
	"fmt"
	"sync"
)

// DOCUMENT: Add a proper README for Pocket's GenericFifoSet & Mempool. See the source code for as an interface reference for now
type GenericFIFOSet[TIdx comparable, TData any] struct {
	set      map[TIdx]struct{}
	queue    *list.List
	capacity int
	m        sync.Mutex

	indexerFn     func(any) TIdx
	isOverflowing func(*GenericFIFOSet[TIdx, TData]) bool

	onAdd       func(TData, *GenericFIFOSet[TIdx, TData])
	onRemove    func(TData, *GenericFIFOSet[TIdx, TData])
	onCollision func(TData, *GenericFIFOSet[TIdx, TData]) error
}

func NewGenericFIFOSet[TIdx comparable, TData any](capacity int, options ...func(*GenericFIFOSet[TIdx, TData])) *GenericFIFOSet[TIdx, TData] {
	gfs := &GenericFIFOSet[TIdx, TData]{
		set:      make(map[TIdx]struct{}, capacity),
		queue:    list.New(),
		capacity: capacity,
		indexerFn: func(item any) TIdx {
			// by default we use the item itself as index (for simple cases like nonce deduping)
			return item.(TIdx)
		},
		isOverflowing: defaultIsOverflowing[TIdx, TData],
		onAdd: func(item TData, g *GenericFIFOSet[TIdx, TData]) {
			// noop
		},
		onRemove: func(item TData, g *GenericFIFOSet[TIdx, TData]) {
			// noop
		},
		onCollision: func(item TData, g *GenericFIFOSet[TIdx, TData]) error {
			return fmt.Errorf("item %v already exists", item)
		},
	}

	for _, o := range options {
		o(gfs)
	}

	return gfs
}

func (g *GenericFIFOSet[TIdx, TData]) Push(item TData) error {
	g.m.Lock()
	defer g.m.Unlock()

	index := g.indexerFn(item)
	if _, ok := g.set[index]; ok {
		if g.onCollision != nil {
			return g.onCollision(item, g)
		}
		return fmt.Errorf("item %v already exists", item)
	}

	g.set[index] = struct{}{}
	g.queue.PushBack(item)
	g.onAdd(item, g)

	if g.isOverflowing != nil && g.isOverflowing(g) {
		front := g.queue.Front()
		delete(g.set, g.indexerFn(front.Value.(TData)))
		g.queue.Remove(front)
	}
	return nil
}

func (g *GenericFIFOSet[TIdx, TData]) Pop() (TData, error) {
	g.m.Lock()
	defer g.m.Unlock()

	if g.queue.Len() == 0 {
		return any(nil).(TData), fmt.Errorf("empty set")
	}

	front := g.queue.Front()
	item := front.Value.(TData)
	delete(g.set, g.indexerFn(item))
	g.queue.Remove(front)
	g.onRemove(item, g)
	return item, nil
}

func (g *GenericFIFOSet[TIdx, TData]) Remove(item TData) {
	g.m.Lock()
	defer g.m.Unlock()

	itemIndex := g.indexerFn(item)
	if _, ok := g.set[itemIndex]; ok {
		delete(g.set, itemIndex)
		for e := g.queue.Front(); e != nil; e = e.Next() {
			if g.indexerFn(e.Value.(TData)) == itemIndex {
				g.queue.Remove(e)
				break
			}
		}
	}
}

func (g *GenericFIFOSet[TIdx, TData]) Len() int {
	g.m.Lock()
	defer g.m.Unlock()

	return g.queue.Len()
}

func (g *GenericFIFOSet[TIdx, TData]) Clear() {
	g.m.Lock()
	defer g.m.Unlock()

	g.set = make(map[TIdx]struct{}, g.capacity)
	g.queue = list.New()
}

func (g *GenericFIFOSet[TIdx, TData]) IsEmpty() bool {
	g.m.Lock()
	defer g.m.Unlock()

	return g.queue.Len() == 0
}

func (g *GenericFIFOSet[TIdx, TData]) Contains(item TData) bool {
	g.m.Lock()
	defer g.m.Unlock()

	_, ok := g.set[g.indexerFn(item)]
	return ok
}

func (g *GenericFIFOSet[TIdx, TData]) ContainsIndex(index TIdx) bool {
	g.m.Lock()
	defer g.m.Unlock()

	_, ok := g.set[index]
	return ok
}

// Options

func WithIndexerFn[TIdx comparable, TData any](fn func(any) TIdx) func(*GenericFIFOSet[TIdx, TData]) {
	return func(g *GenericFIFOSet[TIdx, TData]) {
		g.indexerFn = fn
	}
}

func WithCustomIsOverflowingFn[TIdx comparable, TData any](fn func(*GenericFIFOSet[TIdx, TData]) bool) func(*GenericFIFOSet[TIdx, TData]) {
	return func(g *GenericFIFOSet[TIdx, TData]) {
		g.isOverflowing = fn
	}
}

func WithOnAdd[TIdx comparable, TData any](fn func(TData, *GenericFIFOSet[TIdx, TData])) func(*GenericFIFOSet[TIdx, TData]) {
	return func(g *GenericFIFOSet[TIdx, TData]) {
		g.onAdd = fn
	}
}

func WithOnRemove[TIdx comparable, TData any](fn func(TData, *GenericFIFOSet[TIdx, TData])) func(*GenericFIFOSet[TIdx, TData]) {
	return func(g *GenericFIFOSet[TIdx, TData]) {
		g.onRemove = fn
	}
}

func WithOnCollision[TIdx comparable, TData any](fn func(TData, *GenericFIFOSet[TIdx, TData]) error) func(*GenericFIFOSet[TIdx, TData]) {
	return func(g *GenericFIFOSet[TIdx, TData]) {
		g.onCollision = fn
	}
}

// private methods

func defaultIsOverflowing[TIdx comparable, TData any](g *GenericFIFOSet[TIdx, TData]) bool {
	return g.queue.Len() > g.capacity
}
