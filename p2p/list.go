package poktp2p

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sort"
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
type plist []peer

func (l *plist) add(p peer) {
	slice := []peer(*l)
	slice = append(slice, p)
	*l = plist(slice)
}

func (l *plist) get(pos int) *peer {
	slice := []peer(*l)
	p := slice[pos]
	return &p
}

func (l *plist) sort() {
	slice := []peer(*l)
	less := func(i, j int) bool {
		return slice[i].id < slice[j].id
	}
	sort.SliceStable(slice, less)
	*l = plist(slice)
}

func (l *plist) slice() []peer {
	return []peer(*l)
}

func (l *plist) concat(additional []peer) *plist {
	s := l.slice()

	s = append(s, additional...)
	newl := plist(s)
	return &newl
}

func (l *plist) positionof(id uint64) int {
	var position int = -1

	slice := l.slice()
	for i := 0; i < len(slice); i++ {
		if slice[i].id == id {
			position = i
			break
		}
	}

	return position
}

func (l *plist) size() int {
	slice := []peer(*l)
	return len(slice)
}
