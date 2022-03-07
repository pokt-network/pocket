package p2p

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
)

const (
	ReadBufferSize       = 1024 * 4
	WriteBufferSize      = 1024 * 4
	WireByteHeaderLength = 8
)

func TestIO_NewIO(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)
	if cap(pipe.buffers.read) != ReadBufferSize && cap(pipe.buffers.write) != WriteBufferSize {
		t.Logf("IO pipe is malconfigured")
	} else {
		t.Log("Success")
	}
}

func TestIO_Write(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)

	pipe.buffersState.writeOpen = true // this is usually set to true by pipe.open

	n, _ := pipe.write([]byte("Hello World"), false, 0, false)

	<-pipe.buffersState.writeSignals

	if len(pipe.buffers.write) != int(n)+WireByteHeaderLength {
		t.Errorf("pipe write error: buffer length mismatch, expected: %d, got: %d", len(pipe.buffers.write), n)
	}

	if pipe.buffersState.writeOpen != true {
		t.Errorf("pipe write error: write buffer closed for no reason")
	}
}

func TestIO_WriteConcurrently(t *testing.T) {
	t.Skip()
}

func TestIO_Answer(t *testing.T) {
	t.Skip()

	pipe := NewSocket(1024*4, 8, 100)
	pipe.buffersState.writeOpen = true // usually opened by pipe.open
	pipe.opened.Store(true)

	pipe.network = nil // TODO(derrandz): replace with mockgen

	conn := MockConnM()
	pipe.conn = net.Conn(conn)
	pipe.writer = bufio.NewWriter(pipe.conn)

	go pipe.answer()

	_, open := <-pipe.answering

	if !open {
		t.Errorf("pipe answer routine error: did not signal the start of answering")
	}

	data := GenerateByteLen(1024)
	var buff []byte = make([]byte, 1024+WireByteHeaderLength)

	pipe.write(data, false, 0, false)

	<-conn.signals

	n, err := pipe.conn.Read(buff)

	if err != nil {
		t.Errorf("pipe answer routine error: write error: %s", err.Error())
	}

	if n != len(buff) {
		t.Errorf("pipe answer routine error: written buffer length mismatch, expected: %d, got: %d", len(buff), n)
	}

	_, _, conndata, _, err := pipe.c.decode(buff)
	if bytes.Compare(conndata, data) != 0 {
		t.Errorf("pipe answer routing error: written buffer corrupted")
	}

	pipe.close()

	<-pipe.closed
	_, writerSignalsOpen := <-pipe.buffersState.writeSignals
	if writerSignalsOpen {
		t.Errorf("pipe answer routing error: answer routing still going after pipe close")
	}
}

func TestIO_Read(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)

	conn := MockConn()
	pipe.network = nil // TODO(derrandz): replace with mockgen
	pipe.conn = conn

	pipe.reader = bufio.NewReader(pipe.conn)

	{
		msg := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
		emsg := (&wireCodec{}).encode(Binary, false, 0, msg, false)
		pipe.conn.Write(emsg)

		buff, n, err := pipe.read()

		if err != nil {
			t.Errorf("pipe read error: %s", err.Error())
		}

		if n != ReadBufferSize-WireByteHeaderLength {
			t.Errorf("pipe read error: read buffer length mismatch, expected %d, got %d", ReadBufferSize, n)
		}

		if bytes.Compare(buff[WireByteHeaderLength:], msg) != 0 {
			t.Errorf("pipe read error: read buffer corrupted")
		}
	}

	(conn.(*connM)).Flush() // typecasting to original mock struct type to make use of Flush method

	{
		msg := GenerateByteLen(1024)
		emsg := (&wireCodec{}).encode(Binary, false, 0, msg, false)
		pipe.conn.Write(emsg)

		buff, n, err := pipe.read()

		if err != nil {
			t.Errorf("pipe read error: %s", err.Error())
		}

		if n != 1024 {
			t.Errorf("pipe read error: read buffer length mismatch, expected %d, got %d", 1024, n)
		}

		if bytes.Compare(buff[WireByteHeaderLength:], msg) != 0 {
			t.Errorf("pipe read error: read buffer corrupted")
		}
	}

}

