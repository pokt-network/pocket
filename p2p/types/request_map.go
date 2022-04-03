package types

import sync "sync"

// TODO(derrandz): This structure is called a `RequestMap`, but is using a slice and therefore does
// not take advantage of the o(1) slice indexing.
type RequestMap struct {
	sync.Mutex
	maxCap    uint32
	elements  []*Request
	numNonces uint32
}

func (rm *RequestMap) Get() *Request {
	rm.Lock()
	defer rm.Unlock()

	rm.numNonces++
	nonce := rm.numNonces
	newReq := &Request{Nonce: nonce, ResponsesCh: make(chan Packet)}
	rm.elements = append(rm.elements, newReq)
	return newReq
}

// Nonces are like IDs. Each generated Request is identified by an ID, so that if a peer makes a request
// (i.e. generates a work channel that is identified by ID) it can listen on responses coming from the
// network that have this nonce/ID. It can then redirect this this response to the proper workChannel,
// since the peer has kept this channel open (and blocking) to receive the response.
// This is how we achieve a SEND/ACK behavior in our p2p module.
func (rm *RequestMap) Find(nonce uint32) (uint32, chan Packet, bool) {
	rm.Lock()
	defer rm.Unlock()

	var request *Request

	var ch chan Packet = nil
	var exists bool

	var index int
	for i, element := range rm.elements {
		if element.Nonce == nonce {
			exists = true
			index = i
			request = rm.elements[i]
			break
		}
	}

	if exists {
		rm.elements[index] = nil
		rm.elements = append(
			rm.elements[:index],
			rm.elements[index+1:]...,
		)

		ch = request.ResponsesCh
	}

	return nonce, ch, exists
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
		rm.elements = append(
			rm.elements[:index],
			rm.elements[index+1:]...,
		)
	}

	return exists
}

func (rm *RequestMap) Requests() []*Request {
	rm.Lock()
	defer rm.Unlock()

	return rm.elements
}

func (rm *RequestMap) Len() int {
	return len(rm.elements)
}

func NewRequestMap(cap uint) *RequestMap {
	return &RequestMap{
		maxCap:    uint32(cap),
		elements:  make([]*Request, 0),
		numNonces: uint32(0),
	}
}
