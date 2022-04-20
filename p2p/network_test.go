package p2p

import (
	"net"
	"sync"
	"testing"
	"time"

	cfg "github.com/pokt-network/pocket/shared/config"
	shared "github.com/pokt-network/pocket/shared/config"
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
		TimeoutInMs:      500,
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
		TimeoutInMs:      500,
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

	ready, _, data, _ := ListenAndServe(addr, int(m.config.BufferSize), 200)

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
			pipe.isOpen.Load(),
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
		TimeoutInMs:      500,
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
			pipe.isOpen.Load(),
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
		TimeoutInMs:      500,
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
	ready, _, data, respond := ListenAndServe(addr, int(m.config.BufferSize), 200)

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
		Address:          []byte("0.0.0.0:36341"),
		ExternalIp:       "0.0.0.0:31341",
		Peers:            []string{"0.0.0.0:2221"},
		MaxInbound:       100,
		MaxOutbound:      100,
		BufferSize:       BufferSize,
		WireHeaderLength: WireByteHeaderLength,
		TimeoutInMs:      100,
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
		case <-m.ready:
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

	//go func() {
	c := newWireCodec()
	request := c.encode(Binary, false, 12, msgA, false)
	t.Logf("Mock peer about to write")
	_, werr := conn.Write(request)

	t.Logf("The mock peer has written, err=%s", werr)
	assert.Nil(
		t,
		err,
		"Respond: encountered error while mock peer trying to request the p2p peer.",
	)

	t.Logf("Respond: mock peer has successfully sent a request to the p2p peer: OK")

	//}()

	{
		<-time.After(time.Millisecond * 10)

		t.Logf("Respond: p2p peer waiting on requests...")
		t.Logf("The sink has %d elements", len(m.sink))
		w := <-m.sink // blocking
		t.Logf("Respond: p2p peer has received a request: OK")

		nonce := w.Nonce

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
