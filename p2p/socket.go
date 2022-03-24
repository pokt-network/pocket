package p2p

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/pokt-network/pocket/p2p/types"

	"go.uber.org/atomic"
)

type SocketEventMonitor func(context.Context, *socket) error

// a "socket" (not to be confused with the OS' socket) is an abstraction around the net.Conn go interface, whose purpose is to represent a p2p connection with full "read/write" capabilities
// Both read and write operations are buffered, and both buffer sizes are configurable.
//
// Configuration paramters are directly assigned to the socket struct.
//
// 1 live p2p connection = 1 socket
//
type socket struct {
	// the agent responsible for running/creating/managing sockets
	runner types.Runner

	// configuration parameters
	headerLength uint
	bufferSize   uint
	readTimeout  uint
	addr         string
	kind         types.SocketType // inbound or outbound

	// the actual network socket
	conn net.Conn

	// the wire codec
	codec *wireCodec

	// io buffers
	buffers struct {
		read  *types.Buffer
		write *types.ConcurrentBuffer
	}

	// the io reader/writer
	reader *bufio.Reader
	writer *bufio.Writer

	// the map to track writes that expects acknowledgements
	// we call them requests (as they require responses)
	requests *types.RequestMap

	// turns true when the socket is opened (i.e: the connection is established and IO is on going)
	isOpen atomic.Bool

	ready   chan struct{} // when the socket is opened and IO starts, this channel gets closed to signal readiness
	writing chan struct{} // when the writing starts, this channel receives a new input; closes when done writing (i.e: stopped the socket)
	reading chan struct{} // when the reading starts, this channel receives a new input; closes when done reading (i.e: stopped the socket)
	done    chan struct{} // if this channel is closed or receives and input, it stops the socket and IO operations
	closed  chan struct{} // this channel signals that the socket has been closed by receiving a new input

	errored chan struct{} // on error, this channel receives a new input to signal the happening of an error
	err     struct {      // the reference to store the encountered error.
		sync.Mutex
		error
	}

	logger types.Logger
}

// retrieves the network connection in question through the connector argument
// and starts the IO operations on that connection, while putting in place event handlers for onSocketOpened and onSocketClosed events
// returns an error on failure
func (s *socket) open(ctx context.Context, connector func() net.Conn, onopened SocketEventMonitor, onclosed SocketEventMonitor) error {
	s.buffers.write.Open()

	conn := connector()

	addr := ctx.Value("address").(string)
	kind := ctx.Value("kind").(types.SocketType)

	//TODO(derrandz): retrieve from parameters or config
	if addr == "" {
		return ErrSocketEmptyContextValue("address")
	}

	if kind == "" {
		return ErrSocketEmptyContextValue("kind")
	}

	switch kind {
	case types.Outbound:
		fallthrough
	case types.Inbound:
		fallthrough
	default:
		s.close()
		return ErrSocketUndefinedKind(string(kind))
	}

	go s.startIO(ctx, kind, addr, conn, onopened, onclosed)

	select {
	case <-s.errored:
		return s.err.error
	case <-s.reading:
		s.logger.Debug("Socket has started the reading routine successfully")
	}

	select {
	case <-s.errored:
		return s.err.error
	case <-s.writing:
		s.logger.Debug("Socket has started the reading routine successfully")
	}

	s.isOpen.Store(true)
	close(s.ready)

	return nil
}

// close() stops the ongoing IO operations gracefully (i.e: s.read and s.write routines) and closes the underlying network connection (s.conn),
// as well as all opened channels. Finally, close() signals that the closing process has went successfully by sending a signal on s.closed.
//
// This method does not block.
func (s *socket) close() {
	if !s.isOpen.Load() {
		return
	}
	close(s.done)

	s.err.Lock()
	close(s.errored)
	s.err.Unlock()

	s.buffers.write.Close()

	s.isOpen.Store(false)

	if s.conn != nil {
		s.conn.Close()
	}

	s.closed <- struct{}{}
	s.isOpen.Store(false)
}

