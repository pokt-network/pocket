package p2p

import (
	"bufio"
	"context"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
	"github.com/pokt-network/pocket/p2p/utils"

	"go.uber.org/atomic"
)

type SocketEventMonitor func(context.Context, *socket) error

// A "socket" (not to be confused with the OS' socket) is an abstraction around the net.Conn go interface,
// whose purpose is to represent a p2p connection with full "read/write" capabilities.
//
// Both read and write operations are buffered, and both buffer sizes are configurable.
//
// Configuration paramters are directly assigned to the socket struct.
//
// 1 live p2p connection = 1 socket
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
		// the read buffer is a byte slice of a certain size (configurable: `ReadBufferSize``) which is destined
		// to receive incoming data. This buffer is not concurrent and is primarily used by the s.read routine
		read *types.Buffer

		// the write buffer is a byte slice of a certain size (configurable: `WriteBufferSize``) which is written to
		// by the owner of this socket (i.e: the runner, the peer). When the peer is done writing to the buffer, the s.write routine
		// proceeds to writing this buffer to the concerned connection, resulting in a network send.
		// This buffer is concurrent because two operations are happening in parallel, the writing to the buffer by the peer
		// and the reading off of the buffer by the s.write routine
		write *types.ConcurrentBuffer
	}

	// the io reader/writer
	reader *bufio.Reader
	writer *bufio.Writer

	// the map to track writes that expects acknowledgements
	// we call them requests (as they require responses)
	requests *types.RequestMap

	// turns true when the socket is opened (i.e., the connection is established and IO routines are launched)
	isOpen    atomic.Bool
	isWriting atomic.Bool
	isReading atomic.Bool

	// For reference, see these resources on the use of empty structs in go channels:
	// - https://dave.cheney.net/2014/03/25/the-empty-struct
	// - https://dave.cheney.net/2013/04/30/curious-channels

	ready   chan struct{} // when the socket is opened and IO starts, this channel gets closed to signal readiness
	done    chan struct{} // if this channel is closed or receives and input, it stops the socket and IO operations
	writing chan struct{} // when the writing starts, this channel receives a new input; closes when done writing (i.e: stopped the socket)
	reading chan struct{} // when the reading starts, this channel receives a new input; closes when done reading (i.e: stopped the socket)
	errored chan struct{} // on error, this channel receives a new input to signal the happening of an error

	err struct { // the reference to store the encountered error
		sync.Mutex
		error
	}

	logger types.Logger
}

// A constructor to create a socket
func NewSocket(readBufferSize uint, packetHeaderLength uint, readTimeoutInMs uint) *socket {
	pipe := &socket{
		codec: newWireCodec(),

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

		ready:   make(chan struct{}), // closes to signal the readiness of the socket
		done:    make(chan struct{}), // closes to signal the closing of the socket
		writing: make(chan struct{}), // sends a new input to signal the start of the writing routine, closes when done writing
		reading: make(chan struct{}), // sends new input to signal the start of the reading routine, closes when done reading

		// sends new input to signal the encoutering of an error in running routines
		// closes when the socket closes.
		// Bufferred to 1 to allow non-blocking signaling of errors, and blocking awaiting of error signals.
		// (We are handling 1 error at most, so not more than one signal is expected to be received at a time, establishing a queue of exactly 1 error...)
		errored: make(chan struct{}, 1),

		logger: types.NewLogger(os.Stdout),
	}

	return pipe
}

// Retrieves the underlying TCP socket (net.Conn) in question through the connector argument and starts
// the IO operations on that socket, while also putting in place event handlers for onSocketOpened
// and onSocketClosed events.
func (s *socket) open(ctx context.Context, connector func() (string, types.SocketType, net.Conn), onOpened SocketEventMonitor, onClosed SocketEventMonitor) error {
	s.buffers.write.Open()

	addr, socketType, conn := connector()

	if utils.IsEmpty(addr) {
		return ErrMissingRequiredArg("address")
	}

	if utils.IsEmpty(string(socketType)) {
		return ErrMissingRequiredArg("socketType")
	}

	switch socketType {
	case types.Outbound:
	case types.Inbound:
	default:
		s.close()
		return ErrSocketUndefinedKind(string(socketType))
	}

	go s.startIO(ctx, socketType, addr, conn, onOpened, onClosed)

	select {
	case <-s.errored:
		return s.err.error
	case _, closed := <-s.reading:
		if closed {
			s.logger.Debug("Socket has stopped reading...")
			s.logger.Error("Socket stopped reading imemdiately after opening...")
		} else {
			s.logger.Debug("Socket has started the reading routine successfully")
		}
	}

	select {
	case <-s.errored:
		return s.err.error
	case _, closed := <-s.writing:
		if closed {
			s.logger.Debug("Socket has stopped writing...")
			s.logger.Error("Socket stopped writing imemdiately after opening...")
		} else {
			s.logger.Debug("Socket has started the reading routine successfully")
		}
	}

	s.signalOpen()
	s.signalReady()

	return nil
}

