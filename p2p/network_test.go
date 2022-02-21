package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	stdio "io"
	"net"
	"pocket/p2p/types"
	"sync"
	"testing"
	"time"
)

func TestNetwork_NewP2PModule(t *testing.T) {
	g := NewP2PModule()
	g.peerlist = types.NewPeerlist()
	if g.peerlist.Size() == 0 && g.inbound.maxcap == uint32(MaxInbound) && g.outbound.maxcap == uint32(MaxOutbound) {
		t.Log("Success!")
	} else {
		t.Logf("Gater is malconfigured")
		t.Failed()
	}
}

func TestNetwork_ListenStop(t *testing.T) {
	g := NewP2PModule()
	g.address = "localhost:12345" // g.config
	go g.listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Error listening: gater not ready yet")
	}

	if !g.listening.Load() {
		t.Errorf("Error listening: flag shows false after start")
	}

	g.close()

	_, finished := <-g.closed

	if !finished {
		t.Errorf("Error: not closed after .Close()")
	}

	if err := g.err.error; err != nil {
		t.Errorf("Error listening: %s", err.Error())
	}

	if g.listening.Load() {
		t.Errorf("Error listening: flag shows true after stop")
	}

	g.listener.Lock()
	if g.listener.TCPListener != nil {
		t.Errorf("Error: listener is still active")
	}
	g.listener.Unlock()
}

func TestNetwork_SendOutbound(t *testing.T) {
	g := NewP2PModule()
	g.configure("tcp", "0.0.0.0:3030", "0.0.0.0:3030", []string{})
	go g.listen()

	select {
	case <-g.ready:
	case <-g.errored:
		t.Errorf("Send error: could not start listening, error: %s", g.err.error.Error())
	}

	addr := "0.0.0.0:2111"
	msg := []byte("hello")

	ready, _, data, _ := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	err := g.send(addr, msg, false)

	if err != nil {
		t.Errorf("Send error: Failed to write message to target")
	}

	pipe, exists := g.outbound.find(addr)

	if !exists {
		t.Errorf("Send error: outbound connection not registered")
	}

	if _, down := <-pipe.ready; down != false {
		t.Errorf("Send error: pipe is not ready")
	}

	if !pipe.opened.Load() {
		t.Errorf("Send error: pipe is not open")
	}

	received := <-data
	if received.err != nil {
		t.Errorf("Send error: recipient has received an error while receiving: %s", received.err.Error())
	}

	if bytes.Compare(received.buff[WireByteHeaderLength:], msg) != 0 {
		t.Errorf("Send error: recipient received a corrupted message")
	}
}

func TestNetwork_SendInbound(t *testing.T) {
	g := NewP2PModule()
	g.configure("tcp", "127.0.0.1:30303", "127.0.0.1:30303", []string{})
	go g.listen()
	select {

	case <-g.ready:
	case <-g.errored:
		t.Errorf("Send error: could not start listening, error: %s", g.err.error.Error())
	}

	conn, err := net.Dial("tcp", g.address)

	if err != nil {
		t.Errorf("Send error: could not dial gater")
	}

	<-time.After(time.Millisecond * 10) // let gater catch up and store the new inbound conn

	msg := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
	sent := make(chan int)
	go func() {
		err = g.send(conn.LocalAddr().String(), msg, false)
		sent <- 1
	}()

	<-sent

	if err != nil {
		t.Errorf("Send error: Failed to write message to target")
	}

	pipe, exists := g.inbound.find(conn.LocalAddr().String())

	if !exists {
		t.Errorf("Send error: outbound connection not registered")
	}

	if _, down := <-pipe.ready; down != false {
		t.Errorf("Send error: pipe is not ready")
	}

	if !pipe.opened.Load() {
		t.Errorf("Send error: pipe is not open")
	}

	received := make([]byte, ReadBufferSize)
	n, err := conn.Read(received)
	if err != nil {
		t.Errorf("Send error: recipient has received an error while receiving: %s", err.Error())
	}

	if bytes.Compare(received[WireByteHeaderLength:n], msg) != 0 {
		t.Errorf("Send error: recipient received a corrupted message")
	}
}