/*
 @ io.poll is a continuous read loop that reads incoming messages from a reader/writer/closer (like a network connection)
*/
func TestIO_Poll(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)

	pipe.network = nil // TODO(derrandz): replace with mockgen
	pipe.conn = MockConn()
	pipe.reader = bufio.NewReader(pipe.conn)

	msg := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
	data := pipe.c.encode(Binary, false, 0, msg, false)

	pipe.opened.Store(true)

	go pipe.poll()

	<-pipe.polling
	<-time.After(time.Millisecond * 5)

	pipe.conn.Write(data)

	<-time.After(time.Millisecond * 10)

	buff := pipe.buffers.read

	_, _, dbuff, _, err := pipe.c.decode(buff)

	if err != nil {
		t.Errorf("pipe read error: could not decode read bytes: %s", err.Error())
	}

	if len(buff) != ReadBufferSize {
		t.Errorf("pipe read error: read buffer length mismatch, expected %d, got %d", ReadBufferSize, len(buff))
	}

	if bytes.Compare(dbuff, msg) != 0 {
		t.Errorf("pipe read error: read buffer corrupted")
	}

	<-pipe.network.sink
	pipe.close()

	<-pipe.closed
	_, pollingOpen := <-pipe.polling
	if pollingOpen {
		t.Errorf("pipe poll error: state indicates polling/receiving is active after pipe closed")
	}
}

func TestIO_Inbound(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)

	pipe.network = nil // TODO(derrandz): replace with mockgen
	pipe.buffersState.writeOpen = true

	addr := "dummy-test-host:dummyport"
	conn := MockConnM()

	// did not use MockFunc due to the issues it has with nil errors
	onopenedStub := newFnCallStub()
	onopened := func(p *socket) error {
		onopenedStub.trackCall()
		return nil
	}

	onclosedStub := newFnCallStub()
	onclosed := func(p *socket) error {
		onclosedStub.trackCall()
		return nil
	}

	go pipe.inbound(addr, net.Conn(conn), onopened, onclosed)

	_, answeringOpen := <-pipe.answering
	_, pollingOpen := <-pipe.polling

	if pipe.reader == nil || pipe.writer == nil {
		t.Errorf("pipe inbound error: reader/writter is not initialized after inbound launch")
	}

	if pipe.conn == nil {
		t.Errorf("pipe inbound error: pipe connection is not initialized after inbound launch")
	}

	if !answeringOpen || !pollingOpen {
		t.Errorf("pipe inbound error: pipe is not receiving or sending after inbound launch")
	}

	if !onopenedStub.wasCalled() {
		t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
	}

	if !onopenedStub.wasCalledTimes(1) {
		t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times())
	}

	msg := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
	encoded := pipe.c.encode(Binary, false, 0, msg, false)

	go conn.Write(encoded)
	<-conn.signals

	<-time.After(time.Millisecond * 20)
	w := <-pipe.network.sink
	n := len(w.Bytes())

	//_, _, data, _, err := pipe.c.decode(buff)
	//if err != nil {
	//	t.Errorf("pipe inbound error: failed to decode received inbound buffer: %s", err.Error())
	//}

	if n != ReadBufferSize-WireByteHeaderLength {
		t.Errorf("pipe inbound error (read error): received inbound buffer length mismatch, expected %d, got %d", ReadBufferSize, n)
	}

	if bytes.Compare(w.Bytes(), msg) != 0 {
		t.Errorf("pipe inbound error (read error): received inbound buffer corrupted")
	}

	conn.Flush()
	response := GenerateByteLen((1024 * 4) - WireByteHeaderLength)

	wn, werr := pipe.write(response, false, 0, false)
	if werr != nil {
		t.Errorf("pipe inbound error (write error): error writing to the inbound pipe")
	}

	<-conn.signals

	answer := make([]byte, ReadBufferSize)
	cn, cerr := conn.Read(answer)

	if cerr != nil {
		t.Errorf("pipe inbound error (answer error): inbound peer could not read response, %s", cerr.Error())
	}

	if uint(cn)-9 != wn {
		t.Errorf("pipe inbound error (answer error): inbound peer received wrong number of bytes")
	}

	_, _, answer, _, err := pipe.c.decode(answer)

	if err != nil {
		t.Errorf("pipe inbound error (answer error): inbound peer could not decode response")
	}

	if bytes.Compare(answer, response) != 0 {
		t.Errorf("pipe inbound error (answer error): inbound peer received corrupted response")
	}

	pipe.network.done <- 1

	<-time.After(time.Millisecond * 10) // give time for routines to wrap up

	if !onclosedStub.wasCalled() {
		t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
	}

	if !onclosedStub.wasCalledTimes(1) {
		t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times())
	}
}

