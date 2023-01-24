package list

import (
	"container/list"
	"fmt"
	"sync"
)

type GenericFIFOList[TData any] struct {
	set      map[any]struct{}
	queue    *list.List
	capacity int
	m        sync.Mutex

	indexerFn     func(any) any
	isOverflowing func(*GenericFIFOList[TData]) bool

	onAdd       func(TData, *GenericFIFOList[TData])
	onRemove    func(TData, *GenericFIFOList[TData])
	onCollision func(TData, *GenericFIFOList[TData])
}

func NewGenericFIFOList[TData any](capacity int, options ...func(*GenericFIFOList[TData])) *GenericFIFOList[TData] {
	gfs := &GenericFIFOList[TData]{
		set:      make(map[any]struct{}, capacity),
		queue:    list.New(),
		capacity: capacity,
		indexerFn: func(item any) any {
			return item
		},
		isOverflowing: defaultIsOverflowing[TData],
		onAdd: func(item TData, g *GenericFIFOList[TData]) {
			// do nothing
		},
		onRemove: func(item TData, g *GenericFIFOList[TData]) {
			// do nothing
		},
		onCollision: func(item TData, g *GenericFIFOList[TData]) {
			// do nothing
		},
	}

	for _, o := range options {
		o(gfs)
	}

	return gfs
}

func (g *GenericFIFOList[TData]) Push(item TData) error {
	g.m.Lock()
	defer g.m.Unlock()

	index := g.indexerFn(item)
	if _, ok := g.set[index]; ok {
		if g.onCollision != nil {
			g.onCollision(item, g)
		}
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

func (g *GenericFIFOList[TData]) Pop() (TData, error) {
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

func (g *GenericFIFOList[TData]) Remove(item TData) {
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

func (g *GenericFIFOList[TData]) Len() int {
	g.m.Lock()
	defer g.m.Unlock()

	return g.queue.Len()
}

func (g *GenericFIFOList[TData]) Clear() {
	g.m.Lock()
	defer g.m.Unlock()

	g.set = make(map[any]struct{}, g.capacity)
	g.queue = list.New()
}

func (g *GenericFIFOList[TData]) IsEmpty() bool {
	g.m.Lock()
	defer g.m.Unlock()

	return g.queue.Len() == 0
}

func (g *GenericFIFOList[TData]) Contains(item TData) bool {
	g.m.Lock()
	defer g.m.Unlock()

	_, ok := g.set[g.indexerFn(item)]
	return ok
}

func (g *GenericFIFOList[TData]) ContainsIndex(index any) bool {
	g.m.Lock()
	defer g.m.Unlock()

	_, ok := g.set[index]
	return ok
}

// Options

func WithIndexerFn[TData any](fn func(any) any) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.indexerFn = fn
	}
}

func WithCustomIsOverflowingFn[TData any](fn func(*GenericFIFOList[TData]) bool) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.isOverflowing = fn
	}
}

func WithOnAdd[TData any](fn func(TData, *GenericFIFOList[TData])) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.onAdd = fn
	}
}

func WithOnRemove[TData any](fn func(TData, *GenericFIFOList[TData])) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.onRemove = fn
	}
}

func WithOnCollision[TData any](fn func(TData, *GenericFIFOList[TData])) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.onCollision = fn
	}
}

// private methods

func defaultIsOverflowing[TData any](g *GenericFIFOList[TData]) bool {
	return g.queue.Len() > g.capacity
}
