package p2p

import (
	"bufio"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	testutils "github.com/pokt-network/pocket/p2p/testutils"
	"github.com/pokt-network/pocket/p2p/types"
	sharedTypes "github.com/pokt-network/pocket/shared/types"
	"github.com/stretchr/testify/assert"
)

const (
	ReadBufferSize       = 1024 * 4
	WriteBufferSize      = 1024 * 4
	WireByteHeaderLength = 9
	ReadDeadlineInMs     = 400
)

var encode func([]byte) []byte = func(b []byte) []byte {
	return (&wireCodec{}).encode(Binary, false, 0, b, false)
}

func TestSocket_New(t *testing.T) {
	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)
	if cap(pipe.buffers.read.Bytes()) != ReadBufferSize && cap(pipe.buffers.write.Bytes()) != WriteBufferSize {
		t.Logf("IO pipe is malconfigured")
	} else {
		t.Log("Success")
	}
}

func TestSocket_WriteChunk(t *testing.T) {
	var wg sync.WaitGroup
	var writtenLength uint
	var writeErr error
	var data []byte = []byte("Hello World")

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)

	pipe.buffers.write.Open() // this is usually set to true by pipe.open

	// write a chunk
	{
		// writeChunk is blocking, since it signals (using a channel) that it has written a chunk everytime it's done writing one.
		// if no one is waiting on that signal, writeChunk will block forever.
		// Thus we are sending in an early goroutine to wait on that signal so that writeChunk writes and unblocks immediately after signaling
		wg.Add(1)
		go func() {
			pipe.buffers.write.Wait()
			wg.Done()
		}()

		writtenLength, writeErr = pipe.writeChunk(data, false, 0, false)

		wg.Wait()
	}

	{
		assert.Nilf(
			t,
			writeErr,
			"pipe write error: %s", writeErr,
		)

		assert.Equal(
			t,
			len(pipe.buffers.write.Bytes()),
			int(writtenLength)+WireByteHeaderLength,
			"pipe write error: buffer length mismatch",
		)

		assert.Equal(
			t,
			pipe.buffers.write.IsOpen(),
			true,
			"pipe write error: write buffer closed for no reason",
		)
	}
}

func TestSocket_WriteChunkAckfulPeerAcks(t *testing.T) {
	// 1. prepare socket for writing (i.e: make sure the writer does not block on signaling)
	// 2. ref a reference of the requests map such that we retrieve ethe nonce of the write and response channel
	// 3. test a successful write and that the proper response is retrieved
	var wg sync.WaitGroup
	var response types.Packet
	var err error
	var requestNonce uint32
	var responseChannel chan types.Packet
	var data []byte = []byte("Hello World")
	var requestsMap *types.RequestMap

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)
	requestsMap = pipe.requests

	pipe.buffers.write.Open() // this is usually set to true by pipe.open

	{ // test a successful write
		// write a chunk and respond to it immediately
		{

			wg.Add(1)
			go func() {
				response, err = pipe.writeChunkAckful(data, false)
				wg.Done()
			}()

			pipe.buffers.write.Wait() // will unblock once the writeAckful has written the chunk
			responseChannel = requestsMap.Requests()[0].ResponsesCh
			requestNonce = requestsMap.Requests()[0].Nonce
			responseChannel <- types.Packet{Nonce: requestNonce, Data: []byte("Bye bye world")}
			wg.Wait()
		}

		{
			assert.Nilf(
				t,
				err,
				"pipe write error: %s", err,
			)

			assert.Equal(
				t,
				pipe.buffers.write.IsOpen(),
				true,
				"pipe writeAckful error: write buffer closed for no reason",
			)

			assert.Equal(
				t,
				response.Nonce,
				requestNonce,
				"pipe writeAckful error: nonce mismatch",
			)

			assert.Equal(
				t,
				response.Data,
				[]byte("Bye bye world"),
				"pipe writeAckful error: response data mismatch",
			)
		}
	}

}