func TestNetwork_Request(t *testing.T) {
	g := NewP2PModule()

	g.configure("tcp", "0.0.0.0:4030", "0.0.0.0:4030", []string{})
	go g.listen()
	t.Logf("Started listeningy")
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after.listen")
	}

	t.Logf("Started listenig: OK")

	addr := "localhost:22302"
	ready, _, data, respond := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	t.Logf("Recipient: OK")

	fmt.Println("listening started")
	// send request to addr
	msgA := GenerateByteLen((1024 * 4) - WireByteHeaderLength)

	responses := make(chan []byte)
	errs := make(chan error, 10)

	go func() {
		fmt.Println("Requesting")
		res, err := g.request(addr, msgA, false)
		if err != nil {
			errs <- err
			close(responses)
		}
		responses <- res
	}()

	c := &wcodec{}

	t.Logf("Receiving...")
	d := <-data
	t.Logf("Received: OK")
	nonce, encoding, _, _, err := c.decode(d.buff)

	if err != nil {
		t.Errorf("Request error:  %s", err.Error())
	}

	fmt.Println("Where are you blocked?")
	respond <- c.encode(encoding, false, nonce, msgA, false)

	pipe, _ := g.outbound.find(addr)
	select {
	case err := <-errs:
		t.Errorf("Request error:  %s", err.Error())
	case <-pipe.errored:
		t.Errorf("Request error: error while receiving a response: %s", pipe.err.error.Error())

	case d, open := <-responses:
		if !open {
			err := <-errs
			t.Errorf("Request error: error while receiving a response: %s", err.Error())
		}

		if len(d) != ReadBufferSize-WireByteHeaderLength {
			t.Errorf("Request error: received response buffer length mistach")
		}

		if bytes.Compare(d, msgA) != 0 {
			t.Errorf("Request error: received response buffer is corrupted")
		}
	}
}

func TestNetwork_Respond(t *testing.T) {
	g := NewP2PModule()

	g.configure("tcp", "0.0.0.0:4031", "0.0.0.0:4031", []string{})
	go g.listen()
	t.Logf("Listening...")
	_, waiting := <-g.ready
	t.Logf("Listening: OK")

	if waiting {
		t.Errorf("Request error: gater still not started after.listen")
	}

	t.Logf("Listening: OK")

	<-time.After(time.Millisecond * 2)
	conn, err := net.Dial(g.protocol, g.address)

	if err != nil {
		t.Errorf("Failed to dial gater. Error: %s", err.Error())
	}

	t.Logf("Dial: OK")
	// send to the gater a nonced message (i.e: request)
	addr := conn.LocalAddr().String()
	requestNonce := 12
	msgA := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
	msgB := GenerateByteLen((1024 * 4) - WireByteHeaderLength)

	go func() {
		c := &wcodec{}
		request := c.encode(Binary, false, 12, msgA, false)
		_, err := conn.Write(request)
		t.Logf("Write: OK")
		fmt.Println(err)
	}()

	<-time.After(time.Millisecond * 5)

	t.Logf("Receiving...")
	w := <-g.sink // blocks
	t.Logf("Receiving: OK")

	nonce := w.Nonce()
	t.Logf("Got work %d", nonce)

	err = g.respond(nonce, false, addr, msgB, false)
	if err != nil {
		t.Errorf("Respond error: %s", err.Error())
	}

	buff := make([]byte, ReadBufferSize)
	_, err = conn.Read(buff)

	if err != nil {
		t.Errorf("Respond error: peer failed to read gater's response")
	}

	dnonce, _, decoded, _, err := (&wcodec{}).decode(buff)
	if err != nil {
		t.Errorf("Respond error: could not decode payload. Encountered following error: %s", err.Error())
	}

	if dnonce != uint32(requestNonce) {
		t.Errorf("Respond error: received wrong nonce")
	}

	if len(decoded) != ReadBufferSize-WireByteHeaderLength {
		t.Errorf("Respond error: received response buffer length mistach")
	}

	if bytes.Compare(decoded, msgB) != 0 {
		t.Errorf("Respond error: received response buffer is corrupted")
	}
}

