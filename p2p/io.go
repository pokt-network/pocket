package p2p

import (
	"bufio"
	"errors"
	"fmt"
	stdio "io"
	"math"
	"net"
	"sync"

	"go.uber.org/atomic"
)

type request struct {
	ch chan []byte
	id uint32
}

type io struct {
	g *gater
	c *wcodec

	addr  string
	sense Sense // inbound or outbound

	buffers struct {
		read  []byte
		write []byte
	}
	buffersState struct {
		writeOpen    bool
		writeSignals chan uint
		writeLock    sync.Mutex

		readOpen bool
	}

	dialer Dialer // an intermediary poktp2p interface that returns net.Conn, useful for mocking in testing
	conn   net.Conn

	timeouts struct {
		read int
	}

	reader *bufio.Reader
	writer *bufio.Writer

	requests *reqmap

	ready   chan uint
	done    chan uint
	errored chan uint
	closed  chan uint

	answering chan uint
	polling   chan uint

	opened    atomic.Bool
	sending   bool
	receiving bool

	err error
}

type Sense string

var (
	OutboundIoPipe         Sense = "outbound"
	InboundIoPipe          Sense = "inbound"
	UnspecifiedIoPipeSense Sense = "unspecified"
)

func (p *io) open(sense Sense, addr string, conn net.Conn, onopened func(*io) error, onclosed func(*io) error) error {
	p.buffersState.writeOpen = true
	p.buffersState.readOpen = true

	switch sense {
	case OutboundIoPipe:
		go p.outbound(addr, onopened, onclosed)
	case InboundIoPipe:
		go p.inbound(addr, conn, onopened, onclosed)
	default:
		p.close()
		return errors.New("io pipe open error: undefined sense")
	}

	select {
	case <-p.polling:
	}

	select {
	case <-p.answering:
	}

	p.opened.Store(true)
	close(p.ready)

	return nil
}

func (p *io) close() {
	close(p.done)
	close(p.errored)
	close(p.buffersState.writeSignals)

	p.sending = false
	p.receiving = false
	p.opened.Store(false)
	p.buffersState.writeOpen = false
	p.buffersState.readOpen = false

	p.closed <- 1
}

func (p *io) outbound(addr string, onopened func(p *io) error, onclosed func(p *io) error) {
	defer func() {
		p.close()
	}()

	p.addr = addr
	p.sense = OutboundIoPipe

	if p.dialer == nil {
		p.dialer = NewDialer()
	}

	conn, err := p.dialer.Dial("tcp", addr)
	if err != nil {
		p.err = err
		p.errored <- 1
		close(p.answering)
		close(p.polling)
		return
	}

	p.conn = conn
	p.reader = bufio.NewReader(p.conn)
	p.writer = bufio.NewWriter(p.conn)

	if err := onopened(p); err != nil {
		p.err = err
		p.errored <- 1

		close(p.answering)
		close(p.polling)
		return
	}

	go p.poll()
	p.answer()

	if err := onclosed(p); err != nil {
		p.err = err
		p.errored <- 1
		return
	}
}

func (p *io) inbound(addr string, conn net.Conn, onopened func(*io) error, onclosed func(*io) error) {
	defer func() {
		p.close()
	}()

	p.addr = addr
	p.sense = InboundIoPipe

	p.conn = conn
	p.reader = bufio.NewReader(conn)
	p.writer = bufio.NewWriter(conn)

	if err := onopened(p); err != nil {
		p.err = err
		p.errored <- 1

		close(p.answering)
		close(p.polling)
		return
	}

	go p.poll() // closes p.polling when done
	p.answer()  // closes p.answering when done

	if err := onclosed(p); err != nil {
		p.err = err
		p.errored <- 1
		return
	}
}

func (p *io) read() ([]byte, int, error) {
	var n int
	if _, err := stdio.ReadFull(p.reader, p.buffers.read[:WireByteHeaderLength]); err != nil {
		return nil, 0, err
	}
	_, _, bodylen, err := p.c.decodeHeader(p.buffers.read[:WireByteHeaderLength])
	if err != nil {
		return nil, 0, err
	}

	if bodylen > uint32(ReadBufferSize) { // TODO: replace with configurable max value
		return nil, 0, errors.New(fmt.Sprintf("io pipe error: cannot read a buffer of length %d, the acceptedl length is %d.", bodylen, ReadBufferSize))
	}

	if n, err = stdio.ReadFull(p.reader, p.buffers.read[WireByteHeaderLength:bodylen+uint32(WireByteHeaderLength)]); err != nil {
		return nil, 0, err
	}

	buff := make([]byte, 0)
	buff = append(buff, p.buffers.read[:WireByteHeaderLength+n]...)

	return buff, n, err
}

