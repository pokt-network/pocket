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

type SocketKind string

var (
	Outbound            SocketKind = "outbound"
	Inbound             SocketKind = "inbound"
	UndefinedSocketKind SocketKind = "unspecified"
)

type SocketEventMonitor func(context.Context, *socket) error

type socket struct {
	runner types.Runner
	c      *wireCodec

	headerLength uint
	bufferSize   uint
	readTimeout  uint
	addr         string
	kind         SocketKind // inbound or outbound

	buffers struct {
		read  *types.Buffer
		write *types.ConcurrentBuffer
	}

	conn net.Conn

	reader *bufio.Reader
	writer *bufio.Writer

	requests *types.RequestMap

	isOpen atomic.Bool

	ready   chan uint
	writing chan uint
	reading chan uint
	done    chan uint
	closed  chan uint

	errored chan uint
	err     struct {
		sync.Mutex
		error
	}

	logger struct {
		sync.Mutex
		print func(...interface{}) (int, error)
	}
}

func (s *socket) open(ctx context.Context, connector func() net.Conn, onopened SocketEventMonitor, onclosed SocketEventMonitor) error {
	s.buffers.write.Open()

	conn := connector()

	addr := ctx.Value("address").(string)
	kind := ctx.Value("kind").(SocketKind)

	if addr == "" {
		return ErrSocketEmptyContextValue("address")
	}

	if kind == "" {
		return ErrSocketEmptyContextValue("kind")
	}

	switch kind {
	case Outbound:
	case Inbound:
	default:
		s.close()
		return ErrSocketUndefinedKind(string(kind))
	}

	go s.startIO(ctx, kind, addr, conn, onopened, onclosed)

	select {
	case <-s.reading:
	}

	select {
	case <-s.writing:
	}

	s.isOpen.Store(true)
	close(s.ready)

	return nil
}

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

	s.closed <- 1
	s.isOpen.Store(false)
}

func (s *socket) startIO(ctx context.Context, kind SocketKind, addr string, conn net.Conn, onopened SocketEventMonitor, onclosed SocketEventMonitor) {
	defer func() {
		s.close()
	}()

	s.addr = addr
	s.kind = kind

	s.conn = conn
	s.reader = bufio.NewReader(conn)
	s.writer = bufio.NewWriter(conn)

	if err := onopened(ctx, s); err != nil {
		s.error(err)

		close(s.writing)
		close(s.reading)
		return
	}

	go s.read(ctx) // closes s.reading when done
	s.write(ctx)   // closes s.writing when done

	if err := onclosed(ctx, s); err != nil {
		s.error(err)
		return
	}
}

func (s *socket) handshake() {

}

func (s *socket) readChunk() ([]byte, int, error) {
	var n int

	readBuffer := s.buffers.read.Ref()
	if _, err := io.ReadFull(s.reader, (*readBuffer)[:s.headerLength]); err != nil {
		return nil, 0, err
	}
	_, _, bodylen, err := s.c.decodeHeader((*readBuffer)[:s.headerLength])
	if err != nil {
		return nil, 0, err
	}

	// TODO(derrandz): replace with configurable max value or keep it as is (i.e: max=chunk size) ??
	if bodylen > uint32(s.bufferSize-s.headerLength) {
		return nil, 0, errors.New(fmt.Sprintf("io pipe error: cannot read a buffer of length %d, the accepted body length is %d.", bodylen, s.bufferSize-s.headerLength))
	}

	if n, err = io.ReadFull(s.reader, (*readBuffer)[s.headerLength:bodylen+uint32(s.headerLength)]); err != nil {
		return nil, 0, err
	}

	buff := make([]byte, 0)
	buff = append(buff, (*readBuffer)[:s.headerLength+uint(n)]...)

	return buff, n, err
}

func (s *socket) read(ctx context.Context) {
	defer func() {
		close(s.reading)
		s.closed <- 1
	}()

	{
		s.reading <- 1 // signal start
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

				nonce, _, data, wrapped, err := s.c.decode(buf)
				if err != nil {
					s.error(err)
					break reader
				}

				if nonce != 0 {
					_, ch, found := s.requests.Find(nonce)
					if !found {
						// report that we've received a nonced message whose requested does not exist on our end!
					}

					ch <- types.NewWork(nonce, data, s.addr, wrapped)
					close(ch)
					continue
				}

				s.runner.Sink() <- types.NewWork(nonce, data, s.addr, wrapped)
			}
		}
	}
}

func (s *socket) writeChunk(b []byte, iserroreof bool, reqnum uint32, wrapped bool) (uint, error) {
	defer s.buffers.write.Unlock()
	s.buffers.write.Lock()

	writeBuffer := s.buffers.write.Ref()

	buff := s.c.encode(Binary, iserroreof, reqnum, b, wrapped)
	(*writeBuffer) = append((*writeBuffer), buff...)

	s.buffers.write.Signal()
	return uint(len(b)), nil // TODO(derrandz): should length be of b or of the encoded b
}

// writeChunkAckful is a writeChunk that expects to receive an ACK response for the chunk it has written
// This method will create a request, which is basically a nonce to identify the chunk to write, and a channel on which to receive the response
// the channel is blocking, thus allowing the 'wait to receive the response' behavior.
// the `read` routine takes care of identifying incoming responses (_using the nonce_) and redirecting them to the waiting channels of the currently-open requests.
func (s *socket) writeChunkAckful(b []byte, wrapped bool) (types.Work, error) {
	request := s.requests.Get()
	requestNonce := request.Nonce()

	if _, err := s.writeChunk(b, false, requestNonce, wrapped); err != nil {
		s.requests.Delete(requestNonce)
		return types.NewWork(requestNonce, nil, "", false), err
	}

	var response types.Work

	select {
	case response = <-request.Response():
		return response, nil

	case <-time.After(time.Millisecond * time.Duration(s.readTimeout)):
		return types.Work{}, ErrSocketRequestTimedOut(s.addr, requestNonce)
	}
}

func (s *socket) write(ctx context.Context) {
	defer func() {
		close(s.writing)
		s.closed <- 1
	}()

	{
		s.writing <- 1 // signal start
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

func (s *socket) error(err error) {
	defer s.err.Unlock()
	s.err.Lock()

	if s.err.error == nil {
		s.err.error = err
	}

	s.errored <- 1
}

func NewSocket(readBufferSize uint, packetHeaderLength uint, readTimeoutInMs uint) *socket {
	wc := newWireCodec()
	pipe := &socket{
		c: wc,

		kind:         UndefinedSocketKind,
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

		ready:  make(chan uint),
		closed: make(chan uint, 3),

		done:    make(chan uint, 1),
		errored: make(chan uint, 1),

		writing: make(chan uint),
		reading: make(chan uint),
	}

	return pipe
}
