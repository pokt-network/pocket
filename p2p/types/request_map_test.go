package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestMap_Get(t *testing.T) {
	rMap := NewRequestMap(100)

	request := rMap.Get()

	assert.Equal(
		t,
		request.Nonce,
		uint32(1),
		"Request map error: failed to retrieve new request with valid nonce",
	)

	assert.NotNil(
		t,
		request.ResponsesCh,
		"Request map error: failed to retrieve request with a respond channel",
	)
}

func TestRequestMap_Find(t *testing.T) {
	rMap := NewRequestMap(100)
	request := rMap.Get()
	nonce, ch, exists := rMap.Find(request.Nonce)

	assert.True(
		t,
		exists,
		"Request map error: cannot retrieve/find existing request!",
	)

	assert.Equal(
		t,
		nonce,
		request.Nonce,
		"Request map error: faield to retrieve existing request, found a wrong one with invalid nonce",
	)

	sentPacket := Packet{}
	go request.Respond(Packet{})
	receivedPacket := <-ch

	// assert that the found request is the same as the one sought after by verifying their response channels ref
	assert.Equal(
		t,
		(chan Packet)(ch),
		request.ResponsesCh,
		"Request map error: failed to retrieve an existing request, found a wrong one with a diffrent respond channel",
	)

	// assert that the found request is the same as the one sought after by sending work on the created one
	// and expecting the found one to receive it (will if found==created)
	assert.Equal(
		t,
		sentPacket,
		receivedPacket,
		"Request map error: failed to retrieve existing request, found a wrong one with a diffrent respond channel",
	)
}

func TestRequestMap_Delete(t *testing.T) {
	rMap := NewRequestMap(100)
	request := rMap.Get()

	deleted := rMap.Delete(request.Nonce)

	assert.True(
		t,
		deleted,
		"Request map error: could not delete existing request",
	)

	_, canStillReceiveResponses := <-request.ResponsesCh

	assert.False(
		t,
		canStillReceiveResponses,
		"Request map error: request respond channel is still open after delete",
	)

	_, _, exists := rMap.Find(request.Nonce)
	assert.False(
		t,
		exists,
		"Request map error: the request is still tracked in the reqmap after delete",
	)
}