/*
 @
 @ Might fail (from time to time) due to goroutines synchronization, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless
 @
*/
func TestIO_Outbound(t *testing.T) {
	pipe := NewSocket(1024*4, 8, 100)

	dialer := MockDialer()

	pipe.dialer = types.Dialer(dialer)
	pipe.network = nil // TODO(derrandz): replace with mockgen
	pipe.buffersState.writeOpen = true

	pipe.opened.Store(true)

	addr := "dummy-test-host:dummyport"

	// did not use MockFunc due to the issues it has with nil errors
	onopenedStub := newFnCallStub()
	onopened := func(p *socket) error {
		onopenedStub.trackCall()
		return nil
	}

	onclosedStub := newFnCallStub()
	onclosed := func(p *socket) error {
		onclosedStub.trackCall()
		return nil
	}

	go pipe.outbound(addr, onopened, onclosed)

	_, answeringOpen := <-pipe.answering
	_, pollingOpen := <-pipe.polling

	conn := dialer.conn // will be set after Dial

	if pipe.reader == nil || pipe.writer == nil {
		t.Errorf("pipe outbound error: reader/writter is not initialized after outbound launch")
	}

	if pipe.conn == nil {
		t.Errorf("pipe outbound error: pipe connection is not initialized after outbound launch")
	}

	if !answeringOpen || !pollingOpen {
		t.Errorf("pipe outbound error: pipe is not receiving or sending after outbound launch")
	}

	if !onopenedStub.wasCalled() {
		t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
	}

	if !onopenedStub.wasCalledTimes(1) {
		t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times())
	}

	{
		// send to the outbound peer

		ping := GenerateByteLen((1024 * 4) - WireByteHeaderLength)

		wn, werr := pipe.write(ping, false, 0, false)
		if werr != nil {
			t.Errorf("pipe outbound error (write error): error writing to the outbound pipe")
		}

		<-conn.signals

		rping := make([]byte, 1024*4+WireByteHeaderLength) // buffer for the received ping message
		cn, cerr := conn.Read(rping)

		if cerr != nil {
			t.Errorf("pipe outbound error (answer error): outbound peer could not read response, %s", cerr.Error())
		}

		if uint(cn-WireByteHeaderLength) != wn {
			t.Errorf("pipe outbound error (answer error): outbound peer received wrong number of bytes")
		}

		_, _, rping, _, err := pipe.c.decode(rping)

		if err != nil {
			t.Errorf("pipe outbound error: outbound peer could not decode received buff: %s ", err.Error())
		}

		if bytes.Compare(rping, ping) != 0 {
			t.Errorf("pipe outbound error (answer error): outbound peer received corrupted response")
		}
	}

	// whatever has been written to the conn, will also be read by the pipe
	// since the conn mock is a singal conduit and not two (in/out) as in a real conn
	// so after flushing, we need to make sure to flush out the what's been read by the pipe
	// and sent to the sink. (by emptying the sink)
	conn.Flush()
	<-pipe.network.sink

	{
		rawpong := GenerateByteLen((1024 * 4) - WireByteHeaderLength)
		pong := pipe.c.encode(Binary, false, 0, rawpong, false)
		go conn.Write(pong)

		<-conn.signals

		if bytes.Compare(conn.buff, pong) != 0 {
			t.Errorf("Conn Error: payload mismatch, payload length: %d, buffer length: %d", len(pong), len(conn.buff))
		}

		w := <-pipe.network.sink

		rpong := w.Bytes()

		if len(rpong) != ReadBufferSize-WireByteHeaderLength {
			t.Errorf("pipe outbound error (read error): received outbound buffer length mismatch, expected %d, got %d", ReadBufferSize-WireByteHeaderLength, len(rpong))
		}

		if bytes.Compare(rpong, rawpong) != 0 {
			t.Errorf("pipe outbound error (read error): received corrupted buffer from outbound peer")
		}
	}

	conn.Flush()

	pipe.close()

	<-time.After(time.Millisecond * 10)
	if !onclosedStub.wasCalled() {
		t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
	}

	if !onclosedStub.wasCalledTimes(1) {
		t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times())
	}
}

