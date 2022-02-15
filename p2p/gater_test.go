package p2p

import (
	"bytes"
	"fmt"
	"net"
	"pocket/shared/messages"
	"sync"
	"testing"
	"time"
)

func TestNewGater(t *testing.T) {
	g := NewGater()
	if len(g.peerlist) == 0 && g.inbound.maxcap == uint32(MaxInbound) && g.outbound.maxcap == uint32(MaxOutbound) {
		t.Log("Success!")
	} else {
		t.Logf("Gater is malconfigured")
		t.Failed()
	}
}

func TestListenStop(t *testing.T) {
	g := NewGater()
	g.address = "localhost:12345" // g.Config
	go g.Listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Error listening: gater not ready yet")
	}

	if g.listening != true {
		t.Errorf("Error listening: flag shows false after start")
	}

	g.Close()

	_, finished := <-g.closed

	if finished {
		t.Log("Server closed")
	}

	if err := g.err; err != nil {
		t.Errorf("Error listening: %s", err.Error())
	}

	if g.listener != nil {
		t.Errorf("Error: listener is still active")
	}

	if g.listening != false {
		t.Errorf("Error listening: flag shows true after stop")
	}
}

func TestSendOutbound(t *testing.T) {
	g := NewGater()
	go g.Listen()
	<-g.ready

	addr := "localhost:20202"
	msg := []byte("hello")

	ready, _, data, _ := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	err := g.Send(addr, msg, false)

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

	if pipe.opened != true {
		t.Errorf("Send error: pipe is not open")
	}

	if pipe.receiving != true && pipe.sending != true {
		t.Errorf("Send error: pipe is neither sending nor receiving")
	}

	received := <-data
	if received.err != nil {
		t.Errorf("Send error: recipient has received an error while receiving: %s", err.Error())
	}

	if bytes.Compare(received.buff[WireByteHeaderLength:received.n], msg) != 0 {
		t.Errorf("Send error: recipient received a corrupted message")
	}
}

