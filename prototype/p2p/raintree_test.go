package p2p

import (
	"pocket/p2p/types"
	"sync"
	"testing"
)

func TestRainTree_GetTopLevel(t *testing.T) {

	p := types.NewPeer(0, "")
	err := p.GenerateId()
	if err != nil {
		t.Errorf("Failed to init test, could not generate peer id, err: %s", err.Error())
		t.Failed()
	}

	p.Id()

	list := types.NewPeerlist()
	for i := 0; i < 27; i++ {
		p := types.NewPeer(0, "")

		if err := p.GenerateId(); err != nil {
			t.Errorf("Failed to init test, could not generate peer id for peer list (i: %d). err: %s", i, err.Error())
		}

		list.Add(*p)
	}

	list.Sort()

	maxl := getTopLevel(list)

	if maxl != 4 {
		t.Errorf("Raintree algorithm error: wrong max level value, expected %d, got: %d", 4, maxl)
	}
}

func TestRainTree_GetTargetListSize(t *testing.T) {
	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), "")
		list.Add(*p)
	}

	list.Sort()

	tlsize := int(getTargetListSize(list.Size(), 4, 3))

	if tlsize != 18 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, tlsize)
	}

	tlsize = int(getTargetListSize(list.Size(), 4, 2))

	if tlsize != 12 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 12, tlsize)
	}

	tlsize = int(getTargetListSize(list.Size(), 4, 1))

	if tlsize != 8 {
		t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 8, tlsize)
	}
}

func TestRainTree_GetTargetList(t *testing.T) {
	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), "")
		list.Add(*p)
	}

	list.Sort()
	id := list.Get(18).Id()

	{
		sublist := getTargetList(list, id, 4, 3)

		size := sublist.Size()
		if size != 18 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		slice := sublist.Slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.Id()) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, expected item %v, but did not find it in sublist", expectedpos[i])
				break
			}
		}
	}

	{
		sublist := getTargetList(list, id, 4, 2)

		size := sublist.Size()
		if size != 12 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3}
		slice := sublist.Slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.Id()) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, expected item %v, but did not find it in sublist", expectedpos[i])
				break
			}
		}
	}

	{
		sublist := getTargetList(list, id, 4, 1)

		size := sublist.Size()
		if size != 8 {
			t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, expected list of size %d, got: %d", 18, size)
		}

		expectedpos := []int{19, 20, 21, 22, 23, 24, 25, 26}
		slice := sublist.Slice()
		for i := 0; i < len(slice); i++ {
			elem := slice[i]
			if expectedpos[i] != int(elem.Id()) {
				t.Errorf("Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist, %v", slice)
				break
			}
		}
	}
}

func TestRainTree_PickLeft(t *testing.T) {
	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), "")
		list.Add(*p)
	}

	list.Sort()

	id := list.Get(0).Id()

	l := pickLeft(id, list)

	s := list.Slice()
	left := s[l]
	lid := left.Id()

	if lid != 10 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 10, lid)
		t.Log("list.Size", list.Size(), "top level=", 4, "current level=", 3)
	}
}

func TestRainTree_PickRight(t *testing.T) {
	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), "")
		list.Add(*p)
	}

	list.Sort()

	id := list.Get(0).Id()

	r := pickRight(id, list)

	s := list.Slice()
	right := s[r]
	rid := right.Id()

	if right.Id() != 19 {
		t.Errorf("Raintree algorithm error: failed to pick proper left at provided level, expected %d, got: %d", 19, rid)
		t.Log("list.Size", list.Size(), "top level=", 4, "current level=", 3)
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

	list := types.NewPeerlist()

	for i := 0; i < 9; i++ {
		p := types.NewPeer(uint64(i+1), "")
		list.Add(*p)
		peermap[uint64(i+1)] = make([]struct {
			l uint64
			r uint64
		}, 0)
	}

	t.Logf("List instantiated: OK")
	t.Logf("List: %v", list.Slice())

	act := func(id uint64, l, r *types.Peer, currentlevel int) {
		defer rw.Unlock()
		rw.Lock()

		lid := l.Id()
		rid := r.Id()

		addtopeermap(id, lid, rid)
		queuein(lid, currentlevel, false, id)
		queuein(rid, currentlevel, false, id)

		t.Logf("Queue in left %d and right %d at level %d, by peer %d: OK", lid, rid, currentlevel, id)
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

		p := list.Get(int(id - 1))
		if p == nil {
			t.Errorf("Raintree error: expected peer %d to have received a message during the broadcast, but it did not", id)
			break
		}

	}
}
