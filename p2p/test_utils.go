package p2p

import (
	"sync/atomic"
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