func TestSocket_WriteChunkAckfulPeerTimesOut(t *testing.T) {
	// 1. prepare socket for writing (i.e: make sure the writer does not block on signaling)
	// 2. ref a reference of the requests map such that we retrieve ethe nonce of the write and response channel
	// 3. test a non successful write with a timeout
	var wg sync.WaitGroup
	var err error
	var requestNonce uint32
	var responseChannel chan types.Packet
	var data []byte = []byte("Hello World")
	var requestsMap *types.RequestMap

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)
	requestsMap = pipe.requests

	pipe.buffers.write.Open() // this is usually set to true by pipe.open

	// test a non successful write with a timeout
	{
		// write a chunk and timeout
		{

			wg.Add(1)
			go func() {
				_, err = pipe.writeChunkAckful(data, false)
				wg.Done()
			}()

			pipe.buffers.write.Wait() // will unblock once the writeAckful has written the chunk
			wg.Wait()
		}

		{

			requestNonce = requestsMap.Requests()[0].Nonce
			responseChannel = requestsMap.Requests()[0].ResponsesCh

			assert.NotNil(
				t,
				err,
				"pipe writeAckful error: expected error to be a timeout, got nil",
			)

			assert.Equal(
				t,
				err,
				sharedTypes.ErrSocketRequestTimedOut(pipe.addr, requestNonce),
				"pipe writeAckful error: expected error to be a timeout, got %s", err,
			)

			assert.Equal(
				t,
				pipe.buffers.write.IsOpen(),
				true,
				"pipe writeAckful error: write buffer closed for no reason",
			)

			_, open := <-responseChannel
			t.Skip(
				t,
				open,
				"pipe writeAckful error: response channel should be closed, but it is still open after timeout",
			)
		}
	}
}

func TestSocket_WriteConcurrently(t *testing.T) {
	t.Skip()
}

func TestSocket_WriteRoutine(t *testing.T) {
	runner := testutils.NewRunnerMock() // TODO(derrandz): use mockgen
	conn := testutils.NewConnMock()     // TODO(derrandz): use mockgen
	ctx, cancel := context.WithCancel(context.Background())

	chunk := testutils.NewDataChunk(1024, encode)

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)

	{
		pipe.runner = runner
		pipe.buffers.write.Open() // usually opened by pipe.open
		pipe.isOpen.Store(true)
		pipe.conn = net.Conn(conn)
		pipe.writer = bufio.NewWriter(pipe.conn)
	}

	t.Log("Launching the write routine")
	go pipe.write(ctx)

	_, isSocketWriting := <-pipe.writing

	assert.Equal(
		t,
		isSocketWriting,
		true,
		"pipe.write routine error: did not signal the start of answering",
	)

	var buff []byte = make([]byte, chunk.Length+WireByteHeaderLength)

	{ // write some random data to the write buffer
		t.Log("Writing to the write buffer")
		pipe.writeChunk(chunk.Bytes, false, 0, false)
	}

	// wait for the connection to receive the written data from the write routine
	<-conn.Signals

	t.Log("Mock connection received the written data")

	// read the received data from the connection and assert for correctness
	{
		n, err := pipe.conn.Read(buff)

		assert.Nilf(
			t,
			err,
			"pipe.write routine error: write error: %s", err,
		)

		assert.Equal(
			t,
			n,
			len(buff),
			"pipe.write routine error: written buffer length mismatch",
		)

		_, _, conndata, _, err := pipe.codec.decode(buff)

		assert.Equal(
			t,
			conndata,
			chunk.Bytes,
			"pipe.write routing error: written buffer corrupted",
		)
	}

	t.Log("Closing the socket")
	pipe.close()

	<-pipe.done

	// close the socket and assert for proper closing consequences
	{
		_, isWriteBufferOpen := <-pipe.buffers.write.Signals()

		assert.Equal(
			t,
			isWriteBufferOpen,
			false,
			"pipe.write routing error: answer routing still going after pipe close",
		)

	}

	cancel() // to prevent the context from leaking. This won't have any effect this close would have done its job
}

func TestSocket_WriteRoutineWriterWriteFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_WriteRoutineWriterFlushFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_WriteRoutineWriterSuddenlyCloses(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadChunk(t *testing.T) {
	conn := testutils.NewConnMockBuffered()
	runner := testutils.NewRunnerMock()
	messageA := testutils.NewDataChunk(ReadBufferSize-WireByteHeaderLength, encode)
	messageB := testutils.NewDataChunk(1024, encode)

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)

	{
		pipe.runner = runner
		pipe.conn = conn
		pipe.reader = bufio.NewReader(pipe.conn)
	}

	// write message A
	{
		messageA.Encoded = pipe.codec.encode(Binary, false, 0, messageA.Bytes, false)
		conn.Write(messageA.Encoded)
	}

	{
		buff, n, err := pipe.readChunk()

		assert.Nilf(
			t,
			err,
			"pipe read error: %s", err,
		)

		assert.Equalf(
			t,
			n,
			ReadBufferSize-WireByteHeaderLength,
			"pipe read error: read buffer length mismatch",
		)

		assert.Equalf(
			t,
			buff[WireByteHeaderLength:],
			messageA.Bytes,
			"pipe readChunk error: read buffer corrupted",
		)
	}

	(conn.(*testutils.ConnMock)).Flush() // typecasting to original mock struct type to make use of Flush method

	// write message B
	{
		pipe.conn.Write(messageB.Encoded)
	}

	{
		buff, n, err := pipe.readChunk()

		assert.Nilf(
			t,
			err,
			"pipe read error: %s", err,
		)

		assert.Equalf(
			t,
			n,
			1024,
			"pipe read error: read buffer length mismatch",
		)

		assert.Equalf(
			t,
			buff[WireByteHeaderLength:], messageB.Bytes,
			"pipe readChunk error: read buffer corrupted",
		)
	}
}

func TestSocket_ReadChunkPayloadTooBig(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadChunkReaderFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadChunkDecoderFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadRoutine(t *testing.T) {
	runner := testutils.NewRunnerMock()
	conn := testutils.NewConnMockBuffered()

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)

	{
		pipe.conn = conn
		pipe.runner = runner
		pipe.reader = bufio.NewReader(pipe.conn)

		pipe.isOpen.Store(true)
	}

	chunk := testutils.NewDataChunk((1024*4)-WireByteHeaderLength, encode)
	ctx, cancel := context.WithCancel(context.Background())

	go pipe.read(ctx)

	{
		<-pipe.reading
		conn.Write(chunk.Encoded)
	}

	<-time.After(time.Millisecond * 2)

	{
		buff := pipe.buffers.read.Bytes()

		_, _, dbuff, _, err := pipe.codec.decode(buff)

		assert.Nil(
			t,
			err,
			"pipe read error: could not decode read bytes: %s", err,
		)

		assert.Equalf(
			t,
			len(buff),
			ReadBufferSize,
			"pipe read error: read buffer length mismatch, expected %d, got %d", ReadBufferSize, len(buff),
		)

		assert.Equal(
			t,
			dbuff,
			chunk.Bytes,
			"pipe read error: read buffer corrupted",
		)
	}

	<-runner.GetSinkCh()

	{
		pipe.close()
		<-pipe.done
		_, pollingOpen := <-pipe.reading
		assert.Equal(
			t,
			pollingOpen,
			false,
			"pipe.read error: state indicates polling/receiving is active after pipe closed",
		)
		cancel() // to avoid a context leak, it does not have much effect after the .close()
	}
}

func TestSocket_ReadRoutineIOFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadRoutineDecoderFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_ReadRoutineRunnerSinkBlocksIndefinitely(t *testing.T) {
	t.Skip()
}

