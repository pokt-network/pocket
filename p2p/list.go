package p2p

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sort"
	"sync"
)

/*
 @
 @ peer
 @
*/
type peer struct {
	id      uint64
	address string
}

func Peer(id uint64, addr string) *peer {
	return &peer{
		id:      id,
		address: addr,
	}
}

func (p *peer) generateId() error {
	idbytes := make([]byte, 8)

	_, err := rand.Read(idbytes[:])

	if err != nil {
		return err
	}

	id, n := binary.Uvarint(idbytes)

	if n == 0 {
		return errors.New("peer error: cannot generate id, buffer too small")
	}

	if n < 0 {
		return errors.New("peer error: cannot generate id, buffer overflow")
	}

	p.id = id

	return nil
}

/*
 @
 @ list
 @
*/
type plist struct {
	sync.RWMutex
	elements []peer
}

func (l *plist) add(p peer) {
	l.Lock()
	defer l.Unlock()

	slice := []peer(l.elements)
	slice = append(slice, p)
	l.elements = slice
}

func (l *plist) get(pos int) *peer {
	var p *peer
	defer func() {
		if err := recover(); err != nil {
			p = nil
		}
	}()
	p = &l.elements[pos]
	return p
}

func (l *plist) copy() plist {
	copy := *l
	return copy
}

func (l *plist) update(p []peer) {
	l.Lock()
	defer l.Unlock()

	l.elements = p
}

func (l *plist) sort() {
	l.Lock()
	defer l.Unlock()

	slice := l.elements
	less := func(i, j int) bool {
		return slice[i].id < slice[j].id
	}

	sort.SliceStable(slice, less)
	l.elements = slice
}

func (l *plist) slice() []peer {
	return l.elements
}

func (l *plist) concat(additional []peer) *plist {
	l.Lock()
	defer l.Unlock()

	s := make([]peer, len(l.elements))
	copy(s, l.elements)
	s = append(s, additional...)

	nl := *l
	nl.elements = s

	return &nl
}

func (l *plist) positionof(id uint64) int {
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

func (l *plist) size() int {
	return len(l.elements)
}
