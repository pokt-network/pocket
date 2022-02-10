package poktp2p

import (
	"testing"
	"time"
)

func TestReqMapGet(t *testing.T) {
	rmap := NewReqMap(100)

	request := rmap.get()

	if request.nonce != 1 {
		t.Errorf("reqmap error: failed to retrieve new request with valid nonce")
	}

	if request.ch == nil {
		t.Errorf("reqmap error: failed to retrieve request with a respond channel")
	}
}

func TestReqMapFind(t *testing.T) {

	rmap := NewReqMap(100)
	request := rmap.get()
	nonce, ch, exists := rmap.find(request.nonce)

	if !exists {
		t.Errorf("reqmap error: cannot retrieve/find existing request!")
	}

	if nonce != request.nonce {
		t.Errorf("reqmap error: faield to retrieve existing request, found a wrong one with invalid nonce")
	}

	var isproperch bool

	go func(channel chan work) {
	waiter:
		for {
			select {
			case <-channel:
				isproperch = true
				break waiter
			}
		}
	}(ch)

	<-time.After(time.Millisecond)
	request.ch <- work{data: nil}
	<-time.After(time.Millisecond * 5)
	if isproperch != true {
		t.Errorf("reqmap error: failed to retrieve existing request, found a wrong one with a diffrent respond channel")
	}
}

func TestReqMapDelete(t *testing.T) {
	rmap := NewReqMap(100)
	request := rmap.get()

	deleted := rmap.delete(request.nonce)

	if !deleted {
		t.Errorf("reqmap error: could not delete existing request")
	}

	_, open := <-request.ch
	if open != false {
		t.Errorf("reqmap error: request respond channel is still open after delete")
	}

	_, _, exists := rmap.find(request.nonce)
	if exists {
		t.Errorf("reqmap error: the request is still tracked in the reqmap after delete")
	}
}