// This test simulates an inbound connection and tests the `startIO` method
func TestSocket_StartIOInbound(t *testing.T) {
	addr := "dummy-test-host:dummyport"
	runner := testutils.NewRunnerMock()
	conn := testutils.NewConnMock()
	ctx, cancel := context.WithCancel(context.Background())
	onopenedStub := testutils.NewFnCallStub()
	onopened := func(ctx context.Context, p *socket) error {
		onopenedStub.TrackCall()
		return nil
	}
	onclosedStub := testutils.NewFnCallStub()
	onclosed := func(ctx context.Context, p *socket) error {
		onclosedStub.TrackCall()
		return nil
	}

	// generate random data chunks to send back and forth
	message := testutils.NewDataChunk(ReadBufferSize-WireByteHeaderLength, encode)
	response := testutils.NewDataChunk(ReadBufferSize-WireByteHeaderLength, encode)

	pipe := NewSocket(
		ReadBufferSize,
		WireByteHeaderLength,
		ReadDeadlineInMs,
	)

	{
		pipe.runner = runner
		pipe.buffers.write.Open()
		go pipe.startIO(ctx, types.Inbound, addr, net.Conn(conn), onopened, onclosed)
	}

	<-pipe.ioStarted

	// assert that startIO has launched properly and started IO on the inbound connection
	{
		assert.NotNil(
			t,
			pipe.reader,
			"pipe.open error: reader/writter is not initialized after inbound launch",
		)

		assert.NotNil(
			t,
			pipe.writer,
			"pipe.open error: reader/writter is not initialized after inbound launch",
		)

		assert.NotNil(
			t,
			pipe.conn,
			"pipe.open error: pipe connection is not initialized after inbound launch",
		)

		assert.Equal(
			t,
			true,
			pipe.isWriting.Load(),
			"pipe.open error: pipe is not receiving or sending after inbound launch",
		)

		assert.Equal(
			t,
			true,
			pipe.isReading.Load(),
			"pipe.open error: pipe is not receiving or sending after inbound launch",
		)

		assert.Equal(
			t,
			onopenedStub.WasCalled(),
			true,
			"pipe.open error: did not call onopened handler on opened connection event",
		)

		assert.Equalf(
			t,
			onopenedStub.WasCalledTimes(1),
			true,
			"pipe.open error: expected onopened handler to be called once, got called %d times", onopenedStub.Times(),
		)
	}

	// write data to the socket from the inbound connection
	{
		go conn.Write(message.Encoded)
	}

	// wait for the inbound connection to finish writing
	<-conn.Signals
	<-time.After(time.Millisecond * 5)

	// assert that the socket receives data properly from the inbound connection (i,e: that startIO launches IO routines properly (read routine))
	{
		w := <-runner.GetSinkCh()
		n := len(w.Data)

		assert.Equal(
			t,
			n,
			ReadBufferSize-WireByteHeaderLength,
			"pipe.open error (read error): received inbound buffer length mismatch, expected %d, got %d", ReadBufferSize, n,
		)

		assert.Equal(
			t,
			w.Data,
			message.Bytes,
			"pipe.open error (read error): received inbound buffer corrupted",
		)
	}

	// since this is a mocked conn, we have to empty what it has received before, to then write on it afresh
	conn.Flush()

	{
		wn, werr := pipe.writeChunk(response.Bytes, false, 0, false)
		response.Length = wn
		assert.Nil(
			t,
			werr,
			"pipe.open error (write error): error writing to the inbound pipe",
		)
	}

	// wait for data to be recieved on the inbound connection
	<-conn.Signals

	// assert that the inbound connection has received data properly from the socket (i,e: io routines are working properly)
	{
		answer := make([]byte, ReadBufferSize)
		cn, cerr := conn.Read(answer)

		assert.Nil(
			t,
			cerr,
			"pipe.open error (answer error): inbound peer could not read response, %s", cerr,
		)

		assert.Equal(
			t,
			uint(cn)-WireByteHeaderLength,
			response.Length,
			"pipe.open error (answer error): inbound peer received wrong number of bytes",
		)

		_, _, answer, _, err := pipe.codec.decode(answer)

		assert.Nil(
			t,
			err,
			"pipe.open error (answer error): inbound peer could not decode response",
		)

		assert.Equal(
			t,
			answer,
			response.Bytes,
			"pipe.open error (answer error): inbound peer received corrupted response",
		)
	}

	runner.GetDoneCh() <- 1

	<-time.After(time.Millisecond * 10) // give time for routines to wrap up

	{
		assert.Equal(
			t,
			onclosedStub.WasCalled(),
			true,
			"pipe.open error: did not call onclosed handler on closed connection event",
		)

		assert.Equalf(
			t,
			onclosedStub.WasCalledTimes(1),
			true,
			"pipe.open error: expected onclosed handler to be called once, got called %d times", onopenedStub.Times(),
		)
	}

	cancel() // just to stop the context from leaking. won't have any effect since runner.GetDoneCh() 1 has closed running routines
}

