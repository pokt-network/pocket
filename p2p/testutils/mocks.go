package testutils

import (
	"bytes"
	"context"
	stdio "io"
	"net"
	"sync"
	"time"

	"github.com/pokt-network/pocket/p2p/types"
)

/*
 @
 @ net.Conn mock
 @
*/
type ConnMock struct {
	sync.Mutex
	buff    []byte
	Signals chan int
}

func (c *ConnMock) Read(b []byte) (n int, err error) {
	defer c.Unlock()
	c.Lock()
	if len(c.buff) == 0 {
		return 0, nil
	}
	cbuff := append(make([]byte, 0), c.buff...)
	buff := bytes.NewBuffer(cbuff)
	n, err = buff.Read(b)
	if err == stdio.EOF {
		err = nil
		c.buff = make([]byte, 0)
	}
	return n, err
}

func (c *ConnMock) Write(b []byte) (n int, err error) {
	defer c.Unlock()
	c.Lock()
	cbuff := append(make([]byte, 0), c.buff...)
	buff := bytes.NewBuffer(cbuff)
	n, err = buff.Write(b)
	c.buff = buff.Bytes()

	c.Signals <- 1
	return
}

func (c *ConnMock) Flush() ([]byte, int, error) {
	defer c.Unlock()
	c.Lock()

	flushed := append(make([]byte, 0), c.buff...)
	c.buff = make([]byte, 0)
	return flushed, len(flushed), nil
}

func (c *ConnMock) Close() error {
	defer c.Unlock()
	c.Lock()

	c.buff = nil
	return nil
}

func (c *ConnMock) LocalAddr() net.Addr {
	return nil
}

func (c *ConnMock) RemoteAddr() net.Addr {
	return nil
}

func (c *ConnMock) SetDeadline(t time.Time) error {
	return nil
}

func (c *ConnMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *ConnMock) SetWriteDeadline(t time.Time) error {
	return nil
}

// use when not interested in knowing if the conn received new data
func NewConnMockBuffered() net.Conn {
	return &ConnMock{
		buff:    make([]byte, 0),
		Signals: make(chan int, 100), // so that it may not block
	}
}

// use when interested in knowing if the conn received new data (listen on signals)
func NewConnMock() *ConnMock {
	return &ConnMock{
		buff:    make([]byte, 0),
		Signals: make(chan int),
	}
}

/*
 @
 @  Dialer mock
 @
*/
type Dialer struct {
	Network string
	Address string
	Conn    *ConnMock
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	d.Network = network
	d.Address = address
	return net.Conn(d.Conn), nil
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.Network = network
	d.Address = address
	return net.Conn(d.Conn), nil
}

func MockDialer() *Dialer {
	return &Dialer{
		Conn: NewConnMock(),
	}
}

type RunnerMock struct {
	sink chan types.Packet
	done chan uint
}

func (r *RunnerMock) Sink() chan<- types.Packet {
	return r.sink
}

func (r *RunnerMock) GetSinkCh() chan types.Packet {
	return r.sink
}

func (r *RunnerMock) Done() <-chan uint {
	return r.done
}

func (r *RunnerMock) GetDoneCh() chan uint {
	return r.done
}

func NewRunnerMock() *RunnerMock {
	return &RunnerMock{
		sink: make(chan types.Packet, 1),
		done: make(chan uint, 1),
	}
}
