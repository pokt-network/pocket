package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	stdio "io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
	shared "github.com/pokt-network/pocket/shared/config"
	common "github.com/pokt-network/pocket/shared/types"
)

func TestNetwork_newP2PModule(t *testing.T) {
	m := newP2PModule()
	m.peerlist = types.NewPeerlist()
	m.config = &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       200,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}

	if m.peerlist.Size() == 0 && m.inbound.Capacity() == uint32(100) && m.outbound.Capacity() == uint32(100) {
		t.Log("Success!")
	} else {
		t.Errorf("Gater is malconfigured")
	}
}

func TestNetwork_ListenStop(t *testing.T) {
	m := newP2PModule()
	m.address = "localhost:12345" // m.config
	m.protocol = "tcp"
	go m.listen()

	_, waiting := <-m.ready

	if waiting {
		t.Errorf("Error listening: gater not ready yet")
	}

	if !m.isListening.Load() {
		t.Errorf("Error listening: flag shows false after start")
	}

	t.Log("Server listeninm.")

	m.close()

	_, finished := <-m.closed

	if !finished {
		t.Errorf("Error: not closed after .Close()")
	}

	if err := m.err.error; err != nil {
		t.Errorf("Error listening: %s", err.Error())
	}

	if m.isListening.Load() {
		t.Errorf("Error listening: flag shows true after stop")
	}

	m.listener.Lock()
	if m.listener.TCPListener != nil {
		t.Errorf("Error: listener is still active")
	}
	m.listener.Unlock()

	t.Log("Server closed.")
}

func TestNetwork_SendOutbound(t *testing.T) {
	m := newP2PModule()
	m.config = &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}

	m.configure("tcp", "0.0.0.0:3030", "0.0.0.0:3030", []string{})
	go m.listen()

	select {
	case <-m.isReady():
	case <-m.errored:
		t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
	}

	addr := "0.0.0.0:2111"
	msg := []byte("hello")

	ready, _, data, _ := ListenAndServe(addr, int(m.config.BufferSize))

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	err := m.send(addr, msg, false)

	if err != nil {
		t.Errorf("Send error: Failed to write message to target")
	}

	var pipe *socket
	obj, exists := m.outbound.Find(addr)
	pipe = obj.(*socket)

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

	if bytes.Compare(received.buff[m.config.WireHeaderLength:], msg) != 0 {
		t.Errorf("Send error: recipient received a corrupted message")
	}
}

func TestNetwork_SendInbound(t *testing.T) {
	m := newP2PModule()
	m.config = &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}
	m.configure("tcp", "127.0.0.1:30303", "127.0.0.1:30303", []string{})
	go m.listen()
	select {

	case <-m.ready:
	case <-m.errored:
		t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
	}

	conn, err := net.Dial("tcp", m.address)

	if err != nil {
		t.Errorf("Send error: could not dial gater")
	}

	<-time.After(time.Millisecond * 10) // let gater catch up and store the new inbound conn

	msg := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))
	sent := make(chan int)
	go func() {
		err = m.send(conn.LocalAddr().String(), msg, false)
		sent <- 1
	}()

	<-sent

	if err != nil {
		t.Errorf("Send error: Failed to write message to target")
	}

	var pipe *socket
	obj, exists := m.inbound.Find(conn.LocalAddr().String())
	pipe = obj.(*socket)

	if !exists {
		t.Errorf("Send error: outbound connection not registered")
	}

	if _, down := <-pipe.ready; down != false {
		t.Errorf("Send error: pipe is not ready")
	}

	if !pipe.opened.Load() {
		t.Errorf("Send error: pipe is not open")
	}

	received := make([]byte, m.config.BufferSize)
	n, err := conn.Read(received)
	if err != nil {
		t.Errorf("Send error: recipient has received an error while receiving: %s", err.Error())
	}

	if bytes.Compare(received[m.config.WireHeaderLength:n], msg) != 0 {
		t.Errorf("Send error: recipient received a corrupted message")
	}
}

