package p2p

import (
	"bufio"
	"errors"
	"fmt"
	stdio "io"
	"math"
	"net"
	"sync"
	"time"

	"github.com/pokt-network/pocket/p2p/types"

	"go.uber.org/atomic"
)

type parameters struct {
	headerLength uint
	bufferSize   uint
	timeoutMs    uint
}

type socket struct {
	runner types.Runner
	c      *wireCodec

	params parameters
	addr   string
	sense  Sense // inbound or outbound

	buffers struct {
		read  []byte
		write []byte
	}
	buffersState struct {
		sync.Mutex   // for the actual struct elements, like writeSignals
		writeOpen    bool
		writeSignals chan uint
		writeLock    sync.Mutex
	}

	dialer types.Dialer // an intermediary poktp2p interface that returns net.Conn, useful for mocking in testing
	conn   net.Conn

	timeouts struct {
		read uint
	}

	reader *bufio.Reader
	writer *bufio.Writer

	requests *types.RequestMap

	ready   chan uint
	done    chan uint
	errored chan uint
	closed  chan uint

	answering chan uint
	polling   chan uint

	opened atomic.Bool

	logger struct {
		sync.Mutex
		print func(...interface{}) (int, error)
	}

	err struct {
		sync.Mutex
		error
	}
}
type Sense string

var (
	OutboundIoPipe         Sense = "outbound"
	InboundIoPipe          Sense = "inbound"
	UnspecifiedIoPipeSense Sense = "unspecified"
)

func (s *socket) open(sense Sense, addr string, conn net.Conn, onopened func(*socket) error, onclosed func(*socket) error) error {
	s.buffersState.writeOpen = true

	switch sense {
	case OutboundIoPipe:
		go s.outbound(addr, onopened, onclosed)
	case InboundIoPipe:
		go s.inbound(addr, conn, onopened, onclosed)
	default:
		s.close()
		return errors.New("io pipe open error: undefined sense")
	}

	select {
	case <-s.polling:
	}

	select {
	case <-s.answering:
	}

	s.opened.Store(true)
	close(s.ready)

	return nil
}

func (s *socket) close() {
	if !s.opened.Load() {
		return
	}
	close(s.done)

	s.err.Lock()
	close(s.errored)
	s.err.Unlock()

	s.buffersState.Lock()
	close(s.buffersState.writeSignals)
	s.buffersState.Unlock()

	s.opened.Store(false)

	s.buffersState.writeLock.Lock()
	s.buffersState.writeOpen = false
	s.buffersState.writeLock.Unlock()

	if s.conn != nil {
		s.conn.Close()
	}

	s.closed <- 1
	s.opened.Store(false)
}

func (s *socket) outbound(addr string, onopened func(s *socket) error, onclosed func(s *socket) error) {
	defer func() {
		s.close()
	}()

	s.addr = addr
	s.sense = OutboundIoPipe

	if s.dialer == nil {
		s.dialer = types.NewDialer()
	}

	conn, err := s.dialer.Dial("tcp", addr)
	if err != nil {
		s.error(err)
		close(s.answering)
		close(s.polling)
		return
	}

	s.conn = conn
	s.reader = bufio.NewReader(s.conn)
	s.writer = bufio.NewWriter(s.conn)

	if err := onopened(s); err != nil {
		s.error(err)

		close(s.answering)
		close(s.polling)
		return
	}

	go s.poll()
	s.answer()

	if err := onclosed(s); err != nil {
		s.error(err)
		return
	}
}

func (s *socket) inbound(addr string, conn net.Conn, onopened func(*socket) error, onclosed func(*socket) error) {
	defer func() {
		s.close()
	}()

	s.addr = addr
	s.sense = InboundIoPipe

	s.conn = conn
	s.reader = bufio.NewReader(conn)
	s.writer = bufio.NewWriter(conn)

	if err := onopened(s); err != nil {
		s.error(err)

		close(s.answering)
		close(s.polling)
		return
	}

	go s.poll() // closes s.polling when done
	s.answer()  // closes s.answering when done

	if err := onclosed(s); err != nil {
		s.error(err)
		return
	}
}

func (s *socket) read() ([]byte, int, error) {
	var n int

	if _, err := stdio.ReadFull(s.reader, s.buffers.read[:s.params.headerLength]); err != nil {
		return nil, 0, err
	}
	_, _, bodylen, err := s.c.decodeHeader(s.buffers.read[:s.params.headerLength])
	if err != nil {
		return nil, 0, err
	}

	if bodylen > uint32(s.params.bufferSize-s.params.headerLength) { // TODO(derrandz): replace with configurable max value
		return nil, 0, errors.New(fmt.Sprintf("io pipe error: cannot read a buffer of length %d, the accepted body length is %d.", bodylen, s.params.bufferSize-s.params.headerLength))
	}

	if n, err = stdio.ReadFull(s.reader, s.buffers.read[s.params.headerLength:bodylen+uint32(s.params.headerLength)]); err != nil {
		return nil, 0, err
	}

	buff := make([]byte, 0)
	buff = append(buff, s.buffers.read[:s.params.headerLength+uint(n)]...)

	return buff, n, err
}

