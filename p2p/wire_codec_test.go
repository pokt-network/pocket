package p2p

import (
	"encoding/binary"
	"testing"

	"github.com/pokt-network/pocket/p2p/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWireCodec_Encode(t *testing.T) {
	c := newWireCodec()

	encoding := Binary
	requestNumber := uint32(12)
	isErrorOrEnd := false
	chunk := testutils.NewDataChunk(1024, func(b []byte) []byte {
		return c.encode(encoding, isErrorOrEnd, requestNumber, b, false)
	})

	header := chunk.Encoded[:9]
	body := chunk.Encoded[9:]

	flags := header[0]
	flagswitch, encoding, err := parseFlag(flags)

	if err != nil {
		t.Errorf("Codec error: failed to encode, encountered error while parsing flag: %s", err.Error())
	}

	isWrapped := flagswitch[4]
	isrequest := flagswitch[3]
	isErrOrEnd := flagswitch[2]

	reqnum := header[1:5]
	bodylen := header[5:9]

	assert.False(
		t,
		isWrapped,
		"Codec error: failed to encode, wrong flag for non-wrapped message (not domain encoded)",
	)

	assert.True(
		t,
		isrequest,
		"Codec error: failed to encode, wrong flag for message of type request",
	)

	assert.False(
		t,
		isErrOrEnd,
		"Codec error: failed to encode, wrong flag for non-error message",
	)

	assert.Equal(
		t,
		encoding,
		Binary,
		"Codec error: failed to encode, wrong flag(s) for message encoding type",
	)

	requestNum := binary.BigEndian.Uint32(reqnum)
	assert.Equal(
		t,
		requestNum,
		uint32(12),
		"Codec error: failed to encode, corrupted request number bits in header",
	)

	length := binary.BigEndian.Uint32(bodylen)
	assert.Equal(
		t,
		length,
		uint32(len(chunk.Bytes)),
		"Codec error: failed to encode, corrupted request body length bits in header",
	)

	assert.Equal(
		t,
		body,
		chunk.Bytes,
		"Codec error: failed to encode, corrupted body",
	)
}

func TestWireCodec_Decode(t *testing.T) {
	c := newWireCodec()

	encoding := Binary
	requestNumber := uint32(12)
	isErrorOrEnd := false
	chunk := testutils.NewDataChunk(1024, func(b []byte) []byte {
		return c.encode(encoding, isErrorOrEnd, requestNumber, b, true)
	})

	reqnum, encoding, decodedData, wrapped, err := c.decode(chunk.Encoded)

	assert.Nil(
		t,
		err,
		"Codec error: failed to decode. Encoutered error",
	)

	assert.True(
		t,
		wrapped,
		"Codec error: failed to decode, is_wrapped flag bits are corrupted",
	)

	assert.Nil(
		t,
		err,
		"Codec error: failed to decode, error bits are corrupted",
	)

	assert.Equal(
		t,
		reqnum,
		uint32(12),
		"Codec error: failed to decode, request number bits are corrupted",
	)

	assert.Equal(
		t,
		encoding,
		Binary,
		"Codec error: failed to decode, encoding bits are corrupted",
	)

	assert.Equal(
		t,
		decodedData,
		chunk.Bytes,
		"Codec error: failed to decode, data bits are corrupted",
	)
}
