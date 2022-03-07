package p2p

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWireEncode(t *testing.T) {
	c := newWireCodec()

	encoding := Binary
	requestNumber := uint32(12)
	isErrorOrEnd := false
	data := GenerateByteLen(1024)

	msg := c.encode(encoding, isErrorOrEnd, requestNumber, data, false)

	header := msg[:9]
	body := msg[9:]

	flags := header[0]
	flagswitch, encoding, err := parseFlag(flags)

	if err != nil {
		t.Errorf("Codec error: failed to encode, encountered error while parsing flag: %s", err.Error())
	}

	iswrapped := flagswitch[4]
	isrequest := flagswitch[3]
	iserrorOrEnd := flagswitch[2]

	reqnum := header[1:5]
	bodylen := header[5:9]

	if iswrapped {
		t.Errorf("Codec error: failed to encode, wrong flag for non-wrapped message (not domain encoded)")
	}

	if !isrequest {
		t.Errorf("Codec error: failed to encode, wrong flag for message of type request")
	}

	if iserrorOrEnd {
		t.Errorf("Codec error: failed to encode, wrong flag for non-error message")
	}

	if encoding != Binary {
		t.Errorf("Codec error: failed to encode, wrong flag(s) for message encoding type")
	}

	requestNum := binary.BigEndian.Uint32(reqnum)
	if requestNum != 12 {
		t.Errorf("Codec error: failed to encode, corrupted request number bits in header")
	}

	length := binary.BigEndian.Uint32(bodylen)
	if length != uint32(len(data)) {
		t.Errorf("Codec error: failed to encode, corrupted request body length bits in header")
	}

	if bytes.Compare(body, data) != 0 {
		t.Errorf("Codec error: failed to encode, corrupted body")
	}
}

func TestWireDecode(t *testing.T) {
	c := newWireCodec()

	encoding := Binary
	requestNumber := uint32(12)
	isErrorOrEnd := false
	data := GenerateByteLen(1024)

	msg := c.encode(encoding, isErrorOrEnd, requestNumber, data, true)

	reqnum, encoding, decodedData, wrapped, err := c.decode(msg)

	if err != nil {
		t.Errorf("Codec error: failed to decode. Encoutered error: %s", err.Error())
	}

	if !wrapped {
		t.Errorf("Codec error: failed to decode, is_wrapped flag bits are corrupted")
	}

	if err != nil {
		t.Errorf("Codec error: failed to decode, error bits are corrupted")
	}

	if reqnum != uint32(12) {
		t.Errorf("Codec error: failed to decode, request number bits are corrupted")
	}

	if encoding != Binary {
		t.Errorf("Codec error: failed to decode, encoding bits are corrupted")
	}

	if bytes.Compare(decodedData, data) != 0 {
		t.Errorf("Codec error: failed to decode, data bits are corrupted")
	}
}