// Stops the ongoing IO operations gracefully (i.e., s.read and s.write routines) and closes the
// underlying network connection (s.conn), as well as all opened channels. Finally, close() signals
// that the closing process has went successfully by sending a signal on s.closed.
// NOTE: This method does not block.
func (s *socket) close() {
	if !s.isOpen.Load() {
		return
	}

	close(s.done)

	s.stopErrorReporting()
	s.buffers.write.Close()

	if s.conn != nil {
		s.conn.Close()
	}

	s.signalClose()
}

// Creates a reader and writer for the network connection and kicks off the onSocketOpened event handler.
// This also launches both the reading and writing routines (read, write methods), both of which are blocking.
// When the write routine exists, the onSocketClosed event handler kicks off and returns an error if there is one.
// NOTE: The write and read routines will both exit for the same reasons, so having just one of them block is sufficient.
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

	go s.read(ctx)  // closes s.reading when done
	go s.write(ctx) // closes s.writing when done

waiter:
	for {
		select {
		case _, open := <-s.writing:
			if !open {
				s.logger.Warn("Socket stopped writing...")
				break waiter
			}
		case _, open := <-s.reading:
			if !open {
				s.logger.Warn("Socket stopped reading...")
				break waiter
			}
		}
	}

	if err := onClosed(ctx, s); err != nil {
		s.error(err)
		return
	}
}

// The TLS handshake algorithm to establish encrypted connections
func (s *socket) handshake() {
	panic("Not implemented")
}

// Reads a chunk (of size `readbufferSize`) out of the TCP connection using `s.reader`.
//
// This is used by the read routine (`s.read`) to perform buffered reads.
// To achieve buffered reading, `readChunk` first off reads the header bytes (first bytes from 0 to `headerLength``)
// to retrieve size of received/to-be-read payload. If the payload size exceeds the configured max,
// `readChunk` will error out, the other end will receive a `ErrPayloadTooBig` error.
// TODO(derrandz): implement erroring logic to send this error.
//
// NOTE: After reading the header, readChunk blocks until the full body length is read.
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
		return nil, 0, ErrPayloadTooBig(uint(bodyLen), s.bufferSize-s.headerLength)
	}

	if n, err = io.ReadFull(s.reader, (*readBuffer)[s.headerLength:uint32(s.headerLength)+bodyLen]); err != nil {
		return nil, 0, err
	}

	buff := make([]byte, 0)
	buff = append(buff, (*readBuffer)[:s.headerLength+uint(n)]...)

	return buff, n, err
}

