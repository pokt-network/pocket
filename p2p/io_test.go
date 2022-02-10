package poktp2p

import (
	"bufio"
	"bytes"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestNewIO(t *testing.T) {
	pipe := NewIoPipe()
	if cap(pipe.buffers.read) != ReadBufferSize && cap(pipe.buffers.write) != WriteBufferSize {
		t.Logf("IO pipe is malconfigured")
	} else {
		t.Log("Success")
	}
}

func TestWrite(t *testing.T) {
	pipe := NewIoPipe()

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

func TestWriteConcurrently(t *testing.T) {
	t.Skip()
}

func TestAnswer(t *testing.T) {
	pipe := NewIoPipe()
	pipe.buffersState.writeOpen = true // usually opened by pipe.open

	pipe.g = MockGater()

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

	<-pipe.answering
	_, writerSignalsOpen := <-pipe.buffersState.writeSignals
	if writerSignalsOpen || pipe.sending || pipe.buffersState.writeOpen {
		t.Errorf("pipe answer routing error: answer routing still going after pipe close")
	}
}

func TestRead(t *testing.T) {
	pipe := NewIoPipe()
	pipe.buffersState.readOpen = true // usually opened by pipe.open

	pipe.g = MockGater()
	pipe.conn = MockConn()

	pipe.reader = bufio.NewReader(pipe.conn)

	msg := GenerateByteLen(1024 * 4)
	pipe.conn.Write(msg)

	buff, n, err := pipe.read()

	if err != nil {
		t.Errorf("pipe read error: %s", err.Error())
	}

	if n != ReadBufferSize {
		t.Errorf("pipe read error: read buffer length mismatch, expected %d, got %d", ReadBufferSize, n)
	}

	if bytes.Compare(buff[:ReadBufferSize-WireByteHeaderLength], msg) != 0 {
		t.Errorf("pipe read error: read buffer corrupted")
	}
}

/*
 @ io.poll is a continuous read loop that reads incoming messages from a reader/writer/closer (like a network connection)
*/
func TestPoll(t *testing.T) {
	pipe := NewIoPipe()

	pipe.g = MockGater()
	pipe.conn = MockConn()
	pipe.reader = bufio.NewReader(pipe.conn)

	pipe.buffersState.readOpen = true // usually opened by pipe.open

	msg := GenerateByteLen(1024 * 4)
	data := pipe.c.encode(Binary, false, 0, msg, false)

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

	<-pipe.g.sink
	pipe.close()

	_, pollingOpen := <-pipe.polling
	if pipe.receiving != false || pollingOpen || pipe.buffersState.readOpen {
		t.Errorf("pipe poll error: state indicates polling/receiving is active after pipe closed")
	}
}

func TestInbound(t *testing.T) {
	pipe := NewIoPipe()

	pipe.g = MockGater()
	pipe.buffersState.readOpen = true
	pipe.buffersState.writeOpen = true

	addr := "dummy-test-host:dummyport"
	conn := MockConnM()

	// did not use MockFunc due to the issues it has with nil errors
	onopenedStub := &struct {
		called bool
		times  int
	}{
		called: false,
		times:  0,
	}
	onopened := func(p *io) error {
		onopenedStub.called = true
		onopenedStub.times++
		return nil
	}

	onclosedStub := &struct {
		called bool
		times  int
	}{
		called: false,
		times:  0,
	}
	onclosed := func(p *io) error {
		onclosedStub.called = true
		onclosedStub.times++
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

	if !pipe.receiving || !answeringOpen || !pipe.sending || !pollingOpen {
		t.Errorf("pipe inbound error: pipe is not receiving or sending after inbound launch")
	}

	if onopenedStub.called != true {
		t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
	}

	if onopenedStub.times != 1 {
		t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times)
	}

	msg := GenerateByteLen(1024 * 4)
	encoded := pipe.c.encode(Binary, false, 0, msg, false)

	go conn.Write(encoded)
	<-conn.signals

	<-time.After(time.Millisecond * 20)
	buff := make([]byte, ReadBufferSize)
	n := copy(buff, pipe.buffers.read)

	_, _, data, _, err := pipe.c.decode(buff)
	if err != nil {
		t.Errorf("pipe inbound error: failed to decode received inbound buffer: %s", err.Error())
	}

	if n != ReadBufferSize {
		t.Errorf("pipe inbound error (read error): received inbound buffer length mismatch, expected %d, got %d", ReadBufferSize, n)
	}

	if bytes.Compare(data, msg) != 0 {
		t.Errorf("pipe inbound error (read error): received inbound buffer corrupted")
	}

	conn.Flush()
	response := GenerateByteLen(1024 * 4)

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

	_, _, answer, _, err = pipe.c.decode(answer)

	if err != nil {
		t.Errorf("pipe inbound error (answer error): inbound peer could not decode response")
	}

	if bytes.Compare(answer, response) != 0 {
		t.Errorf("pipe inbound error (answer error): inbound peer received corrupted response")
	}

	pipe.g.done <- 1

	<-time.After(time.Millisecond * 10) // give time for routines to wrap up

	if onclosedStub.called != true {
		t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
	}

	if onclosedStub.times != 1 {
		t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times)
	}
}

/*
 @
 @ Might fail (from time to time) due to goroutines synchronization, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless, expected behavior, the test case is fine nonetheless
 @
*/
func TestOutbound(t *testing.T) {
	pipe := NewIoPipe()

	dialer := MockDialer()

	pipe.dialer = Dialer(dialer)
	pipe.g = MockGater()
	pipe.buffersState.readOpen = true
	pipe.buffersState.writeOpen = true

	addr := "dummy-test-host:dummyport"

	// did not use MockFunc due to the issues it has with nil errors
	onopenedStub := &struct {
		called bool
		times  int
	}{
		called: false,
		times:  0,
	}
	onopened := func(p *io) error {
		onopenedStub.called = true
		onopenedStub.times++
		return nil
	}

	onclosedStub := &struct {
		called bool
		times  int
	}{
		called: false,
		times:  0,
	}
	onclosed := func(p *io) error {
		onclosedStub.called = true
		onclosedStub.times++
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

	if !pipe.receiving || !answeringOpen || !pipe.sending || !pollingOpen {
		t.Errorf("pipe outbound error: pipe is not receiving or sending after outbound launch")
	}

	if onopenedStub.called != true {
		t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
	}

	if onopenedStub.times != 1 {
		t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times)
	}

	// send to the outbound peer

	ping := GenerateByteLen(1024 * 4)

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

	conn.Flush()

	rawpong := GenerateByteLen(1024 * 4)
	pong := pipe.c.encode(Binary, false, 0, rawpong, false)
	go conn.Write(pong)

	<-conn.signals

	if bytes.Compare(conn.buff, pong) != 0 {
		t.Errorf("Conn Error: payload mismatch, payload length: %d, buffer length: %d", len(pong), len(conn.buff))
	}

	if err != nil {
		t.Errorf("pipe outbound error (read error): could not decode received buffer: %s", err.Error())
	}

	w := <-pipe.g.sink

	rpong := w.data

	if len(rpong) != ReadBufferSize-WireByteHeaderLength {
		t.Errorf("pipe outbound error (read error): received outbound buffer length mismatch, expected %d, got %d", ReadBufferSize-WireByteHeaderLength, len(rpong))
	}

	if bytes.Compare(rpong, rawpong) != 0 {
		t.Errorf("pipe outbound error (read error): received corrupted buffer from outbound peer")
	}

	conn.Flush()

	pipe.g.done <- 1

	<-time.After(time.Millisecond * 10)
	if onclosedStub.called != true {
		t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
	}

	if onclosedStub.times != 1 {
		t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times)
	}
}