func (s *socket) poll() {
	defer func() {
		close(s.polling)
		s.closed <- 1
	}()

	{
		s.polling <- 1 // signal start
	}

	for stop := false; !stop; {
		select {
		// TODO(derrandz): replace with passed down context
		case <-s.runner.Done():
			break

		case <-s.done:
			stop = true
			break

		case _, open := <-s.polling:
			if !open {
				stop = true
				break
			}

		default:
			{
				buf, n, err := s.read()
				if err != nil {

					switch err {
					case stdio.EOF:
						s.error(ErrPeerHangUp(err))
						break

					case stdio.ErrUnexpectedEOF:
						s.error(ErrPeerHangUp(err))
						break

					default:
						s.error(ErrUnexpected(err))
						break
					}
				}

				if n == 0 {
					continue
				}

				nonce, _, data, wrapped, err := s.c.decode(buf)
				if err != nil {
					s.error(err)
					break
				}

				if nonce != 0 {
					_, ch, found := s.requests.Find(nonce)
					// TODO(derrandz): this is hacku
					if found {
						ch <- types.NewWork(nonce, data, s.addr, wrapped)
						close(ch)
						continue
					}
				}

				s.runner.Sink() <- types.NewWork(nonce, data, s.addr, wrapped)
			}
		}
	}
}

func (s *socket) write(b []byte, iserroreof bool, reqnum uint32, wrapped bool) (uint, error) {
	defer s.buffersState.writeLock.Unlock()
	s.buffersState.writeLock.Lock()

	buff := s.c.encode(Binary, iserroreof, reqnum, b, wrapped)
	s.buffers.write = append(s.buffers.write, buff...)

	// TODO(derrandz): find a better way, maybe the value itself (the channel) should be an atomic on and off switch to signal writes
	s.buffersState.Lock()
	s.buffersState.writeSignals <- 1
	s.buffersState.Unlock()

	return uint(len(b)), nil // TODO(derrandz): should length be of b or of the encoded b
}

func (s *socket) ackwrite(b []byte, wrapped bool) (types.Work, error) {
	request := s.requests.Get()
	requestNonce := request.Nonce()

	if _, err := s.write(b, false, requestNonce, wrapped); err != nil {
		s.requests.Delete(requestNonce)
		return types.NewWork(requestNonce, nil, "", false), err
	}

	var response types.Work
	select {
	case response = <-request.Response():
	case <-time.After(time.Millisecond * time.Duration(s.params.timeoutMs)):
		return types.Work{}, fmt.Errorf("ackwrite: request timed out. nonce=%d, addr=%s", requestNonce, s.addr)
	}

	return response, nil
}

func (s *socket) answer() {
	defer func() {
		close(s.answering)
		s.closed <- 1
	}()

	{
		s.answering <- 1 // signal start
	}

	for stop := false; !stop; {
		select {
		case <-s.runner.Done():
			stop = true
			break

		case <-s.done:
			stop = true
			break

		case _, open := <-s.answering:
			if !open {
				stop = true
				break
			}

		case <-s.buffersState.writeSignals: // blocks
			{
				if stop {
					break
				}

				s.buffersState.writeLock.Lock()
				buff, open := s.buffers.write, s.buffersState.writeOpen
				s.buffers.write = nil
				s.buffersState.writeLock.Unlock()

				if !open {
					stop = true
					break
				}

				if _, err := s.writer.Write(buff); err != nil {
					s.error(err)
					stop = true
					break
				}

				if err := s.writer.Flush(); err != nil {
					s.error(err)
					stop = true
					break
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
	params := parameters{headerLength: packetHeaderLength, bufferSize: readBufferSize, timeoutMs: readTimeoutInMs}
	wc := newWireCodec()
	pipe := &socket{
		params: params,
		c:      wc,

		requests: types.NewRequestMap(math.MaxUint32),

		buffers: struct {
			read  []byte
			write []byte
		}{
			read:  make([]byte, params.bufferSize),
			write: make([]byte, 0),
		},
		buffersState: struct {
			sync.Mutex
			writeOpen    bool
			writeSignals chan uint
			writeLock    sync.Mutex
		}{
			writeOpen:    false,
			writeSignals: make(chan uint, 1),
			writeLock:    sync.Mutex{},
		},

		timeouts: struct{ read uint }{read: params.timeoutMs},

		sense: UnspecifiedIoPipeSense,

		ready:  make(chan uint),
		closed: make(chan uint, 3),

		done:    make(chan uint, 1),
		errored: make(chan uint, 1),

		answering: make(chan uint),
		polling:   make(chan uint),
	}

	return pipe
}
