package p2p

import (
	"fmt"
	"testing"
)

func TestGetTopLevel(t *testing.T) {

	g := NewGater()

	p := Peer(0, "")
	err := p.generateId()
	if err != nil {
		t.Errorf("Failed to init test, could not generate peer id, err: %s", err.Error())
		t.Failed()
	}

	g.id = p.id

	list := &plist{elements: make([]peer, 0)}
	for i := 0; i < 27; i++ {
		p := Peer(0, "")

		if err := p.generateId(); err != nil {
			t.Errorf("Failed to init test, could not generate peer id for peer list (i: %d). err: %s", i, err.Error())
		}

		list.add(*p)
	}

	list.sort()

	maxl := getTopLevel(*list)

	if maxl != 4 {
		t.Errorf("Raintree algorithm error: wrong max level value, expected %d, got: %d", 4, maxl)
	}
}

func TestGetTargetListSize(t *testing.T) {
	list := &plist{elements: make([]peer, 0)}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	tlsize := int(getTargetListSize(list.size(), 4, 3))

	if tlsize != 18 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, tlsize)
	}

	tlsize = int(getTargetListSize(list.size(), 4, 2))

	if tlsize != 12 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 12, tlsize)
	}

	tlsize = int(getTargetListSize(list.size(), 4, 1))

	if tlsize != 8 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 8, tlsize)
	}
}

func TestGetTargetList(t *testing.T) {
	list := &plist{elements: make([]peer, 0)}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	id := list.get(18).id
	sublist := getTargetList(*list, id, 4, 3)

	size := sublist.size()
	fmt.Println("Size", size)
	if size != 18 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
	}

	expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	slice := sublist.slice()
	for i := 0; i < len(slice); i++ {
		elem := slice[i]
		if expectedpos[i] != int(elem.id) {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, %v", slice)
			t.Failed()
			break
		}
	}
}

func TestPickLeft(t *testing.T) {
	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	id := list.get(0).id

	l := pickLeft(id, *list)

	s := list.slice()
	left := s[l]

	if left.id != 10 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 10, left.id)
		t.Log("list size", list.size(), "top level=", 4, "current level=", 3)
	}
}

func TestPickRight(t *testing.T) {
	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	id := list.get(0).id

	r := pickRight(id, *list)

	s := list.slice()
	right := s[r]

	if right.id != 19 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 19, right.id)
		t.Log("list size", list.size(), "top level=", 4, "current level=", 3)
	}
}
