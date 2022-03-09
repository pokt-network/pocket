package p2p

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
	shared "github.com/pokt-network/pocket/shared/config"
	common "github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type testPeer struct {
	id      uint64
	address string
	ready   chan uint
	done    chan uint
	data    chan struct {
		n    int
		err  error
		buff []byte
	}
	respond chan []byte
}

// This test spawns up 26 mock peers (basically just tcp servers that are able to receive/send with simulated encoded/decoding)
// and 1 real p2p peer.
// The real p2p peer goes ahead and tries to broadcast to the peer list he has access to (the 26 peers)
// Being just mock peers, these peers won't re-broadcast themselves, thus giving this tests the ability to
// isolate the broadcast behavior and assert whether expected to-be-infected peers
// the p2p peer should raintree to the other 26 mock peers
// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
func TestNetworkGossip_Broadcast(t *testing.T) {
	// we will have a gater with id = 1

	var p2pPeer *p2pModule = newP2PModule()

	var mx sync.Mutex
	var rw sync.RWMutex
	var wg sync.WaitGroup

	var peers []*testPeer
	var peerList *types.Peerlist // TODO(derrandz): rename type Peerlist to PeerList
	var peerIpsList []string

	var config *shared.P2PConfig

	var recipients map[uint64]bool

	{ // Prepare mock peers and add them to the peer list
		recipients = map[uint64]bool{}
		peers = make([]*testPeer, 0)
		peerList = types.NewPeerlist()

		for i := 0; i < 27; i++ {
			pId := uint64(i + 1)
			pIp := fmt.Sprintf("127.0.0.1:110%d", i+1)

			peerList.Add( // TODO(derrandz): why not add a pointer
				*types.NewPeer(pId, pIp),
			)

			peerIpsList = append(peerIpsList, pIp)
		}

		for i, p := range peerList.Slice()[1:] {
			wg.Add(1)
			go func(i int, p types.Peer) {
				ready, done, data, respond := ListenAndServe(
					p.Addr(),
					int(BufferSize),
					100,
				)

				<-ready
				t.Logf("Broadcast: mock peer %d has started listening", i)

				mx.Lock()
				peers = append(peers, &testPeer{
					id:      p.Id(),
					address: p.Addr(),
					ready:   ready,
					done:    done,
					data:    data,
					respond: respond,
				})
				recipients[p.Id()] = false
				mx.Unlock()

				wg.Done()
			}(i, p)
		}

		wg.Wait()
	}

	{ // prepare the p2p peer
		p := peerList.Get(0)

		config = &shared.P2PConfig{
			Protocol:         "tcp",
			Address:          []byte(p.Addr()),
			ExternalIp:       p.Addr(),
			Peers:            peerIpsList[1:], // [1:] means minus own ip
			MaxInbound:       100,
			MaxOutbound:      100,
			BufferSize:       BufferSize,
			WireHeaderLength: WireByteHeaderLength,
			TimeoutInMs:      2000,
		}

		err := p2pPeer.initialize(config)

		assert.Nil(
			t,
			err,
			"Broadcast error: could not initialize gater. Error: %s", err,
		)

		p2pPeer.setLogger(fmt.Println)

		p2pPeer.id = p.Id()
		assert.Equal(
			t,
			int(p2pPeer.id),
			1, // keep it explicit
			"Broadcast error: (test setup error) expected gater to have id 1",
		)

		go p2pPeer.listen()

		_, waiting := <-p2pPeer.ready

		assert.Equal(
			t,
			waiting,
			false,
			"Broadcast: Encoutered error while listening: p2p peer not ready yet",
		)

		assert.Equal(
			t,
			p2pPeer.isListening.Load(),
			true,
			"Broadcast: Encountered error while listening: p2p peer not started!",
		)

		t.Log("Broadcast: p2p peer has started listening...")
	}

	var wg_ack sync.WaitGroup
	{ // Prepare the mock peers such that they ACK the received broadcast message
		for i, peer := range peers {
			wg_ack.Add(1)
			go func(i int, p *testPeer) {

				// a node might receive more than once
			waiter:
				for {
					select {
					case d := <-p.data:
						assert.Nilf(
							t,
							d.err,
							"Broadcast: mock peer %d encoutered an error while receiving a broadcast messsage: %s", d.err,
						)
						t.Logf("Peer %d got a message", i)
						rw.Lock()
						recipients[p.id] = true
						rw.Unlock()
						t.Logf("Broadcast: mock peer %d has gotten a message, sending back an ACK.", p.id)

						nonce, _, _, _, err := (&wireCodec{}).decode(d.buff)
						assert.Nilf(
							t,
							err,
							"Broadcast: a mock peer has failed to decode received broadcast message: %s", err,
						)

						ack, err := MakeEncodedAckMsg(nonce, p.address, p2pPeer.externaladdr)
						assert.Nilf(
							t,
							err,
							"Broadcast: a mock peer has failed to encode the ACK message to send back: %s", err,
						)
						<-time.After(time.Millisecond * 2) // a few mock peers time out on ACKs, this aleviates this problem.
						p.respond <- ack
						t.Logf("Broadcast: mock peer %d has sent an ACK back.", p.id)

					case <-p.done:
						break waiter
					default:
						continue waiter
					}
				}
				wg_ack.Done()
			}(i, peer)
		}
	}

	var broadcastErr error
	{ // Launch the broadcast
		wg.Add(1)
		go func() {
			<-p2pPeer.ready
			msg := types.NewP2PMessage(int32(0), int32(0), p2pPeer.address, "", &common.PocketEvent{
				Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
				Data:  nil,
			})

			t.Log("Broadcast: p2p peer is about to start the broadcast")
			broadcastErr = p2pPeer.broadcast(msg, true)
			wg.Done()
		}()

		// Just to dequeue queued respones (ACKS) from mock peers as no handlers to consume from the sink are present
		// It not dequeued, will block.
		go func() {
			<-p2pPeer.sink
		}()
	}

	// wait on the broadcast to finish
	// and signal to the ACKing routines (to stop waiting on messages to ACK)
	wg.Wait()
	for _, peer := range peers {
		peer.done <- 1
	}

	// wait on ACK routines to wrap up after the signal
	wg_ack.Done()

	{ // Assert that the broadcast was carried out successfully

		assert.Nilf(
			t,
			broadcastErr,
			"Broadcast: Encountered error while broadcasting: %s", broadcastErr,
		)

		expectedImpact := [][]uint64{
			{4, 5},  // target list size at level 3 = 18, left = 7, right = 13
			{6, 7},  // target list size at level 2 = 35, left = 13, right = 25 (rolling over involved)
			{9, 13}, // target list size at level 2 = 52, left = 9, right = 19 (rolling over involved)
		}

		for _, level := range expectedImpact {
			l, r := level[0], level[1]

			rw.Lock()

			_, lexists := recipients[l]
			assert.Equalf(
				t,
				lexists,
				true,
				"Broadcast: Expected peer with id %d to be impacted, it was not. Impacted peers were: %v", l, recipients,
			)

			_, rexists := recipients[r]
			assert.Equalf(
				t,
				rexists,
				true,
				"Broadcast: Expected peer with id %d to be impacted, it was not. Impacted peers were: %v", r, recipients,
			)

			rw.Unlock()
		}
	}

}