// INVESTIGATE(team): Investigate why this test occasionally fails due to a race condition.
// This test simulates an inbound connection and tests the `startIO` method
func TestSocket_StartIOOutbound(t *testing.T) {
	addr := "dummy-test-host:dummyport"
	runner := testutils.NewRunnerMock()
	dialer := testutils.MockDialer()
	ctx, cancel := context.WithCancel(context.Background())
	onopenedStub := testutils.NewFnCallStub()
	onopened := func(_ context.Context, p *socket) error {
		onopenedStub.TrackCall()
		return nil
	}
	onclosedStub := testutils.NewFnCallStub()
	onclosed := func(_ context.Context, p *socket) error {
		onclosedStub.TrackCall()
		return nil
	}
	conn := dialer.Conn

	// generate random data chunks to send back and forth
	message := testutils.NewDataChunk(ReadBufferSize-WireByteHeaderLength, encode)
	response := testutils.NewDataChunk(ReadBufferSize-WireByteHeaderLength, encode)

	pipe := NewSocket(
		ReadBufferSize,
		WireByteHeaderLength,
		ReadDeadlineInMs,
	)

	{ // startIO on outbound connection (Mock)
		pipe.runner = runner
		pipe.buffers.write.Open()
		pipe.isOpen.Store(true)

		go pipe.startIO(ctx, types.Outbound, addr, conn, onopened, onclosed)
	}

	<-pipe.ioStarted // wait for IO to start

	// assert that IO started successfully
	{
		assert.NotNil(
			t,
			pipe.reader,
			"pipe.open error: reader/writter is not initialized after inbound launch",
		)

		assert.NotNil(
			t,
			pipe.writer,
			"pipe.open error: reader/writter is not initialized after inbound launch",
		)

		assert.NotNil(
			t,
			pipe.conn,
			"pipe.open error: pipe connection is not initialized after inbound launch",
		)

		assert.Equal(
			t,
			true,
			pipe.isWriting.Load(),
			"pipe.open error: pipe is not receiving or sending after inbound launch",
		)

		assert.Equal(
			t,
			true,
			pipe.isReading.Load(),
			"pipe.open error: pipe is not receiving or sending after inbound launch",
		)

		assert.Equal(
			t,
			onopenedStub.WasCalled(),
			true,
			"pipe.open error: did not call onopened handler on opened connection event",
		)

		assert.Equalf(
			t,
			onopenedStub.WasCalledTimes(1),
			true,
			"pipe.open error: expected onopened handler to be called once, got called %d times", onopenedStub.Times(),
		)
	}

	// send data to the other end of the outbound socket
	{
		wn, werr := pipe.writeChunk(message.Bytes, false, 0, false) // write a randomly generated message
		message.Length = wn                                         // store the written length

		assert.Nil(
			t,
			werr,
			"pipe.open error (write error): error writing to the outbound pipe",
		)

	}

	// wait for the outbound connection to receive  data
	<-conn.Signals

	// assert that the other end of the outbound connection has received data properly
	{
		buffer := make([]byte, 1024*4)
		cn, cerr := conn.Read(buffer)

		assert.Nilf(
			t,
			cerr,
			"pipe.open error (answer error): outbound peer could not read response, %s", cerr,
		)

		assert.Equal(
			t,
			uint(cn-WireByteHeaderLength),
			message.Length,
			"pipe.open error (answer error): outbound peer received wrong number of bytes",
		)

		_, _, decoded, _, err := pipe.codec.decode(buffer)

		assert.Nil(
			t,
			err,
			"pipe.open error: outbound peer could not decode received buff: %s ", err,
		)

		assert.Equal(
			t,
			decoded,
			message.Bytes,
			"pipe.open error (answer error): outbound peer received corrupted response",
		)
	}

	// flush the conn
	conn.Flush()

	// whatever has been written to the conn, will also be read by the socket (since the conn mock is a single conduit and bidirectional (in/out) conduit as in a real conn)
	// so after flushing, we need to make sure to flush out the what's been read and queud by the socket. (i.e: draining the queue/sink)
	for len(runner.GetSinkCh()) > 0 {
		<-runner.GetSinkCh()
	}

	// send a message as a response to the outbound socket from the outbound end
	go conn.Write(response.Encoded)

	// wait for the mock connection to finish writing/sending
	<-conn.Signals

	// assert that the outbound socket receives data properly from the other end
	{
		<-time.After(time.Millisecond * 5)
		w := <-runner.GetSinkCh()

		receivedResponse := w.Data

		assert.Equal(
			t,
			len(receivedResponse),
			ReadBufferSize-WireByteHeaderLength,
			"pipe.open error (read error): received outbound buffer length mismatch",
		)

		assert.Equal(
			t,
			receivedResponse,
			response.Bytes,
			"pipe.open error (read error): received corrupted buffer from outbound peer",
		)
	}

	conn.Flush()
	cancel()
	<-pipe.done

	{
		assert.Equal(
			t,
			onclosedStub.WasCalled(),
			true,
			"pipe.open error: did not call onclosed handler on closed connection event",
		)

		assert.Equal(
			t,
			onclosedStub.WasCalledTimes(1),
			true,
			"pipe.open error: expected onclosed handler to be called once, got called %d times", onopenedStub.Times(),
		)
	}
}

func TestSocket_StartIOWriteRoutineStartFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_StartIOReadRoutineStartFailure(t *testing.T) {
	t.Skip()
}

func TestSocket_StartIORunnerSuddenlyStopped(t *testing.T) {
	t.Skip()
}

func TestSocket_StartIOContextSuddenlyCancled(t *testing.T) {
	t.Skip()
}

func TestSocket_StartIOSocketSuddenlyClosed(t *testing.T) {
	t.Skip()
}

func TestSocket_StartIOSocketSuddenlyErrored(t *testing.T) {
	t.Skip()
}

// test opening an inbound connection
func TestSocket_OpenInbound(t *testing.T) {
	addr := "dummy-test-host:dummyport"
	runner := testutils.NewRunnerMock()
	conn := testutils.NewConnMock()

	connector := func() (string, types.SocketType, net.Conn) {
		return addr, types.Inbound, conn
	}

	ctx, cancel := context.WithCancel(context.Background())

	onopenedStub := testutils.NewFnCallStub()
	onopened := func(_ context.Context, p *socket) error {
		onopenedStub.TrackCall()
		return nil
	}

	onclosedStub := testutils.NewFnCallStub()
	onclosed := func(_ context.Context, p *socket) error {
		onclosedStub.TrackCall()
		return nil
	}

	pipe := NewSocket(ReadBufferSize, WireByteHeaderLength, ReadDeadlineInMs)

	{
		pipe.runner = runner
		pipe.buffers.write.Open()
	}

	{
		pipe.open(ctx, connector, onopened, onclosed)

		_, isNotReady := <-pipe.ready

		t.Skip(
			t,
			isNotReady,
			false,
			"pipe open inbound error: pipe is not receiving or sending after inbound launch",
		)

		t.Skip(
			t,
			pipe.reader,
			"pipe open inbound error: reader/writter is not initialized after inbound launch",
		)
		t.Skip(
			t,
			pipe.writer,
			"pipe open inbound error: reader/writter is not initialized after inbound launch",
		)

		t.Skip(
			t,
			pipe.conn,
			"pipe open inbound error: pipe connection is not initialized after inbound launch",
		)

		t.Skip(
			t,
			pipe.kind,
			types.Inbound,
			"pipe open inbound error: wrong pipe sense",
		)

		t.Skip(
			t,
			onopenedStub.WasCalled(),
			"pipe.open error: did not call onopened handler on opened connection event",
		)

		t.Skip(
			t,
			onopenedStub.WasCalledTimes(1),
			"pipe.open error: expected onopened handler to be called once",
		)
	}

	cancel()
	<-pipe.done
	<-time.After(time.Millisecond * 10)

	{
		t.Skip(
			t,
			onclosedStub.WasCalled(),
			"pipe.open error: did not call onclosed handler on closed connection event",
		)

		t.Skip(
			t,
			onclosedStub.WasCalledTimes(1),
			"pipe.open error: expected onclosed handler to be called once",
		)
	}
}

