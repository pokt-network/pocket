package types

import (
	"sort"
	"sync"
)

/*
 @
 @ list
 @
*/
type Peerlist struct {
	sync.RWMutex
	elements []Peer
}

func (l *Peerlist) Add(p Peer) {
	l.Lock()
	defer l.Unlock()

	slice := []Peer(l.elements)
	slice = append(slice, p)
	l.elements = slice
}

func (l *Peerlist) Get(pos int) *Peer {
	var p *Peer
	defer func() {
		if err := recover(); err != nil {
			p = nil
		}
	}()
	p = &l.elements[pos]
	return p
}

func (l *Peerlist) Copy() Peerlist {
	copy := *l
	return copy
}

func (l *Peerlist) Update(p []Peer) {
	l.Lock()
	defer l.Unlock()

	l.elements = p
}

func (l *Peerlist) Sort() {
	l.Lock()
	defer l.Unlock()

	slice := l.elements
	less := func(i, j int) bool {
		return slice[i].id < slice[j].id
	}

	sort.SliceStable(slice, less)
	l.elements = slice
}

func (l *Peerlist) Slice() []Peer {
	return append(make([]Peer, 0), l.elements...)
}

func (l *Peerlist) Concat(additional []Peer) *Peerlist {
	l.Lock()
	defer l.Unlock()

	s := make([]Peer, len(l.elements))
	copy(s, l.elements)
	s = append(s, additional...)

	nl := *l
	nl.elements = s

	return &nl
}

func (l *Peerlist) PositionOf(id uint64) int {
	var position int = -1

	slice := l.elements
	for i := 0; i < len(slice); i++ {
		if slice[i].id == id {
			position = i
			break
		}
	}

	return position
}

func (l *Peerlist) Size() int {
	return len(l.elements)
}

func NewPeerlist() *Peerlist { return &Peerlist{elements: make([]Peer, 0)} }