func TestNetworkGossip_HandleBroadcast(t *testing.T) {
	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	// var mx sync.Mutex
	// var wg sync.WaitGroup

	// var rw sync.RWMutex
	// receivedMessages := map[uint64][][]byte{}

	// config := &shared.P2PConfig{
	// 	MaxInbound:       100,
	// 	MaxOutbound:      100,
	// 	BufferSize:       1024 * 4,
	// 	WireHeaderLength: 8,
	// 	TimeoutInMs:      200,
	// }

	// iolist := make([]struct {
	// 	id      uint64
	// 	address string
	// 	ready   chan uint
	// 	done    chan uint
	// 	data    chan struct {
	// 		n    int
	// 		err  error
	// 		buff []byte
	// 	}
	// 	respond chan []byte
	// }, 0)

	// list := types.NewPeerlist()

	// for i := 0; i < 27; i++ {
	// 	p := types.NewPeer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
	// 	list.Add(*p)
	// }

	// // mark gater as peer with id=1
	// p := list.Get(0)
	// m := newP2PModule()

	// m.config = config
	// m.id = p.Id()
	// m.address = p.Addr()
	// m.externaladdr = p.Addr()
	// m.peerlist = list

	// err := m.initialize(nil)
	// if err != nil {
	// 	t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	// }

	// if m.id != 1 {
	// 	t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	// }

	// m.setLogger(fmt.Println)
	// for i, p := range list.Slice()[1:] {
	// 	if p.Id() != m.id {
	// 		wg.Add(1)
	// 		go func(i int, p types.Peer) {
	// 			ready, done, data, respond := ListenAndServe(p.Addr(), int(config.BufferSize))
	// 			<-ready

	// 			mx.Lock()
	// 			iolist = append(iolist, struct {
	// 				id      uint64
	// 				address string
	// 				ready   chan uint
	// 				done    chan uint
	// 				data    chan struct {
	// 					n    int
	// 					err  error
	// 					buff []byte
	// 				}
	// 				respond chan []byte
	// 			}{
	// 				id:      p.Id(),
	// 				address: p.Addr(),
	// 				ready:   ready,
	// 				done:    done,
	// 				data:    data,
	// 				respond: respond,
	// 			})
	// 			receivedMessages[p.Id()] = make([][]byte, 0)

	// 			mx.Unlock()

	// 			wg.Done()
	// 		}(i, p)
	// 	}
	// }

	// wg.Wait()

	// go m.listen()

	// _, waiting := <-m.ready

	// if waiting {
	// 	t.Errorf("Broadcast error: error listening: gater not ready yet")
	// }

	// if !m.isListening.Load() {
	// 	t.Errorf("Broadcast error: error listening: flag shows false after start")
	// }

	// <-time.After(time.Millisecond * 10)

	// gossipdone := make(chan int, 1)
	// go func() {
	// 	<-m.ready
	// 	m.on(types.BroadcastDoneEvent, func(args ...interface{}) {
	// 		gossipdone <- 1
	// 	})
	// 	m.handle()
	// }()

	// fanin := make(chan struct {
	// 	n    int
	// 	err  error
	// 	buff []byte
	// }, 30)

	// for i, io := range iolist {
	// 	e := io

	// 	go func(i int) {

	// 	waiter: // a node might receive more than once
	// 		for {

	// 			select {
	// 			case d := <-e.data:
	// 				rw.Lock()
	// 				receivedMessages[e.id] = append(receivedMessages[e.id], d.buff)
	// 				rw.Unlock()

	// 				fmt.Println(e.address, "received data", len(d.buff))
	// 				fanin <- d

	// 				nonce, _, _, _, err := (&wireCodec{}).decode(d.buff)
	// 				fmt.Println("Err", err)
	// 				msgpayload := &common.PocketEvent{
	// 					Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	// 					Data:  nil,
	// 				}
	// 				ack := types.NewP2PMessage(int32(nonce), int32(0), e.address, m.address, msgpayload)
	// 				eack, _ := m.c.Marshal(*ack)
	// 				wack := (&wireCodec{}).encode(Binary, false, nonce, eack, true)

	// 				e.respond <- wack
	// 			case <-e.done:
	// 				break waiter

	// 			default:
	// 			}
	// 		}
	// 	}(i)
	// }

	// conn, _ := net.Dial("tcp", m.address)

	// gm := types.NewP2PMessage(int32(0), int32(4), conn.LocalAddr().String(), m.address, &common.PocketEvent{
	// 	Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
	// 	Data:  nil,
	// })
	// egm, _ := m.c.Marshal(gm)
	// wgm := (&wireCodec{}).encode(Binary, false, 0, egm, true)
	// conn.Write(wgm)

	// fmt.Println("Has written the size of", len(wgm))
	// buff := make([]byte, m.config.BufferSize)
	// conn.Read(buff)
	// fmt.Println("Acked", len(buff))
	// conn.Close()

	// select {
	// case <-gossipdone:
	// }

	// recipients := map[uint64]bool{}
	// rw.Lock()
	// for k, v := range receivedMessages {
	// 	if len(v) > 0 {
	// 		recipients[k] = true
	// 	}
	// }
	// rw.Unlock()

	// expectedImpact := [][]uint64{
	// 	{4, 5},  // target list size at level 3 = 18, left = 7, right = 13
	// 	{6, 7},  // target list size at level 2 = 35, left = 13, right = 25 (rolling over involved)
	// 	{9, 13}, // target list size at level 2 = 52, left = 9, right = 19 (rolling over involved)
	// }

	// for _, level := range expectedImpact {
	// 	l, r := level[0], level[1]
	// 	_, lexists := recipients[l]
	// 	_, rexists := recipients[r]
	// 	if !lexists {
	// 		t.Errorf("Broadcast error: expected peer with id %d to be impacted, it was not. Impacted peers were: %v", l, recipients)
	// 	}

	// 	if !rexists {
	// 		t.Errorf("Broadcast error: expected peer with id %d to be impacted, it was not. Impacted peers were: %v", r, recipients)
	// 	}
	// }
}
func MakeEncodedAckMsg(nonce uint32, ackerAddr string, ackeeAddr string) ([]byte, error) {
	ack := &types.P2PAckMessage{
		Acker: ackerAddr,
		Ackee: ackeeAddr,
	}
	encoded, err := proto.Marshal(ack)
	if err != nil {
		return nil, err
	}
	wireEncoded := (&wireCodec{}).encode(Binary, false, nonce, encoded, true)
	return wireEncoded, nil
}