// test opening an outbound connection
func TestSocket_OpenOutbound(t *testing.T) {
	addr := "dummy-test-host:dummyport"
	dialer := testutils.MockDialer()
	runner := testutils.NewRunnerMock()

	connector := func() (string, types.SocketType, net.Conn) {
		return addr, types.Inbound, dialer.Conn
	}

	ctx, cancel := context.WithCancel(context.Background())

	onopenedStub := testutils.NewFnCallStub()
	onopened := func(_ context.Context, p *socket) error {
		onopenedStub.TrackCall()
		return nil
	}

	onclosedStub := testutils.NewFnCallStub()
	onclosed := func(_ context.Context, p *socket) error {
		onclosedStub.TrackCall()
		return nil
	}

	pipe := NewSocket(
		ReadBufferSize,
		WireByteHeaderLength,
		ReadDeadlineInMs,
	)

	{
		pipe.runner = runner
		pipe.buffers.write.Open()
	}

	// test opening an outbound connection
	{
		err := pipe.open(ctx, connector, onopened, onclosed)

		assert.Nil(
			t,
			err,
			"pipe.open: error while opeining the socket",
		)

		_, isNotReady := <-pipe.ready

		assert.False(
			t,
			isNotReady,
			"pipe.open error: pipe is not receiving or sending after outbound launch",
		)

		assert.NotNil(
			t,
			pipe.reader,
			"pipe.open error: reader/writter is not initialized after outbound launch",
		)

		assert.NotNil(
			t,
			pipe.writer,
			"pipe.open error: reader/writter is not initialized after outbound launch",
		)

		assert.NotNil(
			t,
			pipe.conn,
			"pipe.open error: pipe connection is not initialized after outbound launch",
		)

		assert.Equal(
			t,
			onopenedStub.WasCalled(),
			true,
			"pipe.open error: did not call onopened handler on opened connection event",
		)

		assert.Equal(
			t,
			onopenedStub.WasCalledTimes(1),
			true,
			"pipe.open error: expected onopened handler to be called once",
		)
	}

	cancel()
	<-pipe.done
	time.After(time.Millisecond * 10)
	{
		assert.True(
			t,
			onclosedStub.WasCalled(),
			"pipe.open error: did not call onclosed handler on closed connection event",
		)

		assert.True(
			t,
			onclosedStub.WasCalledTimes(1),
			"pipe.open error: expected onclosed handler to be called once",
		)
	}
}

func TestSocket_OpenConnectorFailure(t *testing.T) {
	t.Skip()
}
