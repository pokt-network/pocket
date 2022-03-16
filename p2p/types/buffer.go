package types

import (
	sync "sync"

	"go.uber.org/atomic"
)

type Buffer struct {
	bytes []byte
}

type ConcurrentBuffer struct {
	Buffer
	sync.Mutex
	open   atomic.Bool
	signal chan struct{}
}

func (b *Buffer) Bytes() []byte {
	return b.bytes
}

func (b *Buffer) Ref() *[]byte {
	return &(b.bytes)
}

func (cb *ConcurrentBuffer) Bytes() []byte {
	defer cb.Unlock()
	cb.Lock()
	return cb.Buffer.Bytes()
}

func (cb *ConcurrentBuffer) DumpBytes() []byte {
	defer cb.Unlock()
	cb.Lock()
	bytes := cb.Buffer.bytes
	cb.Buffer.bytes = nil
	return bytes
}

func (cb *ConcurrentBuffer) Open() {
	cb.open.Store(true)
}

func (cb *ConcurrentBuffer) Close() {
	defer cb.Unlock()
	cb.Lock()

	cb.open.Store(false)
	close(cb.signal)
}

func (cb *ConcurrentBuffer) IsOpen() bool {
	return cb.open.Load() == true
}

func (cb *ConcurrentBuffer) Signal() {
	cb.signal <- struct{}{}
}

func (cb *ConcurrentBuffer) Wait() {
	<-cb.signal
}

func (cb *ConcurrentBuffer) Signals() <-chan struct{} {
	return cb.signal
}

func NewBuffer(size uint) *Buffer {
	return &Buffer{
		bytes: make([]byte, size),
	}
}

func NewConcurrentBuffer(size uint) *ConcurrentBuffer {
	return &ConcurrentBuffer{
		Buffer: *NewBuffer(size),
		signal: make(chan struct{}),
	}
}
