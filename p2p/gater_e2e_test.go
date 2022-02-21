package p2p

import (
	"fmt"
	"pocket/shared/types"
	"sync"
	"testing"
)

func TestE2EBroadcast(t *testing.T) {
	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	var wg sync.WaitGroup

	var rw sync.RWMutex
	receivedMessages := map[uint64][]*types.NetworkMessage{}

	list := &plist{elements: make([]peer, 0)}
	gaters := make([]*gater, 0)

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.add(*p)
	}

	for i := 0; i < 27; i++ {
		// mark gater as peer with id=1
		listcopy := list.copy()
		p := listcopy.get(i)
		g := NewGater()

		g.id = p.id
		g.address = p.address
		g.externaladdr = p.address
		g.peerlist = &listcopy

		gaters = append(gaters, g)

		err := g.Init()
		if err != nil {
			t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
		}
	}

	for _, g := range gaters {
		wg.Add(1)
		go func(gtr *gater) {
			go gtr.Listen()

			_, waiting := <-gtr.ready

			if waiting {
				t.Errorf("Broadcast error: error listening: gater not ready yet")
			}

			if !gtr.listening.Load() {
				t.Errorf("Broadcast error: error listening: flag shows false after start")
			}

			go gtr.Handle()

			wg.Done()
		}(g)
	}

	wg.Wait()

	gossipdone := make(chan int, 1)
	go func() {
		g := gaters[0]

		g.SetLogger(fmt.Println)
		gaters[20].SetLogger(fmt.Println)
		gaters[12].SetLogger(fmt.Println)

		m := (&pbuff{}).message(int32(0), int32(0), types.PocketTopic_CONSENSUS, g.address, "")
		g.On(BroadcastDoneEvent, func() {
			fmt.Println("------------------", g.id)
			rw.Lock()
			receivedMessages[g.id] = append(receivedMessages[g.id], m)
			rw.Unlock()
			gossipdone <- 1
		})
		g.Broadcast(m, true)
		fmt.Println("Done?")
		gossipdone <- 1
	}()

closer:
	for {
		select {
		case <-gossipdone:
			panic("done")
			for _, g := range gaters {
				g.Close()
			}
			break closer
		default:
		}
	}

	recipients := map[uint64]bool{}
	rw.Lock()
	for k, v := range receivedMessages {
		if len(v) > 0 {
			recipients[k] = true
		}
	}
	rw.Unlock()

	expectedImpact := [][]uint64{
		{7, 13},  // target list size at level 3 = 18, left = 7, right = 13
		{13, 25}, // target list size at level 2 = 35, left = 13, right = 25 (rolling over involved)
		{9, 19},  // target list size at level 2 = 52, left = 9, right = 19 (rolling over involved)
	}

	for _, level := range expectedImpact {
		l, r := level[0], level[1]
		_, lexists := recipients[l]
		_, rexists := recipients[r]
		if !lexists {
			t.Errorf("Broadcast error: expected peer with id %d to be impacted, it was not. Impacted peers were: %v", l, recipients)
		}

		if !rexists {
			t.Errorf("Broadcast error: expected peer with id %d to be impacted, it was not. Impacted peers were: %v", r, recipients)
		}
	}
}
