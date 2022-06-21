package testutils

import (
	"sync/atomic"
)

type FnCallStub struct {
	called *int32 // act as bool
	timesc *int32 // act as int
}

func NewFnCallStub() *FnCallStub {
	return &FnCallStub{
		called: new(int32),
		timesc: new(int32),
	}
}

func (f *FnCallStub) TrackCall() {
	atomic.AddInt32(f.called, 1)
	atomic.AddInt32(f.timesc, 1)
}

func (f *FnCallStub) WasCalled() bool {
	v := atomic.LoadInt32(f.called)
	return atomic.CompareAndSwapInt32(f.called, 1, v)
}

func (f *FnCallStub) WasCalledTimes(times int32) bool {
	v := atomic.LoadInt32(f.timesc)
	return atomic.CompareAndSwapInt32(f.timesc, times, v)
}

func (f *FnCallStub) Times() int32 {
	return atomic.LoadInt32(f.timesc)
}