func TestNetwork_Ping(t *testing.T) {
	g := NewP2PModule()

	g.configure("tcp", "0.0.0.0:4032", "0.0.0.0:4032", []string{})
	g.init()

	err := g.init()

	if err != nil {
		t.Logf("Error: failed to initialize gater. %s", err.Error())
	}

	go g.listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	addr := "127.0.0.1:2313"
	ready, _, data, respond := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	// ping addr
	errors := make(chan error)
	responses := make(chan bool)
	go func() {
		t.Logf("Pinging...")
		alive, err := g.ping(addr)
		if err != nil {
			t.Logf("Ping: failed. %s", err.Error())
			errors <- err
		}
		t.Logf("Ping: OK")
		responses <- alive
	}()

	<-time.After(time.Microsecond * 10)
	t.Logf("Receiving...")
	c := &wcodec{}

	select {
	case err := <-errors:
		t.Errorf("err: %s", err.Error())

	case d := <-data:
		{
			t.Logf("Received: OK")
			nonce, encoding, buff, _, err := c.decode(d.buff)
			if err != nil {
				t.Errorf("Ping error: Encountered error while decoding received ping: %s", err.Error())
			}

			m, err := g.c.decode(buff)
			if err != nil {
				t.Errorf("Ping error: failed to decode ping on receipt (domain codec). Encountered error: %s", err.Error())
			}

			msg := m.(types.NetworkMessage)

			if msg.Topic != types.PocketTopic_P2P_PING {
				t.Errorf("Ping error: peer expecte to receive ping message, got %s instead", msg.Topic)
			}

			pongmsg := Message(int32(nonce), 0, types.PocketTopic_P2P_PONG, addr, g.address)
			encoded, err := g.c.encode(pongmsg)

			if err != nil {
				t.Errorf("Ping error: Encountered error while encoding pong message: %s", err.Error())
			}

			respond <- c.encode(encoding, false, nonce, encoded, false)
		}

	case alive, open := <-responses:
		if !open {
			err := <-errors
			t.Errorf("Ping error: error while receiving a response: %s", err.Error())
		}

		if alive != true {
			t.Errorf("Ping error: expected peer to be alive, got the following instead: alive=%v", alive)
		}

	}
}

func TestNetwork_Pong(t *testing.T) {
	g := NewP2PModule()

	g.configure("tcp", "0.0.0.0:4033", "0.0.0.0:4033", []string{})

	err := g.init()
	if err != nil {
		t.Logf("Error: failed to initialize gater. %s", err.Error())
	}

	msg := *Message(0, 0, types.PocketTopic_P2P_PONG, "", g.address)

	go g.listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	addr := "127.0.0.1:22312"
	ready, _, data, _ := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	// ping addr
	var perr *error

	pongnonce := uint32(1)
	signals := make(chan int)
	go func() {
		msg := Message(0, 0, types.PocketTopic_P2P_PING, "", g.address)

		err := g.pong(*msg)
		if err != nil {
			perr = &err
		}
		signals <- 1
	}()

	<-signals

	if perr != nil {
		t.Errorf("Pong error: failed to send a pong message. Error: %s", (*perr).Error())
	}

	d := <-data

	nonce, _, buff, _, err := (&wcodec{}).decode(d.buff)
	if err != nil {
		t.Errorf("Pong error: faield to decode wire bytes received from pong message. Error: %s", err.Error())
	}

	decoded, err := g.c.decode(buff)
	if err != nil {
		t.Errorf("Pong error: failed to decode received pong message. Error: %s", err.Error())
	}

	msg = decoded.(types.NetworkMessage)
	if msg.Topic != types.PocketTopic_P2P_PONG {
		t.Errorf("Pong error: expected to receive a message with action=%s, got: %s instead.", types.PocketTopic_P2P_PONG, msg.Topic)
	}

	if nonce != 1 {
		t.Errorf("Pong error: wrong nonce, expected %d, got: %d", pongnonce, nonce)
	}
}

