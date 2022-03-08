package p2p

import (
	"bufio"
	"fmt"
	stdio "io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
	cfg "github.com/pokt-network/pocket/shared/config"
	shared "github.com/pokt-network/pocket/shared/config"
	common "github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/assert"
)

const (
	WireHeaderLength = 9
	BufferSize       = 1024 * 4
)

func TestNetwork_NewP2PModule(t *testing.T) {
	m := newP2PModule()

	assert.Nil(
		t,
		m.peerlist,
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Nil(
		t,
		m.inbound,
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Nil(
		t,
		m.outbound,
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Equal(
		t,
		m.protocol,
		"",
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Equal(
		t,
		m.address,
		"",
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Equal(
		t,
		m.externaladdr,
		"",
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.NotNil(
		t,
		m.c,
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

	assert.Equal(
		t,
		m.isListening.Load(),
		false,
		"NewP2PModule: Encountered error while instantiating the p2p module",
	)

}

func TestNetwork_ListenStop(t *testing.T) {
	config := &cfg.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:12345"),
		ExternalIp:       "0.0.0.0:12345",
		MaxInbound:       128,
		MaxOutbound:      128,
		Peers:            []string{"0.0.0.0:1111"},
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      100,
	}

	m := newP2PModule()
	err := m.initialize(config)

	assert.Nilf(
		t,
		err,
		"ListenStop: Encountered error while initializing the p2p module: %s", err,
	)

	go m.listen()

	_, waiting := <-m.ready

	assert.Equal(
		t,
		waiting,
		false,
		"Error listening: gater not ready yet",
	)

	assert.Equal(
		t,
		m.isListening.Load(),
		true,
		"Error listening: flag shows false after start",
	)

	t.Log("Server listening.")
	t.Log("Closing...")

	m.close()

	_, finished := <-m.closed

	assert.Equal(
		t,
		finished,
		true,
		"Error: not closed after .Close()",
	)

	assert.Nilf(
		t,
		m.err.error,
		"Error listening: %s", err,
	)

	assert.Equal(
		t,
		m.isListening.Load(),
		false,
		"Error listening: flag shows true after stop",
	)

	m.listener.Lock()
	assert.Nil(
		t,
		m.listener.TCPListener,
		"Error: listener is still active",
	)
	m.listener.Unlock()

	t.Log("Server closed.")
}

func TestNetwork_SendOutbound(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:30301"),
		ExternalIp:       "0.0.0.0:32321",
		Peers:            []string{"0.0.0.0:2221"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      200,
	}
	m := newP2PModule()

	{
		err := m.initialize(config)

		assert.Nilf(
			t,
			err,
			"SendOutbound: failed to initialize the p2p module: %s", err,
		)
	}

	{
		go m.listen()

		select {
		case <-m.isReady():
		case <-m.errored:
			t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
		}
	}

	addr := "0.0.0.0:2111"
	msg := []byte("hello")

	ready, _, data, _ := ListenAndServe(addr, int(m.config.BufferSize))

	select {
	case v := <-ready:
		assert.Equal(
			t,
			v,
			uint(1),
			"Send error: could not start recipient server",
		)
	}

	{
		err := m.send(addr, msg, false)

		assert.Nilf(
			t,
			err,
			"Send error: Failed to write message to target: %s", err,
		)
	}

	{
		var pipe *socket
		obj, exists := m.outbound.Find(addr)
		pipe = obj.(*socket)

		assert.Equal(
			t,
			exists,
			true,
			"Send error: outbound connection not registered",
		)

		_, down := <-pipe.ready
		assert.Equal(
			t,
			down,
			false,
			"Send error: pipe is not ready",
		)

		assert.Equal(
			t,
			pipe.opened.Load(),
			true,
			"Send error: pipe is not open",
		)
	}

	{
		received := <-data
		assert.Nilf(
			t,
			received.err,
			"Send error: recipient has received an error while receiving: %s", received.err,
		)

		assert.Equalf(
			t,
			received.buff[m.config.WireHeaderLength:],
			msg,
			"Send error: recipient received a corrupted message",
		)
	}
}

func TestNetwork_SendInbound(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:31301"),
		ExternalIp:       "0.0.0.0:31321",
		Peers:            []string{"0.0.0.0:2221"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      200,
	}

	m := newP2PModule()

	{
		err := m.initialize(config)

		assert.Nilf(
			t,
			err,
			"SendInbound: failed to initialize the p2p module: %s", err,
		)
	}

	{
		go m.listen()

		select {
		case <-m.isReady():
		case <-m.errored:
			t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
		}
	}

	conn, err := net.Dial("tcp", m.address)

	assert.Nil(
		t,
		err,
		"SendInbound: encountered error while dialing the p2p peer",
	)

	<-time.After(time.Millisecond * 2) // let p2p peer catch up and store the new inbound conn

	msg := GenerateByteLen(int(m.config.BufferSize) - int(m.config.WireHeaderLength))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		err = m.send(conn.LocalAddr().String(), msg, false)
		assert.Nil(
			t,
			err,
			"SendInbound: Failed to write message to target",
		)
		wg.Done()
	}()

	wg.Wait()

	{
		var pipe *socket
		obj, exists := m.inbound.Find(conn.LocalAddr().String())
		pipe = obj.(*socket)

		assert.Equal(
			t,
			exists,
			true,
			"Send error: outbound connection not registered",
		)

		_, down := <-pipe.ready
		assert.Equal(
			t,
			down,
			false,
			"Send error: pipe is not ready",
		)

		assert.Equal(
			t,
			pipe.opened.Load(),
			true,
			"Send error: pipe is not open",
		)

		received := make([]byte, m.config.BufferSize)
		n, err := conn.Read(received)

		assert.Nil(
			t,
			err,
			"Send error: recipient has received an error while receiving: %s", err,
		)

		assert.Equal(
			t,
			received[m.config.WireHeaderLength:n],
			msg,
			"Send error: recipient received a corrupted message",
		)
	}
}

func TestNetwork_Request(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:36301"),
		ExternalIp:       "0.0.0.0:31361",
		Peers:            []string{"0.0.0.0:2221"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      200,
	}

	m := newP2PModule()

	{
		err := m.initialize(config)

		assert.Nilf(
			t,
			err,
			"Request: failed to initialize the p2p module: %s", err,
		)
	}

	{
		go m.listen()

		select {
		case <-m.isReady():
		case <-m.errored:
			t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
		}
	}

	t.Logf("Started listenig: OK")

	addr := "localhost:22302"
	ready, _, data, respond := ListenAndServe(addr, int(m.config.BufferSize))

	select {
	case v := <-ready:
		assert.Equal(
			t,
			v,
			uint(1),
			"Request: Encountered error while trying to start the mock peer",
		)
	}

	t.Logf("Request: Successfully started the mock peer: OK")

	msgA := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))

	wg := &sync.WaitGroup{}
	responses := make(chan []byte, 10)

	wg.Add(1)
	go func() {
		t.Log("Request: p2p peer is initiating the request")
		res, err := m.request(addr, msgA, false) // false indicates that no types encoding is taking place: i.e raw payload
		if err != nil {
			assert.Failf(
				t,
				"Request: p2p peer failed to perform request: %s", err.Error(),
			)
			close(responses)
			wg.Done()
			return
		}
		responses <- res
		t.Logf("Request: p2p peer has gotten a response")
		wg.Done()
	}()

	{
		wg.Add(1)
		go func() {
			c := newWireCodec()

			t.Logf("Request: mock peer receiving request...")
			d := <-data
			t.Logf("Request: mock peer received request: OK")

			nonce, encoding, _, _, err := c.decode(d.buff)

			assert.Nil(
				t,
				err,
				"Request:  mock peer encoutered error while decoding the received request %s", err,
			)

			respond <- c.encode(encoding, false, nonce, msgA, false)
			t.Logf("Request: mock peer has sent a response.")
			wg.Done()
		}()
	}

	wg.Wait()

	t.Log("Past the wait")
	{
		var pipe *socket
		obj, _ := m.outbound.Find(addr)
		pipe = obj.(*socket)

		wg.Add(1)
		go func() {
			select {
			case <-pipe.errored:
				assert.Nilf(
					t,
					pipe.err.error,
					"Request error: error while receiving a response: %s", pipe.err.error,
				)
			case <-pipe.ready:
			}
			wg.Done()
		}()
	}

	{

		wg.Add(1)
		go func() {
			t.Log("Parsing the responses")
			select {
			case d, _ := <-responses:
				t.Log("fioatch")
				assert.Equal(
					t,
					len(d),
					int(m.config.BufferSize-m.config.WireHeaderLength),
					"Request error: received response buffer length mistach",
				)

				assert.Equal(
					t,
					d,
					msgA,
					"Request error: received response buffer is corrupted",
				)
			}

			wg.Done()
		}()

		wg.Wait()
	}
}

func TestNetwork_Respond(t *testing.T) {
	config := &shared.P2PConfig{
		Protocol:         "tcp",
		Address:          []byte("0.0.0.0:36301"),
		ExternalIp:       "0.0.0.0:31361",
		Peers:            []string{"0.0.0.0:2221"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      200,
	}

	m := newP2PModule()

	{
		err := m.initialize(config)

		assert.Nilf(
			t,
			err,
			"Request: failed to initialize the p2p module: %s", err,
		)
	}

	{
		go m.listen()

		select {
		case <-m.isReady():
		case <-m.errored:
			t.Errorf("Send error: could not start listening, error: %s", m.err.error.Error())
		}
	}

	t.Logf("Respond: p2p peer has started listening: OK")

	conn, err := net.Dial(m.protocol, m.externaladdr)

	assert.Nil(
		t,
		err,
		"Failed to dial gater. Error: %s", err,
	)

	t.Logf("Respond: mock peer has dialed the p2p peer successfully: OK")

	// send to the gater a nonced message (i.e: request)
	addr := conn.LocalAddr().String()
	requestNonce := 12
	msgA := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))
	msgB := GenerateByteLen((1024 * 4) - int(m.config.WireHeaderLength))

	go func() {
		c := newWireCodec()
		request := c.encode(Binary, false, 12, msgA, false)
		_, err := conn.Write(request)
		assert.Nil(
			t,
			err,
			"Respond: encountered error while mock peer trying to request the p2p peer.",
		)
		t.Logf("Respond: mock peer has successfully sent a request to the p2p peer: OK")
	}()

	{
		<-time.After(time.Millisecond * 5)

		t.Logf("Respond: p2p peer waiting on requests...")
		w := <-m.sink // blocks
		t.Logf("Respond: p2p peer has received a request: OK")

		nonce := w.Nonce()

		t.Logf("Respond: p2p peer responding...")

		err = m.respond(nonce, false, addr, msgB, false)

		assert.Nil(
			t,
			err,
			"Respond error: %s", err,
		)

		t.Logf("Respond: p2p peer has sent a response")
	}

	{
		buff := make([]byte, m.config.BufferSize)
		_, err = conn.Read(buff)

		assert.Nil(
			t,
			err,
			"Respond: mock peer encountered error while trying to read the response", err,
		)

		t.Logf("Respond: mock peer has received the response")

		c := newWireCodec()

		dnonce, _, decoded, _, err := c.decode(buff)

		assert.Nil(
			t,
			err,
			"Respond error: could not decode payload. Encountered following error: %s", err,
		)

		t.Logf("Respond: p2p peer has sent a response")

		assert.Equal(
			t,
			dnonce,
			uint32(requestNonce),
			"Respond error: received wrong nonce",
		)

		assert.Equal(
			t,
			len(decoded),
			int(m.config.BufferSize-m.config.WireHeaderLength),
			"Respond error: received response buffer length mistach",
		)

		assert.Equal(
			t,
			decoded,
			msgB,
			"Respond error: received response buffer is corrupted",
		)
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

	err := m.initialize(nil)
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

	err := m.initialize(nil)
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
				n, err := stdio.ReadFull(creader, buffer[:WireHeaderLength]) // TODO(derrandz): parameterize this
				if err != nil {
					if isErrTimeout(err) {
						readerClosed = true
						continue reader
					}
					data <- datapoint(n, err, buffer)
					break reader
				}

				_, _, bodylength, derr := codec.decodeHeader(buffer[:WireHeaderLength]) // TODO(derrandz): DITTO line 861
				if derr != nil {
					data <- datapoint(len(buffer), err, buffer)
					break reader
				}

				n, err = stdio.ReadAtLeast(creader, buffer[WireHeaderLength:], int(bodylength)) // TODO(derrandz): DITTO line 861
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
					hl := 9
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
