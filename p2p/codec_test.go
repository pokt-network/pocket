package poktp2p

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestWireEncode(t *testing.T) {
	c := &wcodec{}

	encoding := Binary
	requestNumber := uint32(12)
	isErrorOrEnd := false
	data := GenerateByteLen(1024)

	msg := c.encode(encoding, isErrorOrEnd, requestNumber, data, false)

	header := msg[:9]
	body := msg[9:]

	flags := header[0]
	flagswitch, encoding, err := parseflag(flags)

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
	c := &wcodec{}

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

func TestDomainEncode(t *testing.T) {
	c := NewDomainCodec()

	msg := struct {
		topic  string
		action string
	}{
		topic:  "membership",
		action: "ping",
	}

	encoder := func(m struct {
		topic  string
		action string
	}) ([]byte, error) {
		topic := []byte(m.topic)
		action := []byte(m.action)

		tlen := uint16(len(topic))
		alen := uint16(len(action))

		tlenb := make([]byte, 2)
		binary.BigEndian.PutUint16(tlenb[:2], tlen)

		alenb := make([]byte, 2)
		binary.BigEndian.PutUint16(alenb[:2], alen)

		buff := make([]byte, 0)

		buff = append(buff, tlenb...)
		buff = append(buff, alenb...)
		buff = append(buff, topic...)
		buff = append(buff, action...)

		return buff, nil
	}

	decoder := func(d []byte) (struct {
		topic  string
		action string
	}, error) {
		tlenb := d[:2]

		tlen := binary.BigEndian.Uint16(tlenb)

		msg := d[4:]
		topic := string(msg[:tlen])
		action := string(msg[tlen:])

		return struct {
			topic  string
			action string
		}{topic: topic, action: action}, nil
	}

	c.register(msg, encoder, decoder)

	encoding, err := c.encode(msg)
	if err != nil {
		t.Errorf("Domain Codec error: failed to encode message: %s", err.Error())
	}

	expectedlen := 2 + 2 + len(msg.topic) + len(msg.action)
	if len(encoding) == expectedlen {
		t.Errorf("Domain Codec error: wrong encoding byte length, expected %d, got %d", expectedlen, len(encoding))
	}

	decoding, err := decoder(encoding[2:])
	if err != nil {
		t.Errorf("Domain Codec error: failed to decode encoding bytes")
	}

	m := decoding

	if m.topic != msg.topic && m.action != msg.action {
		t.Errorf("Domain Codec error: corrupted decoded message, expected %s, %s, got %s, %s", msg.topic, msg.action, m.topic, m.action)
	}
}

func TestDomainDecode(t *testing.T) {
	c := NewDomainCodec()

	msg := struct {
		topic  string
		action string
	}{
		topic:  "membership",
		action: "ping",
	}

	encoder := func(m struct {
		topic  string
		action string
	}) ([]byte, error) {
		topic := []byte(m.topic)
		action := []byte(m.action)

		tlen := uint16(len(topic))
		alen := uint16(len(action))

		tlenb := make([]byte, 2)
		binary.BigEndian.PutUint16(tlenb[:2], tlen)

		alenb := make([]byte, 2)
		binary.BigEndian.PutUint16(alenb[:2], alen)

		buff := make([]byte, 0)

		buff = append(buff, tlenb...)
		buff = append(buff, alenb...)
		buff = append(buff, topic...)
		buff = append(buff, action...)

		return buff, nil
	}

	decoder := func(d []byte) (struct {
		topic  string
		action string
	}, error) {
		tlenb := d[:2]

		tlen := binary.BigEndian.Uint16(tlenb)

		msg := d[4:]
		topic := string(msg[:tlen])
		action := string(msg[tlen:])

		return struct {
			topic  string
			action string
		}{topic: topic, action: action}, nil
	}

	id, err := c.register(msg, encoder, decoder)
	idb := make([]byte, 2)
	binary.BigEndian.PutUint16(idb[:2], id)
	encoding, err := encoder(msg)
	decoding, err := c.decode(append(idb, encoding...))
	if err != nil {
		t.Errorf("Domain Codec error: failed to decode encoding bytes: %s", err.Error())
	}

	m := decoding.(struct {
		topic  string
		action string
	})

	if m.topic != msg.topic && m.action != msg.action {
		t.Errorf("Domain Codec error: corrupted decoded message, expected %s, %s, got %s, %s", msg.topic, msg.action, m.topic, m.action)
	}
}