func TestOpen(t *testing.T) {

	// test opening an outbound connection
	{
		pipe := NewIoPipe()

		dialer := MockDialer()

		pipe.dialer = Dialer(dialer)
		pipe.g = MockGater()
		pipe.buffersState.readOpen = true
		pipe.buffersState.writeOpen = true

		addr := "dummy-test-host:dummyport"

		// did not use MockFunc due to the issues it has with nil errors
		onopenedStub := &struct {
			called bool
			times  int
		}{
			called: false,
			times:  0,
		}
		onopened := func(p *io) error {
			onopenedStub.called = true
			onopenedStub.times++
			return nil
		}

		onclosedStub := &struct {
			called bool
			times  int
		}{
			called: false,
			times:  0,
		}
		onclosed := func(p *io) error {
			onclosedStub.called = true
			onclosedStub.times++
			return nil
		}

		pipe.open(OutboundIoPipe, addr, nil, onopened, onclosed)

		_, closed := <-pipe.ready
		ready := !closed

		if !ready || !pipe.receiving || !pipe.sending {
			t.Errorf("pipe outbound error: pipe is not receiving or sending after outbound launch")
		}

		if pipe.reader == nil || pipe.writer == nil {
			t.Errorf("pipe outbound error: reader/writter is not initialized after outbound launch")
		}

		if pipe.conn == nil {
			t.Errorf("pipe outbound error: pipe connection is not initialized after outbound launch")
		}

		if onopenedStub.called != true {
			t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
		}

		if onopenedStub.times != 1 {
			t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times)
		}

		pipe.g.done <- 1
		<-time.After(time.Millisecond * 10)

		if onclosedStub.called != true {
			t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
		}

		if onclosedStub.times != 1 {
			t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times)
		}
	}

	// test opening an inbound connection
	{
		pipe := NewIoPipe()

		pipe.g = MockGater()
		pipe.buffersState.readOpen = true
		pipe.buffersState.writeOpen = true

		addr := "dummy-test-host:dummyport"

		// did not use MockFunc due to the issues it has with nil errors
		onopenedStub := &struct {
			called bool
			times  int
		}{
			called: false,
			times:  0,
		}
		onopened := func(p *io) error {
			onopenedStub.called = true
			onopenedStub.times++
			return nil
		}

		onclosedStub := &struct {
			called bool
			times  int
		}{
			called: false,
			times:  0,
		}
		onclosed := func(p *io) error {
			onclosedStub.called = true
			onclosedStub.times++
			return nil
		}

		pipe.open(InboundIoPipe, addr, MockConn(), onopened, onclosed)

		_, closed := <-pipe.ready
		ready := !closed

		if !ready || !pipe.receiving || !pipe.sending {
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

		if onopenedStub.called != true {
			t.Errorf("pipe inbound error: did not call onopened handler on opened connection event")
		}

		if onopenedStub.times != 1 {
			t.Errorf("pipe inbound error: expected onopened handler to be called once, got called %d times", onopenedStub.times)
		}

		pipe.g.done <- 1
		<-time.After(time.Millisecond * 10)

		if onclosedStub.called != true {
			t.Errorf("pipe inbound error: did not call onclosed handler on closed connection event")
		}

		if onclosedStub.times != 1 {
			t.Errorf("pipe inbound error: expected onclosed handler to be called once, got called %d times", onopenedStub.times)
		}
	}

}

/*
 @ test utils
*/
func GenerateByteLen(size int) []byte {
	buff := make([]byte, size)
	rand.Read(buff)
	return buff
}
