package types

import sync "sync"

type RequestMap struct {
	sync.Mutex
	maxcap   uint32
	elements []*Request
	nonces   uint32
}

func (rm *RequestMap) Get() *Request {
	rm.Lock()
	defer rm.Unlock()

	rm.nonces++
	nonce := rm.nonces
	newreq := &Request{Nonce: nonce, ResponsesCh: make(chan Packet)}
	rm.elements = append(rm.elements, newreq)
	return newreq
}

func (rm *RequestMap) Find(nonce uint32) (uint32, chan Packet, bool) {
	rm.Lock()
	defer rm.Unlock()

	var request *Request
	var exists bool
	var index int

	for i := 0; i < len(rm.elements); i++ {
		if rm.elements[i].Nonce == nonce {
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
		return request.Nonce, request.ResponsesCh, exists
	}

	return nonce, nil, false
}

func (rm *RequestMap) Delete(nonce uint32) bool {
	defer rm.Unlock()
	rm.Lock()

	var exists bool
	var index int

	for i := 0; i < len(rm.elements); i++ {
		if rm.elements[i].Nonce == nonce {
			exists = true
			index = i
			break
		}
	}

	if exists {
		close(rm.elements[index].ResponsesCh)
		rm.elements[index] = nil
		fhalf := rm.elements[:index]
		shalf := rm.elements[index+1:]
		rm.elements = append(fhalf, shalf...)
	}

	return exists
}

func NewRequestMap(cap uint) *RequestMap {
	return &RequestMap{maxcap: uint32(cap), elements: make([]*Request, 0), nonces: uint32(0)}
}
