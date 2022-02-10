package poktp2p

import (
	"fmt"
	"sync"
)

type iomap struct {
	sync.Mutex
	// add a mutex to synchronize goroutines
	maxcap   uint32
	elements map[string]*io
}

func (im *iomap) get(id string) (*io, bool) {
	defer im.Unlock()
	im.Lock()

	var pipe *io
	var exists bool

	pipe, exists = im.elements[id]
	if !exists {
		// create a new iopipe
		// TODO: add logic to check for maxcap if reached
		// TODO: add logic to swap old connections for new one on maxcap reached
		pipe = NewIoPipe()
		im.elements[id] = pipe
	}

	return pipe, exists
}

func (im *iomap) find(id string) (*io, bool) {
	defer im.Unlock()
	im.Lock()

	el, exists := im.elements[id]
	return el, exists
}

func (im *iomap) peak(id string) bool {
	defer im.Unlock()
	im.Lock()

	_, exists := im.elements[id]
	return exists
}

func (im *iomap) remove(id string) (bool, error) {
	defer im.Unlock()
	im.Lock()

	return false, nil
}

func NewIoMap(cap uint) *iomap {
	return &iomap{
		maxcap:   uint32(cap),
		elements: make(map[string]*io),
	}
}

/*
 @
 @ reqmap: request maps
 @
*/
type req struct {
	nonce uint32
	ch    chan work
}

type reqmap struct {
	sync.Mutex
	maxcap   uint32
	elements []*req
	nonces   uint32
}

func (rm *reqmap) get() *req {
	rm.Lock()
	defer rm.Unlock()

	rm.nonces++
	nonce := rm.nonces
	fmt.Println("nonce=", nonce)
	newreq := &req{nonce: nonce, ch: make(chan work, 1)}
	rm.elements = append(rm.elements, newreq)
	return newreq
}

func (rm *reqmap) find(nonce uint32) (uint32, chan work, bool) {
	rm.Lock()
	defer rm.Unlock()

	var request *req
	var exists bool
	var index int

	for i := 0; i < len(rm.elements); i++ {
		if rm.elements[i].nonce == nonce {
			exists = true
			index = i
			request = rm.elements[i]
		}
	}

	if exists {
		rm.elements[index] = nil
		fhalf := rm.elements[:index]
		shalf := rm.elements[index+1:]
		rm.elements = append(fhalf, shalf...)
		return request.nonce, request.ch, exists
	}

	return nonce, nil, false
}

func (rm *reqmap) delete(nonce uint32) bool {
	defer rm.Unlock()
	rm.Lock()

	var exists bool
	var index int

	for i := 0; i < len(rm.elements); i++ {
		if rm.elements[i].nonce == nonce {
			exists = true
			index = i
			break
		}
	}

	if exists {
		close(rm.elements[index].ch)
		rm.elements[index] = nil
		fhalf := rm.elements[:index]
		shalf := rm.elements[index+1:]
		rm.elements = append(fhalf, shalf...)
	}

	return exists
}

func NewReqMap(cap uint) *reqmap {
	return &reqmap{maxcap: uint32(cap), elements: make([]*req, 0), nonces: uint32(0)}
}
