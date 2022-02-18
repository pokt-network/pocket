package p2p

import (
	"fmt"
	"os"
	"pocket/shared/types"
	"sync"
	"testing"
	"time"
)

func TestE2EBroadcast(t *testing.T) {
	if os.Getenv("E2E") != "true" {
		t.Skip("Skipping e2e test")
	}

	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	var wg sync.WaitGroup
	var rw sync.RWMutex

	var receipts map[uint64][]*types.NetworkMessage

	list := &plist{}
	fmt.Println(list.slice())
	for i := 0; i < 27; i++ {
		id := uint64(i + 1)
		addr := fmt.Sprintf("127.0.0.1:110%d", i+1)
		fmt.Println("Inserting", i+1, addr)
		p := Peer(id, addr)
		list.add(*p)
	}

	fmt.Println(list.get(0))

	gaters := make([]*gater, 0)
	for i := 0; i < 27; i++ {
		// mark gater as peer with id=1
		listcopy := list.copy()

		g := NewGater()
		p := listcopy.get(i)

		g.id = p.id
		g.address = p.address
		g.externaladdr = p.address
		g.peerlist = &listcopy

		gaters = append(gaters, g)

		err := g.Init()
		if err != nil {
			t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
		}

		g.On(BroadcastDoneEvent, func(args ...interface{}) {
			rw.Lock()
			if receipts[g.id] != nil { // finished a previous broadcast round
				return
			}

			receipts[g.id] = make([]*types.NetworkMessage, 0)

			m := args[0]
			msg := m.(*types.NetworkMessage)

			receipts[g.id] = append(receipts[g.id], msg)
			rw.Unlock()
		})
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

	receipts = map[uint64][]*types.NetworkMessage{}
	gossiperr := make(chan error, 1)
	go func() {
		g := gaters[0]
		g.SetLogger(fmt.Println)

		m := (&pbuff{}).message(int32(0), int32(0), types.PocketTopic_CONSENSUS, g.address, "")
		err := g.Broadcast(m, true)
		if err != nil {
			gossiperr <- err
			panic("dd")
		}
		fmt.Println("Done?")
	}()

closer:
	for {
		select {
		case err := <-gossiperr:
			t.Errorf("Broadcast error: %s", err.Error())
			for _, g := range gaters {
				g.Close()
			}
			break closer

		default:
		}
	}

	<-time.After(time.Second * 2)
	recipients := map[uint64]bool{}
	rw.Lock()
	for k, v := range receipts {
		if len(v) > 0 {
			recipients[k] = true
		}
	}
	rw.Unlock()

	expectedImpact := [][]uint64{
		{4, 5},  // target list size at level 3 = 18, left = 7, right = 13
		{6, 7},  // target list size at level 2 = 35, left = 13, right = 25 (rolling over involved)
		{9, 13}, // target list size at level 2 = 52, left = 9, right = 19 (rolling over involved)
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
