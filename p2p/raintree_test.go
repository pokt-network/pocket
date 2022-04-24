package p2p

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Setup_TestRainTree_Case(root int, listSize int) RainTree {
	list := make([]peerInfo, 0)
	for i := 0; i < listSize; i++ {
		list = append(list, peerInfo{
			ID:      i + 1,
			address: "127.0.0.1:" + strconv.Itoa(10000+i+1),
		})
	}

	rt := NewRainTree()
	rt.SetLeafs(list)
	rt.SetRoot(root)

	return rt
}

func TestRainTree_GetTopLevel(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 27)

	maxl := tree.GetTopLevel()

	assert.Equal(
		t,
		4,
		maxl,
		"Raintree algorithm error: wrong max level value",
	)
}

func TestRainTree_GetTargetListSize(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 27)

	tlsize := tree.GetTargetListSize(4, 3)

	assert.Equal(
		t,
		18,
		int(tlsize),
		"Raintree algorithm error: failed to retrieve proper sublist",
	)

	tlsize = tree.GetTargetListSize(4, 2)

	assert.Equal(
		t,
		12,
		int(tlsize),
		"Raintree algorithm error: failed to retrieve proper sublist",
	)

	tlsize = tree.GetTargetListSize(4, 1)

	assert.Equal(
		t,
		8,
		int(tlsize),
		"Raintree algorithm error: failed to retrieve proper sublist",
	)
}

func TestRainTree_GetTargetList(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 27)

	t.Run("Raintree calculates the correct target list for a 4-level tree, at the 3rd level", func(t *testing.T) {
		tree.SetRoot(19)
		sublist := tree.GetTargetList(4, 3)

		assert.Equal(
			t,
			18,
			len(sublist),
			"Raintree algorithm error: failed to retrieve proper sublist",
		)

		expected := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		assert.Equal(
			t,
			expected,
			sublist,
			"Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist",
		)
	})

	t.Run("Raintree calculates the correct target list for a 4-level tree, at the 2nd level", func(t *testing.T) {
		tree.SetRoot(19)
		sublist := tree.GetTargetList(4, 2)

		assert.Equal(
			t,
			12,
			len(sublist),
			"Raintree algorithm error: failed to retrieve proper sublist",
		)

		expected := []int{19, 20, 21, 22, 23, 24, 25, 26, 27, 1, 2, 3}
		assert.Equal(
			t,
			expected,
			sublist,
			"Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist",
		)
	})

	t.Run("Raintree calculates the correct target list for a 4-level tree, at the 1st level", func(t *testing.T) {
		tree.SetRoot(19)
		sublist := tree.GetTargetList(4, 1)

		assert.Equal(
			t,
			8,
			len(sublist),
			"Raintree algorithm error: failed to retrieve proper sublist",
		)

		expected := []int{19, 20, 21, 22, 23, 24, 25, 26}
		assert.Equal(
			t,
			expected,
			sublist,
			"Raintree algorithm error: failed to retrieve proper sublist, wrong elements of sublist",
		)
	})
}

func TestRainTree_PickLeft(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 27)

	l := tree.PickLeft(tree.GetSortedList())
	left := tree.GetByPosition(l)

	assert.Equal(
		t,
		10,
		left.ID,
		"Raintree algorithm error: failed to pick proper left at provided level",
	)
}

func TestRainTree_PickRight(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 27)

	l := tree.PickRight(tree.GetSortedList())
	left := tree.GetByPosition(l)

	assert.Equal(
		t,
		19,
		left.ID,
		"Raintree algorithm error: failed to pick proper right at provided level",
	)
}

func TestRainTree_Traverse(t *testing.T) {
	tree := Setup_TestRainTree_Case(1, 9)

	var rw sync.RWMutex

	peermap := map[int][]struct {
		l int
		r int
	}{}

	for _, id := range tree.GetSortedList() {
		peermap[id] = make([]struct {
			l int
			r int
		}, 0)
	}

	queue := make([]struct {
		id          int
		level       int
		root        bool
		contactedby int
	}, 0)

	addtopeermap := func(id, l, r int) {
		peermap[id] = append(peermap[id], struct {
			l int
			r int
		}{l, r})
	}
	queuein := func(id int, level int, root bool, contactedby int) {
		queue = append(queue, struct {
			id          int
			level       int
			root        bool
			contactedby int
		}{id, level, root, contactedby})
	}
	queuepop := func() struct {
		id          int
		level       int
		root        bool
		contactedby int
	} {
		popped := queue[0]
		queue = queue[1:]
		return popped
	}

	t.Logf("Tree instantiated: OK")
	t.Logf("Tree: %v", tree)

	act := func(originator int, l, r peerInfo, currentlevel int) error {
		defer rw.Unlock()
		rw.Lock()

		lid := l.ID
		rid := r.ID

		addtopeermap(originator, lid, rid)
		queuein(lid, currentlevel, false, originator)
		queuein(rid, currentlevel, false, originator)

		t.Logf("Queue in left %d and right %d at level %d, by peer %d: OK", lid, rid, currentlevel, originator)
		return nil
	}

	// uncomment to play step by step
	//var signalc chan os.Signal
	//var signalg chan os.Signal

	//signalc = make(chan os.Signal)
	//signalg = make(chan os.Signal)

	//signal.Notify(signalc, syscall.SIGQUIT)
	//signal.Notify(signalg, os.Interrupt)

	// queue in node id=5 to start raining
	queuein(int(5), 3, true, 0)

	// uncomment to play step by step
	// runner:
	for {
		currentpeer := queuepop()
		//t.Logf("Peer %d will start raining...", currentpeer.id)
		//t.Logf("Is root %v", currentpeer.root)
		tree.SetRoot(currentpeer.id)
		tree.Traverse(currentpeer.root, currentpeer.level, act)
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

	for _, id := range tree.GetSortedList() {
		_, exists := peermap[id]
		assert.Truef(
			t,
			exists,
			"Raintree error: expected peer %d to have received a message during the broadcast, but it did not", id,
		)
	}
}