func TestSendInbound(t *testing.T) {
	g := NewGater()
	go g.Listen()
	<-g.ready

	conn, err := net.Dial("tcp", g.address)

	if err != nil {
		t.Errorf("Send error: could not dial gater")
	}

	<-time.After(time.Millisecond * 10) // let gater catch up and store the new inbound conn

	msg := GenerateByteLen(1024 * 4)
	sent := make(chan int)
	go func() {
		err = g.Send(conn.LocalAddr().String(), msg, false)
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

	if pipe.opened != true {
		t.Errorf("Send error: pipe is not open")
	}

	if pipe.receiving != true && pipe.sending != true {
		t.Errorf("Send error: pipe is neither sending nor receiving")
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

func TestRequest(t *testing.T) {
	g := NewGater()

	go g.Listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	addr := "localhost:22302"
	ready, _, data, respond := ListenAndServe(addr, ReadBufferSize)

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	// send request to addr
	msgA := GenerateByteLen(1024 * 4)
	responses := make(chan []byte)
	errs := make(chan error, 10)
	go func() {
		res, err := g.Request(addr, msgA, false)
		if err != nil {
			errs <- err
			close(responses)
		}
		responses <- res
	}()

	go func() {
		c := &wcodec{}
		d := <-data
		nonce, encoding, _, _, err := c.decode(d.buff)
		if err != nil {
			fmt.Println("Error decoding", err.Error())
		}
		respond <- c.encode(encoding, false, nonce, msgA, false)
	}()

	d, open := <-responses
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

func TestRespond(t *testing.T) {
	g := NewGater()

	go g.Listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	<-time.After(time.Millisecond * 20)
	conn, err := net.Dial(g.protocol, g.address)

	if err != nil {
		t.Errorf("Failed to dial gater. Error: %s", err.Error())
	}

	// send to the gater a nonced message (i.e: request)
	addr := conn.LocalAddr().String()
	requestNonce := 12
	msgA := GenerateByteLen(1024 * 4)
	msgB := GenerateByteLen(1024 * 4)
	errs := make(chan error, 10)
	signals := make(chan int)

	go func() {
		<-time.After(time.Millisecond * 50)
		w := <-g.sink
		nonce := w.nonce

		err := g.Respond(nonce, false, addr, msgB, false)
		if err != nil {
			errs <- err
		}
		signals <- 1
	}()

	go func() {
		c := &wcodec{}
		request := c.encode(Binary, false, 12, msgA, false)
		conn.Write(request)
	}()

	<-signals
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

func TestPing(t *testing.T) {
	g := NewGater()

	g.address = "127.0.0.1:30303"
	g.Init()

	go g.Listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	addr := "127.0.0.1:22302"
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
		alive, err := g.Ping(addr)
		if err != nil {
			errors <- err
		}
		responses <- alive
	}()

	sink := make(chan []byte)
	<-time.After(time.Millisecond * 20)
	go func() {
		c := &wcodec{}
		d := <-data
		nonce, encoding, buff, _, err := c.decode(d.buff)
		if err != nil {
			errors <- err
		}
		sink <- buff

		if err != nil {
			errors <- err
		}

		pongmsg := (&churnmgmt{}).message(nonce, Pong, 0, addr, g.address)
		encoded, err := g.c.encode(pongmsg)

		if err != nil {
			errors <- err
		}

		respond <- c.encode(encoding, false, nonce, encoded, false)
	}()

	select {

	case err := <-errors:
		t.Errorf("Ping error: Encountered error: %s", err.Error())
		t.Failed()

	case buff := <-sink:
		{
			m, err := g.c.decode(buff)
			if err != nil {
				t.Errorf("Ping error: failed to decode ping on receipt (domain codec). Encountered error: %s", err.Error())
			}

			msg := m.(message)

			if msg.action != Ping {
				t.Errorf("Ping error: peer expecte to receive ping message, got %s instead", msg.action)
			}

		}

	case alive, open := <-responses:
		{
			if !open {
				err := <-errors
				t.Errorf("Ping error: error while receiving a response: %s", err.Error())
			}

			if alive != true {
				t.Errorf("Ping error: expected peer to be alive, got the following instead: alive=%v", alive)
			}
		}
	}
}

func TestPong(t *testing.T) {
	g := NewGater()

	g.address = "127.0.0.1:30303"
	g.Init()

	go g.Listen()
	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Request error: gater still not started after Listen")
	}

	addr := "127.0.0.1:22302"
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
		msg := (&churnmgmt{}).message(pongnonce, Ping, 0, addr, g.address)
		err := g.Pong(msg)
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

	msg := decoded.(message)
	if msg.action != Pong {
		t.Errorf("Pong error: expected to receive a message with action=%s, got: %s instead.", Pong, msg.action)
	}

	if nonce != 1 {
		t.Errorf("Pong error: wrong nonce, expected %d, got: %d", pongnonce, nonce)
	}
}

func TestBroadcast(t *testing.T) {
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

	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.add(*p)
	}

	// mark gater as peer with id=1
	p := list.get(0)
	g := NewGater()

	g.id = p.id
	g.address = p.address
	g.peerlist = *list

	err := g.Init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if g.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	for i, p := range list.slice()[1:] {
		wg.Add(1)
		go func(i int, p peer) {
			ready, done, data, respond := ListenAndServe(p.address, ReadBufferSize)
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
				id:      p.id,
				address: p.address,
				ready:   ready,
				done:    done,
				data:    data,
				respond: respond,
			})
			receivedMessages[p.id] = make([][]byte, 0)

			m.Unlock()

			wg.Done()
		}(i, p)
	}

	wg.Wait()

	go g.Listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if g.listening != true {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	gossipdone := make(chan int)
	go func() {
		<-g.ready
		m := (&pbuff{}).message(int32(0), int32(0), messages.PocketTopic_P2P, g.address, "")
		g.Broadcast(m, true)
		gossipdone <- 1
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
					e.respond <- d.buff

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

func TestHandleBroadcast(t *testing.T) {
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

	list := &plist{}

	for i := 0; i < 27; i++ {
		p := Peer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.add(*p)
	}

	// mark gater as peer with id=1
	p := list.get(0)
	g := NewGater()

	g.id = p.id
	g.address = p.address
	g.peerlist = *list

	err := g.Init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if g.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	for i, p := range list.slice()[1:] {
		wg.Add(1)
		go func(i int, p peer) {
			ready, done, data, respond := ListenAndServe(p.address, ReadBufferSize)
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
				id:      p.id,
				address: p.address,
				ready:   ready,
				done:    done,
				data:    data,
				respond: respond,
			})
			receivedMessages[p.id] = make([][]byte, 0)

			m.Unlock()

			wg.Done()
		}(i, p)
	}

	wg.Wait()

	go g.Listen()

	_, waiting := <-g.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if g.listening != true {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	go func() {
		<-g.ready
		g.Handle()
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

					nonce, _, _, _, _ := (&wcodec{}).decode(d.buff)
					ack := (&gossip{}).message(nonce, GossipACK, 0, e.address, g.address)
					eack, _ := g.c.encode(ack)
					eack = append(eack, make([]byte, ReadBufferSize-len(eack))...)
					wack := (&wcodec{}).encode(Binary, false, nonce, eack, true)

					e.respond <- wack[:ReadBufferSize]
					//e.respond <- d.buff
					<-time.After(time.Millisecond * 300)

				case <-e.done:
					break waiter

				default:
				}
			}
		}(i)
	}

	conn, _ := net.Dial("tcp", g.address)

	level := uint16(4)
	gm := (&gossip{}).message(0, Gossip, level, conn.LocalAddr().String(), g.address)
	egm, _ := g.c.encode(gm)
	egm = append(egm, make([]byte, ReadBufferSize-len(egm))...)
	wgm := (&wcodec{}).encode(Binary, false, 0, egm, true)

	conn.Write(wgm)
	buff := make([]byte, ReadBufferSize)
	conn.Read(buff)
	fmt.Println("Acked", len(buff))
	conn.Close()

	counter := 0
closer:
	for {
		select {

		case <-fanin:
			counter++
		default:
			if counter > 35 {

				for _, io := range iolist {
					io.done <- 1
				}
				break closer
			}
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
	rw := sync.RWMutex{}

	go func() {
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
				fmt.Println("New connection from", conn.RemoteAddr().String())
				ready <- 0
			}

			go func(c net.Conn) {
				buff := make([]byte, readbufflen)
			reader:
				for {

					rw.Lock()
					select {
					case <-done:
						fmt.Println("A done")
						break reader
					case msg := <-response:
						fmt.Println("A write")
						_, err := c.Write(msg)
						fmt.Println("Written", err)
						if err != nil {
							close(ready)
							close(done)
							close(data)
							close(response)
						}
					default:
						c.SetReadDeadline(time.Now().Add(time.Millisecond * 2))
						n, err := c.Read(buff)
						cp := append(make([]byte, 0), buff...)
						data <- struct {
							n    int
							err  error
							buff []byte
						}{n, err, cp}
					}
					rw.Unlock()
				}
			}(conn)
		}
	}()
	return
}