// The read routine.
//
// This routine performs buffered reads (the size of the buffer is the config param: `readBufferSize`) on the
// established connection and does two things as a consequence of a read operation:
//    1. If it's a response to a request out of this socket, it will redirect the response to the request's response channel (see `types/request.go`)
//    2. If not, it will handover the read (past-tense participle) buffer as a Packet (check `types/packet.go`) to the runner (check `types/runner.go`)
//
// This routine halts if:
//   - The cancelable context cancels
//   - If the runner stops (i.e: s.runner.Done() receives)
//   - If the socket closes
//   - If some party signals that this routine should stop (by closing the s.reading channel)
//   - If there is an IO error: EOF, UnexpectedEOF, peer hang up, unexpected error
func (s *socket) read(ctx context.Context) {
	defer s.signalReadingStop()
	s.signalReadingStart()

reader:
	for {
		select {
		case <-ctx.Done(): // stop if context cancels
			break reader
		case <-s.runner.Done(): // stop if runner stops
			break reader
		case <-s.done: // stop if the socket closes
			break reader
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
					s.logger.Warn("Read 0 bytes on socket:", s.addr)
					continue
				}

				nonce, _, data, wrapped, err := s.codec.decode(buf)
				if err != nil {
					s.error(err)
					break reader
				}

				// A non-zero nonce happens on nonced-respones (i.e., responses to already sent requests).
				// Using the non-zero nonce, we are able to fetch the existing (waiting) request from
				// the request map and pull out the channel on which this request expects to receive a response.
				if nonce != 0 {
					_, ch, found := s.requests.Find(nonce)
					if !found {
						s.logger.Warn("Received response with nonce but no request found:", nonce)
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

// TODO(derrandz): Add buffered write by writing exactly `WriteBufferSize` amount of bytes, and splitting if larger.
//                 Will need to think about how to split big payloads and sequence them.
//
// Writes a chunk=writeBufferSize to the writer (s.writer).
// This operation is blocking, and blocks until a waiter is ready to receive the signal from the write buffer (signaling that is has written).
// This is used by send/request/broadcast operations.
// Upon each send, the write routine will receive a signal so that it may proceed to send the write over the network.
func (s *socket) writeChunk(b []byte, isErrorOf bool, reqNum uint32, wrapped bool) (uint, error) {
	defer s.buffers.write.Unlock()
	s.buffers.write.Lock()

	writeBuffer := s.buffers.write.Ref()

	buff := s.codec.encode(Binary, isErrorOf, reqNum, b, wrapped)
	*writeBuffer = append(*writeBuffer, buff...)

	s.buffers.write.Signal()
	return uint(len(b)), nil // TODO(derrandz): should length be of b or of the encoded b
}

// writeChunkAckful is a writeChunk that expects to receive an ACK response for the chunk it has written.
// This method will create a request, which is basically a nonce and a channel, the nonce to identify the
// written chunk and the channel to receive the response for that particular written chunk.
//
// The channel - on which the response is expected to be received - is blocking, thus enables the 'wait to receive the response' behavior.
// The `read` routine takes care of identifying incoming responses (_using the nonce_) and redirecting them to the waiting channels of the currently-open requests.
func (s *socket) writeChunkAckful(b []byte, wrapped bool) (types.Packet, error) {
	panic("Not used or tested at the moment")

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

// The write routine.
//
// This routine performs buffered writes (`writeBufferSize``) on the established connection.
// This routine halts if:
//   - the cancelable routine cancels
//   - if the runner stops
//   - if the socket closes
//   - if some party signals that this routine should stop (by closing the s.writing channel)
//   - if there is a write error
func (s *socket) write(ctx context.Context) {
	defer s.signalWritingStop()
	s.signalWritingStart()

writer:
	for {
		select {
		case <-ctx.Done(): // stop if context cancels
			break writer
		case <-s.runner.Done(): // stop if runner stops
			break writer
		case <-s.done: // stop if the socket closes
			break writer
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

// Tracks and stores the encountered error
func (s *socket) error(err error) {
	defer s.err.Unlock()
	s.err.Lock()

	if s.err.error == nil {
		s.err.error = err
	}

	s.errored <- struct{}{}
}

// signal the readiness of the socket, called when everything has been perofrmed successfully when opening the socket
func (s *socket) signalReady() {
	close(s.ready)
}

// signal that the socket has been opened. Flip the flag.
func (s *socket) signalOpen() {
	s.isOpen.Store(true)
}

// signal that the socket has been closed.
func (s *socket) signalClose() {
	s.isOpen.Store(false)
}

func (s *socket) stopErrorReporting() {
	defer s.err.Unlock()

	s.err.Lock()
	close(s.errored)
}

func (s *socket) signalWritingStart() {
	s.writing <- struct{}{}
	s.isWriting.Store(true)
}

func (s *socket) signalWritingStop() {
	close(s.writing)
	s.isWriting.Store(false)
}

func (s *socket) signalReadingStart() {
	s.reading <- struct{}{}
	s.isReading.Store(true)
}

func (s *socket) signalReadingStop() {
	close(s.reading)
	s.isReading.Store(false)
}