func (p *io) poll() {
	defer func() {
		close(p.polling)
		p.closed <- 1
	}()

	{
		p.receiving = true
		p.polling <- 1 // signal start
	}

	for stop := false; !stop; {
		select {
		case <-p.g.done:
			p.receiving = false
			p.buffersState.readOpen = false
			stop = true
			break

		case <-p.done:
			p.receiving = false
			p.buffersState.readOpen = false
			stop = true
			break

		case _, open := <-p.polling:
			if !open {
				p.receiving = false
				p.buffersState.readOpen = false
				stop = true
				break
			}

		default:
			{
				buf, n, err := p.read()
				if err != nil {
					fmt.Println("Error!", err)
					p.err = err
					p.errored <- 1
					break
				}

				if n == 0 {
					continue
				}

				nonce, _, data, wrapped, err := p.c.decode(buf)
				if err != nil {
					p.err = err
					p.errored <- 1
					break
				}

				if nonce != 0 {
					_, ch, found := p.requests.find(nonce)
					// TODO: this is hacku
					if found {
						ch <- work{nonce: nonce, decode: wrapped, data: data, addr: p.addr}
						close(ch)
						continue
					}
				}

				p.g.sink <- work{nonce: nonce, decode: wrapped, data: data, addr: p.addr}
			}
		}
	}
}

func (p *io) write(b []byte, iserroreof bool, reqnum uint32, wrapped bool) (uint, error) {
	defer p.buffersState.writeLock.Unlock()
	p.buffersState.writeLock.Lock()
	buff := p.c.encode(Binary, iserroreof, reqnum, b, wrapped)
	p.buffers.write = append(p.buffers.write, buff...)
	p.buffersState.writeSignals <- 1
	return uint(len(b)), nil // TODO: should length be of b or of the encoded b
}

func (p *io) ackwrite(b []byte, wrapped bool) (work, error) {
	request := p.requests.get()

	if _, err := p.write(b, false, request.nonce, wrapped); err != nil {
		p.requests.delete(request.nonce)
		return work{data: nil}, err
	}

	var response work
	select {
	case response = <-request.ch:
	}

	return response, nil
}

func (p *io) answer() {
	defer func() {
		close(p.answering)
		p.closed <- 1
	}()

	{
		p.sending = true
		p.answering <- 1 // signal start
	}

	for stop := false; !stop; {
		select {
		case <-p.g.done:
			p.sending = false
			p.buffersState.writeOpen = false
			stop = true
			break

		case <-p.done:
			p.sending = false
			p.buffersState.writeOpen = false
			stop = true
			break

		case _, open := <-p.answering:
			if !open {
				p.sending = false
				p.buffersState.writeOpen = false
				stop = true
				break
			}

		case <-p.buffersState.writeSignals: // blocks
			{
				if stop {
					break
				}

				p.buffersState.writeLock.Lock()
				buff, open := p.buffers.write, p.buffersState.writeOpen
				p.buffers.write = nil
				p.buffersState.writeLock.Unlock()

				if !open {
					stop = true
					break
				}

				if _, err := p.writer.Write(buff); err != nil {
					p.err = err
					p.errored <- 1
					stop = true
					break
				}

				if err := p.writer.Flush(); err != nil {
					p.err = err
					p.errored <- 1
					stop = true
					break
				}
			}
		}
	}
}

func NewIoPipe() *io {
	pipe := &io{
		c: &wcodec{},

		requests: NewReqMap(math.MaxUint32),

		buffers: struct {
			read  []byte
			write []byte
		}{
			read:  make([]byte, ReadBufferSize),
			write: make([]byte, 0),
		},
		buffersState: struct {
			writeOpen    bool
			writeSignals chan uint
			writeLock    sync.Mutex

			readOpen bool
		}{
			writeOpen:    false,
			writeSignals: make(chan uint, 1),
			writeLock:    sync.Mutex{},

			readOpen: false,
		},

		timeouts: struct{ read int }{read: ReadDeadlineMs},

		sense: UnspecifiedIoPipeSense,

		ready:  make(chan uint),
		closed: make(chan uint, 3),

		done:    make(chan uint, 1),
		errored: make(chan uint, 1),

		answering: make(chan uint),
		polling:   make(chan uint),

		sending:   false,
		receiving: false,

		err: nil,
	}

	return pipe
}