func TestNetwork_Request(t *testing.T) {
	m := newP2PModule()
	m.config = &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}
	m.configure("tcp", "0.0.0.0:4030", "0.0.0.0:4030", []string{})

	go m.listen()

	t.Logf("Started listeningy")
	_, waiting := <-m.ready

	if waiting {
		t.Errorf("Request error: gater still not started after.listen")
	}

	t.Logf("Started listenig: OK")

	addr := "localhost:22302"
	ready, _, data, respond := ListenAndServe(addr, int(m.config.BufferSize))

	select {
	case v := <-ready:
		if v == 0 {
			t.Errorf("Send error: could not start recipient server")
		}
	}

	t.Logf("Recipient: OK")

	fmt.Println("listening started")
	// send request to addr
	msgA := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))

	responses := make(chan []byte)
	errs := make(chan error, 10)

	go func() {
		fmt.Println("Requesting")
		res, err := m.request(addr, msgA, false)
		if err != nil {
			errs <- err
			close(responses)
		}
		responses <- res
	}()

	c := newWireCodec()

	t.Logf("Receiving...")
	d := <-data
	t.Logf("Received: OK")
	nonce, encoding, _, _, err := c.decode(d.buff)

	if err != nil {
		t.Errorf("Request error:  %s", err.Error())
	}

	fmt.Println("Where are you blocked?")
	respond <- c.encode(encoding, false, nonce, msgA, false)

	var pipe *socket
	obj, _ := m.outbound.Find(addr)
	pipe = obj.(*socket)

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

		if len(d) != int(m.config.BufferSize-m.config.WireHeaderLength) {
			t.Errorf("Request error: received response buffer length mistach")
		}

		if bytes.Compare(d, msgA) != 0 {
			t.Errorf("Request error: received response buffer is corrupted")
		}
	}
}

func TestNetwork_Respond(t *testing.T) {
	m := newP2PModule()
	m.config = &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}
	m.configure("tcp", "0.0.0.0:4031", "0.0.0.0:4031", []string{})
	go m.listen()
	t.Logf("Listening...")
	_, waiting := <-m.ready
	t.Logf("Listening: OK")

	if waiting {
		t.Errorf("Request error: gater still not started after.listen")
	}

	t.Logf("Listening: OK")

	<-time.After(time.Millisecond * 2)
	conn, err := net.Dial(m.protocol, m.address)

	if err != nil {
		t.Errorf("Failed to dial gater. Error: %s", err.Error())
	}

	t.Logf("Dial: OK")
	// send to the gater a nonced message (i.e: request)
	addr := conn.LocalAddr().String()
	requestNonce := 12
	msgA := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))
	msgB := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))

	go func() {
		c := newWireCodec() // two instances better than one with locks (for a test)
		request := c.encode(Binary, false, 12, msgA, false)
		_, err := conn.Write(request)
		t.Logf("Write: OK")
		fmt.Println(err)
	}()

	<-time.After(time.Millisecond * 5)

	t.Logf("Receiving...")
	w := <-m.sink // blocks
	t.Logf("Receiving: OK")

	nonce := w.Nonce()
	t.Logf("Got work %d", nonce)

	err = m.respond(nonce, false, addr, msgB, false)
	if err != nil {
		t.Errorf("Respond error: %s", err.Error())
	}

	buff := make([]byte, m.config.BufferSize)
	_, err = conn.Read(buff)

	if err != nil {
		t.Errorf("Respond error: peer failed to read gater's response")
	}

	c := newWireCodec()
	dnonce, _, decoded, _, err := c.decode(buff)
	if err != nil {
		t.Errorf("Respond error: could not decode payload. Encountered following error: %s", err.Error())
	}

	if dnonce != uint32(requestNonce) {
		t.Errorf("Respond error: received wrong nonce")
	}

	if len(decoded) != int(m.config.BufferSize-m.config.WireHeaderLength) {
		t.Errorf("Respond error: received response buffer length mistach")
	}

	if bytes.Compare(decoded, msgB) != 0 {
		t.Errorf("Respond error: received response buffer is corrupted")
	}
}

