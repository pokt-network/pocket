package list

import (
	"container/list"
	"fmt"
	"sync"
)

type GenericFIFOList[TData comparable] struct {
	queue    *list.List
	capacity int
	m        sync.Mutex

	isEqual       func(TData, TData) bool
	isOverflowing func(*GenericFIFOList[TData]) bool

	onAdd    func(TData, *GenericFIFOList[TData])
	onRemove func(TData, *GenericFIFOList[TData])
}

func NewGenericFIFOList[TData comparable](capacity int, options ...func(*GenericFIFOList[TData])) *GenericFIFOList[TData] {
	gfs := &GenericFIFOList[TData]{
		queue:         list.New(),
		capacity:      capacity,
		isOverflowing: defaultIsOverflowing[TData],
		isEqual:       defaultIsEqual[TData],
		onAdd: func(item TData, g *GenericFIFOList[TData]) {
			// noop
		},
		onRemove: func(item TData, g *GenericFIFOList[TData]) {
			// noop
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

	g.queue.PushBack(item)
	g.onAdd(item, g)

	if g.isOverflowing != nil && g.isOverflowing(g) {
		front := g.queue.Front()
		g.queue.Remove(front)
		g.onRemove(item, g)
	}
	return nil
}

func (g *GenericFIFOList[TData]) Pop() (v TData, err error) {
	g.m.Lock()
	defer g.m.Unlock()

	if g.queue.Len() == 0 {
		return v, fmt.Errorf("empty set")
	}

	front := g.queue.Front()
	item := front.Value.(TData)
	g.queue.Remove(front)
	g.onRemove(item, g)
	return item, nil
}

func (g *GenericFIFOList[TData]) Remove(item TData) {
	g.m.Lock()
	defer g.m.Unlock()

	for e := g.queue.Front(); e != nil; e = e.Next() {
		if g.isEqual(e.Value.(TData), item) {
			g.queue.Remove(e)
			g.onRemove(item, g)
			break
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

	g.queue.Init()
}

func (g *GenericFIFOList[TData]) IsEmpty() bool {
	g.m.Lock()
	defer g.m.Unlock()

	return g.queue.Len() == 0
}

func (g *GenericFIFOList[TData]) Contains(item TData) bool {
	g.m.Lock()
	defer g.m.Unlock()

	for e := g.queue.Front(); e != nil; e = e.Next() {
		if g.isEqual(e.Value.(TData), item) {
			return true
		}
	}

	return false
}

func (g *GenericFIFOList[TData]) GetAll() []TData {
	g.m.Lock()
	defer g.m.Unlock()

	items := make([]TData, 0, g.queue.Len())
	for e := g.queue.Front(); e != nil; e = e.Next() {
		items = append(items, e.Value.(TData))
	}
	return items
}

// Options

func WithCustomIsOverflowingFn[TData comparable](fn func(*GenericFIFOList[TData]) bool) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.isOverflowing = fn
	}
}

func WithOnAdd[TData comparable](fn func(TData, *GenericFIFOList[TData])) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.onAdd = fn
	}
}

func WithOnRemove[TData comparable](fn func(TData, *GenericFIFOList[TData])) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.onRemove = fn
	}
}

func WithIsEqual[TData comparable](fn func(TData, TData) bool) func(*GenericFIFOList[TData]) {
	return func(g *GenericFIFOList[TData]) {
		g.isEqual = fn
	}
}

// private methods

func defaultIsOverflowing[TData comparable](g *GenericFIFOList[TData]) bool {
	return g.queue.Len() > g.capacity
}

func defaultIsEqual[TData comparable](a, b TData) bool {
	return a == b
}
