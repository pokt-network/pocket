package p2p

import (
	"bytes"
	"context"
	stdio "io"
	"net"
	"reflect"
	"sync"
	"time"
)

/*
 @
 @ net.Conn mock
 @
*/

type connM struct {
	sync.Mutex
	buff    []byte
	signals chan int
}

func (c *connM) Read(b []byte) (n int, err error) {
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

func (c *connM) Write(b []byte) (n int, err error) {
	defer c.Unlock()
	c.Lock()
	cbuff := append(make([]byte, 0), c.buff...)
	buff := bytes.NewBuffer(cbuff)
	n, err = buff.Write(b)
	c.buff = buff.Bytes()

	c.signals <- 1
	return
}

func (c *connM) Flush() ([]byte, int, error) {
	defer c.Unlock()
	c.Lock()

	flushed := append(make([]byte, 0), c.buff...)
	c.buff = make([]byte, 0)
	return flushed, len(flushed), nil
}

func (c *connM) Close() error {
	defer c.Unlock()
	c.Lock()

	c.buff = nil
	return nil
}

func (c *connM) LocalAddr() net.Addr {
	return nil
}

func (c *connM) RemoteAddr() net.Addr {
	return nil
}

func (c *connM) SetDeadline(t time.Time) error {
	return nil
}

func (c *connM) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *connM) SetWriteDeadline(t time.Time) error {
	return nil
}

// use when not interested in knowing if the conn received new data
func MockConn() net.Conn {
	return &connM{
		buff:    make([]byte, 0),
		signals: make(chan int, 100), // so that it may not block
	}
}

// use when interested in knowing if the conn received new data (listen on signals)
func MockConnM() *connM {
	return &connM{
		buff:    make([]byte, 0),
		signals: make(chan int),
	}
}

/*
 @
 @  Dialer mock
 @
*/

type dialer struct {
	network string
	address string
	conn    *connM
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	d.network = network
	d.address = address
	return net.Conn(d.conn), nil
}

func (d *dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.network = network
	d.address = address
	return net.Conn(d.conn), nil
}

func MockDialer() *dialer {
	return &dialer{
		conn: MockConnM(),
	}
}

/*
 @
 @ Functions mock
 @ Has issues with functions that return nil err
 @
*/
func MockFunc(rts []interface{}, fn interface{}) *struct {
	called bool
	times  int
} {
	stub := struct {
		called bool
		times  int
	}{
		called: false,
		times:  0,
	}

	template := func(in []reflect.Value) []reflect.Value {
		returns := make([]reflect.Value, 0)
		stub.called = true
		stub.times++

		for i := 0; i < len(rts); i++ {
			elem := reflect.ValueOf(&rts[i]).Elem()
			returns = append(returns, reflect.Zero(reflect.TypeOf(elem)))
		}
		return returns
	}

	f := reflect.ValueOf(fn).Elem()
	v := reflect.MakeFunc(f.Type(), template)
	f.Set(v)

	return &stub
}