func TestIO_Open(t *testing.T) {

	// test opening an outbound connection
	{
		pipe := NewSocket(1024*4, 8, 100)

		dialer := MockDialer()

		pipe.dialer = types.Dialer(dialer)
		pipe.network = nil // TODO(derrandz): replace with mockgen
		pipe.buffersState.writeOpen = true

		addr := "dummy-test-host:dummyport"

		// did not use MockFunc due to the issues it has with nil errors
		onopenedStub := newFnCallStub()
		onopened := func(p *socket) error {
			onopenedStub.trackCall()
			return nil
		}

		onclosedStub := newFnCallStub()
		onclosed := func(p *socket) error {
			onclosedStub.trackCall()
			return nil
		}

		pipe.open(OutboundIoPipe, addr, nil, onopened, onclosed)

		_, closed := <-pipe.ready
		ready := !closed

		if !ready {
			t.Errorf("pipe outbound error: pipe is not receiving or sending after outbound launch")
		}

		if pipe.reader == nil || pipe.writer == nil {
			t.Errorf("pipe outbound error: reader/writter is not initialized after outbound launch")
		}

		if pipe.conn == nil {
			t.Errorf("pipe outbound error: pipe connection is not initialized after outbound launch")
		}

		if !onopenedStub.wasCalled() {
			t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
		}

		if !onopenedStub.wasCalledTimes(1) {
			t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times())
		}

		pipe.network.done <- 1
		<-time.After(time.Millisecond * 10)

		if !onclosedStub.wasCalled() {
			t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
		}

		if !onclosedStub.wasCalledTimes(1) {
			t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times())
		}
	}

	// test opening an inbound connection
	{
		pipe := NewSocket(1024*4, 8, 100)

		pipe.network = nil // TODO(derrandz): replace with mockgen
		pipe.buffersState.writeOpen = true

		addr := "dummy-test-host:dummyport"

		// did not use MockFunc due to the issues it has with nil errors
		onopenedStub := newFnCallStub()
		onopened := func(p *socket) error {
			onopenedStub.trackCall()
			return nil
		}

		onclosedStub := newFnCallStub()
		onclosed := func(p *socket) error {
			onclosedStub.trackCall()
			return nil
		}

		pipe.open(InboundIoPipe, addr, MockConn(), onopened, onclosed)

		_, closed := <-pipe.ready
		ready := !closed

		if !ready {
			t.Errorf("pipe open inbound error: pipe is not receiving or sending after inbound launch")
		}

		if pipe.reader == nil || pipe.writer == nil {
			t.Errorf("pipe open inbound error: reader/writter is not initialized after inbound launch")
		}

		if pipe.conn == nil {
			t.Errorf("pipe open inbound error: pipe connection is not initialized after inbound launch")
		}

		if pipe.sense != InboundIoPipe {
			t.Errorf("pipe open inbound error: wrong pipe sense")
		}

		if !onopenedStub.wasCalled() {
			t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
		}

		if !onopenedStub.wasCalledTimes(1) {
			t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times())
		}

		pipe.network.done <- 1
		<-time.After(time.Millisecond * 10)

		if !onclosedStub.wasCalled() {
			t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
		}

		if !onclosedStub.wasCalledTimes(1) {
			t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times())
		}
	}

}

func TestReqMapGet(t *testing.T) {
	rmap := types.NewRequestMap(100)

	request := rmap.Get()

	if request.Nonce() != 1 {
		t.Errorf("reqmap error: failed to retrieve new request with valid nonce")
	}

	if request.Response() == nil {
		t.Errorf("reqmap error: failed to retrieve request with a respond channel")
	}
}

func TestReqMapFind(t *testing.T) {

	rmap := types.NewRequestMap(100)
	request := rmap.Get()
	nonce, ch, exists := rmap.Find(request.Nonce())

	if !exists {
		t.Errorf("reqmap error: cannot retrieve/find existing request!")
	}

	if nonce != request.Nonce() {
		t.Errorf("reqmap error: faield to retrieve existing request, found a wrong one with invalid nonce")
	}

	var isproperch bool

	go func(channel chan types.Work) {
	waiter:
		for {
			select {
			case <-channel:
				isproperch = true
				break waiter
			}
		}
	}(ch)

	<-time.After(time.Millisecond)
	request.Respond(types.Work{})
	<-time.After(time.Millisecond * 5)
	if isproperch != true {
		t.Errorf("reqmap error: failed to retrieve existing request, found a wrong one with a diffrent respond channel")
	}
}

func TestReqMapDelete(t *testing.T) {
	rmap := types.NewRequestMap(100)
	request := rmap.Get()

	deleted := rmap.Delete(request.Nonce())

	if !deleted {
		t.Errorf("reqmap error: could not delete existing request")
	}

	_, open := <-request.Response()
	if open != false {
		t.Errorf("reqmap error: request respond channel is still open after delete")
	}

	_, _, exists := rmap.Find(request.Nonce())
	if exists {
		t.Errorf("reqmap error: the request is still tracked in the reqmap after delete")
	}
}