// creates a reader and writer for the network connection
// kicks off the onSocketOpened event handler, returns if there is an error
// launches both the reading and writing routines (read, write methods): both are blocking
// when the write routine exists, the onSocketClosed event handler kicks off (returns an error if there is one)
// the write and read routines will both exist for the same reasons, so having just one of them block is sufficient.
func (s *socket) startIO(ctx context.Context, kind types.SocketType, addr string, conn net.Conn, onOpened SocketEventMonitor, onClosed SocketEventMonitor) {
	defer s.close()

	s.addr = addr
	s.kind = kind

	s.conn = conn
	s.reader = bufio.NewReader(conn)
	s.writer = bufio.NewWriter(conn)

	if err := onOpened(ctx, s); err != nil {
		s.error(err)
		close(s.writing)
		close(s.reading)
		return
	}

	go s.read(ctx) // closes s.reading when done
	s.write(ctx)   // closes s.writing when done

	if err := onClosed(ctx, s); err != nil {
		s.error(err)
		return
	}
}

// the TLS handshake algorithm to establish encrypted connections
func (s *socket) handshake() {
	panic("Not implemented")
}

// s.readChunk reads a chunk (of size readbufferSize) out of the TCP connection using the s.reader.
//
// This is used by the read routine (s.read) to perform buffered reads.
// To achieve buffered reading, readChunk first off reads the header bytes (first bytes from 0 to headerLength) to retrieve size of
// received/to-be-read payload. If the payload size exceeds the configured max, readChunk will error out, the other end will receive a ErrPayloadTooBig error. (TODO: implement erroring logic to send this error)
//
// After reading the header, readChunk blocks until the full body length is read.
//
func (s *socket) readChunk() ([]byte, int, error) {
	var n int

	readBuffer := s.buffers.read.Ref()
	if _, err := io.ReadFull(s.reader, (*readBuffer)[:s.headerLength]); err != nil {
		return nil, 0, err
	}
	_, _, bodyLen, err := s.codec.decodeHeader((*readBuffer)[:s.headerLength])
	if err != nil {
		return nil, 0, err
	}

	// TODO(derrandz): replace with configurable max value or keep it as is (i.e: max=chunk size) ??
	if bodyLen > uint32(s.bufferSize-s.headerLength) {
		// TODO(derrandz): move error to socket_err.go
		return nil, 0, errors.New(fmt.Sprintf("io pipe error: cannot read a buffer of length %d, the accepted body length is %d.", bodyLen, s.bufferSize-s.headerLength))
	}

	if n, err = io.ReadFull(s.reader, (*readBuffer)[s.headerLength:uint32(s.headerLength)+bodyLen]); err != nil {
		return nil, 0, err
	}

	buff := make([]byte, 0)
	buff = append(buff, (*readBuffer)[:s.headerLength+uint(n)]...)

	return buff, n, err
}

// the read routine
// this routine performs buffered reads (readBufferSize) on the established connection
// and does two things as a consequence of a read operation:
//    1- if it's a response to a request out of this socket, it will redirect the response to the request's response channel (see types/request.go)
// .  2- if not, it will handover the read buffer as Packet (check types/work.go) to the runner (check types/runner.go)
// this routine halfs if:
//   - the cancelable routine cancels
//   - if the runner stops
//   - if the socket closes
//   - if some party signals that this routine should stop (by closing the s.reading channel)
//   - if there is an IO error: EOF, UnexpectedEOF, peer hang up, unexpected error
func (s *socket) read(ctx context.Context) {
	defer func() {
		close(s.reading)
		s.closed <- struct{}{}
	}()

	{
		s.reading <- struct{}{} // signal start
	}

reader:
	for {
		select {
		// TODO(derrandz): replace with passed down context ?
		case <-ctx.Done():
			break reader

		case <-s.runner.Done():
			break reader

		case <-s.done:
			break reader

		case _, open := <-s.reading:
			if !open {
				break reader
			}

		default:
			{
				buf, n, err := s.readChunk()
				if err != nil {

					switch err {
					case io.EOF:
						s.error(ErrPeerHangUp(err))
						break reader

					case io.ErrUnexpectedEOF:
						s.error(ErrPeerHangUp(err))
						break reader

					default:
						s.error(ErrUnexpected(err))
						break reader
					}
				}

				if n == 0 {
					continue
				}

				nonce, _, data, wrapped, err := s.codec.decode(buf)
				if err != nil {
					s.error(err)
					break reader
				}

				if nonce != 0 {
					_, ch, found := s.requests.Find(nonce)
					if !found {
						// report that we've received a nonced message whose requested does not exist on our end!
					}

					ch <- types.NewPacket(nonce, data, s.addr, wrapped)
					close(ch)
					continue
				}

				s.runner.Sink() <- types.NewPacket(nonce, data, s.addr, wrapped)
			}
		}
	}
}

