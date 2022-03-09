package p2p

import (
	"bufio"
	"crypto/rand"
	stdio "io"
	"net"
	"sync/atomic"
	"time"
)

const (
	HeaderLength = 9
)

type fnCallStub struct {
	called *int32 // act as bool
	timesc *int32 // act as int
}

func newFnCallStub() *fnCallStub {
	return &fnCallStub{
		called: new(int32),
		timesc: new(int32),
	}
}

func (f *fnCallStub) trackCall() {
	atomic.AddInt32(f.called, 1)
	atomic.AddInt32(f.timesc, 1)
}

func (f *fnCallStub) wasCalled() bool {
	v := atomic.LoadInt32(f.called)
	return atomic.CompareAndSwapInt32(f.called, 1, v)
}

func (f *fnCallStub) wasCalledTimes(times int32) bool {
	v := atomic.LoadInt32(f.timesc)
	return atomic.CompareAndSwapInt32(f.timesc, times, v)
}

func (f *fnCallStub) times() int32 {
	return atomic.LoadInt32(f.timesc)
}

func GenerateByteLen(size int) []byte {
	buff := make([]byte, size)
	rand.Read(buff)
	return buff
}

func ListenAndServe(addr string, readbufflen int, timeoutinMs int) (ready, done chan uint, data chan struct {
	n    int
	err  error
	buff []byte
}, response chan []byte) {

	ready = make(chan uint)
	done = make(chan uint)
	data = make(chan struct {
		n    int
		err  error
		buff []byte
	}, 10)
	response = make(chan []byte, 1)

	datapoint := func(n int, err error, buff []byte) struct {
		n    int
		err  error
		buff []byte
	} {
		return struct {
			n    int
			err  error
			buff []byte
		}{n: n, err: err, buff: buff}
	}

	readwriteconn := func(c net.Conn) {
		readerClosed := false

		codec := (&wireCodec{})
		creader := bufio.NewReader(c)
		buffer := make([]byte, readbufflen)

	reader:
		for {
			select {
			case <-done:
				break reader

			case msg := <-response:
				_, err := c.Write(msg)
				if err != nil {
					close(ready)
					close(done)
					close(data)
					close(response)
				}

			default:
				if readerClosed {
					continue reader
				}
				c.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeoutinMs)))
				n, err := stdio.ReadFull(creader, buffer[:HeaderLength]) // TODO(derrandz): parameterize this
				if err != nil {
					if isErrTimeout(err) {
						readerClosed = true
						continue reader
					}
					data <- datapoint(n, err, buffer)
					break reader
				}

				_, _, bodylength, derr := codec.decodeHeader(buffer[:HeaderLength]) // TODO(derrandz): DITTO line 861
				if derr != nil {
					data <- datapoint(len(buffer), err, buffer)
					break reader
				}

				n, err = stdio.ReadAtLeast(creader, buffer[HeaderLength:], int(bodylength)) // TODO(derrandz): DITTO line 861
				if err != nil {
					if isErrEOF(err) {
						err = nil
					}

					if isErrTimeout(err) {
						readerClosed = true
						continue
					}

					data <- datapoint(n, err, buffer)
					break reader
				}

				if n > 0 {
					hl := 9
					bl := int(bodylength)
					buff := buffer[:hl+bl]
					data <- datapoint(n, err, buff)
				}
			}
		}
		close(ready)
	}
	accept := func() {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			ready <- 0
			return
		}
		ready <- 1

	listener:
		for {
			select {
			case <-done:
				break listener
			default:
			}

			conn, err := l.Accept()
			if err != nil {
				ready <- 0
			}

			go readwriteconn(conn)
		}
	}

	go accept()

	return
}
