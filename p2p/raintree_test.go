package p2p

import (
	"sync"
	"testing"
)

func TestRainTree_GetTopLevel(t *testing.T) {

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

	maxl := getTopLevel(list)

	if maxl != 4 {
		t.Errorf("Raintree algorithm error: wrong max level value, expected %d, got: %d", 4, maxl)
	}
}

func TestRainTree_GetTargetListSize(t *testing.T) {
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

func TestRainTree_GetTargetList(t *testing.T) {
	list := &plist{elements: make([]peer, 0)}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()
	id := list.get(18).id

	{
		sublist := getTargetList(list, id, 4, 3)

		size := sublist.size()
		if size != 18 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		slice := sublist.slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.id) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, expected item %v, but did not find it in sublist", expectedpos[i])
				break
			}
		}
	}

	{
		sublist := getTargetList(list, id, 4, 2)

		size := sublist.size()
		if size != 12 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3}
		slice := sublist.slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.id) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, expected item %v, but did not find it in sublist", expectedpos[i])
				break
			}
		}
	}

	{
		sublist := getTargetList(list, id, 4, 1)

		size := sublist.size()
		if size != 8 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26}
		slice := sublist.slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.id) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, %v", slice)
				break
			}
		}
	}
}

func TestRainTree_PickLeft(t *testing.T) {
	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	id := list.get(0).id

	l := pickLeft(id, list)

	s := list.slice()
	left := s[l]

	if left.id != 10 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 10, left.id)
		t.Log("list size", list.size(), "top level=", 4, "current level=", 3)
	}
}

func TestRainTree_PickRight(t *testing.T) {
	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
	}

	list.sort()

	id := list.get(0).id

	r := pickRight(id, list)

	s := list.slice()
	right := s[r]

	if right.id != 19 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 19, right.id)
		t.Log("list size", list.size(), "top level=", 4, "current level=", 3)
	}
}

func TestRainTree_Rain(t *testing.T) {
	var rw sync.RWMutex

	peermap := map[uint64][]struct {
		l uint64
		r uint64
	}{}
	addtopeermap := func(id, l, r uint64) {
		peermap[id] = append(peermap[id], struct {
			l uint64
			r uint64
		}{l, r})
	}

	queue := make([]struct {
		id          uint64
		level       int
		root        bool
		contactedby uint64
	}, 0)
	queuein := func(id uint64, level int, root bool, contactedby uint64) {
		queue = append(queue, struct {
			id          uint64
			level       int
			root        bool
			contactedby uint64
		}{id, level, root, contactedby})
	}
	queuepop := func() struct {
		id          uint64
		level       int
		root        bool
		contactedby uint64
	} {
		popped := queue[0]
		queue = queue[1:]
		return popped
	}

	list := &plist{}

	for i := 0; i < 9; i++ {
		p := Peer(uint64(i+1), "")
		list.add(*p)
		peermap[uint64(i+1)] = make([]struct {
			l uint64
			r uint64
		}, 0)
	}

	t.Logf("List instantiated: OK")
	t.Logf("List: %v", list.elements)

	act := func(id uint64, l, r *peer, currentlevel int) {
		defer rw.Unlock()
		rw.Lock()

		addtopeermap(id, l.id, r.id)
		queuein(l.id, currentlevel, false, id)
		queuein(r.id, currentlevel, false, id)

		t.Logf("Queue in left %d and right %d at level %d, by peer %d: OK", l.id, r.id, currentlevel, id)
	}

	// uncomment to play step by step
	//var signalc chan os.Signal
	//var signalg chan os.Signal

	//signalc = make(chan os.Signal)
	//signalg = make(chan os.Signal)

	//signal.Notify(signalc, syscall.SIGQUIT)
	//signal.Notify(signalg, os.Interrupt)

	// queue in node id=5 to start raining
	queuein(uint64(5), 3, true, 0)

	// uncomment to play step by step
	// runner:
	for {
		currentpeer := queuepop()
		//t.Logf("Peer %d will start raining...", currentpeer.id)
		//t.Logf("Is root %v", currentpeer.root)
		rain(currentpeer.id, list, act, currentpeer.root, currentpeer.level)
		t.Logf("Peer %d done raining: OK", currentpeer.id)
		if len(queue) == 0 {
			t.Logf("End of queue.")
			break
		}
		// uncomment to play step by step
		//select {
		//// CTRL+\ to continue
		//case <-signalg:
		//	fmt.Println("got user input")
		//	continue
		//case <-signalc:
		//	t.Logf("Interrupting...")
		//	break runner
		//}
	}

	for id, _ := range peermap {

		p := list.get(int(id - 1))
		if p == nil {
			t.Errorf("Raintree error: expected peer %d to have received a message during the broadcast, but it did not", id)
			break
		}

	}
}