func TestNetwork_Broadcast(t *testing.T) {
	// we will have a gater with id = 1
	// it should raintree to the other 27 peers
	// such that it performs SEND/ACK/RESEND on the full list with no redundancy/no cleanup
	// Atm no RESEND on NACK is implemented, so it's just SEND/ACK
	var mx sync.Mutex
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
	config := &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}

	list := types.NewPeerlist()

	for i := 0; i < 27; i++ {
		p := types.NewPeer(uint64(i+1), fmt.Sprintf("127.0.0.1:110%d", i+1))
		list.Add(*p)
	}

	fmt.Println("Starting")
	// mark gater as peer with id=1
	p := list.Get(0)
	m := newP2PModule()

	m.config = config
	m.id = p.Id()
	m.address = p.Addr()
	m.externaladdr = p.Addr()
	m.peerlist = list

	err := m.init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if m.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	m.setLogger(fmt.Println)
	fmt.Println("here?")

	for i, p := range list.Slice()[1:] {
		wg.Add(1)
		go func(i int, p types.Peer) {
			ready, done, data, respond := ListenAndServe(p.Addr(), int(config.BufferSize))
			<-ready

			mx.Lock()
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

			mx.Unlock()

			wg.Done()
		}(i, p)
	}

	wg.Wait()

	go m.listen()

	_, waiting := <-m.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if !m.isListening.Load() {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	gossipdone := make(chan int)
	go func() {
		<-m.ready
		msgpayload := &common.PocketEvent{
			Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
			Data:  nil,
		}
		msg := types.Message(int32(0), int32(0), m.address, "", msgpayload)

		fmt.Println("Starting gossip")
		m.broadcast(msg, true)
		gossipdone <- 1
	}()
	go func() {
		<-m.sink
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
	var mx sync.Mutex
	var wg sync.WaitGroup

	var rw sync.RWMutex
	receivedMessages := map[uint64][][]byte{}

	config := &shared.P2PConfig{
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       1024 * 4,
		WireHeaderLength: 8,
		TimeoutInMs:      200,
	}

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
	m := newP2PModule()

	m.config = config
	m.id = p.Id()
	m.address = p.Addr()
	m.externaladdr = p.Addr()
	m.peerlist = list

	err := m.init()
	if err != nil {
		t.Errorf("Broadcast error: could not initialize gater. Error: %s", err.Error())
	}

	if m.id != 1 {
		t.Errorf("Broadcast error: (test setup error) expected gater to have id 1")
	}

	m.setLogger(fmt.Println)
	for i, p := range list.Slice()[1:] {
		if p.Id() != m.id {
			wg.Add(1)
			go func(i int, p types.Peer) {
				ready, done, data, respond := ListenAndServe(p.Addr(), int(config.BufferSize))
				<-ready

				mx.Lock()
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

				mx.Unlock()

				wg.Done()
			}(i, p)
		}
	}

	wg.Wait()

	go m.listen()

	_, waiting := <-m.ready

	if waiting {
		t.Errorf("Broadcast error: error listening: gater not ready yet")
	}

	if !m.isListening.Load() {
		t.Errorf("Broadcast error: error listening: flag shows false after start")
	}

	<-time.After(time.Millisecond * 10)

	gossipdone := make(chan int, 1)
	go func() {
		<-m.ready
		m.on(types.BroadcastDoneEvent, func(args ...interface{}) {
			gossipdone <- 1
		})
		m.handle()
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

					nonce, _, _, _, err := (&wireCodec{}).decode(d.buff)
					fmt.Println("Err", err)
					msgpayload := &common.PocketEvent{
						Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
						Data:  nil,
					}
					ack := types.Message(int32(nonce), int32(0), e.address, m.address, msgpayload)
					eack, _ := m.c.Encode(*ack)
					wack := (&wireCodec{}).encode(Binary, false, nonce, eack, true)

					e.respond <- wack
				case <-e.done:
					break waiter

				default:
				}
			}
		}(i)
	}

	conn, _ := net.Dial("tcp", m.address)

	gm := types.Message(int32(0), int32(4), conn.LocalAddr().String(), m.address, &common.PocketEvent{
		Topic: common.PocketTopic_CONSENSUS_MESSAGE_TOPIC,
		Data:  nil,
	})
	egm, _ := m.c.Encode(gm)
	wgm := (&wireCodec{}).encode(Binary, false, 0, egm, true)
	conn.Write(wgm)

	fmt.Println("Has written the size of", len(wgm))
	buff := make([]byte, m.config.BufferSize)
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

		codec := (&wireCodec{})
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
				n, err := stdio.ReadFull(creader, buffer[:8]) // TODO(derrandz): parameterize this
				if err != nil {
					if isErrTimeout(err) {
						readerClosed = true
						continue reader
					}
					data <- datapoint(n, err, buffer)
					break reader
				}

				_, _, bodylength, derr := codec.decodeHeader(buffer[:8]) // TODO(derrandz): DITTO line 861
				if derr != nil {
					data <- datapoint(len(buffer), err, buffer)
					break reader
				}

				n, err = stdio.ReadAtLeast(creader, buffer[8:], int(bodylength)) // TODO(derrandz): DITTO line 861
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
					hl := 8
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
