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
	respond  chan []byte
	infected bool
}

func Setup_TestNetworkGossip_Broadcast() (*p2pModule, []*testPeer, *types.Peerlist) {
	var p2pPeer *p2pModule = newP2PModule()

	var mx sync.Mutex
	var wg sync.WaitGroup

	var peers []*testPeer
	var peerList *types.Peerlist // TODO(derrandz): rename type Peerlist to PeerList

	var config *shared.P2PConfig

	{ // Prepare mock peers and add them to the peer list
		peers = make([]*testPeer, 0)
		peerList = generatePeerList(27)

		// launch tcp servers for each mock peer
		for i, p := range peerList.Slice()[1:] {
			wg.Add(1)
			go func(i int, p types.Peer) {
				ready, done, data, respond := ListenAndServe(p.Addr(), int(BufferSize), 100)

				fmt.Println(p)
				mx.Lock()
				peers = append(peers, &testPeer{
					id:      p.Id(),
					address: p.Addr(),
					ready:   ready,
					done:    done,
					data:    data,
					respond: respond,
				})
				mx.Unlock()

				<-ready   // wait for the tcp server to be ready
				wg.Done() // notify waitgroup
			}(i, p)
		}
	}

	wg.Wait() // wait for the tcp servers to come up

	{ // prepare the p2p peer
		p := peerList.Get(0)

		config = &shared.P2PConfig{
			Protocol:         "tcp",
			Address:          []byte(p.Addr()),
			ExternalIp:       p.Addr(),
			Peers:            extractIps(peerList),
			MaxInbound:       100,
			MaxOutbound:      100,
			BufferSize:       BufferSize,
			WireHeaderLength: WireByteHeaderLength,
			TimeoutInMs:      2000,
		}

		err := p2pPeer.initialize(config)

		if err != nil {
			panic(fmt.Sprintf("Failed to setup test: %s", err))
		}

		p2pPeer.id = p.Id()
		p2pPeer.setLogger(fmt.Println)

		go p2pPeer.listen()

		<-p2pPeer.ready
	}

	return p2pPeer, peers, peerList
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
	p2pPeer, mockPeers, _ := Setup_TestNetworkGossip_Broadcast()

	var wg_ack sync.WaitGroup
	{ // Prepare the mock peers such that they ACK the received broadcast message
		for _, peer := range mockPeers {
			wg_ack.Add(1)
			go sendAckOnBroadcastReceived(&wg_ack, peer, p2pPeer.externaladdr)
		}
	}

	var wg sync.WaitGroup
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
	time.After(time.Millisecond * 100)
	for _, peer := range mockPeers {
		peer.done <- 1
	}

	// wait on ACK routines to wrap up after the signal
	wg_ack.Wait()

	{ // Assert that the broadcast was carried out successfully

		assert.Nilf(
			t,
			broadcastErr,
			"Broadcast: Encountered error while broadcasting: %s", broadcastErr,
		)

		actualInfectedPeers := getInfectedPeers(mockPeers)
		expectedInfectedPeers := getExpectedInfectedPeers(p2pPeer.id, p2pPeer.peerlist)

		fmt.Println("Expected Infected Peers", expectedInfectedPeers)
		fmt.Println("Actual Infected Peers", actualInfectedPeers)

		for id, _ := range expectedInfectedPeers {
			_, wasInfected := actualInfectedPeers[id]
			assert.Equalf(
				t,
				true,
				wasInfected,
				"Broadcast: Expected peer with id %d to be impacted, it was not. Impacted peers were: %v", id, actualInfectedPeers,
			)
		}
	}

	Teardown_TestNetworkGossip_Broadcast(p2pPeer, mockPeers)
}

func Teardown_TestNetworkGossip_Broadcast(p2pPeer *p2pModule, mockPeers []*testPeer) {
	p2pPeer.Stop()
	<-p2pPeer.done
	for i, p := range mockPeers {
		close(p.done)
	}
}

func TestNetworkGossip_HandleBroadcast(t *testing.T) {
	t.Skip()
}

func getInfectedPeers(l []*testPeer) map[uint64]bool {
	infectedPeers := map[uint64]bool{}
	for _, p := range l {
		if p.infected {
			infectedPeers[p.id] = true
		}
	}
	return infectedPeers
}

func getExpectedInfectedPeers(originatorId uint64, list *types.Peerlist) map[uint64]bool {
	expectedPeers := map[uint64]bool{}

	act := func(id uint64, l, r *types.Peer, currentlevel int) error {
		lid := l.Id()
		rid := r.Id()
		expectedPeers[lid] = true
		expectedPeers[rid] = true
		return nil
	}

	fmt.Println("<<<<<< starting parameters for expectation mock rain:", originatorId, list.Size(), true, 0)
	fmt.Println("<<<<<<<list:", list)
	rain(originatorId, list, act, true, 0)

	return expectedPeers
}

func generatePeerList(length int) *types.Peerlist {
	peerList := types.NewPeerlist()

	for i := 0; i < length; i++ {
		pId := uint64(i + 1)
		pIp := fmt.Sprintf("127.0.0.1:110%d", i+1)

		peerList.Add( // TODO(derrandz): why not add a pointer
			*types.NewPeer(pId, pIp),
		)
	}

	return peerList
}

func extractIps(pl *types.Peerlist) []string {
	ips := []string{}
	for _, peer := range pl.Slice() {
		ips = append(ips, peer.Addr())
	}
	return ips
}

func sendAckOnBroadcastReceived(wg *sync.WaitGroup, p *testPeer, broadcasterAddr string) {
	wCodec := newWireCodec()
waiter:
	for {
		select {
		case d := <-p.data:
			p.infected = true
			nonce, _, _, _, err := wCodec.decode(d.buff)
			if err != nil {
				wg.Done()
				panic(fmt.Sprintf("Fatal!: Mock peer failed to decode received broadcast message."))
			}
			ack, err := MakeEncodedAckMsg(nonce, p.address, broadcasterAddr)
			<-time.After(time.Millisecond * 2) // a few mock peers time out on ACKs, this aleviates this problem.
			p.respond <- ack
		case <-p.done:
			break waiter
		}
	}
	wg.Done()
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