// TODO(derrandz): add buffered write by writing exactly the writeBufferSize. Will need to think about how to split big payloads and sequence them
// writes a chunk=writeBufferSize to the writer
// this operation is blocking, and blocks until a waiter is ready to receive the signal from the write buffer (signaling that is has written)
// This is used by send/request/broadcast operations
// upon each send, the write routine will receive a signal so that it may proceed to send the write over the network
func (s *socket) writeChunk(b []byte, iserroreof bool, reqnum uint32, wrapped bool) (uint, error) {
	defer s.buffers.write.Unlock()
	s.buffers.write.Lock()

	writeBuffer := s.buffers.write.Ref()

	buff := s.codec.encode(Binary, iserroreof, reqnum, b, wrapped)
	(*writeBuffer) = append((*writeBuffer), buff...)

	s.buffers.write.Signal()
	return uint(len(b)), nil // TODO(derrandz): should length be of b or of the encoded b
}

// writeChunkAckful is a writeChunk that expects to receive an ACK response for the chunk it has written
// This method will create a request, which is basically a nonce to identify the chunk to write, and a channel on which to receive the response
// the channel is blocking, thus allowing the 'wait to receive the response' behavior.
// the `read` routine takes care of identifying incoming responses (_using the nonce_) and redirecting them to the waiting channels of the currently-open requests.
func (s *socket) writeChunkAckful(b []byte, wrapped bool) (types.Packet, error) {
	request := s.requests.Get()
	requestNonce := request.Nonce

	if _, err := s.writeChunk(b, false, requestNonce, wrapped); err != nil {
		s.requests.Delete(requestNonce)
		return types.NewPacket(requestNonce, nil, "", false), err
	}

	var response types.Packet

	select {
	case response = <-request.ResponsesCh:
		return response, nil

	case <-time.After(time.Millisecond * time.Duration(s.readTimeout)):
		return types.Packet{}, ErrSocketRequestTimedOut(s.addr, requestNonce)
	}
}

// the write routine
// this routine performs buffered writes (writeBufferSize) on the established connection
// this routine halfs if:
//   - the cancelable routine cancels
//   - if the runner stops
//   - if the socket closes
//   - if some party signals that this routine should stop (by closing the s.writing channel)
//   - if there is a write error
func (s *socket) write(ctx context.Context) {
	defer func() {
		close(s.writing)
		s.closed <- struct{}{}
	}()

	{
		s.writing <- struct{}{} // signal start
	}

writer:
	for {
		select {
		case <-ctx.Done():
			break writer
		case <-s.runner.Done():
			break writer

		case <-s.done:
			break writer

		case _, open := <-s.writing:
			if !open {
				break writer
			}

		case <-s.buffers.write.Signals(): // blocks
			{
				if !s.buffers.write.IsOpen() {
					break writer
				}

				buff := s.buffers.write.DumpBytes()

				if _, err := s.writer.Write(buff); err != nil {
					s.error(err)
					break writer
				}

				if err := s.writer.Flush(); err != nil {
					s.error(err)
					break writer
				}
			}
		}
	}
}

// tracks and stores the encountered error
func (s *socket) error(err error) {
	defer s.err.Unlock()
	s.err.Lock()

	if s.err.error == nil {
		s.err.error = err
	}

	s.errored <- struct{}{}
}

// A constructor to create a socket
func NewSocket(readBufferSize uint, packetHeaderLength uint, readTimeoutInMs uint) *socket {
	wc := newWireCodec()
	pipe := &socket{
		codec: wc,

		kind:         types.UndefinedSocketType,
		headerLength: packetHeaderLength,
		bufferSize:   readBufferSize,
		readTimeout:  readTimeoutInMs,

		buffers: struct {
			read  *types.Buffer
			write *types.ConcurrentBuffer
		}{
			read:  types.NewBuffer(readBufferSize),
			write: types.NewConcurrentBuffer(0),
		},

		requests: types.NewRequestMap(math.MaxUint32),

		ready: make(chan struct{}),

		// to allow for graceful shutdown, we buffer this to 3 since there are 3 routines that we need to have them signal their closing.
		//Upon their signals, this channel becomes blocking thus allowing the parent routine to wait until all 3 are closed
		closed: make(chan struct{}, 3),

		done: make(chan struct{}),

		// any reporting routine will immediately halt after signaling on this channel.
		//buffering this channel to 1 allows the erroring routine to not block, but the parent to wait if it needs to confirm error or no error
		errored: make(chan struct{}, 1),

		writing: make(chan struct{}),
		reading: make(chan struct{}),
	}

	return pipe
}