func TestNetwork_Broadcast(t *testing.T) {
	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	var m sync.Mutex
	var wg sync.WaitGroup

	var rw sync.RWMutex
	receivedMessages := map[uint64][][]byte{}

	iolist := make([]struct {
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
	}, 0)

	fmt.Println("Startin hereg")
	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.Add(*p)
	}

	fmt.Println("Starting")
	// mark gater as peer with id=1
	p := list.Get(0)
	g := NewP2PModule()

	g.id = p.Id()
	g.address = p.Addr()
	g.externaladdr = p.Addr()
	g.peerlist = list

	err := g.init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if g.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	g.setLogger(fmt.Println)
	fmt.Println("here?")

	for i, p := range list.Slice()[1:] {
		wg.Add(1)
		go func(i int, p types.Peer) {
			ready, done, data, respond := ListenAndServe(p.Addr(), ReadBufferSize)
			<-ready

			m.Lock()
			iolist = append(iolist, struct {
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
			}{
				id:      p.Id(),
				address: p.Addr(),
				ready:   ready,
				done:    done,
				data:    data,
				respond: respond,
			})
			receivedMessages[p.Id()] = make([][]byte, 0)

			m.Unlock()

			wg.Done()
		}(i, p)
	}

	wg.Wait()

	go g.listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if !g.listening.Load() {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	gossipdone := make(chan int)
	go func() {
		<-g.ready
		m := Message(int32(0), int32(0), types.PocketTopic_CONSENSUS, g.address, "")

		fmt.Println("Starting gossip")
		g.broadcast(m, true)
		gossipdone <- 1
	}()
	go func() {
		<-g.sink
	}()

	fanin := make(chan struct {
		n    int
		err  error
		buff []byte
	}, 30)

	for i, io := range iolist {
		e := io
		go func(i int) {
		waiter: // a node might receive more than once
			for {

				select {
				case d := <-e.data:
					rw.Lock()
					receivedMessages[e.id] = append(receivedMessages[e.id], d.buff)
					rw.Unlock()
					fanin <- d
					fmt.Println("Sending back", len(d.buff), e.address)
					e.respond <- d.buff
					fmt.Println("Should now have responded")

				case <-e.done:
					break waiter

				default:
				}

			}
		}(i)
	}

fan:
	for {
		select {
		case <-fanin:
		case <-gossipdone:
			for _, io := range iolist {
				io.done <- 1
			}

			break fan
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
		{4, 5},  // target list size at level 3 = 18, left = 7, right = 13
		{6, 7},  // target list size at level 2 = 35, left = 13, right = 25 (rolling over involved)
		{9, 13}, // target list size at level 2 = 52, left = 9, right = 19 (rolling over involved)
	}

	fmt.Println(recipients)

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

func TestNetwork_HandleBroadcast(t *testing.T) {
	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	var m sync.Mutex
	var wg sync.WaitGroup

	var rw sync.RWMutex
	receivedMessages := map[uint64][][]byte{}

	iolist := make([]struct {
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
	}, 0)

	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.Add(*p)
	}

	// mark gater as peer with id=1
	p := list.Get(0)
	g := NewP2PModule()

	g.id = p.Id()
	g.address = p.Addr()
	g.externaladdr = p.Addr()
	g.peerlist = list

	err := g.init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if g.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	g.setLogger(fmt.Println)
	for i, p := range list.Slice()[1:] {
		if p.Id() != g.id {
			wg.Add(1)
			go func(i int, p types.Peer) {
				ready, done, data, respond := ListenAndServe(p.Addr(), ReadBufferSize)
				<-ready

				m.Lock()
				iolist = append(iolist, struct {
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
				}{
					id:      p.Id(),
					address: p.Addr(),
					ready:   ready,
					done:    done,
					data:    data,
					respond: respond,
				})
				receivedMessages[p.Id()] = make([][]byte, 0)

				m.Unlock()

				wg.Done()
			}(i, p)
		}
	}

	wg.Wait()

	go g.listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if !g.listening.Load() {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	gossipdone := make(chan int, 1)
	go func() {
		<-g.ready
		g.on(types.BroadcastDoneEvent, func(args ...interface{}) {
			gossipdone <- 1
		})
		g.handle()
	}()

	fanin := make(chan struct {
		n    int
		err  error
		buff []byte
	}, 30)

	for i, io := range iolist {
		e := io

		go func(i int) {

		waiter: // a node might receive more than once
			for {

				select {
				case d := <-e.data:
					rw.Lock()
					receivedMessages[e.id] = append(receivedMessages[e.id], d.buff)
					rw.Unlock()

					fmt.Println(e.address, "received data", len(d.buff))
					fanin <- d

					nonce, _, _, _, err := (&wcodec{}).decode(d.buff)
					fmt.Println("Err", err)
					ack := Message(int32(nonce), int32(0), types.PocketTopic_CONSENSUS, e.address, g.address)
					eack, _ := g.c.encode(*ack)
					wack := (&wcodec{}).encode(Binary, false, nonce, eack, true)

					e.respond <- wack
				case <-e.done:
					break waiter

				default:
				}
			}
		}(i)
	}

	conn, _ := net.Dial("tcp", g.address)

	gm := Message(int32(0), int32(4), types.PocketTopic_CONSENSUS, conn.LocalAddr().String(), g.address)
	egm, _ := g.c.encode(*gm)
	wgm := (&wcodec{}).encode(Binary, false, 0, egm, true)

	conn.Write(wgm)
	fmt.Println("Has written the size of", len(wgm))
	buff := make([]byte, ReadBufferSize)
	conn.Read(buff)
	fmt.Println("Acked", len(buff))
	conn.Close()

	select {
	case <-gossipdone:
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

/*
 @
 @ Utils
 @
*/

func ListenAndServe(addr string, readbufflen int) (ready, done chan uint, data chan struct {
	n    int
	err  error
	buff []byte
}, response chan []byte) {

	ready = make(chan uint)
	done = make(chan uint)
	data = make(chan struct {
		n    int
		err  error
		buff []byte
	}, 10)
	response = make(chan []byte, 1)

	datapoint := func(n int, err error, buff []byte) struct {
		n    int
		err  error
		buff []byte
	} {
		return struct {
			n    int
			err  error
			buff []byte
		}{n: n, err: err, buff: buff}
	}

	readwriteconn := func(c net.Conn) {
		readerClosed := false

		codec := (&wcodec{})
		creader := bufio.NewReader(c)
		buffer := make([]byte, readbufflen)

	reader:
		for {
			select {
			case <-done:
				break reader

			case msg := <-response:
				_, err := c.Write(msg)
				if err != nil {
					close(ready)
					close(done)
					close(data)
					close(response)
				}

			default:
				if readerClosed {
					continue reader
				}
				c.SetReadDeadline(time.Now().Add(time.Millisecond * 2))
				n, err := stdio.ReadFull(creader, buffer[:WireByteHeaderLength])
				if err != nil {
					if isErrTimeout(err) {
						readerClosed = true
						continue reader
					}
					data <- datapoint(n, err, buffer)
					break reader
				}

				_, _, bodylength, derr := codec.decodeHeader(buffer[:WireByteHeaderLength])
				if derr != nil {
					data <- datapoint(len(buffer), err, buffer)
					break reader
				}

				n, err = stdio.ReadAtLeast(creader, buffer[WireByteHeaderLength:], int(bodylength))
				if err != nil {
					if isErrEOF(err) {
						err = nil
					}

					if isErrTimeout(err) {
						readerClosed = true
						continue
					}

					data <- datapoint(n, err, buffer)
					break reader
				}

				if n > 0 {
					hl := WireByteHeaderLength
					bl := int(bodylength)
					buff := buffer[:hl+bl]
					data <- datapoint(n, err, buff)
				}
			}
		}
		close(ready)
	}
	accept := func() {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			ready <- 0
			return
		}
		ready <- 1

	listener:
		for {
			select {
			case <-done:
				break listener
			default:
			}

			conn, err := l.Accept()
			if err != nil {
				ready <- 0
			}

			go readwriteconn(conn)
		}
	}

	go accept()

	return
}
